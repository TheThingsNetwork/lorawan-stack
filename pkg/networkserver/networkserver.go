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
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

const (
	// recentDownlinkCount is the maximum amount of recent downlinks stored per device.
	recentDownlinkCount = 20

	// fOptsCapacity is the maximum length of FOpts in bytes.
	fOptsCapacity = 15

	// classCTimeout represents the time interval, within which class C
	// device should acknowledge the downlink or answer MAC command.
	classCTimeout = 5 * time.Minute
)

func timePtr(t time.Time) *time.Time {
	return &t
}

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
		LoRaWANVersion: dev.LoRaWANVersion,
		CurrentParameters: ttnpb.MACParameters{
			ADRAckDelay:       uint32(band.ADRAckDelay),
			ADRAckLimit:       uint32(band.ADRAckLimit),
			ADRNbTrans:        1,
			MaxDutyCycle:      ttnpb.DUTY_CYCLE_1,
			MaxEIRP:           band.DefaultMaxEIRP,
			Rx1Delay:          ttnpb.RxDelay(band.ReceiveDelay1.Seconds()),
			Rx1DataRateOffset: 0,
			Rx2DataRateIndex:  band.DefaultRx2Parameters.DataRateIndex,
			Rx2Frequency:      band.DefaultRx2Parameters.Frequency,
		},
	}

	// NOTE: dev.MACState.CurrentParameters must not contain pointer values at this point.
	dev.MACState.DesiredParameters = dev.MACState.CurrentParameters

	if len(band.DownlinkChannels) > len(band.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		return errInvalidFrequencyPlan
	}

	dev.MACState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, len(band.UplinkChannels))
	dev.MACState.DesiredParameters.Channels = make([]*ttnpb.MACParameters_Channel, int(math.Max(float64(len(dev.MACState.CurrentParameters.Channels)), float64(len(fp.UplinkChannels)))))

	for i, upCh := range band.UplinkChannels {
		if len(upCh.DataRateIndexes) == 0 {
			return errInvalidFrequencyPlan
		}

		ch := &ttnpb.MACParameters_Channel{
			UplinkFrequency:  upCh.Frequency,
			MinDataRateIndex: ttnpb.DataRateIndex(upCh.DataRateIndexes[0]),
			MaxDataRateIndex: ttnpb.DataRateIndex(upCh.DataRateIndexes[len(upCh.DataRateIndexes)-1]),
		}
		dev.MACState.CurrentParameters.Channels[i] = ch

		chCopy := *ch
		dev.MACState.DesiredParameters.Channels[i] = &chCopy
	}

	for i, downCh := range band.DownlinkChannels {
		if i >= len(dev.MACState.CurrentParameters.Channels) {
			return errInvalidFrequencyPlan
		}
		dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency = downCh.Frequency
	}

	for i, upCh := range fp.UplinkChannels {
		ch := dev.MACState.DesiredParameters.Channels[i]
		if ch == nil {
			dev.MACState.DesiredParameters.Channels[i] = &ttnpb.MACParameters_Channel{
				MinDataRateIndex: ttnpb.DataRateIndex(upCh.MinDataRate),
				MaxDataRateIndex: ttnpb.DataRateIndex(upCh.MaxDataRate),
				UplinkFrequency:  upCh.Frequency,
			}
			continue
		}

		if ch.MinDataRateIndex > ttnpb.DataRateIndex(upCh.MinDataRate) || ttnpb.DataRateIndex(upCh.MaxDataRate) > ch.MaxDataRateIndex {
			return errInvalidFrequencyPlan
		}
		ch.MinDataRateIndex = ttnpb.DataRateIndex(upCh.MinDataRate)
		ch.MaxDataRateIndex = ttnpb.DataRateIndex(upCh.MaxDataRate)
		ch.UplinkFrequency = upCh.Frequency
	}

	for i, downCh := range fp.DownlinkChannels {
		if i >= len(dev.MACState.DesiredParameters.Channels) {
			return errInvalidFrequencyPlan
		}
		dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency = downCh.Frequency
	}

	dev.MACState.DesiredParameters.UplinkDwellTime = fp.DwellTime.GetUplinks()
	dev.MACState.DesiredParameters.DownlinkDwellTime = fp.DwellTime.GetDownlinks()

	if fp.Rx2 != nil {
		dev.MACState.DesiredParameters.Rx2Frequency = fp.Rx2.Frequency
	}
	if fp.DefaultRx2DataRate != nil {
		dev.MACState.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	}

	if fp.PingSlot != nil {
		dev.MACState.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	}
	if fp.DefaultPingSlotDataRate != nil {
		dev.MACState.DesiredParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)
	}

	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 {
		dev.MACState.DesiredParameters.MaxEIRP = float32(math.Min(float64(dev.MACState.CurrentParameters.MaxEIRP), float64(*fp.MaxEIRP)))
	}

	if dev.DefaultMACParameters != nil {
		dev.MACState.CurrentParameters = deepcopy.Copy(*dev.DefaultMACParameters).(ttnpb.MACParameters)
	}

	return nil
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

