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
	"hash"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// recentUplinkCount is the maximium amount of recent uplinks stored per device.
const recentUplinkCount = 20

const maxFCntGap = 16384
const accumulationCapacity = 20

// NetworkServer implements the network server component.
//
// The network server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component
	*deviceregistry.RegistryRPC
	registry deviceregistry.Interface

	NetID types.NetID

	joinServers          []ttnpb.NsJsClient
	gateways             *sync.Map // gtwID -> ttnpb.NsGsClient
	applicationServersMu *sync.RWMutex
	applicationServers   map[string]*applicationUplinkStream

	metadataAccumulators *sync.Map

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	deduplicationWindow time.Duration
	cooldownWindow      time.Duration
}

// Config represents the NetworkServer configuration.
type Config struct {
	Registry            deviceregistry.Interface `name:"-"`
	JoinServers         []ttnpb.NsJsClient       `name:"-"`
	DeduplicationWindow time.Duration            `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`
	CooldownWindow      time.Duration            `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"`
}

// New returns new *NetworkServer.
func New(c *component.Component, conf *Config) *NetworkServer {
	ns := &NetworkServer{
		Component:               c,
		RegistryRPC:             deviceregistry.NewRPC(c, conf.Registry), // TODO: Add checks https://github.com/TheThingsIndustries/ttn/issues/558
		registry:                conf.Registry,
		applicationServersMu:    &sync.RWMutex{},
		applicationServers:      make(map[string]*applicationUplinkStream),
		gateways:                &sync.Map{},
		metadataAccumulators:    &sync.Map{},
		metadataAccumulatorPool: &sync.Pool{},
		hashPool:                &sync.Pool{},
		deduplicationWindow:     conf.DeduplicationWindow,
		cooldownWindow:          conf.CooldownWindow,
		joinServers:             conf.JoinServers,
	}
	ns.hashPool.New = func() interface{} { return fnv.New64a() }
	ns.metadataAccumulatorPool.New = func() interface{} {
		return &metadataAccumulator{newAccumulator()}
	}
	c.RegisterGRPC(ns)
	return ns
}

type applicationUplinkStream struct {
	ttnpb.AsNs_LinkApplicationServer
	closeCh chan struct{}
}

func (s applicationUplinkStream) Close() error {
	close(s.closeCh)
	return nil
}

// LinkApplication is called by the application server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(id *ttnpb.ApplicationIdentifiers, stream ttnpb.AsNs_LinkApplicationServer) error {
	ws := &applicationUplinkStream{
		AsNs_LinkApplicationServer: stream,
		closeCh:                    make(chan struct{}),
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
		return ErrNewSubscription.New(nil)
	}
}

// DownlinkQueueReplace is called by the application server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	// TODO: authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = req.Downlinks
	return &pbtypes.Empty{}, dev.Store("QueuedApplicationDownlinks")
}

// DownlinkQueuePush is called by the application server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	// TODO: authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = append(dev.EndDevice.QueuedApplicationDownlinks, req.Downlinks...)
	return &pbtypes.Empty{}, dev.Store("QueuedApplicationDownlinks")
}

// DownlinkQueueList is called by the application server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	// TODO: authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, id)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{Downlinks: dev.EndDevice.QueuedApplicationDownlinks}, nil
}

// DownlinkQueueClear is called by the application server to clear the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueClear(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	// TODO: authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, id)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = nil
	return &pbtypes.Empty{}, dev.Store("QueuedApplicationDownlinks")
}

// StartServingGateway is called by the gateway server to indicate that it is serving a gateway.
func (ns *NetworkServer) StartServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// StopServingGateway is called by the gateway server to indicate that it is no longer serving a gateway.
func (ns *NetworkServer) StopServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

type accumulator struct {
	accumulation *sync.Map
}

func newAccumulator() *accumulator {
	return &accumulator{
		accumulation: &sync.Map{},
	}
}

