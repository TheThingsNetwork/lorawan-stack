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

// Package networkserver provides a LoRaWAN 1.1-compliant Network Server implementation.
package networkserver

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// recentUplinkCount is the maximium amount of recent uplinks stored per device.
	recentUplinkCount = 20

	// accumulationCapacity is the initial capacity of the accumulator
	accumulationCapacity = 20
)

// WindowEndFunc is a function, which is used by Network Server to determine the end of deduplication and cooldown windows.
type WindowEndFunc func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time

// NewWindowEndAfterFunc returns a WindowEndFunc, which closes
// the returned channel after at least duration d after msg.ServerTime or if the context is done.
func NewWindowEndAfterFunc(d time.Duration) WindowEndFunc {
	return func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
		ch := make(chan time.Time, 1)

		start := msg.GetReceivedAt()
		if start.IsZero() {
			start = time.Now()
		}

		end := start.Add(d)
		if end.Before(time.Now()) {
			ch <- end
			return ch
		}

		go func() {
			time.Sleep(time.Until(start.Add(d)))
			ch <- end
		}()
		return ch
	}
}

// NsGsClientFunc is the function used to get gateway server.
type NsGsClientFunc func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error)

// PeerGetter is the interface, which wraps GetPeer method.
type PeerGetter interface {
	GetPeer(role ttnpb.PeerInfo_Role, tags []string, shardKey []byte) cluster.Peer
}

// NewGatewayServerPeerGetterFunc returns a NsGsClientFunc, which uses g to retrieve gateway server clients.
func NewGatewayServerPeerGetterFunc(g PeerGetter) NsGsClientFunc {
	return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
		p := g.GetPeer(
			ttnpb.PeerInfo_GATEWAY_SERVER,
			[]string{fmt.Sprintf("gtw=%s", id.UniqueID(ctx))},
			nil,
		)
		if p == nil {
			return nil, ErrGatewayServerNotFound.New(nil)
		}
		return ttnpb.NewNsGsClient(p.Conn()), nil
	}
}

// Config represents the NetworkServer configuration.
type Config struct {
	Registry            deviceregistry.Interface `name:"-"`
	JoinServers         []ttnpb.NsJsClient       `name:"-"`
	DeduplicationWindow time.Duration            `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`
	CooldownWindow      time.Duration            `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"`
}

// NetworkServer implements the Network Server component.
//
// The Network Server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component
	*deviceregistry.RegistryRPC
	registry deviceregistry.Interface

	NetID types.NetID

	joinServers []ttnpb.NsJsClient

	applicationServersMu *sync.RWMutex
	applicationServers   map[string]*applicationUpStream

	metadataAccumulators *sync.Map

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	deduplicationDone WindowEndFunc
	collectionDone    WindowEndFunc

	gsClient NsGsClientFunc
}

// Option configures the NetworkServer.
type Option func(ns *NetworkServer)

// WithDeduplicationDoneFunc overrides the default WindowEndFunc, which
// is used to determine the end of uplink metadata deduplication.
func WithDeduplicationDoneFunc(fn WindowEndFunc) Option {
	return func(ns *NetworkServer) {
		ns.deduplicationDone = fn
	}
}

// WithCollectionDoneFunc overrides the default WindowEndFunc, which
// is used to determine the end of uplink duplicate collection.
func WithCollectionDoneFunc(fn WindowEndFunc) Option {
	return func(ns *NetworkServer) {
		ns.collectionDone = fn
	}
}

// WithNsGsClientFunc overrides the default NsGsClientFunc, which
// is used to get the gateway server for a gateway identifiers.
func WithNsGsClientFunc(fn NsGsClientFunc) Option {
	return func(ns *NetworkServer) {
		ns.gsClient = fn
	}
}

