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
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// recentUplinkCount is the maximium amount of recent uplinks stored per device.
	recentUplinkCount = 20

	// accumulationCapacity is the initial capacity of the accumulator.
	accumulationCapacity = 20

	// fOptsCapacity is the maximum length of FOpts in bytes.
	fOptsCapacity = 15

	// classCTimeout represents the time interval, within which class C
	// device should acknowledge the downlink or answer MAC command.
	classCTimeout = 5 * time.Minute
)

var (
	// appQueueUpdateTimeout represents the time interval, within which AS
	// shall update the application queue after receiving the uplink.
	appQueueUpdateTimeout = 200 * time.Millisecond
)

func resetMACState(fps *frequencyplans.Store, dev *ttnpb.EndDevice) error {
	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return err
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return err
	}

	dev.MACState = &ttnpb.MACState{
		LoRaWANVersion: dev.EndDeviceVersion.LoRaWANVersion,
		MACParameters: ttnpb.MACParameters{
			ADRAckDelay:       uint32(band.ADRAckDelay),
			ADRAckLimit:       uint32(band.ADRAckLimit),
			ADRNbTrans:        1,
			DutyCycle:         ttnpb.DUTY_CYCLE_1,
			MaxEIRP:           band.DefaultMaxEIRP,
			Rx1DataRateOffset: 0,
			Rx2DataRateIndex:  band.DefaultRx2Parameters.DataRateIndex,
			Rx2Frequency:      uint64(band.DefaultRx2Parameters.Frequency),
			Rx1Delay:          uint32(band.ReceiveDelay1.Seconds()),
			DownlinkDwellTime: band.DwellTime > 0,
			UplinkDwellTime:   band.DwellTime > 0,
		},
	}

	// NOTE: dev.MACState.MACParameters must not contain pointer values at this point.
	dev.MACState.DesiredMACParameters = dev.MACState.MACParameters

	dev.MACState.MACParameters.Channels = make([]*ttnpb.MACParameters_Channel, len(band.UplinkChannels))
	dev.MACState.DesiredMACParameters.Channels = make([]*ttnpb.MACParameters_Channel, int(math.Max(float64(len(dev.MACState.MACParameters.Channels)), float64(len(fp.Channels)))))

	if len(band.DownlinkChannels) > len(band.UplinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		return errInvalidFrequencyPlan
	}

	for i, upCh := range band.UplinkChannels {
		ch := &ttnpb.MACParameters_Channel{
			UplinkFrequency:   upCh.Frequency,
			DownlinkFrequency: upCh.Frequency,
			MinDataRateIndex:  ttnpb.DataRateIndex(upCh.DataRateIndexes[0]),
			MaxDataRateIndex:  ttnpb.DataRateIndex(upCh.DataRateIndexes[len(upCh.DataRateIndexes)-1]),
		}
		if len(band.DownlinkChannels) > i {
			ch.DownlinkFrequency = band.DownlinkChannels[i].Frequency
		}

		dev.MACState.MACParameters.Channels[i] = ch

		chCopy := *ch
		dev.MACState.DesiredMACParameters.Channels[i] = &chCopy
	}

	for i, fpCh := range fp.Channels {
		ch := dev.MACState.DesiredMACParameters.Channels[i]
		if ch == nil {
			ch = &ttnpb.MACParameters_Channel{}
			dev.MACState.DesiredMACParameters.Channels[i] = ch
		}

		ch.UplinkFrequency = uint64(fpCh.Frequency)
		ch.DownlinkFrequency = uint64(fpCh.Frequency)

		if ch.MinDataRateIndex > ttnpb.DataRateIndex(fpCh.DataRate.GetIndex()) || ttnpb.DataRateIndex(fpCh.DataRate.GetIndex()) > ch.MaxDataRateIndex {
			return errInvalidFrequencyPlan
		}
		// TODO: This should fixed once https://github.com/TheThingsIndustries/lorawan-stack/issues/927 is resolved.
		ch.MinDataRateIndex = ttnpb.DataRateIndex(fpCh.DataRate.GetIndex())
		ch.MaxDataRateIndex = ttnpb.DataRateIndex(fpCh.DataRate.GetIndex())
	}

	if fp.UplinkDwellTime != nil {
		dev.MACState.DesiredMACParameters.UplinkDwellTime = true
	}

	if fp.DownlinkDwellTime != nil {
		dev.MACState.DesiredMACParameters.DownlinkDwellTime = true
	}

	if fp.Rx2 != nil {
		dev.MACState.DesiredMACParameters.Rx2Frequency = uint64(fp.Rx2.Frequency)
		dev.MACState.DesiredMACParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(fp.Rx2.DataRate.Index)
	}

	if fp.PingSlot != nil {
		if fp.PingSlot.DataRate != nil {
			dev.MACState.DesiredMACParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(fp.PingSlot.DataRate.Index)
		}
		dev.MACState.DesiredMACParameters.PingSlotFrequency = uint64(fp.PingSlot.Frequency)
	}

	if fp.MaxEIRP > 0 {
		dev.MACState.DesiredMACParameters.MaxEIRP = float32(math.Min(float64(dev.MACState.MaxEIRP), float64(fp.MaxEIRP)))
	}

	if dev.EndDeviceVersion.DefaultMACParameters != nil {
		dev.MACState.MACParameters = deepcopy.Copy(*dev.EndDeviceVersion.DefaultMACParameters).(ttnpb.MACParameters)
	}

	return nil
}