func (a *accumulator) Add(v interface{}) {
	a.accumulation.Store(v, struct{}{})
}

func (a *accumulator) Range(f func(v interface{})) {
	a.accumulation.Range(func(k, _ interface{}) bool {
		f(k)
		return true
	})
}

func (a *accumulator) Reset() {
	a.Range(a.accumulation.Delete)
}

type metadataAccumulator struct {
	*accumulator
}

func (a *metadataAccumulator) Accumulated() []*ttnpb.RxMetadata {
	md := make([]*ttnpb.RxMetadata, 0, accumulationCapacity)
	a.accumulator.Range(func(k interface{}) {
		md = append(md, k.(*ttnpb.RxMetadata))
	})
	return md
}

func (a *metadataAccumulator) Add(mds ...*ttnpb.RxMetadata) {
	for _, md := range mds {
		a.accumulator.Add(md)
	}
}

func (ns *NetworkServer) deduplicateUplink(msg *ttnpb.UplinkMessage) (*metadataAccumulator, bool) {
	start := time.Now()

	h := ns.hashPool.Get().(hash.Hash)
	h.Write(msg.GetRawPayload())

	k := string(h.Sum(nil))

	h.Reset()
	ns.hashPool.Put(h)

	a := ns.metadataAccumulatorPool.Get().(*metadataAccumulator)
	lv, isDup := ns.metadataAccumulators.LoadOrStore(k, a)
	lv.(*metadataAccumulator).Add(msg.RxMetadata...)

	if isDup {
		ns.metadataAccumulatorPool.Put(a)
		return nil, true
	}

	go func() {
		time.Sleep(time.Until(start.Add(ns.deduplicationWindow + ns.cooldownWindow)))
		ns.metadataAccumulators.Delete(k)
	}()
	return a, false
}