// New returns new NetworkServer.
func New(c *component.Component, conf *Config, opts ...Option) (*NetworkServer, error) {
	ns := &NetworkServer{
		Component:               c,
		registry:                conf.Registry,
		joinServers:             conf.JoinServers,
		applicationServersMu:    &sync.RWMutex{},
		applicationServers:      make(map[string]*applicationUpStream),
		metadataAccumulators:    &sync.Map{},
		metadataAccumulatorPool: &sync.Pool{},
		hashPool:                &sync.Pool{},
	}
	ns.hashPool.New = func() interface{} { return fnv.New64a() }
	ns.metadataAccumulatorPool.New = func() interface{} {
		return &metadataAccumulator{newAccumulator()}
	}

	for _, opt := range opts {
		opt(ns)
	}

	registryRPC, err := deviceregistry.NewRPC(c, conf.Registry, deviceregistry.ForComponents(ttnpb.PeerInfo_NETWORK_SERVER)) // TODO: Add checks https://github.com/TheThingsIndustries/ttn/issues/558
	if err != nil {
		return nil, errors.NewWithCausef(err, "Could not initialize the Network Server's device registry RPC")
	}
	ns.RegistryRPC = registryRPC

	switch {
	case ns.deduplicationDone == nil && conf.DeduplicationWindow == 0:
		return nil, ErrInvalidConfiguration.NewWithCause(nil, errors.New("DeduplicationWindow is zero and WithDeduplicationDoneFunc not specified"))

	case ns.collectionDone == nil && conf.DeduplicationWindow == 0:
		return nil, ErrInvalidConfiguration.NewWithCause(nil, errors.New("DeduplicationWindow is zero and WithCollectionDoneFunc not specified"))

	case ns.collectionDone == nil && conf.CooldownWindow == 0:
		return nil, ErrInvalidConfiguration.NewWithCause(nil, errors.New("CooldownWindow is zero and WithCollectionDoneFunc not specified"))
	}

	if ns.deduplicationDone == nil {
		ns.deduplicationDone = NewWindowEndAfterFunc(conf.DeduplicationWindow)
	}
	if ns.collectionDone == nil {
		ns.collectionDone = NewWindowEndAfterFunc(conf.DeduplicationWindow + conf.CooldownWindow)
	}
	if ns.gsClient == nil {
		ns.gsClient = NewGatewayServerPeerGetterFunc(ns.Component)
	}

	c.RegisterGRPC(ns)
	return ns, nil
}

type applicationUpStream struct {
	ttnpb.AsNs_LinkApplicationServer
	closeCh chan struct{}
}

func (s applicationUpStream) Close() error {
	close(s.closeCh)
	return nil
}

// LinkApplication is called by the Application Server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(id *ttnpb.ApplicationIdentifiers, stream ttnpb.AsNs_LinkApplicationServer) error {
	ws := &applicationUpStream{
		AsNs_LinkApplicationServer: stream,
		closeCh:                    make(chan struct{}),
	}

	ctx := stream.Context()
	uid := id.UniqueID(ctx)

	ns.applicationServersMu.Lock()
	cl, ok := ns.applicationServers[uid]
	ns.applicationServers[uid] = ws
	if ok {
		if err := cl.Close(); err != nil {
			ns.applicationServersMu.Unlock()
			return err
		}
	}
	ns.applicationServersMu.Unlock()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		ns.applicationServersMu.Lock()
		cl, ok := ns.applicationServers[uid]
		if ok && cl == ws {
			delete(ns.applicationServers, uid)
		}
		ns.applicationServersMu.Unlock()
		return err
	case <-ws.closeCh:
		return ErrNewSubscription.New(nil)
	}
}

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	// TODO: authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}

	dev.EndDevice.QueuedApplicationDownlinks = req.Downlinks
	if err = dev.Store("QueuedApplicationDownlinks"); err != nil {
		return nil, err
	}

	if dev.EndDevice.GetMACInfo().GetDeviceClass() == ttnpb.CLASS_C {
		// TODO: Schedule the next downlink (https://github.com/TheThingsIndustries/ttn/issues/728)
	}
	return ttnpb.Empty, nil
}