// WindowEndFunc is a function, which is used by Network Server to determine the end of deduplication and cooldown windows.
type WindowEndFunc func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time

// NewWindowEndAfterFunc returns a WindowEndFunc, which closes
// the returned channel after at least duration d after msg.ServerTime or if the context is done.
func NewWindowEndAfterFunc(d time.Duration) WindowEndFunc {
	return func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
		ch := make(chan time.Time, 1)

		if msg.ReceivedAt.IsZero() {
			msg.ReceivedAt = time.Now()
		}

		end := msg.ReceivedAt.Add(d)
		if end.Before(time.Now()) {
			ch <- end
			return ch
		}

		go func() {
			time.Sleep(time.Until(msg.ReceivedAt.Add(d)))
			ch <- end
		}()
		return ch
	}
}

// NsGsClientFunc is the function used to get Gateway Server.
type NsGsClientFunc func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error)

// PeerGetter is the interface, which wraps GetPeer method.
type PeerGetter interface {
	GetPeer(role ttnpb.PeerInfo_Role, tags []string, shardKey []byte) cluster.Peer
}

// NewGatewayServerPeerGetterFunc returns a NsGsClientFunc, which uses g to retrieve Gateway Server clients.
func NewGatewayServerPeerGetterFunc(g PeerGetter) NsGsClientFunc {
	return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
		p := g.GetPeer(
			ttnpb.PeerInfo_GATEWAY_SERVER,
			[]string{fmt.Sprintf("gtw=%s", unique.ID(ctx, id))},
			nil,
		)
		if p == nil {
			return nil, errGatewayServerNotFound
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

	metadataAccumulators *sync.Map // uint64 -> *metadataAccumulator

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	deduplicationDone WindowEndFunc
	collectionDone    WindowEndFunc

	gsClient NsGsClientFunc

	macHandlers *sync.Map // ttnpb.MACCommandIdentifier -> MACHandler
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
// is used to get the Gateway Server for a gateway identifiers.
func WithNsGsClientFunc(fn NsGsClientFunc) Option {
	return func(ns *NetworkServer) {
		ns.gsClient = fn
	}
}

// WithMACHandler registers a MACHandler for specified CID in 0x80-0xFF range.
// WithMACHandler panics if a MACHandler for the CID is already registered, or if
// the CID is out of range.
func WithMACHandler(cid ttnpb.MACCommandIdentifier, fn MACHandler) Option {
	if cid < 0x80 || cid > 0xFF {
		panic(errCIDOutOfRange.WithAttributes(
			"min", fmt.Sprintf("0x%X", 0x80),
			"max", fmt.Sprintf("0x%X", 0xFF),
		))
	}

	return func(ns *NetworkServer) {
		_, ok := ns.macHandlers.LoadOrStore(cid, fn)
		if ok {
			panic(errDuplicateCIDHandler.WithAttributes(
				"cid", fmt.Sprintf("0x%X", int32(cid)),
			))
		}
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
		macHandlers:             &sync.Map{},
	}
	ns.hashPool.New = func() interface{} {
		return fnv.New64a()
	}
	ns.metadataAccumulatorPool.New = func() interface{} {
		return &metadataAccumulator{newAccumulator()}
	}

	for _, opt := range opts {
		opt(ns)
	}

	registryRPC, err := deviceregistry.NewRPC(
		c,
		conf.Registry,
		deviceregistry.ForComponents(ttnpb.PeerInfo_NETWORK_SERVER),
		deviceregistry.WithSetDeviceProcessor(ns.setDeviceProcessor),
	)
	if err != nil {
		return nil, errDeviceRegistryInitialize.WithCause(err)
	}
	ns.RegistryRPC = registryRPC

	switch {
	case ns.deduplicationDone == nil && conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow is zero and WithDeduplicationDoneFunc not specified"))

	case ns.collectionDone == nil && conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow is zero and WithCollectionDoneFunc not specified"))

	case ns.collectionDone == nil && conf.CooldownWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("CooldownWindow is zero and WithCollectionDoneFunc not specified"))
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

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterStreamHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthStreamHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsApplicationDownlinkQueue", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(ns)
	return ns, nil
}

func (ns *NetworkServer) setDeviceProcessor(_ context.Context, create bool, dev *ttnpb.EndDevice, fields ...string) (*ttnpb.EndDevice, []string, error) {
	switch {
	case len(dev.RecentDownlinks) > 0,
		len(dev.RecentUplinks) > 0,
		dev.MACState != nil,
		dev.QueuedApplicationDownlinks != nil:
		return nil, nil, errors.New("trying to override internal field")
	}

	if !create {
		return dev, fields, nil
	}

	if err := resetMACState(ns.Component.FrequencyPlans, dev); err != nil {
		return nil, nil, err
	}

	if !dev.SupportsJoin {
		dev.Session = &ttnpb.Session{
			// TODO: Populate session for ABP.
			// (https://github.com/TheThingsIndustries/lorawan-stack/issues/291)
		}
	}

	if len(fields) == 0 {
		return dev, nil, nil
	}
	return dev, append(fields,
		"MACState",
		"Session",
	), nil
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
func (ns *NetworkServer) LinkApplication(id *ttnpb.ApplicationIdentifiers, stream ttnpb.AsNs_LinkApplicationServer) (err error) {
	ws := &applicationUpStream{
		AsNs_LinkApplicationServer: stream,
		closeCh:                    make(chan struct{}),
	}

	ctx := stream.Context()
	uid := unique.ID(ctx, id)

	events.Publish(evtStartApplicationLink(ctx, id, nil))
	defer events.Publish(evtEndApplicationLink(ctx, id, err))

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
		return errDuplicateSubscription
	}
}

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}

	dev.QueuedApplicationDownlinks = req.Downlinks
	if err = dev.Store(
		"QueuedApplicationDownlinks",
	); err != nil {
		return nil, err
	}

	if dev.MACState != nil && dev.MACState.DeviceClass == ttnpb.CLASS_C {
		// TODO: Schedule the next downlink (https://github.com/TheThingsIndustries/ttn/issues/728)
	}
	return ttnpb.Empty, nil
}