var errNoDownlink = errors.Define("no_downlink", "no downlink to send")

// generateDownlink attempts to generate a downlink.
// generateDownlink returns the marshaled payload of the downlink and error if any.
// If no downlink could be generated - nil, errNoDownlink is returned.
// generateDownlink does not perform validation of dev.MACState.DesiredParameters,
// hence, it could potentially generate MAC command(s), which are not suported by the
// regional parameters the device operates in.
// For example, a sequence of 'NewChannel' MAC commands could be generated for a
// device operating in a region where a fixed channel plan is defined in case
// dev.MACState.CurrentParameters.Channels is not equal to dev.MACState.DesiredParameters.Channels.
func generateDownlink(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) ([]byte, error) {
	if dev.MACState == nil {
		return nil, errUnknownMACState
	}

	if dev.Session == nil {
		return nil, errEmptySession
	}

	spec := lorawan.DefaultMACCommands

	cmds := make([]*ttnpb.MACCommand, 0, len(dev.MACState.QueuedResponses)+len(dev.MACState.PendingRequests))
	for _, cmd := range dev.MACState.QueuedResponses {
		desc := spec[cmd.CID]
		if desc == nil {
			maxDownLen = 0
			continue
		}
		if desc.DownlinkLength > maxDownLen {
			continue
		}
		cmds = append(cmds, cmd)
		maxDownLen -= desc.DownlinkLength
	}

	dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]

	var fPending bool
	for _, f := range []func(context.Context, *ttnpb.EndDevice, uint16, uint16) (uint16, uint16, bool){
		// LoRaWAN 1.0+
		enqueueNewChannelReq,
		enqueueLinkADRReq,
		enqueueDutyCycleReq,
		enqueueRxParamSetupReq,
		enqueueDevStatusReq,
		enqueueRxTimingSetupReq,
		enqueuePingSlotChannelReq,
		enqueueBeaconFreqReq,

		// LoRaWAN 1.0.2+
		enqueueTxParamSetupReq,
		enqueueDLChannelReq,

		// LoRaWAN 1.1+
		enqueueADRParamSetupReq,
		enqueueForceRejoinReq,
		enqueueRejoinParamSetupReq,
	} {
		var ok bool
		maxDownLen, maxUpLen, ok = f(ctx, dev, maxDownLen, maxUpLen)
		fPending = fPending || !ok
	}
	cmds = append(cmds, dev.MACState.PendingRequests...)

	cmdBuf := make([]byte, 0, maxDownLen)
	for _, cmd := range cmds {
		var err error
		cmdBuf, err = spec.AppendDownlink(cmdBuf, *cmd)

		if err != nil {
			return nil, errEncodeMAC.WithCause(err)
		}
	}

	var up *ttnpb.UplinkMessage
	for i := len(dev.RecentUplinks) - 1; i >= 0; i-- {
		switch up = dev.RecentUplinks[i]; up.Payload.MHDR.MType {
		case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP, ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
			break
		default:
			log.FromContext(ctx).WithFields(log.Fields(
				"m_type", up.Payload.MHDR.MType,
				"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
			)).Warn("Unknown MType stored in RecentUplinks")
		}
	}

	if len(dev.MACState.PendingRequests) == 0 &&
		len(dev.MACState.QueuedResponses) == 0 &&
		len(dev.QueuedApplicationDownlinks) == 0 &&
		!(up.GetPayload() != nil &&
			(up.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP ||
				up.Payload.MHDR.MType == ttnpb.MType_UNCONFIRMED_UP && up.Payload.GetMACPayload().FCtrl.ADRAckReq)) {
		return nil, errNoDownlink
	}

	pld := &ttnpb.MACPayload{
		FHDR: ttnpb.FHDR{
			DevAddr: *dev.EndDeviceIdentifiers.DevAddr,
			FCtrl: ttnpb.FCtrl{
				Ack: up != nil && up.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP,
			},
			FCnt: dev.Session.LastNFCntDown + 1,
		},
	}

	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if len(cmdBuf) <= fOptsCapacity && len(dev.QueuedApplicationDownlinks) > 0 {
		var down *ttnpb.ApplicationDownlink
		down, dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[0], dev.QueuedApplicationDownlinks[1:]

		if len(down.FRMPayload) > int(maxDownLen) {
			// TODO: Inform AS that payload is too long(https://github.com/TheThingsIndustries/lorawan-stack/issues/377)
		} else {
			pld.FHDR.FCnt = down.FCnt
			pld.FPort = down.FPort
			pld.FRMPayload = down.FRMPayload
			if down.Confirmed {
				dev.MACState.PendingApplicationDownlink = down
				dev.Session.LastConfFCntDown = pld.FCnt

				mType = ttnpb.MType_CONFIRMED_DOWN
			}
		}
	}

	if len(cmdBuf) > 0 && (pld.FPort == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || len(dev.Session.NwkSEncKey.Key) == 0 {
			return nil, errUnknownNwkSEncKey
		}

		var key types.AES128Key
		if dev.Session.NwkSEncKey.KEKLabel != "" {
			// TODO: (https://github.com/TheThingsIndustries/lorawan-stack/issues/271)
			panic("Unsupported")
		}
		copy(key[:], dev.Session.NwkSEncKey.Key[:])

		var err error
		cmdBuf, err = crypto.EncryptDownlink(key, *dev.EndDeviceIdentifiers.DevAddr, pld.FHDR.FCnt, cmdBuf)
		if err != nil {
			return nil, errEncryptMAC.WithCause(err)
		}
	}

	if pld.FPort == 0 {
		pld.FRMPayload = cmdBuf
		dev.Session.LastNFCntDown = pld.FCnt
	} else {
		pld.FHDR.FOpts = cmdBuf
	}

	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && pld.FCnt > dev.Session.LastNFCntDown {
		dev.Session.LastNFCntDown = pld.FCnt
	}

	pld.FHDR.FCtrl.FPending = fPending || len(dev.QueuedApplicationDownlinks) > 0

	switch {
	case dev.MACState.DeviceClass != ttnpb.CLASS_C,
		mType != ttnpb.MType_CONFIRMED_DOWN && len(dev.MACState.PendingRequests) == 0:
		break

	case dev.MACState.LastConfirmedDownlinkAt.Add(classCTimeout).After(time.Now()):
		return nil, errScheduleTooSoon

	default:
		dev.MACState.LastConfirmedDownlinkAt = timePtr(time.Now().UTC())
	}

	b, err := lorawan.MarshalMessage(ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: pld,
		},
	})
	if err != nil {
		return nil, errEncodePayload.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	var key types.AES128Key
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		if dev.Session.FNwkSIntKey == nil || len(dev.Session.FNwkSIntKey.Key) == 0 {
			return nil, errUnknownFNwkSIntKey
		}

		if dev.Session.NwkSEncKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(key[:], dev.Session.NwkSEncKey.Key[:])
	} else {
		if dev.Session.SNwkSIntKey == nil || len(dev.Session.SNwkSIntKey.Key) == 0 {
			return nil, errUnknownSNwkSIntKey
		}
		if dev.Session.SNwkSIntKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(key[:], dev.Session.SNwkSIntKey.Key[:])
	}

	var confFCnt uint32
	if pld.Ack {
		confFCnt = up.GetPayload().GetMACPayload().GetFCnt()
	}
	mic, err := crypto.ComputeDownlinkMIC(
		key,
		*dev.EndDeviceIdentifiers.DevAddr,
		confFCnt,
		b,
	)
	if err != nil {
		return nil, errComputeMIC
	}
	return append(b, mic[:]...), nil
}