// DownlinkQueuePush is called by the Application Server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	// TODO: Authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}

	dev.EndDevice.QueuedApplicationDownlinks = append(dev.EndDevice.QueuedApplicationDownlinks, req.Downlinks...)
	if err := dev.Store("QueuedApplicationDownlinks"); err != nil {
		return nil, err
	}

	if dev.EndDevice.GetMACInfo().GetDeviceClass() == ttnpb.CLASS_C {
		// TODO: Schedule the next downlink (https://github.com/TheThingsIndustries/ttn/issues/728)
	}
	return ttnpb.Empty, nil
}

// DownlinkQueueList is called by the Application Server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	// TODO: Authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, devID)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{Downlinks: dev.EndDevice.QueuedApplicationDownlinks}, nil
}

// DownlinkQueueClear is called by the Application Server to clear the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueClear(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	// TODO: Authentication https://github.com/TheThingsIndustries/ttn/issues/558
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, devID)
	if err != nil {
		return nil, err
	}
	dev.EndDevice.QueuedApplicationDownlinks = nil
	return ttnpb.Empty, dev.Store("QueuedApplicationDownlinks")
}

// StartServingGateway is called by the Gateway Server to indicate that it is serving a gateway.
func (ns *NetworkServer) StartServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	uid := gtwID.UniqueID(ctx)
	if uid == "" {
		return nil, ErrMissingGatewayID.New(nil)
	}

	gsID := rpcmetadata.FromIncomingContext(ctx).ID
	// TODO: Associate the GS ID with the gateway uid in the cluster once
	// https://github.com/TheThingsIndustries/ttn/issues/506#issuecomment-385963158 is resolved
	_ = gsID
	return ttnpb.Empty, nil
}

// StopServingGateway is called by the Gateway Server to indicate that it is no longer serving a gateway.
func (ns *NetworkServer) StopServingGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	uid := id.UniqueID(ctx)
	if uid == "" {
		return nil, ErrMissingGatewayID.New(nil)
	}

	gsID := rpcmetadata.FromIncomingContext(ctx).ID
	// TODO: Deassociate the GS ID with the gateway uid in the cluster once
	// https://github.com/TheThingsIndustries/ttn/issues/506#issuecomment-385963158 is resolved
	_ = gsID
	return ttnpb.Empty, nil
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

func (ns *NetworkServer) deduplicateUplink(ctx context.Context, msg *ttnpb.UplinkMessage) (*metadataAccumulator, bool) {
	h := ns.hashPool.Get().(hash.Hash)
	_, _ = h.Write(msg.GetRawPayload())

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

	go func(msg *ttnpb.UplinkMessage) {
		<-ns.collectionDone(ctx, msg)
		ns.metadataAccumulators.Delete(k)
	}(deepcopy.Copy(msg).(*ttnpb.UplinkMessage))
	return a, false
}

func setDownlinkModulation(s *ttnpb.TxSettings, dr band.DataRate) (err error) {
	if dr.Rate.LoRa != "" && dr.Rate.FSK > 0 {
		return errors.New("Both LoRa and FSK present - inconsistent data rate")
	}

	if dr.Rate.LoRa == "" {
		s.Modulation = ttnpb.Modulation_FSK
		s.BitRate = dr.Rate.FSK
		s.SpreadingFactor = 0
		s.Bandwidth = 0
		return nil
	}

	sf, err := dr.Rate.SpreadingFactor()
	if err != nil {
		return err
	}

	bw, err := dr.Rate.Bandwidth()
	if err != nil {
		return err
	}

	s.Modulation = ttnpb.Modulation_LORA
	s.SpreadingFactor = uint32(sf)
	s.Bandwidth = bw
	s.BitRate = 0
	return nil
}