// DownlinkQueuePush is called by the Application Server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}

	dev.QueuedApplicationDownlinks = append(dev.QueuedApplicationDownlinks, req.Downlinks...)
	if err := dev.Store(
		"QueuedApplicationDownlinks",
	); err != nil {
		return nil, err
	}

	if dev.MACState != nil && dev.MACState.DeviceClass == ttnpb.CLASS_C {
		// TODO: Schedule the next downlink (https://github.com/TheThingsIndustries/ttn/issues/728)
	}
	return ttnpb.Empty, nil
}

// DownlinkQueueList is called by the Application Server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, devID)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{Downlinks: dev.QueuedApplicationDownlinks}, nil
}

// DownlinkQueueClear is called by the Application Server to clear the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueClear(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	dev, err := deviceregistry.FindByIdentifiers(ns.registry, devID)
	if err != nil {
		return nil, err
	}
	dev.QueuedApplicationDownlinks = nil
	return ttnpb.Empty, dev.Store(
		"QueuedApplicationDownlinks",
	)
}

// StartServingGateway is called by the Gateway Server to indicate that it is serving a gateway.
func (ns *NetworkServer) StartServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, gtwID)
	if uid == "" {
		return nil, errMissingGatewayID
	}

	gsID := rpcmetadata.FromIncomingContext(ctx).ID
	// TODO: Associate the GS ID with the gateway uid in the cluster once
	// https://github.com/TheThingsIndustries/ttn/issues/506#issuecomment-385963158 is resolved
	_ = gsID
	return ttnpb.Empty, nil
}

// StopServingGateway is called by the Gateway Server to indicate that it is no longer serving a gateway.
func (ns *NetworkServer) StopServingGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, id)
	if uid == "" {
		return nil, errMissingGatewayID
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

func (ns *NetworkServer) deduplicateUplink(ctx context.Context, msg *ttnpb.UplinkMessage) (*metadataAccumulator, func(), bool) {
	h := ns.hashPool.Get().(hash.Hash64)
	_, _ = h.Write(msg.RawPayload)

	k := h.Sum64()

	h.Reset()
	ns.hashPool.Put(h)

	a := ns.metadataAccumulatorPool.Get().(*metadataAccumulator)
	lv, isDup := ns.metadataAccumulators.LoadOrStore(k, a)
	lv.(*metadataAccumulator).Add(msg.RxMetadata...)

	if isDup {
		ns.metadataAccumulatorPool.Put(a)
		return nil, nil, true
	}
	return a, func() {
		ns.metadataAccumulators.Delete(k)
	}, false
}

func setDownlinkModulation(s *ttnpb.TxSettings, dr band.DataRate) (err error) {
	if dr.Rate.LoRa != "" && dr.Rate.FSK > 0 {
		return errLoRaAndFSK
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
	if dev.MACState == nil {
		return errUnknownMACState
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", dev.EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID,
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
	))

	msg := &ttnpb.DownlinkMessage{
		RawPayload:           b,
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
	}

	type tx struct {
		ttnpb.TxSettings
		Delay time.Duration
	}
	slots := make([]tx, 0, 2)

	fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return errUnknownFrequencyPlan.WithCause(err)
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return errUnknownBand.WithCause(err)
	}

	var mds []*ttnpb.RxMetadata
	if up == nil {
		// Class C
		if len(dev.RecentUplinks) == 0 {
			return errUplinkNotFound
		}
		mds = dev.RecentUplinks[len(dev.RecentUplinks)-1].RxMetadata
	} else {
		drIdx, err := band.Rx1DataRate(up.Settings.DataRateIndex, dev.MACState.Rx1DataRateOffset, dev.MACState.DownlinkDwellTime)
		if err != nil {
			return err
		}

		chIdx, err := band.Rx1Channel(up.Settings.ChannelIndex)
		if err != nil {
			return err
		}
		if uint(chIdx) >= uint(len(dev.MACState.Channels)) {
			return errChannelIndexTooHigh
		}

		ch := dev.MACState.Channels[int(chIdx)]
		if ch == nil || ch.DownlinkFrequency == 0 {
			return errUnknownChannel
		}

		rx1 := tx{
			TxSettings: ttnpb.TxSettings{
				DataRateIndex:         drIdx,
				CodingRate:            "4/5",
				PolarizationInversion: true,
				ChannelIndex:          chIdx,
				Frequency:             ch.DownlinkFrequency,
				TxPower:               int32(band.DefaultMaxEIRP),
			},
		}
		if isJoinAccept {
			rx1.Delay = band.JoinAcceptDelay1
		} else {
			rx1.Delay = time.Second * time.Duration(dev.MACState.Rx1Delay)
		}

		if err = setDownlinkModulation(&rx1.TxSettings, band.DataRates[drIdx]); err != nil {
			return err
		}

		mds = up.RxMetadata
		slots = append(slots, rx1)
	}

	if uint(dev.MACState.Rx2DataRateIndex) > uint(len(band.DataRates)) {
		return errInvalidRx2DataRateIndex
	}

	rx2 := tx{
		TxSettings: ttnpb.TxSettings{
			DataRateIndex:         dev.MACState.Rx2DataRateIndex,
			CodingRate:            "4/5",
			PolarizationInversion: true,
			Frequency:             dev.MACState.Rx2Frequency,
			TxPower:               int32(band.DefaultMaxEIRP),
		},
	}
	if isJoinAccept {
		rx2.Delay = band.JoinAcceptDelay2
	} else {
		rx2.Delay = time.Second * time.Duration(1+dev.MACState.Rx1Delay)
	}

	if err = setDownlinkModulation(&rx2.TxSettings, band.DataRates[dev.MACState.Rx2DataRateIndex]); err != nil {
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
		return mds[i].SNR > mds[j].SNR
	})

	var errs []error
	for _, s := range slots {
		msg.Settings = s.TxSettings

		for _, md := range mds {
			logger := logger.WithField(
				"gateway_id", md.GatewayIdentifiers.GatewayID,
			)

			cl, err := ns.gsClient(ctx, md.GatewayIdentifiers)
			if err != nil {
				logger.WithError(err).Debug("Could not get Gateway Server")
				continue
			}

			msg.TxMetadata = ttnpb.TxMetadata{
				GatewayIdentifiers: md.GatewayIdentifiers,
				Timestamp:          md.Timestamp + uint64(s.Delay.Nanoseconds()),
			}

			_, err = cl.ScheduleDownlink(ctx, msg, ns.WithClusterAuth())
			if err != nil {
				errs = append(errs, err)
				continue
			}

			dev.RecentDownlinks = append(dev.RecentDownlinks, msg)
			if err = dev.Store(
				"MACState",
				"QueuedApplicationDownlinks",
				"RecentDownlinks",
				"Session",
			); err != nil {
				return errDeviceStoreFailed.WithCause(err)
			}
			return nil
		}
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Warn("all Gateway Servers failed to schedule the downlink")
	return errScheduleFailed
}