// matchDevice tries to match the uplink message with a device.
// If the uplink message matches a fallback session, that fallback session is recovered.
// If successful, the FCnt in the uplink message is set to the full FCnt. The NextFCntUp in the session is updated accordingly.
func (ns *NetworkServer) matchDevice(msg *ttnpb.UplinkMessage) (*deviceregistry.Device, error) {
	mac := msg.Payload.GetMACPayload()

	devs, err := ns.registry.FindBy(
		&ttnpb.EndDevice{
			Session: &ttnpb.Session{
				DevAddr: mac.DevAddr,
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
				DevAddr: mac.DevAddr,
			},
		},
		"SessionFallback.DevAddr",
	)
	if err != nil {
		return nil, err
	}

	for _, dev := range fb {
		dev.EndDevice.Session = dev.EndDevice.SessionFallback
		dev.EndDevice.EndDeviceIdentifiers.DevAddr = &dev.EndDevice.Session.DevAddr
	}
	devs = append(devs, fb...)

	type device struct {
		*deviceregistry.Device
		fCnt uint32
		gap  uint32
	}
	matching := make([]device, 0, len(devs))

outer:
	for _, dev := range devs {
		dev.EndDevice.SessionFallback = nil

		fCnt := mac.GetFCnt()
		next := dev.EndDevice.GetSession().GetNextFCntUp()

		switch {
		case dev.FCntIs16Bit, fCnt >= next:
		case fCnt > next&0xffff:
			fCnt |= next &^ 0xffff
		case next < 0xffff<<16:
			fCnt |= (next + 1<<16) &^ 0xffff
		case !dev.FCntResets:
			continue outer
		}

		var gap uint32
		if dev.FCntResets {
			gap = math.MaxUint32
		} else {
			gap = fCnt - next

			switch dev.EndDevice.GetLoRaWANVersion() {
			case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
				if gap > maxFCntGap {
					continue outer
				}
			}
		}

		matching = append(matching, device{
			Device: dev,
			fCnt:   fCnt,
			gap:    gap,
		})
		if dev.FCntResets && fCnt != mac.GetFCnt() {
			matching = append(matching, device{
				Device: dev,
				fCnt:   mac.GetFCnt(),
				gap:    gap,
			})
		}
	}
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	if len(msg.RawPayload) < 4 {
		return nil, errors.New("Length of RawPayload must not be less than 4")
	}
	pld := msg.RawPayload[:len(msg.RawPayload)-4]

	for _, dev := range matching {
		ses := dev.GetSession()

		ke := ses.GetFNwkSIntKey()
		if ke == nil || ke.Key.IsZero() {
			return nil, ErrCorruptRegistry.NewWithCause(nil, ErrMissingFNwkSIntKey.New(nil))
		}
		fNwkSIntKey := *ke.Key

		if mac.GetAck() {
			if len(dev.RecentDownlinks) == 0 {
				// Uplink acknowledges a downlink, but no downlink was sent by the device,
				// hence it must be the wrong device.
				continue
			}
		}

		var computedMIC [4]byte
		switch dev.LoRaWANVersion {
		case ttnpb.MAC_V1_1:
			ke := ses.GetSNwkSIntKey()
			if ke == nil || ke.Key.IsZero() {
				return nil, ErrCorruptRegistry.NewWithCause(nil, ErrMissingSNwkSIntKey.New(nil))
			}
			sNwkSIntKey := *ke.Key

			var confFCnt uint32
			if mac.GetAck() {
				confFCnt = dev.GetRecentDownlinks()[len(dev.RecentDownlinks)-1].Payload.GetMACPayload().GetFCnt()
			}
			set := msg.GetSettings()
			computedMIC, err = crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey,
				confFCnt, uint8(set.GetDataRateIndex()), uint8(set.GetChannelIndex()),
				mac.DevAddr, dev.fCnt, pld)
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(fNwkSIntKey, mac.DevAddr, dev.fCnt, pld)
		default:
			return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("Unmatched LoRaWAN version"))
		}
		if err != nil {
			return nil, ErrMICComputeFailed.NewWithCause(nil, err)
		}
		if !bytes.Equal(msg.Payload.GetMIC(), computedMIC[:]) {
			continue
		}
		if dev.fCnt == math.MaxUint32 {
			return nil, ErrFCntTooHigh.New(nil)
		}
		ses.NextFCntUp = dev.fCnt + 1
		return dev.Device, nil
	}
	return nil, ErrDeviceNotFound.New(nil)
}