func (ns *NetworkServer) scheduleDownlink(ctx context.Context, dev *deviceregistry.Device, up *ttnpb.UplinkMessage, acc *metadataAccumulator, b []byte, isJoinAccept bool) error {
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", dev.EndDevice.EndDeviceIdentifiers.ApplicationIdentifiers.GetApplicationID(),
		"device_id", dev.EndDevice.EndDeviceIdentifiers.GetDeviceID(),
	))

	msg := &ttnpb.DownlinkMessage{
		RawPayload:           b,
		EndDeviceIdentifiers: dev.EndDevice.EndDeviceIdentifiers,
	}

	type tx struct {
		ttnpb.TxSettings
		Delay uint32
	}
	slots := make([]tx, 0, 2)

	fp, err := ns.Component.FrequencyPlans.GetByID(dev.EndDevice.GetFrequencyPlanID())
	if err != nil {
		return common.ErrCorruptRegistry.NewWithCause(nil, err)
	}

	band, err := band.GetByID(fp.GetBandID())
	if err != nil {
		return common.ErrCorruptRegistry.NewWithCause(nil, err)
	}

	st := dev.EndDevice.GetMACState()

	var mds []*ttnpb.RxMetadata
	if up == nil {
		// Class C
		ups := dev.EndDevice.GetRecentUplinks()
		if len(ups) == 0 {
			return ErrUplinkNotFound.New(nil)
		}
		mds = ups[len(ups)-1].GetRxMetadata()
	} else {
		sets := up.GetSettings()

		chIdx, err := band.Rx1Channel(sets.GetChannelIndex())
		if err != nil {
			return err
		}
		if uint(chIdx) >= uint(len(fp.Channels)) {
			return ErrChannelIndexTooHigh.New(nil)
		}

		drIdx, err := band.Rx1DataRate(sets.GetDataRateIndex(), st.GetRx1DataRateOffset(), st.GetDownlinkDwellTime())
		if err != nil {
			return err
		}

		rx1 := tx{
			TxSettings: ttnpb.TxSettings{
				DataRateIndex:         drIdx,
				CodingRate:            "4/5",
				PolarizationInversion: true,
				ChannelIndex:          chIdx,
				Frequency:             uint64(fp.Channels[chIdx].Frequency),
				TxPower:               int32(band.DefaultMaxEIRP),
			},
		}
		if isJoinAccept {
			rx1.Delay = uint32(band.JoinAcceptDelay1.Nanoseconds())
		} else {
			rx1.Delay = st.GetRxDelay()
		}

		if err = setDownlinkModulation(&rx1.TxSettings, band.DataRates[drIdx]); err != nil {
			return err
		}

		mds = up.GetRxMetadata()
		slots = append(slots, rx1)
	}

	drIdx := st.GetRx2DataRateIndex()
	if uint(drIdx) > uint(len(band.DataRates)) {
		return common.ErrCorruptRegistry.NewWithCause(nil, errors.Errorf("RX2 data rate index must be lower or equal to %d", len(band.DataRates)-1))
	}

	rx2 := tx{
		TxSettings: ttnpb.TxSettings{
			DataRateIndex:         drIdx,
			CodingRate:            "4/5",
			PolarizationInversion: true,
			Frequency:             st.GetRx2Frequency(),
			TxPower:               int32(band.DefaultMaxEIRP),
		},
	}
	if isJoinAccept {
		rx2.Delay = uint32(band.JoinAcceptDelay2.Nanoseconds())
	} else {
		rx2.Delay = st.GetRxDelay() + uint32(time.Second.Nanoseconds())
	}

	if err = setDownlinkModulation(&rx2.TxSettings, band.DataRates[drIdx]); err != nil {
		return err
	}

	slots = append(slots, rx2)

	if acc != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ns.deduplicationDone(ctx, up):
		}

		mds = acc.Accumulated()
	}

	sort.SliceStable(mds, func(i, j int) bool {
		// TODO: Improve the sorting algorithm (https://github.com/TheThingsIndustries/ttn/issues/729)
		return mds[i].GetSNR() > mds[j].GetSNR()
	})

	var errs []error
	for _, s := range slots {
		msg.Settings = s.TxSettings

		for _, md := range mds {
			logger := logger.WithField(
				"gateway_id", md.GatewayIdentifiers.GetGatewayID(),
			)

			cl, err := ns.gsClient(ctx, md.GatewayIdentifiers)
			if err != nil {
				logger.WithError(err).Debug("Could not get gateway server")
				continue
			}

			msg.TxMetadata = ttnpb.TxMetadata{
				GatewayIdentifiers: md.GatewayIdentifiers,
				Timestamp:          md.GetTimestamp() + uint64(s.Delay),
			}

			_, err = cl.ScheduleDownlink(ctx, msg)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			dev.EndDevice.RecentDownlinks = append(dev.EndDevice.GetRecentDownlinks(), msg)
			if err = dev.Store("RecentDownlinks"); err != nil {
				logger.WithError(err).Error("Failed to store device")
				return err
			}
			return nil
		}
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Debug("Failed to schedule downlink")
	return errors.New("Failed to schedule downlink")
}

