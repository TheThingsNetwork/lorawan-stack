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
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// recentUplinkCount is the maximum amount of recent uplinks stored per device.
	recentUplinkCount = 20

	// recentDownlinkCount is the maximum amount of recent downlinks stored per device.
	recentDownlinkCount = 20

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
		case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP, ttnpb.MType_JOIN_REQUEST:
			break
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

		if len(pld.FRMPayload) > int(maxDownLen) {
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
	Devices             DeviceRegistry `name:"-"`
	DeduplicationWindow time.Duration  `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`
	CooldownWindow      time.Duration  `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"`
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

	deduplicationDone WindowEndFunc
	collectionDone    WindowEndFunc

	gsClient NsGsClientFunc
	jsClient NsJsClientFunc

	macHandlers *sync.Map // ttnpb.MACCommandIdentifier -> MACHandler

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

	c.RegisterGRPC(ns)
	return ns, nil
}

func (ns *NetworkServer) deduplicateUplink(ctx context.Context, up *ttnpb.UplinkMessage) (*metadataAccumulator, func(), bool) {
	h := ns.hashPool.Get().(hash.Hash64)
	_, _ = h.Write(up.RawPayload)

	k := h.Sum64()

	h.Reset()
	ns.hashPool.Put(h)

	acc := ns.metadataAccumulatorPool.Get().(*metadataAccumulator)
	lv, isDup := ns.metadataAccumulators.LoadOrStore(k, acc)
	lv.(*metadataAccumulator).Add(up.RxMetadata...)

	if isDup {
		ns.metadataAccumulatorPool.Put(acc)
		return nil, nil, true
	}
	return acc, func() {
		ns.metadataAccumulators.Delete(k)
	}, false
}

