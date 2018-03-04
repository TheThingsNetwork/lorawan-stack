// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package networkserver provides a LoRaWAN 1.1-compliant network server implementation.
package networkserver

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mitchellh/hashstructure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const recentUplinkCount = 20
const deduplicationWindow = 5 * time.Second
const maxFCntGap = 16384

// NetworkServer implements the network server component.
//
// The network server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component
	*deviceregistry.RegistryRPC
	registry deviceregistry.Interface

	NetID types.NetID

	joinServers          *sync.Map // devEUI prefix -> ttnpb.NsJsClient
	gateways             *sync.Map // gtwID -> ttnpb.NsGsClient
	applicationServersMu *sync.RWMutex
	applicationServers   map[string]*applicationUplinkStream

	claimedUplinks   *sync.Map // []byte -> struct{}
	claimedDownlinks *sync.Map // []byte -> struct{}

	hashOptions *hashstructure.HashOptions
}

// Config represents the NetworkServer configuration.
type Config struct {
	Registry    deviceregistry.Interface
	JoinServers []ttnpb.NsJsClient
}

// New returns new *NetworkServer.
func New(c *component.Component, conf *Config) *NetworkServer {
	ns := &NetworkServer{
		Component:            c,
		RegistryRPC:          deviceregistry.NewRPC(c, conf.Registry), // TODO: Add checks
		registry:             conf.Registry,
		joinServers:          &sync.Map{},
		applicationServersMu: &sync.RWMutex{},
		applicationServers:   make(map[string]*applicationUplinkStream),
		gateways:             &sync.Map{},
		claimedUplinks:       &sync.Map{},
		claimedDownlinks:     &sync.Map{},
		hashOptions: &hashstructure.HashOptions{
			Hasher: fnv.New64a(),
		},
	}
	c.RegisterGRPC(ns)
	return ns
}

// StartServingGateway is called by the gateway server to indicate that it is serving a gateway.
func (ns *NetworkServer) StartServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	// A GatewayServer (or proxy) links to ALL NetworkServers, so this better be efficient
	// TODO: Announce to cluster
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// StopServingGateway is called by the gateway server to indicate that it is no longer serving a gateway.
func (ns *NetworkServer) StopServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	// TODO: Announce to cluster
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// matchDevice tries to match the uplink message with a device.
// If the uplink message matches a fallback session, that fallback session is recovered.
// If successful, the FCnt in the uplink message is set to the full FCnt. The NextFCntUp in the session is updated accordingly.
func (ns *NetworkServer) matchDevice(devAddr types.DevAddr, txDRIdx, txChIdx uint8, fCnt uint32, ack bool, mic []byte, pld []byte) (*deviceregistry.Device, error) {
	devs, err := ns.registry.FindBy(
		&ttnpb.EndDevice{
			Session: &ttnpb.Session{
				DevAddr: devAddr,
			},
		},
		"Session.DevAddr",
	)
	if err != nil {
		return nil, err
	}

	fb, err := ns.registry.FindBy(
		&ttnpb.EndDevice{
			SessionFallback: &ttnpb.Session{
				DevAddr: devAddr,
			},
		},
		"SessionFallback.DevAddr",
	)
	if err != nil {
		return nil, err
	}

	for _, dev := range fb {
		dev.EndDevice.Session = dev.EndDevice.SessionFallback
	}
	// Given that DevAddr never repeat in sequential sessions,
	// we can assume that devs does not contain duplicates.
	devs = append(devs, fb...)

	type device struct {
		*deviceregistry.Device
		fCnt uint32
		gap  uint32
	}
	matching := make([]device, 0, len(devs))

	for _, dev := range devs {
		dev.EndDevice.SessionFallback = nil

		next := dev.EndDevice.GetSession().GetNextFCntUp()

		if !dev.FCntIs16Bit &&
			next&0xffff > fCnt &&
			next < 0xffff<<16 {
			fCnt |= (next + 1<<16) &^ 0xffff
		} else {
			fCnt |= next &^ 0xffff
		}

		gap := fCnt - next

		switch dev.EndDevice.GetLoRaWANVersion() {
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			if gap > maxFCntGap {
				continue
			}
		}

		matching = append(matching, device{
			Device: dev,
			fCnt:   fCnt,
			gap:    gap,
		})
	}
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	for _, dev := range matching {
		ses := dev.GetSession()

		ke := ses.GetFNwkSIntKey()
		if ke == nil || ke.Key == nil || ke.Key.IsZero() {
			return nil, ErrCorruptRegistry.NewWithCause(nil, ErrMissingFNwkSIntKey.New(nil))
		}
		fNwkSIntKey := *ke.Key

		var confFCnt uint32
		if ack {
			if len(dev.RecentDownlinks) == 0 {
				continue
			}
			pld := dev.GetRecentDownlinks()[len(dev.RecentDownlinks)-1].GetPayload()
			confFCnt = pld.GetMACPayload().GetFCnt()
		}

		var computedMIC [4]byte
		switch dev.LoRaWANVersion {
		case ttnpb.MAC_V1_1:
			ke := ses.GetSNwkSIntKey()
			if ke == nil || ke.Key == nil || ke.Key.IsZero() {
				return nil, ErrCorruptRegistry.NewWithCause(nil, ErrMissingSNwkSIntKey.New(nil))
			}
			sNwkSIntKey := *ke.Key

			computedMIC, err = crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, confFCnt, txDRIdx, txChIdx, devAddr, dev.fCnt, pld)
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(fNwkSIntKey, devAddr, dev.fCnt, pld)
		default:
			return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("Unmatched LoRaWAN version"))
		}
		if err != nil {
			return nil, ErrMICComputeFailed.NewWithCause(nil, err)
		}
		if !bytes.Equal(mic, computedMIC[:]) {
			continue
		}
		if dev.fCnt == math.MaxUint32 {
			return nil, ErrFCntTooHigh.New(nil)
		}
		ses.NextFCntUp = dev.fCnt + 1
		return dev.Device, nil
	}
	return nil, ErrNotFound.New(nil)
}