// generateAndScheduleDownlink loads dev, tries to generate a downlink and schedule it.
// If no downlink could be generated, nothing is scheduled and nil is returned.
// If downlink was generated, but could not be scheduled, an error is returned describing why.
func (ns *NetworkServer) generateAndScheduleDownlink(ctx context.Context, dev *deviceregistry.Device, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", dev.EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID,
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
	))

	if dev.MACState == nil {
		return errUnknownMACState
	}

	if dev.Session == nil {
		logger.Debug("No active session found for device")
		return errEmptySession
	}

	dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]

	// TODO: Diff MACParameters and DesiredMACParameters and add commands to dev.MACState.PendingRequests.
	// (https://github.com/TheThingsIndustries/ttn/issues/837)

	if len(dev.MACState.PendingRequests) == 0 &&
		len(dev.MACState.QueuedResponses) == 0 &&
		len(dev.QueuedApplicationDownlinks) == 0 &&
		(up == nil || up.Payload.MType != ttnpb.MType_CONFIRMED_UP) {
		logger.Debug("Nothing to schedule")
		return nil
	}

	sort.SliceStable(dev.MACState.PendingRequests, func(i, j int) bool {
		// NOTE: The ordering of a sequence of commands with identical CIDs shall not be changed.

		ci := dev.MACState.PendingRequests[i].CID
		cj := dev.MACState.PendingRequests[j].CID
		switch {
		case ci >= 0x80: // Proprietary
			return false
		case cj >= 0x80:
			return true

		case ci > 0x0F: // >1.1
			return false
		case cj > 0x0F:
			return true

		case ci < 0x02, ci > 0x0A: // >1.0.2
			return false
		case cj < 0x02, cj > 0x0A:
			return true

		case ci > 0x08: // >1.0.1
			return false
		case cj > 0x08:
			return true
		}
		return false
	})

	cmds := append(dev.MACState.QueuedResponses, dev.MACState.PendingRequests...)

	// TODO: Ensure cmds can be answered in one frame
	// (https://github.com/TheThingsIndustries/ttn/issues/836)

	cmdBuf := make([]byte, 0, len(cmds))
	for _, cmd := range cmds {
		cmdBuf, err = cmd.AppendLoRaWAN(cmdBuf)
		if err != nil {
			return errMACEncodeFailed.WithCause(err)
		}
	}

	pld := &ttnpb.MACPayload{
		FHDR: ttnpb.FHDR{
			DevAddr: *dev.EndDeviceIdentifiers.DevAddr,
			FCtrl: ttnpb.FCtrl{
				Ack: up != nil && up.Payload.MType == ttnpb.MType_CONFIRMED_UP,
			},
			FCnt: dev.Session.NextNFCntDown,
		},
	}

	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if len(cmdBuf) <= fOptsCapacity && len(dev.QueuedApplicationDownlinks) > 0 {
		var down *ttnpb.ApplicationDownlink
		down, dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[0], dev.QueuedApplicationDownlinks[1:]

		pld.FHDR.FCnt = down.FCnt
		pld.FPort = down.FPort
		pld.FRMPayload = down.FRMPayload
		if down.Confirmed {
			dev.MACState.PendingApplicationDownlink = down
			dev.Session.LastConfFCntDown = pld.FCnt

			mType = ttnpb.MType_CONFIRMED_DOWN
		}
	}

	if len(cmdBuf) > 0 && (pld.FPort == 0 || dev.EndDevice.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || dev.Session.NwkSEncKey.Key.IsZero() {
			return errMissingNwkSEncKey
		}

		cmdBuf, err = crypto.EncryptDownlink(*dev.Session.NwkSEncKey.Key, *dev.EndDeviceIdentifiers.DevAddr, pld.FHDR.FCnt, cmdBuf)
		if err != nil {
			return errMACEncryptFailed.WithCause(err)
		}
	}

	if pld.FPort == 0 {
		pld.FRMPayload = cmdBuf
		dev.Session.NextNFCntDown++
	} else {
		pld.FHDR.FOpts = cmdBuf
	}

	// TODO: Set to true if commands were trimmed.
	// (https://github.com/TheThingsIndustries/ttn/issues/836)
	pld.FHDR.FCtrl.FPending = len(dev.QueuedApplicationDownlinks) > 0

	switch {
	case dev.MACState.DeviceClass != ttnpb.CLASS_C,
		mType != ttnpb.MType_CONFIRMED_DOWN && len(dev.MACState.PendingRequests) == 0:
		break

	case dev.MACState.NextConfirmedDownlinkAt.After(time.Now()):
		return errScheduleTooSoon

	default:
		t := time.Now().Add(classCTimeout).UTC()
		dev.MACState.NextConfirmedDownlinkAt = &t
	}

	b, err := (ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: pld,
		},
	}).MarshalLoRaWAN()
	if err != nil {
		return errMarshalPayloadFailed.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	var key types.AES128Key
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		if dev.Session.FNwkSIntKey == nil || dev.Session.FNwkSIntKey.Key.IsZero() {
			return errMissingFNwkSIntKey
		}
		key = *dev.Session.FNwkSIntKey.Key

	} else {
		if dev.Session.SNwkSIntKey == nil || dev.Session.SNwkSIntKey.Key.IsZero() {
			return errMissingSNwkSIntKey
		}
		key = *dev.Session.SNwkSIntKey.Key
	}

	mic, err := crypto.ComputeDownlinkMIC(
		key,
		*dev.EndDeviceIdentifiers.DevAddr,
		pld.FCnt,
		b,
	)
	if err != nil {
		return errComputeMIC
	}
	return ns.scheduleDownlink(ctx, dev, up, acc, append(b, mic[:]...), false)
}