// matchDevice tries to match the uplink message with a device and returns the matched device and session.
// The LastFCntUp in the matched session is updated according to the FCnt in up.
func (ns *NetworkServer) matchDevice(ctx context.Context, up *ttnpb.UplinkMessage) (*ttnpb.EndDevice, *ttnpb.Session, error) {
	pld := up.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithField("dev_addr", pld.DevAddr)

	type device struct {
		*ttnpb.EndDevice

		matchedSession *ttnpb.Session
		fCnt           uint32
		gap            uint32
	}

	var devs []device
	if err := ns.devices.RangeByAddr(pld.DevAddr,
		[]string{
			"frequency_plan_id",
			"mac_state",
			"pending_session",
			"recent_downlinks",
			"recent_uplinks",
			"resets_f_cnt",
			"session",
			"uses_32_bit_f_cnt",
		},
		func(dev *ttnpb.EndDevice) bool {
			if dev.MACState == nil || (dev.Session == nil && dev.PendingSession == nil) {
				return true
			}

			ses := dev.Session
			if ses == nil {
				ses = dev.PendingSession
			}
			devs = append(devs, device{
				EndDevice:      dev,
				matchedSession: ses,
			})
			return true

		}); err != nil {
		logger.WithError(err).Warn("Failed to find devices in registry by DevAddr")
		return nil, nil, err
	}

	matching := make([]device, 0, len(devs))

outer:
	for _, dev := range devs {
		fCnt := pld.FCnt

		switch {
		case !dev.Uses32BitFCnt, fCnt > dev.matchedSession.LastFCntUp:
		case fCnt > dev.matchedSession.LastFCntUp&0xffff:
			fCnt |= dev.matchedSession.LastFCntUp &^ 0xffff
		case dev.matchedSession.LastFCntUp < 0xffff0000:
			fCnt |= (dev.matchedSession.LastFCntUp + 0x10000) &^ 0xffff
		}

		gap := uint32(math.MaxUint32)
		if fCnt == 0 && dev.matchedSession.LastFCntUp == 0 && len(dev.RecentUplinks) == 0 {
			gap = 0
		} else if !dev.ResetsFCnt {
			if fCnt <= dev.matchedSession.LastFCntUp {
				continue outer
			}

			gap = fCnt - dev.matchedSession.LastFCntUp

			if dev.MACState.LoRaWANVersion.HasMaxFCntGap() {
				fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
				if err != nil {
					logger.WithError(err).WithFields(log.Fields(
						"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
					)).Warn("Failed to get the frequency plan of the device in registry")
					continue
				}

				band, err := band.GetByID(fp.BandID)
				if err != nil {
					logger.WithError(err).WithFields(log.Fields(
						"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
					)).Warn("Failed to get the band of the device in registry")
					continue
				}

				if gap > uint32(band.MaxFCntGap) {
					continue outer
				}
			}
		}

		matching = append(matching, device{
			EndDevice:      dev.EndDevice,
			matchedSession: dev.matchedSession,
			gap:            gap,
			fCnt:           fCnt,
		})
		if dev.ResetsFCnt && fCnt != pld.FCnt {
			matching = append(matching, device{
				EndDevice:      dev.EndDevice,
				matchedSession: dev.matchedSession,
				gap:            gap,
				fCnt:           pld.FCnt,
			})
		}
	}

	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	if len(up.RawPayload) < 4 {
		return nil, nil, errRawPayloadTooShort
	}
	b := up.RawPayload[:len(up.RawPayload)-4]

	for _, dev := range matching {
		if pld.Ack {
			if len(dev.RecentDownlinks) == 0 {
				// Uplink acknowledges a downlink, but no downlink was sent to the device,
				// hence it must be the wrong device.
				continue
			}
		}

		if dev.matchedSession.FNwkSIntKey == nil || len(dev.matchedSession.FNwkSIntKey.Key) == 0 {
			logger.WithFields(log.Fields(
				"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
			)).Warn("Device missing FNwkSIntKey in registry")
			continue
		}

		var fNwkSIntKey types.AES128Key
		if dev.matchedSession.FNwkSIntKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(fNwkSIntKey[:], dev.matchedSession.FNwkSIntKey.Key[:])

		var computedMIC [4]byte
		var err error
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(
				fNwkSIntKey,
				pld.DevAddr,
				dev.fCnt,
				b,
			)

		} else {
			if dev.matchedSession.SNwkSIntKey == nil || len(dev.matchedSession.SNwkSIntKey.Key) == 0 {
				logger.WithFields(log.Fields(
					"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
				)).Warn("Device missing SNwkSIntKey in registry")
				continue
			}

			var sNwkSIntKey types.AES128Key
			if dev.matchedSession.SNwkSIntKey.KEKLabel != "" {
				// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
				panic("Unsupported")
			}
			copy(sNwkSIntKey[:], dev.matchedSession.SNwkSIntKey.Key[:])

			var confFCnt uint32
			if pld.Ack {
				confFCnt = dev.matchedSession.LastConfFCntDown
			}

			computedMIC, err = crypto.ComputeUplinkMIC(
				sNwkSIntKey,
				fNwkSIntKey,
				confFCnt,
				uint8(up.Settings.DataRateIndex),
				uint8(up.Settings.ChannelIndex),
				pld.DevAddr,
				dev.fCnt,
				b,
			)
		}
		if err != nil {
			logger.WithError(err).WithFields(log.Fields(
				"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
			)).Error("Failed to compute MIC")
			continue
		}
		if !bytes.Equal(up.Payload.MIC, computedMIC[:]) {
			continue
		}

		if dev.fCnt == math.MaxUint32 {
			return nil, nil, errFCntTooHigh
		}
		dev.matchedSession.LastFCntUp = dev.fCnt
		return dev.EndDevice, dev.matchedSession, nil
	}
	return nil, nil, errDeviceNotFound
}

// MACHandler defines the behavior of a MAC command on a device.
type MACHandler func(ctx context.Context, dev *ttnpb.EndDevice, pld []byte, up *ttnpb.UplinkMessage) error