func (ns *NetworkServer) handleUplink(ctx context.Context, msg *ttnpb.UplinkMessage) error {
	logger := ns.Logger()

	pld := msg.Payload.GetMACPayload()
	if pld == nil {
		return ErrMissingPayload.New(nil)
	}
	if pld.DevAddr.IsZero() {
		return ErrMissingDevAddr.New(nil)
	}

	fCnt := pld.GetFCnt()
	if fCnt > math.MaxUint16 {
		return ErrCorruptMessage.NewWithCause(nil, errors.Errorf("FCnt must be lower or equal to %d", math.MaxUint16))
	}

	md := msg.GetRxMetadata()
	if len(md) == 0 {
		return ErrCorruptMessage.NewWithCause(nil, errors.New("Empty rx metadata"))
	}

	txChIdx := md[0].GetChannelIndex()
	if txChIdx > math.MaxUint8 {
		return ErrCorruptMessage.NewWithCause(nil, errors.Errorf("TxChIdx must be lower or equal to %d", math.MaxUint8))
	}

	settings := msg.GetSettings()
	txDRIdx := settings.GetDataRateIndex()
	if txDRIdx > math.MaxUint8 {
		return ErrCorruptMessage.NewWithCause(nil, errors.Errorf("TxDRIdx must be lower or equal to %d", math.MaxUint8))
	}

	dev, err := ns.matchDevice(pld.DevAddr, uint8(txDRIdx), uint8(txChIdx), fCnt, pld.GetAck(), msg.Payload.GetMIC(), pld.GetFRMPayload())
	if err != nil {
		return errors.NewWithCause(err, "Failed to match device")
	}

	if len(dev.RecentUplinks) >= recentUplinkCount {
		dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
	}
	dev.RecentUplinks = append(dev.RecentUplinks, msg)

	if err := dev.Update("Session", "RecentUplinks"); err != nil {
		logger.WithError(err).Error("Failed to update device")
		return err
	}

	go func() {
		ns.applicationServersMu.RLock()
		cl, ok := ns.applicationServers[dev.ApplicationID]
		if ok {
			up := &ttnpb.ApplicationUp{&ttnpb.ApplicationUp_UplinkMessage{&ttnpb.ApplicationUplink{
				FPort:      pld.FPort,
				FCnt:       dev.Session.NextFCntUp - 1,
				FRMPayload: pld.FRMPayload,
			}}}
			logger := logger.WithFields(log.Fields("application_id", dev.ApplicationID, "f_cnt", up.GetUplinkMessage().GetFCnt()))

			if err := cl.Send(up); err != nil {
				logger.WithError(err).Warn("Failed to send uplink to application server")
			}
		}
		ns.applicationServersMu.RUnlock()
	}()
	return nil
}