// matchDevice tries to match the uplink message with a device.
// If the uplink message matches a fallback session, that fallback session is recovered.
// If successful, the FCnt in the uplink message is set to the full FCnt. The NextFCntUp in the session is updated accordingly.
func (ns *NetworkServer) matchDevice(ctx context.Context, msg *ttnpb.UplinkMessage) (*deviceregistry.Device, error) {
	pld := msg.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithField("dev_addr", pld.DevAddr)

	var devs []*deviceregistry.Device

	_, err := ns.registry.Range(
		&ttnpb.EndDevice{
			Session: &ttnpb.Session{
				DevAddr: pld.DevAddr,
			},
		},
		"", 0, 0,
		func(d *deviceregistry.Device) bool {
			if d.MACState == nil {
				return true
			}
			devs = append(devs, d)
			return true
		},
		"Session.DevAddr",
	)
	if err != nil {
		logger.WithError(err).Warn("Failed to search for device in registry by active DevAddr")
		return nil, err
	}

	_, err = ns.registry.Range(
		&ttnpb.EndDevice{
			SessionFallback: &ttnpb.Session{
				DevAddr: pld.DevAddr,
			},
		},
		"", 0, 0,
		func(d *deviceregistry.Device) bool {
			d.EndDeviceIdentifiers.DevAddr = &d.SessionFallback.DevAddr
			d.Session = d.SessionFallback
			d.SessionFallback = nil
			if d.MACState == nil {
				return true
			}
			devs = append(devs, d)
			return true
		},
		"SessionFallback.DevAddr",
	)
	if err != nil {
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
		fCnt := pld.FCnt

		switch {
		case !dev.EndDeviceVersion.Supports32BitFCnt, fCnt >= dev.Session.NextFCntUp:
		case fCnt > dev.Session.NextFCntUp&0xffff:
			fCnt |= dev.Session.NextFCntUp &^ 0xffff
		case dev.Session.NextFCntUp < 0xffff<<16:
			fCnt |= (dev.Session.NextFCntUp + 1<<16) &^ 0xffff
		}

		gap := uint32(math.MaxUint32)
		if !dev.EndDeviceVersion.FCntResets {
			if dev.Session.NextFCntUp > fCnt {
				continue outer
			}

			gap = fCnt - dev.Session.NextFCntUp

			if dev.MACState.LoRaWANVersion.HasMaxFCntGap() {
				fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
				if err != nil {
					return nil, errUnknownFrequencyPlan.WithCause(err)
				}

				band, err := band.GetByID(fp.BandID)
				if err != nil {
					return nil, errUnknownBand.WithCause(err)
				}

				if gap > uint32(band.MaxFCntGap) {
					continue outer
				}
			}
		}

		matching = append(matching, device{
			Device: dev,
			gap:    gap,
			fCnt:   fCnt,
		})
		if dev.EndDeviceVersion.FCntResets && fCnt != pld.FCnt {
			matching = append(matching, device{
				Device: dev,
				gap:    gap,
				fCnt:   pld.FCnt,
			})
		}
	}

	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	if len(msg.RawPayload) < 4 {
		return nil, errRawPayloadTooLong
	}
	b := msg.RawPayload[:len(msg.RawPayload)-4]

	for _, dev := range matching {
		if pld.Ack {
			if len(dev.RecentDownlinks) == 0 {
				// Uplink acknowledges a downlink, but no downlink was sent to the device,
				// hence it must be the wrong device.
				continue
			}
		}

		if dev.Session.FNwkSIntKey == nil || dev.Session.FNwkSIntKey.Key.IsZero() {
			return nil, errMissingFNwkSIntKey
		}

		var computedMIC [4]byte
		var err error
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(
				*dev.Session.FNwkSIntKey.Key,
				pld.DevAddr,
				dev.fCnt,
				b,
			)

		} else {
			if dev.Session.SNwkSIntKey == nil || dev.Session.SNwkSIntKey.Key.IsZero() {
				return nil, errMissingSNwkSIntKey
			}

			var confFCnt uint32
			if pld.Ack {
				confFCnt = dev.Session.LastConfFCntDown
			}

			computedMIC, err = crypto.ComputeUplinkMIC(
				*dev.Session.SNwkSIntKey.Key,
				*dev.Session.FNwkSIntKey.Key,
				confFCnt,
				uint8(msg.Settings.DataRateIndex),
				uint8(msg.Settings.ChannelIndex),
				pld.DevAddr,
				dev.fCnt,
				b,
			)
		}
		if err != nil {
			return nil, errComputeMIC.WithCause(err)
		}
		if !bytes.Equal(msg.Payload.MIC, computedMIC[:]) {
			continue
		}

		if dev.fCnt == math.MaxUint32 {
			return nil, errFCntTooHigh
		}
		dev.Session.NextFCntUp = dev.fCnt + 1
		return dev.Device, nil
	}
	return nil, errDeviceNotFound
}