func (ns *NetworkServer) handleUplink(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()

	dev, ses, err := ns.matchDevice(ctx, up)
	if err != nil {
		return errDeviceNotFound.WithCause(err)
	}

	logger := log.FromContext(ctx).WithField("device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))

	pld := up.Payload.GetMACPayload()
	if pld == nil {
		return errNoPayload
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
		} else {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkNack{
				DownlinkNack: dev.MACState.PendingApplicationDownlink,
			}
		}
		asUp.CorrelationIDs = append(asUp.CorrelationIDs, up.CorrelationIDs...)

		logger.Debug("Sending downlink (n)ack to Application Server...")
		if ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, asUp); err != nil {
			return err
		} else if !ok {
			logger.Warn("Application Server not found, downlink (n)ack not sent")
		}
	}

	mac := pld.FOpts
	if len(mac) == 0 && pld.FPort == 0 {
		mac = pld.FRMPayload
	}

	if len(mac) > 0 && (len(pld.FOpts) == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if ses.NwkSEncKey == nil || len(ses.NwkSEncKey.Key) == 0 {
			return errUnknownNwkSEncKey
		}

		var key types.AES128Key
		if ses.NwkSEncKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(key[:], ses.NwkSEncKey.Key[:])

		mac, err = crypto.DecryptUplink(key, *dev.EndDeviceIdentifiers.DevAddr, pld.FCnt, mac)
		if err != nil {
			return errDecrypt.WithCause(err)
		}
	}

	var cmds []*ttnpb.MACCommand
	for r := bytes.NewReader(mac); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadUplink(r, cmd); err != nil {
			logger.
				WithField("unmarshaled", len(cmds)).
				WithError(err).
				Warn("Failed to unmarshal MAC commands")
			break
		}
		cmds = append(cmds, cmd)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, dev, up)

	var handleErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"default_mac_parameters",
			"downlink_margin",
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_version",
			"mac_settings",
			"mac_state",
			"pending_session",
			"recent_uplinks",
			"resets_f_cnt",
			"session",
			"supports_join",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				logger.Warn("Device deleted during uplink handling, dropping...")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			paths := make([]string, 0, 9)

			storedSes := stored.Session
			if ses != dev.Session {
				storedSes = stored.PendingSession
			}

			if storedSes.GetSessionKeyID() != ses.SessionKeyID {
				logger.Warn("Device changed session during uplink handling, dropping...")
				handleErr = true
				return nil, nil, errOutdatedData
			}
			if storedSes.GetLastFCntUp() > ses.LastFCntUp && !stored.ResetsFCnt {
				logger.WithFields(log.Fields(
					"stored_f_cnt", storedSes.GetLastFCntUp(),
					"got_f_cnt", ses.LastFCntUp,
				)).Warn("A more recent uplink was received by device during uplink handling, dropping...")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			if ses == dev.Session {
				stored.Session = ses
				paths = append(paths, "session")
			} else {
				stored.PendingSession = ses
				paths = append(paths, "pending_session")
			}

			stored.RecentUplinks = append(stored.RecentUplinks, up)
			if len(stored.RecentUplinks) >= recentUplinkCount {
				stored.RecentUplinks = stored.RecentUplinks[len(stored.RecentUplinks)-recentUplinkCount+1:]
			}
			paths = append(paths, "recent_uplinks")

			if stored.MACState != nil {
				stored.MACState.PendingApplicationDownlink = nil
			} else if err := resetMACState(ns.Component.FrequencyPlans, stored); err != nil {
				handleErr = true
				return nil, nil, err
			}
			paths = append(paths, "mac_state")

			stored.MACState.QueuedResponses = stored.MACState.QueuedResponses[:0]

		outer:
			for len(cmds) > 0 {
				cmd, cmds := cmds[0], cmds[1:]
				switch cmd.CID {
				case ttnpb.CID_RESET:
					err = handleResetInd(ctx, stored, cmd.GetResetInd(), ns.Component.FrequencyPlans)
				case ttnpb.CID_LINK_CHECK:
					err = handleLinkCheckReq(ctx, stored, up)
				case ttnpb.CID_LINK_ADR:
					pld := cmd.GetLinkADRAns()
					dupCount := 0
					if stored.MACState.LoRaWANVersion == ttnpb.MAC_V1_0_2 {
						for _, dup := range cmds {
							if dup.CID != ttnpb.CID_LINK_ADR {
								break
							}
							if *dup.GetLinkADRAns() != *pld {
								err = errInvalidPayload
								break
							}
							dupCount++
						}
					}
					if err != nil {
						break
					}
					cmds = cmds[dupCount:]
					err = handleLinkADRAns(ctx, stored, pld, uint(dupCount), ns.Component.FrequencyPlans)
				case ttnpb.CID_DUTY_CYCLE:
					err = handleDutyCycleAns(ctx, stored)
				case ttnpb.CID_RX_PARAM_SETUP:
					err = handleRxParamSetupAns(ctx, stored, cmd.GetRxParamSetupAns())
				case ttnpb.CID_DEV_STATUS:
					err = handleDevStatusAns(ctx, stored, cmd.GetDevStatusAns(), ses.LastFCntUp, up.ReceivedAt)
					paths = append(paths,
						"battery_percentage",
						"downlink_margin",
						"last_dev_status_received_at",
						"power_state",
					)
				case ttnpb.CID_NEW_CHANNEL:
					err = handleNewChannelAns(ctx, stored, cmd.GetNewChannelAns())
				case ttnpb.CID_RX_TIMING_SETUP:
					err = handleRxTimingSetupAns(ctx, stored)
				case ttnpb.CID_TX_PARAM_SETUP:
					err = handleTxParamSetupAns(ctx, stored)
				case ttnpb.CID_DL_CHANNEL:
					err = handleDLChannelAns(ctx, stored, cmd.GetDLChannelAns())
				case ttnpb.CID_REKEY:
					err = handleRekeyInd(ctx, stored, cmd.GetRekeyInd())
				case ttnpb.CID_ADR_PARAM_SETUP:
					err = handleADRParamSetupAns(ctx, stored)
				case ttnpb.CID_DEVICE_TIME:
					err = handleDeviceTimeReq(ctx, stored, up)
				case ttnpb.CID_REJOIN_PARAM_SETUP:
					err = handleRejoinParamSetupAns(ctx, stored, cmd.GetRejoinParamSetupAns())
				case ttnpb.CID_PING_SLOT_INFO:
					err = handlePingSlotInfoReq(ctx, stored, cmd.GetPingSlotInfoReq())
				case ttnpb.CID_PING_SLOT_CHANNEL:
					err = handlePingSlotChannelAns(ctx, stored, cmd.GetPingSlotChannelAns())
				case ttnpb.CID_BEACON_TIMING:
					err = handleBeaconTimingReq(ctx, stored)
				case ttnpb.CID_BEACON_FREQ:
					err = handleBeaconFreqAns(ctx, stored, cmd.GetBeaconFreqAns())
				case ttnpb.CID_DEVICE_MODE:
					err = handleDeviceModeInd(ctx, stored, cmd.GetDeviceModeInd())
				default:
					h, ok := ns.macHandlers.Load(cmd.CID)
					if !ok {
						logger.WithField("cid", cmd.CID).Warn("Unknown MAC command received, skipping the rest...")
						break outer
					}
					err = h.(MACHandler)(ctx, stored, cmd.GetRawPayload(), up)
				}
				if err != nil {
					logger.WithField("cid", cmd.CID).WithError(err).Warn("Failed to process MAC command")
					handleErr = true
					return nil, nil, err
				}
			}
			if stored.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				stored.Session = ses
				stored.PendingSession = nil
			} else if stored.PendingSession != nil {
				handleErr = true
				return nil, nil, errNoRekey
			}
			paths = append(paths,
				"pending_session",
				"session",
			)

			if stored.Session != ses {
				// Sanity check
				panic(fmt.Errorf("Session mismatch"))
			}
			return stored, paths, nil
		})
	if err != nil && !handleErr {
		logger.WithError(err).Error("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
	}
	if err != nil {
		return err
	}

	registerForwardUplink(ctx, dev, up)
	ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		CorrelationIDs:       up.CorrelationIDs,
		Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
			FCnt:         dev.Session.LastFCntUp,
			FPort:        pld.FPort,
			FRMPayload:   pld.FRMPayload,
			RxMetadata:   up.RxMetadata,
			SessionKeyID: dev.Session.SessionKeyID,
			Settings:     up.Settings,
		}},
	})

	updateTimeout := appQueueUpdateTimeout
	if !ok {
		logger.Warn("Application Server not found, not forwarding uplink")
		updateTimeout = 0
	}

	time.AfterFunc(updateTimeout, func() {
		// TODO: Decouple Class A downlink from uplink. (https://github.com/TheThingsIndustries/lorawan-stack/issues/905)
		var schedErr bool
		_, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID, []string{
			"frequency_plan_id",
			"mac_state",
			"queued_application_downlinks",
			"recent_downlinks",
			"recent_uplinks",
			"session",
		}, func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if err := ns.scheduleDownlink(ctx, stored, acc, nil); err != nil {
				schedErr = true
				return nil, nil, err
			}
			return stored, []string{
				"mac_state",
				"queued_application_downlinks",
				"recent_downlinks",
				"recent_uplinks",
				"session",
			}, nil
		})
		if schedErr && !errors.Resemble(err, errNoDownlink) {
			logger.Debug("No downlink to schedule")
		} else if schedErr {
			logger.WithError(err).Error("Failed to schedule downlink in reception slot")
		} else {
			logger.WithError(err).Error("Failed to update device in registry")
			// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
		}
	})
	return err
}