// WindowEndFunc is a function, which is used by Network Server to determine the end of deduplication and cooldown windows.
type WindowEndFunc func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time

// NewWindowEndAfterFunc returns a WindowEndFunc, which closes
// the returned channel after at least duration d after up.ServerTime or if the context is done.
func NewWindowEndAfterFunc(d time.Duration) WindowEndFunc {
	return func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
		ch := make(chan time.Time, 1)

		if up.ReceivedAt.IsZero() {
			up.ReceivedAt = time.Now()
		}

		end := up.ReceivedAt.Add(d)
		if end.Before(time.Now()) {
			ch <- end
			return ch
		}

		go func() {
			time.Sleep(time.Until(up.ReceivedAt.Add(d)))
			ch <- end
		}()
		return ch
	}
}

// PeerGetter is the interface, which wraps GetPeer method.
type PeerGetter interface {
	GetPeer(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer
}

// NsGsClientFunc is the function used to get Gateway Server.
type NsGsClientFunc func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error)

// NewGatewayServerPeerGetterFunc returns a NsGsClientFunc, which uses g to retrieve Gateway Server clients.
func NewGatewayServerPeerGetterFunc(g PeerGetter) NsGsClientFunc {
	return func(ctx context.Context, ids ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
		p := g.GetPeer(ctx, ttnpb.PeerInfo_GATEWAY_SERVER, ids)
		if p == nil {
			return nil, errGatewayServerNotFound
		}
		return ttnpb.NewNsGsClient(p.Conn()), nil
	}
}