func (ns *NetworkServer) scheduleApplicationDownlink(ctx context.Context, dev *deviceregistry.Device, up *ttnpb.UplinkMessage, acc *metadataAccumulator) error {
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", dev.EndDevice.EndDeviceIdentifiers.ApplicationIdentifiers.GetApplicationID(),
		"device_id", dev.EndDevice.EndDeviceIdentifiers.GetDeviceID(),
	))

	dev, err := dev.Load()
	if err != nil {
		logger.WithError(err).Error("Failed to load device")
		return err
	}

	devAddr := *dev.EndDevice.EndDeviceIdentifiers.DevAddr

	queue := dev.EndDevice.GetQueuedApplicationDownlinks()
	if len(queue) == 0 {
		logger.Debug("Downlink queue empty")
		return nil
	}
	down, queue := queue[0], queue[1:]

	fCnt := down.GetFCnt()

	b, err := (ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: ttnpb.MType_UNCONFIRMED_DOWN, // TODO: Support confirmed downlinks (https://github.com/TheThingsIndustries/ttn/issues/730)
		},
		Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
			FHDR: ttnpb.FHDR{
				DevAddr: devAddr,
				FCtrl: ttnpb.FCtrl{
					ADR:      false,
					Ack:      up != nil && up.Payload.GetMType() == ttnpb.MType_CONFIRMED_UP,
					FPending: len(queue) > 0,
				},
				FCnt:  fCnt,
				FOpts: nil, // TODO: MAC handling (https://github.com/TheThingsIndustries/ttn/issues/292)
			},
			FPort:      down.GetFPort(),
			FRMPayload: down.GetFRMPayload(),
		}},
	}).MarshalLoRaWAN()
	if err != nil {
		logger.WithError(err).Error("Failed to marshal payload")
		return err
	}
	// NOTE: It is assumed, that b does not contain MIC.

	ses := dev.EndDevice.GetSession()
	if ses == nil {
		logger.Debug("No active session found for device")
		return nil
	}

	ke := ses.SessionKeys.GetSNwkSIntKey()
	if ke == nil || ke.Key.IsZero() {
		return common.ErrCorruptRegistry.NewWithCause(nil, ErrMissingSNwkSIntKey.New(nil))
	}

	mic, err := crypto.ComputeDownlinkMIC(*ke.Key, devAddr, fCnt, b)
	if err != nil {
		logger.WithError(err).Error("Failed to compute downlink MIC")
		return err
	}

	if err := ns.scheduleDownlink(ctx, dev, up, acc, append(b, mic[:]...), false); err != nil {
		return err
	}

	dev.EndDevice.Session.NextAFCntDown++
	dev.EndDevice.QueuedApplicationDownlinks = queue
	if err = dev.Store("QueuedApplicationDownlinks", "Session.NextAFCntDown"); err != nil {
		logger.WithError(err).Error("Failed to store device")
		return err
	}

	if len(queue) > 0 && dev.EndDevice.GetMACInfo().GetDeviceClass() == ttnpb.CLASS_C {
		// TODO: Schedule the next downlink (https://github.com/TheThingsIndustries/ttn/issues/728)
	}
	return nil
}