// newDevAddr generates a DevAddr for specified EndDevice.
func (ns *NetworkServer) newDevAddr(context.Context, *ttnpb.EndDevice) types.DevAddr {
	nwkAddr := make([]byte, types.NwkAddrLength(ns.NetID))
	random.Read(nwkAddr)
	nwkAddr[0] &= 0xff >> (8 - types.NwkAddrBits(ns.NetID)%8)
	devAddr, err := types.NewDevAddr(ns.NetID, nwkAddr)
	if err != nil {
		panic(errors.New("failed to create new DevAddr").WithCause(err))
	}
	return devAddr
}

func (ns *NetworkServer) handleJoin(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()

	pld := up.Payload.GetJoinRequestPayload()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", pld.DevEUI,
		"join_eui", pld.JoinEUI,
	))

	dev, err := ns.devices.GetByEUI(ctx, pld.JoinEUI, pld.DevEUI,
		[]string{
			"frequency_plan_id",
			"lorawan_version",
			"mac_state",
			"session",
		},
	)
	if err != nil {
		logger.WithError(err).Error("Failed to load device from registry")
		return err
	}

	devAddr := ns.newDevAddr(ctx, dev)
	for dev.Session != nil && devAddr.Equal(dev.Session.DevAddr) {
		devAddr = ns.newDevAddr(ctx, dev)
	}

	fp, err := ns.FrequencyPlans.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return errUnknownFrequencyPlan.WithCause(err)
	}

	req := &ttnpb.JoinRequest{
		RawPayload: up.RawPayload,
		Payload:    up.Payload,
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DevEUI:  &pld.DevEUI,
			JoinEUI: &pld.JoinEUI,
			DevAddr: &devAddr,
		},
		NetID:              ns.NetID,
		SelectedMACVersion: dev.LoRaWANVersion, // Assume NS version is always higher than the version of the device
		RxDelay:            dev.MACState.DesiredParameters.Rx1Delay,
		CFList:             frequencyplans.CFList(*fp, dev.LoRaWANPHYVersion),
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: dev.MACState.DesiredParameters.Rx1DataRateOffset,
			Rx2DR:       dev.MACState.DesiredParameters.Rx2DataRateIndex,
			OptNeg:      true,
		},
	}

	js, err := ns.jsClient(ctx, dev.EndDeviceIdentifiers)
	if err != nil {
		logger.WithError(err).Debug("Could not get Join Server")
		return err
	}

	resp, err := js.HandleJoin(ctx, req, ns.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Join Server failed to handle join-request")
		return err
	}
	registerForwardUplink(ctx, dev, up)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, dev, up)

	var invalidatedQueue []*ttnpb.ApplicationDownlink
	var resetErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"default_mac_parameters",
			"frequency_plan_id",
			"lorawan_version",
			"mac_state",
			"queued_application_downlinks",
			"recent_uplinks",
			"session",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			paths := make([]string, 0, 5)

			dev.Session = &ttnpb.Session{
				DevAddr:     devAddr,
				SessionKeys: resp.SessionKeys,
				StartedAt:   time.Now(),
			}
			paths = append(paths, "session")

			dev.EndDeviceIdentifiers.DevAddr = &devAddr
			paths = append(paths, "ids.dev_addr")

			if err := resetMACState(ns.Component.FrequencyPlans, dev); err != nil {
				resetErr = true
				return nil, nil, err
			}

			if req.DownlinkSettings.OptNeg && dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) > 0 {
				// The version will be further negotiated via RekeyInd/RekeyConf
				dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
			}
			dev.MACState.CurrentParameters.Rx1Delay = req.RxDelay
			dev.MACState.CurrentParameters.Rx1DataRateOffset = req.DownlinkSettings.Rx1DROffset
			dev.MACState.CurrentParameters.Rx2DataRateIndex = req.DownlinkSettings.Rx2DR
			dev.MACState.DesiredParameters.Rx1Delay = dev.MACState.CurrentParameters.Rx1Delay
			dev.MACState.DesiredParameters.Rx1DataRateOffset = dev.MACState.CurrentParameters.Rx1DataRateOffset
			dev.MACState.DesiredParameters.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
			paths = append(paths, "mac_state")

			dev.RecentUplinks = append(dev.RecentUplinks, up)
			if len(dev.RecentUplinks) > recentUplinkCount {
				dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
			}
			paths = append(paths, "recent_uplinks")

			invalidatedQueue = dev.QueuedApplicationDownlinks
			dev.QueuedApplicationDownlinks = nil
			paths = append(paths, "queued_application_downlinks")

			return dev, paths, nil
		})
	if err != nil && !resetErr {
		logger.WithError(err).Error("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
	}
	if err != nil {
		return err
	}

	var schedErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"frequency_plan_id",
			"mac_state",
			"queued_application_downlinks",
			"recent_downlinks",
			"recent_uplinks",
			"session",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			err := ns.scheduleDownlink(ctx, stored, acc, resp.RawPayload)
			if err != nil {
				schedErr = true
				return nil, nil, err
			}
			return stored, []string{
				"mac_state",
				"queued_application_downlinks",
				"recent_downlinks",
				"session",
			}, nil
		})
	if schedErr {
		logger.WithError(err).Debug("Failed to schedule join-accept")
		return err
	} else if err != nil {
		logger.WithError(err).Error("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
		return err
	}

	logger = logger.WithField(
		"application_uid", unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers),
	)

	logger.Debug("Sending join-accept to AS...")
	_, err = ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		CorrelationIDs:       up.CorrelationIDs,
		Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
			AppSKey:              resp.SessionKeys.AppSKey,
			InvalidatedDownlinks: invalidatedQueue,
			SessionKeyID:         dev.Session.SessionKeyID,
			SessionStartedAt:     dev.Session.StartedAt,
		}},
	})
	if err != nil {
		logger.WithError(err).Errorf("Failed to send join-accept to AS")
		return err
	}

	return nil
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()
	// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// scheduleDownlink schedules downlink with payload b for device dev.