// NsJsClientFunc is the function used to get Join Server.
type NsJsClientFunc func(ctx context.Context, id ttnpb.EndDeviceIdentifiers) (ttnpb.NsJsClient, error)

// NewJoinServerPeerGetterFunc returns a NsJsClientFunc, which uses g to retrieve Join Server clients.
func NewJoinServerPeerGetterFunc(g PeerGetter) NsJsClientFunc {
	return func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (ttnpb.NsJsClient, error) {
		p := g.GetPeer(ctx, ttnpb.PeerInfo_JOIN_SERVER, ids)
		if p == nil {
			return nil, errJoinServerNotFound
		}
		return ttnpb.NewNsJsClient(p.Conn()), nil
	}
}

// Config represents the NetworkServer configuration.
type Config struct {
	Devices             DeviceRegistry    `name:"-"`
	DownlinkTasks       DownlinkTaskQueue `name:"-"`
	DeduplicationWindow time.Duration     `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`
	CooldownWindow      time.Duration     `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"`
}

// NetworkServer implements the Network Server component.
//
// The Network Server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component

	devices DeviceRegistry

	NetID types.NetID

	applicationServersMu *sync.RWMutex
	applicationServers   map[string]*applicationUpStream

	metadataAccumulators *sync.Map // uint64 -> *metadataAccumulator

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	macHandlers   *sync.Map // ttnpb.MACCommandIdentifier -> MACHandler
	downlinkTasks DownlinkTaskQueue

	deduplicationDone WindowEndFunc
	collectionDone    WindowEndFunc

	gsClient NsGsClientFunc
	jsClient NsJsClientFunc

	handleASUplink func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (bool, error)
}