// newDevAddr generates a DevAddr for specified EndDevice.
func (ns *NetworkServer) newDevAddr(*ttnpb.EndDevice) types.DevAddr {
	nwkAddr := make([]byte, types.NwkAddrLength(ns.NetID))
	rand.Read(nwkAddr)
	nwkAddr[0] &= 0xff >> (8 - types.NwkAddrBits(ns.NetID)%8)
	devAddr, err := types.NewDevAddr(ns.NetID, nwkAddr)
	if err != nil {
		panic(errors.NewWithCause(err, "Failed to create new DevAddr"))
	}
	return devAddr
}

func (ns *NetworkServer) handleJoin(ctx context.Context, uplink *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	logger := ns.Logger()

	pld := uplink.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, ErrMissingPayload.New(nil)
	}

	if pld.DevEUI.IsZero() {
		return nil, ErrMissingDevEUI.New(nil)
	}
	if pld.JoinEUI.IsZero() {
		return nil, ErrMissingJoinEUI.New(nil)
	}

	var joinServers []ttnpb.NsJsClient
	ns.joinServers.Range(func(prefix interface{}, cl interface{}) bool {
		if prefix.(types.EUI64Prefix).Matches(pld.DevEUI) {
			joinServers = append(joinServers, cl.(ttnpb.NsJsClient))
		}
		return true
	})
	if len(joinServers) == 0 {
		return nil, ErrUnconfiguredDevEUI.New(errors.Attributes{
			"dev_eui": pld.DevEUI,
		})
	}

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI:  &pld.DevEUI,
		JoinEUI: &pld.JoinEUI,
	})
	if err != nil {
		logger.WithError(err).WithFields(log.Fields(
			"dev_eui", pld.DevEUI,
			"join_eui", pld.JoinEUI,
		)).Warn("Failed to search for device in registry")
		return nil, err
	}

	devAddr := ns.newDevAddr(dev.EndDevice)
	for devAddr.Equal(dev.GetSession().DevAddr) {
		devAddr = ns.newDevAddr(dev.EndDevice)
	}

	state := dev.GetMACState()
	req := &ttnpb.JoinRequest{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DevEUI:  &pld.DevEUI,
			JoinEUI: &pld.JoinEUI,
			DevAddr: &devAddr,
		},
		SelectedMacVersion: dev.GetLoRaWANVersion(),
		RxDelay:            state.GetRxDelay(),
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: state.GetRx1DataRateOffset(),
			Rx2DR:       state.GetRx2DataRateIndex(),
		},
	}

	var errs []error
	for _, js := range joinServers {
		resp, err := js.HandleJoin(ctx, req)
		if err != nil {
			errs = append(errs, err)
			logger.WithError(err).Warn("Join failed")
			continue
		}

		dev.SessionFallback = nil
		dev.Session = &ttnpb.Session{
			DevAddr:     devAddr,
			SessionKeys: resp.SessionKeys,
			StartedAt:   time.Now(),
		}

		if err = dev.Update("Session", "SessionFallback"); err != nil {
			logger.WithField("device", dev).WithError(err).Error("Failed to update device")
		}
		return &pbtypes.Empty{}, nil
	}
	return nil, errors.NewWithCause(errs[0], "Failed to perform join procedure")
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, uplink *ttnpb.UplinkMessage) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the gateway server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, msg *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	logger := ns.Logger()

	if msg.GetPayload().Payload == nil {
		b := msg.GetRawPayload()
		if len(b) == 0 {
			return nil, ErrMissingPayload.New(nil)
		}

		if err := msg.Payload.UnmarshalLoRaWAN(b); err != nil {
			return nil, ErrUnmarshalFailed.NewWithCause(nil, err)
		}
	}

	sum, err := hashstructure.Hash(msg.GetPayload().Payload, ns.hashOptions)
	if err != nil {
		panic(errors.NewWithCause(err, "Failed to hash payload"))
	}

	_, claimed := ns.claimedUplinks.LoadOrStore(sum, struct{}{})
	if claimed {
		logger.Info("Dropping duplicate uplink")
		return &pbtypes.Empty{}, nil
	}
	defer time.AfterFunc(deduplicationWindow, func() { ns.claimedUplinks.Delete(sum) })

	pld := msg.GetPayload()
	if pld.GetMajor() != ttnpb.Major_LORAWAN_R1 {
		return nil, ErrUnsupportedLoRaWANMajorVersion.New(errors.Attributes{
			"major": pld.GetMajor(),
		})
	}

	switch t := pld.GetMType(); t {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return &pbtypes.Empty{}, ns.handleUplink(ctx, msg)
	case ttnpb.MType_JOIN_REQUEST:
		return &pbtypes.Empty{}, ns.handleJoin(ctx, msg)
	case ttnpb.MType_REJOIN_REQUEST:
		return &pbtypes.Empty{}, ns.handleRejoin(ctx, msg)
	default:
		return nil, ErrWrongPayloadType.New(errors.Attributes{
			"type": t,
		})
	}
}