// matchDevice tries to match the uplink message with a device.
// If the uplink message matches a fallback session, that fallback session is recovered.
// If successful, the FCnt in the uplink message is set to the full FCnt. The NextFCntUp in the session is updated accordingly.
func (ns *NetworkServer) matchDevice(ctx context.Context, msg *ttnpb.UplinkMessage) (*deviceregistry.Device, error) {
	pld := msg.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithField("dev_addr", pld.DevAddr)

	var devs []*deviceregistry.Device
	if err := ns.registry.Range(
		&ttnpb.EndDevice{
			Session: &ttnpb.Session{
				DevAddr: pld.DevAddr,
			},
		},
		0,
		func(d *deviceregistry.Device) bool {
			devs = append(devs, d)
			return true
		},
		"Session.DevAddr",
	); err != nil {
		logger.WithError(err).Warn("Failed to search for device in registry by active DevAddr")
		return nil, err
	}

	if err := ns.registry.Range(
		&ttnpb.EndDevice{
			SessionFallback: &ttnpb.Session{
				DevAddr: pld.DevAddr,
			},
		},
		0,
		func(d *deviceregistry.Device) bool {
			d.EndDevice.Session = d.EndDevice.SessionFallback
			d.EndDevice.EndDeviceIdentifiers.DevAddr = &d.EndDevice.Session.DevAddr
			devs = append(devs, d)
			return true
		},
		"SessionFallback.DevAddr",
	); err != nil {
		logger.WithError(err).Warn("Failed to search for device in registry by fallback DevAddr")
		return nil, err
	}

	type device struct {
		*deviceregistry.Device
		fCnt uint32
		gap  uint32
	}
	matching := make([]device, 0, len(devs))

outer:
	for _, dev := range devs {
		dev.EndDevice.SessionFallback = nil

		fCnt := pld.GetFCnt()
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
				fp, err := ns.Component.FrequencyPlans.GetByID(dev.GetFrequencyPlanID())
				if err != nil {
					return nil, common.ErrCorruptRegistry.NewWithCause(nil, err)
				}

				band, err := band.GetByID(fp.GetBandID())
				if err != nil {
					return nil, common.ErrCorruptRegistry.NewWithCause(nil, err)
				}

				if gap > uint32(band.MaxFCntGap) {
					continue outer
				}
			}
		}

		matching = append(matching, device{
			Device: dev,
			fCnt:   fCnt,
			gap:    gap,
		})
		if dev.FCntResets && fCnt != pld.GetFCnt() {
			matching = append(matching, device{
				Device: dev,
				fCnt:   pld.GetFCnt(),
				gap:    gap,
			})
		}
	}
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	if len(msg.RawPayload) < 4 {
		logger.Debug("Payload specified is too short")
		return nil, errors.New("Length of RawPayload must not be less than 4")
	}
	b := msg.RawPayload[:len(msg.RawPayload)-4]

	for _, dev := range matching {
		ses := dev.GetSession()

		ke := ses.GetFNwkSIntKey()
		if ke == nil || ke.Key.IsZero() {
			return nil, common.ErrCorruptRegistry.NewWithCause(nil, ErrMissingFNwkSIntKey.New(nil))
		}
		fNwkSIntKey := *ke.Key

		if pld.GetAck() {
			if len(dev.RecentDownlinks) == 0 {
				// Uplink acknowledges a downlink, but no downlink was sent by the device,
				// hence it must be the wrong device.
				continue
			}
		}

		var computedMIC [4]byte
		var err error
		switch dev.LoRaWANVersion {
		case ttnpb.MAC_V1_1:
			ke := ses.GetSNwkSIntKey()
			if ke == nil || ke.Key.IsZero() {
				return nil, common.ErrCorruptRegistry.NewWithCause(nil, ErrMissingSNwkSIntKey.New(nil))
			}
			sNwkSIntKey := *ke.Key

			var confFCnt uint32
			if pld.GetAck() {
				confFCnt = dev.GetRecentDownlinks()[len(dev.RecentDownlinks)-1].Payload.GetMACPayload().GetFCnt()
			}
			set := msg.GetSettings()
			computedMIC, err = crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey,
				confFCnt, uint8(set.GetDataRateIndex()), uint8(set.GetChannelIndex()),
				pld.DevAddr, dev.fCnt, b)
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(fNwkSIntKey, pld.DevAddr, dev.fCnt, b)
		default:
			return nil, common.ErrCorruptRegistry.NewWithCause(nil, errors.New("Unmatched LoRaWAN version"))
		}
		if err != nil {
			return nil, common.ErrComputeMIC.NewWithCause(nil, err)
		}
		if !bytes.Equal(msg.Payload.GetMIC(), computedMIC[:]) {
			continue
		}
		if dev.fCnt == math.MaxUint32 {
			return nil, common.ErrFCntTooHigh.New(nil)
		}
		ses.NextFCntUp = dev.fCnt + 1
		return dev.Device, nil
	}
	return nil, ErrDeviceNotFound.New(nil)
}