// Option configures the NetworkServer.
type Option func(ns *NetworkServer)

// WithDeduplicationDoneFunc overrides the default WindowEndFunc, which
// is used to determine the end of uplink metadata deduplication.
func WithDeduplicationDoneFunc(f WindowEndFunc) Option {
	return func(ns *NetworkServer) {
		ns.deduplicationDone = f
	}
}

// WithCollectionDoneFunc overrides the default WindowEndFunc, which
// is used to determine the end of uplink duplicate collection.
func WithCollectionDoneFunc(f WindowEndFunc) Option {
	return func(ns *NetworkServer) {
		ns.collectionDone = f
	}
}

// WithNsGsClientFunc overrides the default NsGsClientFunc, which
// is used to get the Gateway Server by gateway identifiers.
func WithNsGsClientFunc(f NsGsClientFunc) Option {
	return func(ns *NetworkServer) {
		ns.gsClient = f
	}
}

// WithNsJsClientFunc overrides the default NsJsClientFunc, which
// is used to get the Join Server by end device identifiers.
func WithNsJsClientFunc(f NsJsClientFunc) Option {
	return func(ns *NetworkServer) {
		ns.jsClient = f
	}
}

// WithASUplinkHandler overrides the default function called, for sending the uplink to AS.
func WithASUplinkHandler(f func(context.Context, ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationUp) (bool, error)) Option {
	return func(ns *NetworkServer) {
		ns.handleASUplink = f
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
		devices:                 conf.Devices,
		downlinkTasks:           conf.DownlinkTasks,
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
		return &metadataAccumulator{}
	}

	for _, opt := range opts {
		opt(ns)
	}

	switch {
	case ns.deduplicationDone == nil && conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow is zero and WithDeduplicationDoneFunc not specified"))

	case ns.collectionDone == nil && conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow is zero and WithCollectionDoneFunc not specified"))

	case ns.collectionDone == nil && conf.CooldownWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("CooldownWindow is zero and WithCollectionDoneFunc not specified"))
	}

	if ns.downlinkTasks == nil {
		return nil, errInvalidConfiguration.WithCause(errors.New("DownlinkTasks is not specified"))
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
	if ns.jsClient == nil {
		ns.jsClient = NewJoinServerPeerGetterFunc(ns.Component)
	}

	if ns.handleASUplink == nil {
		ns.handleASUplink = func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (ok bool, err error) {
			ns.applicationServersMu.RLock()
			as, ok := ns.applicationServers[unique.ID(ctx, ids)]
			if ok {
				err = as.Send(up)
			}
			ns.applicationServersMu.RUnlock()
			return ok, err
		}
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterStreamHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthStreamHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsApplicationDownlinkQueue", cluster.HookName, c.ClusterAuthUnaryHook())

	ns.Component.RegisterTask(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err := ns.processDownlinkTask(ctx); err != nil {
				return err
			}
		}
	}, component.TaskRestartOnFailure)

	c.RegisterGRPC(ns)
	return ns, nil
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context) error {
	var scheduleErr bool
	var setErr bool
	var addErr bool
	err := ns.downlinkTasks.Pop(ctx, func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"delay", time.Now().Sub(t),
			"device_uid", unique.ID(ctx, devID),
			"start_at", t,
		))
		logger.Debug("Processing downlink task...")

		dev, err := ns.devices.SetByID(ctx, devID.ApplicationIdentifiers, devID.DeviceID,
			[]string{
				"frequency_plan_id",
				"mac_state",
				"queued_application_downlinks",
				"recent_downlinks",
				"recent_uplinks",
				"session",
			},
			func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					return nil, nil, nil
				}

				if dev.MACState == nil {
					return nil, nil, errUnknownMACState
				}

				// TODO: Support RxRequest and proper class-based downlink here once https://github.com/TheThingsIndustries/lorawan-stack/pull/1376 is merged.

				type tx struct {
					ttnpb.TxSettings
					Delay time.Duration
				}
				slots := make([]tx, 0, 2)

				fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
				if err != nil {
					return nil, nil, errUnknownFrequencyPlan.WithCause(err)
				}

				band, err := band.GetByID(fp.BandID)
				if err != nil {
					return nil, nil, errUnknownBand.WithCause(err)
				}

				if len(dev.RecentUplinks) == 0 {
					return nil, nil, errUplinkNotFound
				}
				up := dev.RecentUplinks[len(dev.RecentUplinks)-1]

				var upADR bool
				for i := len(dev.RecentUplinks) - 1; i >= 0; i-- {
					switch up := dev.RecentUplinks[i]; up.Payload.MHDR.MType {
					case ttnpb.MType_JOIN_REQUEST:
						break
					case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP:
						upADR = up.Payload.GetMACPayload().FHDR.FCtrl.ADR
						break
					}
				}

				drIdx, err := band.Rx1DataRate(up.Settings.DataRateIndex, dev.MACState.CurrentParameters.Rx1DataRateOffset, dev.MACState.CurrentParameters.DownlinkDwellTime)
				if err != nil {
					return nil, nil, err
				}

				chIdx, err := band.Rx1Channel(up.Settings.ChannelIndex)
				if err != nil {
					return nil, nil, err
				}
				if uint(chIdx) < uint(len(dev.MACState.CurrentParameters.Channels)) &&
					dev.MACState.CurrentParameters.Channels[int(chIdx)] != nil &&
					dev.MACState.CurrentParameters.Channels[int(chIdx)].DownlinkFrequency != 0 {
					rx1 := tx{
						TxSettings: ttnpb.TxSettings{
							DataRateIndex:      drIdx,
							CodingRate:         "4/5",
							InvertPolarization: true,
							ChannelIndex:       chIdx,
							Frequency:          dev.MACState.CurrentParameters.Channels[int(chIdx)].DownlinkFrequency,
							TxPower:            int32(band.DefaultMaxEIRP),
						},
					}
					if up.Payload.MHDR.MType == ttnpb.MType_JOIN_REQUEST {
						rx1.Delay = band.JoinAcceptDelay1
					} else {
						rx1.Delay = time.Second * time.Duration(dev.MACState.CurrentParameters.Rx1Delay)
					}

					if err = setDownlinkModulation(&rx1.TxSettings, band.DataRates[drIdx]); err != nil {
						return nil, nil, err
					}
					slots = append(slots, rx1)
				}

				if uint(dev.MACState.CurrentParameters.Rx2DataRateIndex) > uint(len(band.DataRates)) {
					return nil, nil, errInvalidRx2DataRateIndex
				}

				rx2 := tx{
					TxSettings: ttnpb.TxSettings{
						DataRateIndex:      dev.MACState.CurrentParameters.Rx2DataRateIndex,
						CodingRate:         "4/5",
						InvertPolarization: true,
						Frequency:          dev.MACState.CurrentParameters.Rx2Frequency,
						TxPower:            int32(band.DefaultMaxEIRP),
					},
				}
				if up != nil && up.Payload.MHDR.MType == ttnpb.MType_JOIN_REQUEST {
					rx2.Delay = band.JoinAcceptDelay2
				} else {
					rx2.Delay = time.Second * time.Duration(1+dev.MACState.CurrentParameters.Rx1Delay)
				}

				if err = setDownlinkModulation(&rx2.TxSettings, band.DataRates[dev.MACState.CurrentParameters.Rx2DataRateIndex]); err != nil {
					return nil, nil, err
				}
				slots = append(slots, rx2)

				sort.SliceStable(up.RxMetadata, func(i, j int) bool {
					// TODO: Improve the sorting algorithm (https://github.com/TheThingsIndustries/ttn/issues/729)
					return up.RxMetadata[i].SNR > up.RxMetadata[j].SNR
				})

				ctx = events.ContextWithCorrelationID(ctx, append(
					up.CorrelationIDs,
					fmt.Sprintf("ns:downlink:%s", events.NewCorrelationID()),
				)...)

				var errs []error
				for _, s := range slots {
					// generateDownlink may mutate device, hence we may need to create a copy
					sDev := dev

					down := &ttnpb.DownlinkMessage{
						EndDeviceIDs:   &sDev.EndDeviceIdentifiers,
						Settings:       s.TxSettings,
						CorrelationIDs: events.CorrelationIDsFromContext(ctx),
					}

					if up.Payload.MHDR.MType == ttnpb.MType_JOIN_REQUEST || up.Payload.MHDR.MType == ttnpb.MType_REJOIN_REQUEST {
						down.RawPayload = sDev.MACState.QueuedJoinAccept

					} else {
						var maxUpDR ttnpb.DataRateIndex
						if upADR {
							maxUpDR = up.Settings.DataRateIndex
						}

						sDev = deepcopy.Copy(sDev).(*ttnpb.EndDevice)

						down.RawPayload, err = generateDownlink(ctx, sDev,
							band.DataRates[down.Settings.DataRateIndex].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
							band.DataRates[maxUpDR].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
						)
						if err != nil {
							return nil, nil, err
						}
					}
					if down.RawPayload == nil {
						return nil, nil, errNoPayload
					}

					for _, md := range up.RxMetadata {
						logger := logger.WithField(
							"gateway_uid", unique.ID(ctx, md.GatewayIdentifiers),
						)

						gs, err := ns.gsClient(ctx, md.GatewayIdentifiers)
						if err != nil {
							logger.WithError(err).Debug("Could not get Gateway Server")
							continue
						}

						down.TxMetadata = ttnpb.TxMetadata{
							GatewayIdentifiers: md.GatewayIdentifiers,
							Timestamp:          md.Timestamp + uint64(s.Delay.Nanoseconds()),
						}

						_, err = gs.ScheduleDownlink(ctx, down, ns.WithClusterAuth())
						if err != nil {
							errs = append(errs, err)
							continue
						}

						dev = sDev
						dev.MACState.QueuedJoinAccept = nil
						dev.RecentDownlinks = append(dev.RecentDownlinks, down)
						if len(dev.RecentDownlinks) > recentDownlinkCount {
							dev.RecentDownlinks = append(dev.RecentDownlinks[:0], dev.RecentDownlinks[len(dev.RecentDownlinks)-recentDownlinkCount:]...)
						}
						return dev, []string{
							"mac_state",
							"queued_application_downlinks",
							"recent_downlinks",
							"session",
						}, nil
					}
				}

				for i, err := range errs {
					logger = logger.WithField(
						fmt.Sprintf("error_%d", i), err,
					)
				}
				scheduleErr = true
				logger.Warn("All Gateway Servers failed to schedule the downlink")
				return nil, nil, errSchedule
			})

		switch {
		case scheduleErr:
			return err

		case err != nil && errors.Resemble(err, errNoDownlink):
			return nil

		case err != nil:
			setErr = true
			logger.WithError(err).Error("Failed to update device in registry")
			return err
		}

		if dev.GetMACState().GetDeviceClass() != ttnpb.CLASS_C {
			return nil
		}

		if err := ns.downlinkTasks.Add(ctx, devID, time.Now()); err != nil {
			addErr = true
			logger.WithError(err).Error("Failed to add class C device to downlink schedule")
			return err
		}
		return nil
	})
	if err != nil && !setErr && !addErr && !scheduleErr {
		ns.Logger().WithError(err).Error("Failed to pop device from downlink schedule")
	}
	return err
}

// RegisterServices registers services provided by ns at s.
func (ns *NetworkServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsNsServer(s, ns)
	ttnpb.RegisterAsNsServer(s, ns)
	ttnpb.RegisterNsEndDeviceRegistryServer(s, ns)
}

// RegisterHandlers registers gRPC handlers.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterNsEndDeviceRegistryHandler(ns.Context(), s, conn)
}

// Roles returns the roles that the Network Server fulfills.
func (ns *NetworkServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_NETWORK_SERVER}
}