// up represents the uplink message, which triggered the downlink(if such exists).
// acc represents the metadata accumulator used to deduplicate up.
// scheduleDownlink returns the downlink scheduled and error if any.
func (ns *NetworkServer) scheduleDownlink(ctx context.Context, dev *ttnpb.EndDevice, acc *metadataAccumulator, b []byte) error {
	if dev.MACState == nil {
		return errUnknownMACState
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
	))

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

	if len(dev.RecentUplinks) == 0 {
		return errUplinkNotFound
	}
	up := dev.RecentUplinks[len(dev.RecentUplinks)-1]

	if b == nil && up.Payload.MHDR.MType == ttnpb.MType_JOIN_REQUEST || up.Payload.MHDR.MType == ttnpb.MType_REJOIN_REQUEST {
		return errNoPayload
	}

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
		return err
	}

	chIdx, err := band.Rx1Channel(up.Settings.ChannelIndex)
	if err != nil {
		return err
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
			return err
		}
		slots = append(slots, rx1)
	}

	if uint(dev.MACState.CurrentParameters.Rx2DataRateIndex) > uint(len(band.DataRates)) {
		return errInvalidRx2DataRateIndex
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
		return err
	}
	slots = append(slots, rx2)

	mds := up.RxMetadata
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

	ctx = events.ContextWithCorrelationID(ctx, append(
		up.CorrelationIDs,
		fmt.Sprintf("ns:downlink:%s", events.NewCorrelationID()),
	)...)

	var errs []error
	for _, s := range slots {
		down := &ttnpb.DownlinkMessage{
			EndDeviceIDs:   &dev.EndDeviceIdentifiers,
			Settings:       s.TxSettings,
			RawPayload:     b,
			CorrelationIDs: events.CorrelationIDsFromContext(ctx),
		}

		if down.RawPayload == nil {
			var maxUpDR ttnpb.DataRateIndex
			if upADR {
				maxUpDR = up.Settings.DataRateIndex
			}

			down.RawPayload, err = generateDownlink(ctx, dev,
				band.DataRates[down.Settings.DataRateIndex].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
				band.DataRates[maxUpDR].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
			)
			if err != nil {
				return err
			}
		}

		for _, md := range mds {
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

			dev.RecentDownlinks = append(dev.RecentDownlinks, down)
			if len(dev.RecentDownlinks) > recentDownlinkCount {
				dev.RecentDownlinks = append(dev.RecentDownlinks[:0], dev.RecentDownlinks[len(dev.RecentDownlinks)-recentDownlinkCount:]...)
			}
			return nil
		}
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Warn("All Gateway Servers failed to schedule the downlink")
	return errSchedule
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