func (ns *NetworkServer) handleUplink(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator) error {
	dev, err := ns.matchDevice(ctx, msg)
	if err != nil {
		return errors.NewWithCause(err, "Failed to match device")
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", dev.EndDeviceIdentifiers.ApplicationIdentifiers.GetApplicationID(),
		"device_id", dev.EndDeviceIdentifiers.GetDeviceID(),
	))

	go ns.scheduleApplicationDownlink(ctx, dev, msg, acc)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, msg):
	}

	msg.RxMetadata = acc.Accumulated()

	ups := dev.GetRecentUplinks()
	if len(ups) >= recentUplinkCount {
		ups = ups[len(dev.RecentUplinks)-recentUplinkCount+1:]
	}
	dev.RecentUplinks = append(ups, msg)

	if err := dev.Store("EndDeviceIdentifiers.DevAddr", "Session", "SessionFallback", "RecentUplinks"); err != nil {
		logger.WithError(err).Error("Failed to update device")
		return err
	}
	uid := dev.EndDeviceIdentifiers.ApplicationIdentifiers.UniqueID(ctx)
	if uid == "" {
		return common.ErrCorruptRegistry.NewWithCause(nil, ErrMissingApplicationID.New(nil))
	}

	ns.applicationServersMu.RLock()
	cl, ok := ns.applicationServers[uid]
	ns.applicationServersMu.RUnlock()

	if !ok {
		return nil
	}

	pld := msg.Payload.GetMACPayload()
	ses := dev.GetSession()
	return cl.Send(&ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		SessionKeyID:         ses.GetSessionKeyID(),
		Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
			FCnt:       ses.GetNextFCntUp() - 1,
			FPort:      pld.GetFPort(),
			FRMPayload: pld.GetFRMPayload(),
			RxMetadata: msg.GetRxMetadata(),
		}},
	})
}

// newDevAddr generates a DevAddr for specified EndDevice.
func (ns *NetworkServer) newDevAddr(*ttnpb.EndDevice) types.DevAddr {
	nwkAddr := make([]byte, types.NwkAddrLength(ns.NetID))
	random.Read(nwkAddr)
	nwkAddr[0] &= 0xff >> (8 - types.NwkAddrBits(ns.NetID)%8)
	devAddr, err := types.NewDevAddr(ns.NetID, nwkAddr)
	if err != nil {
		panic(errors.NewWithCause(err, "Failed to create new DevAddr"))
	}
	return devAddr
}