func (ns *NetworkServer) handleUplink(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator, start time.Time) error {
	logger := ns.Logger()

	dev, err := ns.matchDevice(msg)
	if err != nil {
		return errors.NewWithCause(err, "Failed to match device")
	}

	time.Sleep(time.Until(start.Add(ns.deduplicationWindow)))
	msg.RxMetadata = acc.Accumulated()

	dev.RecentUplinks = append(dev.GetRecentUplinks(), msg)
	if len(dev.RecentUplinks) > recentUplinkCount {
		dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
	}

	if err := dev.Store("EndDeviceIdentifiers.DevAddr", "Session", "SessionFallback", "RecentUplinks"); err != nil {
		logger.WithError(err).Error("Failed to update device")
		return err
	}

	mac := msg.Payload.GetMACPayload()
	go func() {
		ns.applicationServersMu.RLock()
		cl, ok := ns.applicationServers[dev.EndDeviceIdentifiers.GetApplicationID()]
		ns.applicationServersMu.RUnlock()

		if ok {
			up := &ttnpb.ApplicationUp{&ttnpb.ApplicationUp_UplinkMessage{&ttnpb.ApplicationUplink{
				FCnt:       dev.GetSession().GetNextFCntUp() - 1,
				FPort:      mac.GetFPort(),
				FRMPayload: mac.GetFRMPayload(),
				RxMetadata: msg.GetRxMetadata(),
			}}}
			if err := cl.Send(up); err != nil {
				logger.WithFields(log.Fields(
					"application_id", dev.EndDeviceIdentifiers.GetApplicationID(),
					"f_cnt", up.GetUplinkMessage().GetFCnt(),
				)).WithError(err).Warn("Failed to send uplink to application server")
			}
		}
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

func (ns *NetworkServer) handleJoin(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator, start time.Time) error {
	pld := msg.Payload.GetJoinRequestPayload()

	logger := ns.Logger().WithFields(log.Fields(
		"dev_eui", pld.DevEUI,
		"join_eui", pld.JoinEUI,
	))

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI:  &pld.DevEUI,
		JoinEUI: &pld.JoinEUI,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to search for device in registry")
		return err
	}

	devAddr := ns.newDevAddr(dev.EndDevice)
	for s := dev.GetSession(); s != nil && devAddr.Equal(s.DevAddr); {
		devAddr = ns.newDevAddr(dev.EndDevice)
	}

	stDes := dev.GetMACStateDesired()

	req := &ttnpb.JoinRequest{
		RawPayload: msg.GetRawPayload(),
		Payload:    msg.GetPayload(),
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DevEUI:  &pld.DevEUI,
			JoinEUI: &pld.JoinEUI,
			DevAddr: &devAddr,
		},
		NetID:              ns.NetID,
		SelectedMacVersion: dev.GetLoRaWANVersion(),
		RxDelay:            stDes.GetRxDelay(),
		CFList:             nil, // TODO: Add if required https://github.com/TheThingsIndustries/ttn/issues/559
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: stDes.GetRx1DataRateOffset(),
			Rx2DR:       stDes.GetRx2DataRateIndex(),
		},
	}

	var errs []error
	for _, js := range ns.joinServers {
		resp, err := js.HandleJoin(ctx, req)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		dev.SessionFallback = nil
		dev.Session = &ttnpb.Session{
			DevAddr:     devAddr,
			SessionKeys: resp.SessionKeys,
			StartedAt:   time.Now(),
		}
		dev.EndDeviceIdentifiers.DevAddr = &devAddr

		time.Sleep(time.Until(start.Add(ns.deduplicationWindow)))
		msg.RxMetadata = acc.Accumulated()

		dev.RecentUplinks = append(dev.RecentUplinks, msg)
		if len(dev.RecentUplinks) > recentUplinkCount {
			dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
		}

		if err = dev.Store("EndDeviceIdentifiers.DevAddr", "Session", "SessionFallback", "RecentUplinks"); err != nil {
			logger.WithError(err).Error("Failed to update device")
			return err
		}
		return nil
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Warn("Join failed")
	return errors.NewWithCause(errors.New("No join server could handle join request"), "Failed to perform join procedure")
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator, start time.Time) error {
	// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the gateway server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, msg *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	start := time.Now()
	logger := ns.Logger()

	b := msg.GetRawPayload()

	pld := msg.GetPayload()
	if pld.Payload == nil {
		if err := msg.Payload.UnmarshalLoRaWAN(b); err != nil {
			return nil, ErrUnmarshalFailed.NewWithCause(nil, err)
		}
	}

	if pld.GetMajor() != ttnpb.Major_LORAWAN_R1 {
		return nil, ErrUnsupportedLoRaWANMajorVersion.New(errors.Attributes{
			"major": pld.GetMajor(),
		})
	}

	acc, ok := ns.deduplicateUplink(msg)
	if ok {
		logger.Debug("Dropping duplicate uplink")
		return &pbtypes.Empty{}, nil
	}

	switch pld.GetMType() {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return &pbtypes.Empty{}, ns.handleUplink(ctx, msg, acc, start)
	case ttnpb.MType_JOIN_REQUEST:
		return &pbtypes.Empty{}, ns.handleJoin(ctx, msg, acc, start)
	case ttnpb.MType_REJOIN_REQUEST:
		return &pbtypes.Empty{}, ns.handleRejoin(ctx, msg, acc, start)
	default:
		logger.Error("Unmatched MType")
		return &pbtypes.Empty{}, nil
	}
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