type MACHandler func(ctx context.Context, dev *ttnpb.EndDevice, pld []byte, msg *ttnpb.UplinkMessage) error

func (ns *NetworkServer) handleUplink(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, msg, err)
		}
	}()

	dev, err := ns.matchDevice(ctx, msg)
	if err != nil {
		return errDeviceNotFound.WithCause(err)
	}
	if dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		// LoRaWAN1.1+ device will send a RekeyInd.
		dev.SessionFallback = nil
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", dev.EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID,
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
	))

	updateTimeout := appQueueUpdateTimeout

	uid := unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers)
	if uid == "" {
		return errMissingApplicationID
	}

	ns.applicationServersMu.RLock()
	asCl, asOk := ns.applicationServers[uid]
	ns.applicationServersMu.RUnlock()

	if !asOk {
		updateTimeout = 0
	}

	defer time.AfterFunc(updateTimeout, func() {
		// TODO: Decouple Class A downlink from uplink.
		// https://github.com/TheThingsIndustries/lorawan-stack/issues/905

		dev, err := dev.Load()
		if err != nil {
			logger.WithError(err).Error("Failed to load device")
			return
		}

		if err := ns.generateAndScheduleDownlink(ctx, dev, msg, acc); err != nil {
			logger.WithError(err).Error("Failed to schedule downlink in reception slot")
		}
	})

	pld := msg.Payload.GetMACPayload()
	if pld == nil {
		return errMissingPayload
	}

	if dev.MACState != nil && dev.MACState.PendingApplicationDownlink != nil {
		asUp := &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			CorrelationIDs:       dev.MACState.PendingApplicationDownlink.CorrelationIDs,
		}

		if pld.Ack {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: dev.MACState.PendingApplicationDownlink,
			}
			asUp.CorrelationIDs = append(asUp.CorrelationIDs, msg.CorrelationIDs...)
		} else {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkNack{
				DownlinkNack: dev.MACState.PendingApplicationDownlink,
			}
		}

		if err := asCl.Send(asUp); err != nil {
			return err
		}
		dev.MACState.PendingApplicationDownlink = nil
	}

	mac := pld.FOpts
	if len(mac) == 0 && pld.FPort == 0 {
		mac = pld.FRMPayload
	}

	if len(mac) > 0 && (len(pld.FOpts) == 0 || dev.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || dev.Session.NwkSEncKey.Key.IsZero() {
			return errMissingNwkSEncKey
		}

		mac, err = crypto.DecryptUplink(*dev.Session.NwkSEncKey.Key, *dev.EndDeviceIdentifiers.DevAddr, pld.FCnt, mac)
		if err != nil {
			return errDecryptionFailed.WithCause(err)
		}
	}

	var cmds []*ttnpb.MACCommand
	for r := bytes.NewReader(mac); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := ttnpb.ReadMACCommand(r, true, cmd); err != nil {
			logger.
				WithField("unmarshaled", len(cmds)).
				WithError(err).
				Warn("Failed to unmarshal MAC commands")
			break
		}
		cmds = append(cmds, cmd)
	}

	if dev.MACState == nil {
		if err := resetMACState(ns.Component.FrequencyPlans, dev.EndDevice); err != nil {
			return err
		}
	}
	dev.MACState.ADRDataRateIndex = msg.Settings.DataRateIndex

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, msg):
	}

	msg.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, dev.EndDevice, msg)

	dev.MACState.QueuedResponses = dev.MACState.QueuedResponses[:0]