func (ns *NetworkServer) handleJoin(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator) error {
	pld := msg.Payload.GetJoinRequestPayload()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", pld.DevEUI,
		"join_eui", pld.JoinEUI,
	))

	dev, err := deviceregistry.FindByIdentifiers(ns.registry, &ttnpb.EndDeviceIdentifiers{
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

	fp, err := ns.FrequencyPlans.GetByID(dev.GetFrequencyPlanID())
	if err != nil {
		return common.ErrCorruptRegistry.NewWithCause(nil, err)
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
		CFList:             frequencyplans.CFList(fp, dev.GetLoRaWANPHYVersion()),
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

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ns.deduplicationDone(ctx, msg):
		}

		msg.RxMetadata = acc.Accumulated()

		dev.RecentUplinks = append(dev.RecentUplinks, msg)
		if len(dev.RecentUplinks) > recentUplinkCount {
			dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
		}

		if err := ns.scheduleDownlink(ctx, dev, msg, nil, resp.GetRawPayload(), true); err != nil {
			logger.WithError(err).Debug("Failed to schedule join accept")
			return err
		}

		band, err := band.GetByID(fp.GetBandID())
		if err != nil {
			return common.ErrCorruptRegistry.NewWithCause(nil, err)
		}

		dev.EndDevice.MACState = &ttnpb.MACState{
			MaxTxPower:        uint32(dev.EndDevice.MaxTxPower),
			UplinkDwellTime:   fp.DwellTime != nil,
			DownlinkDwellTime: false, // TODO: Get this from band (https://github.com/TheThingsIndustries/ttn/issues/774)
			ADRNbTrans:        1,
			ADRAckLimit:       uint32(band.ADRAckLimit),
			ADRAckDelay:       uint32(band.ADRAckDelay),
			DutyCycle:         ttnpb.DUTY_CYCLE_1,
			RxDelay:           req.GetRxDelay(),
			Rx1DataRateOffset: req.DownlinkSettings.GetRx1DROffset(),
			Rx2DataRateIndex:  req.DownlinkSettings.GetRx2DR(),
			Rx2Frequency:      uint64(band.DefaultRx2Parameters.Frequency),
		}
		dev.EndDevice.MACStateDesired = dev.EndDevice.MACState

		if err = dev.Store("EndDeviceIdentifiers.DevAddr", "Session", "SessionFallback", "RecentUplinks", "MACState", "MACStateDesired"); err != nil {
			logger.WithError(err).Error("Failed to update device")
			return err
		}

		uid := dev.EndDeviceIdentifiers.ApplicationIdentifiers.UniqueID(ctx)
		if uid == "" {
			return common.ErrCorruptRegistry.NewWithCause(nil, ErrMissingApplicationID.New(nil))
		}

		go func() {
			ns.applicationServersMu.RLock()
			cl, ok := ns.applicationServers[uid]
			ns.applicationServersMu.RUnlock()

			if !ok {
				return
			}

			err := cl.Send(&ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				SessionKeyID:         dev.GetSession().GetSessionKeyID(),
				Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
					AppSKey: resp.SessionKeys.GetAppSKey(),
				}},
			})
			if err != nil {
				logger.WithField(
					"application_id", dev.EndDeviceIdentifiers.ApplicationIdentifiers.GetApplicationID(),
				).WithError(err).Errorf("Failed to send Join Accept to AS")
			}
		}()
		return nil
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Warn("Join failed")
	return errors.NewWithCause(errors.New("No Join Server could handle join request"), "Failed to perform join procedure")
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator) error {
	// TODO: Implement (https://github.com/TheThingsIndustries/ttn/issues/557)
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the Gateway Server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, msg *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	msg.ReceivedAt = time.Now()

	logger := log.FromContext(ctx)

	b := msg.GetRawPayload()

	pld := msg.GetPayload()
	if pld.Payload == nil {
		if err := msg.Payload.UnmarshalLoRaWAN(b); err != nil {
			return nil, common.ErrUnmarshalPayloadFailed.NewWithCause(nil, err)
		}
	}

	if pld.GetMajor() != ttnpb.Major_LORAWAN_R1 {
		return nil, common.ErrUnsupportedLoRaWANVersion.New(errors.Attributes{
			"version": pld.GetMajor(),
		})
	}

	acc, ok := ns.deduplicateUplink(ctx, msg)
	if ok {
		logger.Debug("Dropping duplicate uplink")
		return ttnpb.Empty, nil
	}

	switch pld.GetMType() {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return ttnpb.Empty, ns.handleUplink(ctx, msg, acc)
	case ttnpb.MType_JOIN_REQUEST:
		return ttnpb.Empty, ns.handleJoin(ctx, msg, acc)
	case ttnpb.MType_REJOIN_REQUEST:
		return ttnpb.Empty, ns.handleRejoin(ctx, msg, acc)
	default:
		logger.Error("Unmatched MType")
		return ttnpb.Empty, nil
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

// Roles returns the roles that the Network Server fulfills.
func (ns *NetworkServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_NETWORK_SERVER}
}