type applicationUplinkStream struct {
	ttnpb.AsNs_LinkApplicationServer
	closeCh chan struct{}
}

func (s applicationUplinkStream) Close() error {
	close(s.closeCh)
	return nil
}

type closer interface {
	Close() error
}

// LinkApplication is called by the application server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(id *ttnpb.ApplicationIdentifier, stream ttnpb.AsNs_LinkApplicationServer) error {
	ws := &applicationUplinkStream{
		stream,
		make(chan struct{}),
	}
	appID := id.GetApplicationID()

	ns.applicationServersMu.Lock()
	cl, ok := ns.applicationServers[appID]
	ns.applicationServers[appID] = ws
	if ok {
		if err := cl.Close(); err != nil {
			ns.applicationServersMu.Unlock()
			return err
		}
	}
	ns.applicationServersMu.Unlock()

	ctx := stream.Context()
	select {
	case <-ctx.Done():
		err := ctx.Err()
		ns.applicationServersMu.Lock()
		cl, ok := ns.applicationServers[appID]
		if ok && cl == ws {
			delete(ns.applicationServers, appID)
		}
		ns.applicationServersMu.Unlock()
		return err
	case <-ws.closeCh:
		return errors.New("Another uplink subscription started for application")
	}
}

// DownlinkQueueReplace is called by the application server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if req.EndDeviceIdentifiers.IsZero() {
		return nil, errors.New("Empty identifiers specified")
	}
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = req.Downlinks
	return &pbtypes.Empty{}, dev.Update("QueuedApplicationDownlinks")
}

// DownlinkQueuePush is called by the application server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if req.EndDeviceIdentifiers.IsZero() {
		return nil, errors.New("Empty identifiers specified")
	}
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = append(dev.EndDevice.QueuedApplicationDownlinks, req.Downlinks...)
	return &pbtypes.Empty{}, dev.Update("QueuedApplicationDownlinks")
}

// DownlinkQueueList is called by the application server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	if id.IsZero() {
		return nil, errors.New("Empty identifiers specified")
	}
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, id)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{Downlinks: dev.EndDevice.QueuedApplicationDownlinks}, nil
}

// DownlinkQueueClear is called by the application server to clear the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueClear(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if id.IsZero() {
		return nil, errors.New("Empty identifiers specified")
	}
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, id)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = nil
	return &pbtypes.Empty{}, dev.Update("QueuedApplicationDownlinks")
}

// RegisterServices registers services provided by ns at s.
func (ns *NetworkServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsNsServer(s, ns)
	ttnpb.RegisterAsNsServer(s, ns)
	ttnpb.RegisterNsApplicationDownlinkQueueServer(s, ns)
	ttnpb.RegisterNsDeviceRegistryServer(s, ns)
}

// RegisterHandlers registers gRPC handlers.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterNsDeviceRegistryHandler(ns.Context(), s, conn)
}

// Roles returns the roles that the network server fulfils
func (ns *NetworkServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_NETWORK_SERVER}
}