outer:
	for _, cmd := range cmds {
		cid := cmd.CID
		switch cid {
		case ttnpb.CID_RESET:
			err = handleResetInd(ctx, dev.EndDevice, cmd.GetResetInd(), ns.Component.FrequencyPlans)
		case ttnpb.CID_LINK_CHECK:
			err = handleLinkCheckReq(ctx, dev.EndDevice, msg)
		case ttnpb.CID_LINK_ADR:
			err = handleLinkADRAns(ctx, dev.EndDevice, cmd.GetLinkADRAns())
		case ttnpb.CID_DUTY_CYCLE:
			err = handleDutyCycleAns(ctx, dev.EndDevice)
		case ttnpb.CID_RX_PARAM_SETUP:
			err = handleRxParamSetupAns(ctx, dev.EndDevice, cmd.GetRxParamSetupAns())
		case ttnpb.CID_DEV_STATUS:
			err = handleDevStatusAns(ctx, dev.EndDevice, cmd.GetDevStatusAns())
		case ttnpb.CID_NEW_CHANNEL:
			err = handleNewChannelAns(ctx, dev.EndDevice, cmd.GetNewChannelAns())
		case ttnpb.CID_RX_TIMING_SETUP:
			err = handleRxTimingSetupAns(ctx, dev.EndDevice)
		case ttnpb.CID_TX_PARAM_SETUP:
			err = handleTxParamSetupAns(ctx, dev.EndDevice)
		case ttnpb.CID_DL_CHANNEL:
			err = handleDLChannelAns(ctx, dev.EndDevice, cmd.GetDlChannelAns())
		case ttnpb.CID_REKEY:
			err = handleRekeyInd(ctx, dev.EndDevice, cmd.GetRekeyInd())
		case ttnpb.CID_ADR_PARAM_SETUP:
			err = handleADRParamSetupAns(ctx, dev.EndDevice)
		case ttnpb.CID_DEVICE_TIME:
			err = handleDeviceTimeReq(ctx, dev.EndDevice, msg)
		case ttnpb.CID_REJOIN_PARAM_SETUP:
			err = handleRejoinParamSetupAns(ctx, dev.EndDevice, cmd.GetRejoinParamSetupAns())
		case ttnpb.CID_PING_SLOT_INFO:
			err = handlePingSlotInfoReq(ctx, dev.EndDevice, cmd.GetPingSlotInfoReq())
		case ttnpb.CID_PING_SLOT_CHANNEL:
			err = handlePingSlotChannelAns(ctx, dev.EndDevice, cmd.GetPingSlotChannelAns())
		case ttnpb.CID_BEACON_TIMING:
			err = handleBeaconTimingReq(ctx, dev.EndDevice)
		case ttnpb.CID_BEACON_FREQ:
			err = handleBeaconFreqAns(ctx, dev.EndDevice, cmd.GetBeaconFreqAns())
		case ttnpb.CID_DEVICE_MODE:
			err = handleDeviceModeInd(ctx, dev.EndDevice, cmd.GetDeviceModeInd())
		default:
			h, ok := ns.macHandlers.Load(cid)
			if !ok {
				logger.WithField("cid", cid).Warn("Unknown MAC command received, skipping the rest...")
				break outer
			}
			err = h.(MACHandler)(ctx, dev.EndDevice, cmd.GetRawPayload(), msg)
		}
		if err != nil {
			logger.WithField("cid", cid).WithError(err).Warn("Failed to process MAC command")
		}
	}

	if len(dev.RecentUplinks) >= recentUplinkCount {
		dev.RecentUplinks = dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount+1:]
	}
	dev.RecentUplinks = append(dev.RecentUplinks, msg)

	if err := dev.Store(
		"MACState",
		"RecentUplinks",
		"Session",
		"SessionFallback",
	); err != nil {
		logger.WithError(err).Error("Failed to store device")
		return err
	}

	if !asOk {
		return nil
	}

	registerForwardUplink(ctx, dev.EndDevice, msg)
	return asCl.Send(&ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		CorrelationIDs:       msg.CorrelationIDs,
		Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
			FCnt:         dev.Session.NextFCntUp - 1,
			FPort:        pld.FPort,
			FRMPayload:   pld.FRMPayload,
			RxMetadata:   msg.RxMetadata,
			SessionKeyID: dev.Session.SessionKeyID,
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
		panic(errors.New("failed to create new DevAddr").WithCause(err))
	}
	return devAddr
}

func (ns *NetworkServer) handleJoin(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, msg, err)
		}
	}()

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
	for dev.Session != nil && devAddr.Equal(dev.Session.DevAddr) {
		devAddr = ns.newDevAddr(dev.EndDevice)
	}

	fp, err := ns.FrequencyPlans.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return errUnknownFrequencyPlan.WithCause(err)
	}

	req := &ttnpb.JoinRequest{
		RawPayload: msg.RawPayload,
		Payload:    msg.Payload,
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DevEUI:  &pld.DevEUI,
			JoinEUI: &pld.JoinEUI,
			DevAddr: &devAddr,
		},
		NetID:              ns.NetID,
		SelectedMacVersion: dev.LoRaWANVersion,
		RxDelay:            dev.MACState.DesiredMACParameters.Rx1Delay,
		CFList:             frequencyplans.CFList(fp, dev.LoRaWANPHYVersion),
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: dev.MACState.DesiredMACParameters.Rx1DataRateOffset,
			Rx2DR:       dev.MACState.DesiredMACParameters.Rx2DataRateIndex,
			OptNeg:      true,
		},
	}

	var errs []error
	for _, js := range ns.joinServers {
		registerForwardUplink(ctx, dev.EndDevice, msg)
		resp, err := js.HandleJoin(ctx, req, ns.WithClusterAuth())
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if dev.SessionFallback == nil {
			dev.SessionFallback = dev.Session
		}
		dev.Session = &ttnpb.Session{
			DevAddr:     devAddr,
			SessionKeys: resp.SessionKeys,
			StartedAt:   time.Now(),
		}
		dev.EndDeviceIdentifiers.DevAddr = &devAddr

		if err := resetMACState(ns.Component.FrequencyPlans, dev.EndDevice); err != nil {
			return err
		}
		dev.MACState.Rx1Delay = req.RxDelay
		dev.MACState.Rx1DataRateOffset = req.DownlinkSettings.Rx1DROffset
		dev.MACState.Rx2DataRateIndex = req.DownlinkSettings.Rx2DR
		if req.DownlinkSettings.OptNeg && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) > 0 {
			// The version will be further negotiated via RekeyInd/RekeyConf
			dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
		}

		dev.MACState.DesiredMACParameters.Rx1Delay = dev.MACState.Rx1Delay
		dev.MACState.DesiredMACParameters.Rx1DataRateOffset = dev.MACState.Rx1DataRateOffset
		dev.MACState.DesiredMACParameters.Rx2DataRateIndex = dev.MACState.Rx2DataRateIndex

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ns.deduplicationDone(ctx, msg):
		}

		msg.RxMetadata = acc.Accumulated()
		registerMergeMetadata(ctx, dev.EndDevice, msg)

		dev.RecentUplinks = append(dev.RecentUplinks, msg)
		if len(dev.RecentUplinks) > recentUplinkCount {
			dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
		}

		if err := ns.scheduleDownlink(ctx, dev, msg, nil, resp.RawPayload, true); err != nil {
			logger.WithError(err).Debug("Failed to schedule join-accept")
			return err
		}

		if err = dev.Store(
			"EndDeviceIdentifiers.DevAddr",
			"MACState",
			"RecentUplinks",
			"Session",
			"SessionFallback",
		); err != nil {
			logger.WithError(err).Error("Failed to update device")
			return err
		}

		uid := unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers)
		if uid == "" {
			return errMissingApplicationID
		}

		go func() {
			ns.applicationServersMu.RLock()
			cl, ok := ns.applicationServers[uid]
			ns.applicationServersMu.RUnlock()

			if !ok {
				return
			}

			if err := cl.Send(&ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       msg.CorrelationIDs,
				Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
					AppSKey:      resp.SessionKeys.AppSKey,
					SessionKeyID: dev.Session.SessionKeyID,
				}},
			}); err != nil {
				logger.WithField(
					"application_id", dev.EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID,
				).WithError(err).Errorf("Failed to send Join-accept to AS")
			}
		}()
		return nil
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Warn("Join failed")
	return errJoinFailed
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, msg *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, msg, err)
		}
	}()
	// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the Gateway Server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, msg *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		msg.CorrelationIDs,
		fmt.Sprintf("ns:uplink:%s", events.NewCorrelationID()),
	)...)
	msg.CorrelationIDs = events.CorrelationIDsFromContext(ctx)

	msg.ReceivedAt = time.Now()

	logger := log.FromContext(ctx)

	if msg.Payload.Payload == nil {
		if err := msg.Payload.UnmarshalLoRaWAN(msg.RawPayload); err != nil {
			return nil, errUnmarshalPayloadFailed.WithCause(err)
		}
	}

	if msg.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"major", msg.Payload.Major,
		)
	}

	acc, stopDedup, ok := ns.deduplicateUplink(ctx, msg)
	if ok {
		registerReceiveUplinkDuplicate(ctx, msg)
		return ttnpb.Empty, nil
	}
	registerReceiveUplink(ctx, msg)

	defer func(msg *ttnpb.UplinkMessage) {
		<-ns.collectionDone(ctx, msg)
		stopDedup()
	}(msg)

	msg = deepcopy.Copy(msg).(*ttnpb.UplinkMessage)
	switch msg.Payload.MType {
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
