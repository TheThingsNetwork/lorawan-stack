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
	"go.thethings.network/lorawan-stack/pkg/errors"
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
func generateDownlink(ctx context.Context, dev *ttnpb.EndDevice, ack bool, confFCnt uint32) (b []byte, err error) {
	if !ack && confFCnt > 0 {
		panic("confFCnt must be 0 if ack is false")
	}

	if dev.MACState == nil {
		return nil, errUnknownMACState
	}

	if dev.Session == nil {
		return nil, errEmptySession
	}

	dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]

	// TODO: Queue LinkADRReq(https://github.com/TheThingsIndustries/ttn/issues/837)

	if dev.MACState.DesiredParameters.MaxDutyCycle != dev.MACState.CurrentParameters.MaxDutyCycle {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_DutyCycleReq{
			MaxDutyCycle: dev.MACState.DesiredParameters.MaxDutyCycle,
		}).MACCommand())
	}

	if dev.MACState.DesiredParameters.Rx2Frequency != dev.MACState.CurrentParameters.Rx2Frequency ||
		dev.MACState.DesiredParameters.Rx2DataRateIndex != dev.MACState.CurrentParameters.Rx2DataRateIndex ||
		dev.MACState.DesiredParameters.Rx1DataRateOffset != dev.MACState.CurrentParameters.Rx1DataRateOffset {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_RxParamSetupReq{
			Rx2Frequency:      dev.MACState.DesiredParameters.Rx2Frequency,
			Rx2DataRateIndex:  dev.MACState.DesiredParameters.Rx2DataRateIndex,
			Rx1DataRateOffset: dev.MACState.DesiredParameters.Rx1DataRateOffset,
		}).MACCommand())
	}

	if dev.LastStatusReceivedAt == nil ||
		dev.MACSettings.StatusCountPeriodicity > 0 && dev.NextStatusAfter == 0 ||
		dev.MACSettings.StatusTimePeriodicity > 0 && dev.LastStatusReceivedAt.Add(dev.MACSettings.StatusTimePeriodicity).Before(time.Now()) {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, ttnpb.CID_DEV_STATUS.MACCommand())
		dev.NextStatusAfter = dev.MACSettings.StatusCountPeriodicity

	} else if dev.NextStatusAfter != 0 {
		dev.NextStatusAfter--
	}

	for i := len(dev.MACState.DesiredParameters.Channels); i < len(dev.MACState.CurrentParameters.Channels); i++ {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_NewChannelReq{
			ChannelIndex: uint32(i),
		}).MACCommand())
	}

	for i, ch := range dev.MACState.DesiredParameters.Channels {
		if len(dev.MACState.CurrentParameters.Channels) <= i ||
			ch.MinDataRateIndex != dev.MACState.CurrentParameters.Channels[i].MinDataRateIndex ||
			ch.MaxDataRateIndex != dev.MACState.CurrentParameters.Channels[i].MaxDataRateIndex ||
			ch.UplinkFrequency != dev.MACState.CurrentParameters.Channels[i].UplinkFrequency {

			dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_NewChannelReq{
				ChannelIndex:     uint32(i),
				MinDataRateIndex: ch.MinDataRateIndex,
				MaxDataRateIndex: ch.MaxDataRateIndex,
			}).MACCommand())
		}

		if (len(dev.MACState.CurrentParameters.Channels) <= i && ch.UplinkFrequency != ch.DownlinkFrequency) ||
			ch.DownlinkFrequency != dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency {

			dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_DLChannelReq{
				ChannelIndex: uint32(i),
				Frequency:    ch.DownlinkFrequency,
			}).MACCommand())
		}
	}

	if dev.MACState.DesiredParameters.Rx1Delay != dev.MACState.CurrentParameters.Rx1Delay {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_RxTimingSetupReq{
			Delay: dev.MACState.DesiredParameters.Rx1Delay,
		}).MACCommand())
	}

	if dev.MACState.DesiredParameters.MaxEIRP != dev.MACState.CurrentParameters.MaxEIRP ||
		dev.MACState.DesiredParameters.DownlinkDwellTime != dev.MACState.CurrentParameters.DownlinkDwellTime ||
		dev.MACState.DesiredParameters.UplinkDwellTime != dev.MACState.CurrentParameters.UplinkDwellTime {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_TxParamSetupReq{
			MaxEIRPIndex:      ttnpb.Float32ToDeviceEIRP(dev.MACState.DesiredParameters.MaxEIRP),
			DownlinkDwellTime: dev.MACState.DesiredParameters.DownlinkDwellTime,
			UplinkDwellTime:   dev.MACState.DesiredParameters.UplinkDwellTime,
		}).MACCommand())
	}

	if dev.MACState.DesiredParameters.ADRAckLimit != dev.MACState.CurrentParameters.ADRAckLimit ||
		dev.MACState.DesiredParameters.ADRAckDelay != dev.MACState.CurrentParameters.ADRAckDelay {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_ADRParamSetupReq{
			ADRAckLimitExponent: ttnpb.Uint32ToADRAckLimitExponent(dev.MACState.DesiredParameters.ADRAckLimit),
			ADRAckDelayExponent: ttnpb.Uint32ToADRAckDelayExponent(dev.MACState.DesiredParameters.ADRAckDelay),
		}).MACCommand())
	}

	// TODO: Queue ForceRejoinReq(https://github.com/TheThingsIndustries/ttn/issues/837)

	if dev.MACState.DesiredParameters.RejoinTimePeriodicity != dev.MACState.CurrentParameters.RejoinTimePeriodicity {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_RejoinParamSetupReq{
			MaxTimeExponent:  dev.MACState.DesiredParameters.RejoinTimePeriodicity,
			MaxCountExponent: dev.MACState.DesiredParameters.RejoinCountPeriodicity,
		}).MACCommand())
	}

	if dev.MACState.DesiredParameters.PingSlotDataRateIndex != dev.MACState.CurrentParameters.PingSlotDataRateIndex ||
		dev.MACState.DesiredParameters.PingSlotFrequency != dev.MACState.CurrentParameters.PingSlotFrequency {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_PingSlotChannelReq{
			Frequency:     dev.MACState.DesiredParameters.PingSlotFrequency,
			DataRateIndex: dev.MACState.DesiredParameters.PingSlotDataRateIndex,
		}).MACCommand())
	}

	if dev.MACState.DesiredParameters.BeaconFrequency != dev.MACState.CurrentParameters.BeaconFrequency {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_BeaconFreqReq{
			Frequency: dev.MACState.DesiredParameters.BeaconFrequency,
		}).MACCommand())
	}

	if !ack &&
		len(dev.MACState.PendingRequests) == 0 &&
		len(dev.MACState.QueuedResponses) == 0 &&
		len(dev.QueuedApplicationDownlinks) == 0 {
		return nil, errNoDownlink
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

	// TODO: Ensure cmds can be answered in one frame (https://github.com/TheThingsIndustries/ttn/issues/836)

	cmdBuf := make([]byte, 0, len(cmds))
	for _, cmd := range cmds {
		cmdBuf, err = cmd.AppendLoRaWAN(cmdBuf)
		if err != nil {
			return nil, errEncodeMAC.WithCause(err)
		}
	}

	pld := &ttnpb.MACPayload{
		FHDR: ttnpb.FHDR{
			DevAddr: *dev.EndDeviceIdentifiers.DevAddr,
			FCtrl: ttnpb.FCtrl{
				Ack: ack,
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
		cmdBuf, err = crypto.EncryptDownlink(key, *dev.EndDeviceIdentifiers.DevAddr, pld.FHDR.FCnt, cmdBuf)
		if err != nil {
			return nil, errEncryptMAC.WithCause(err)
		}
	}

	if pld.FPort == 0 {
		pld.FRMPayload = cmdBuf
		dev.Session.NextNFCntDown++
	} else {
		// TODO: Ensure that maxPayloadSize of the data rate is not exceeded. (https://github.com/TheThingsIndustries/lorawan-stack/issues/995)
		pld.FHDR.FOpts = cmdBuf
	}

	// TODO: Set to true if commands were trimmed. (https://github.com/TheThingsIndustries/ttn/issues/836)
	pld.FHDR.FCtrl.FPending = len(dev.QueuedApplicationDownlinks) > 0

	switch {
	case dev.MACState.DeviceClass != ttnpb.CLASS_C,
		mType != ttnpb.MType_CONFIRMED_DOWN && len(dev.MACState.PendingRequests) == 0:
		break

	case dev.MACState.NextConfirmedDownlinkAt.After(time.Now()):
		return nil, errScheduleTooSoon

	default:
		t := time.Now().Add(classCTimeout).UTC()
		dev.MACState.NextConfirmedDownlinkAt = &t
	}

	b, err = (ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: pld,
		},
	}).MarshalLoRaWAN()
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

// NsGsClientFunc is the function used to get Gateway Server.
type NsGsClientFunc func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error)

// PeerGetter is the interface, which wraps GetPeer method.
type PeerGetter interface {
	GetPeer(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer
}

// NewGatewayServerPeerGetterFunc returns a NsGsClientFunc, which uses g to retrieve Gateway Server clients.
func NewGatewayServerPeerGetterFunc(g PeerGetter) NsGsClientFunc {
	return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
		p := g.GetPeer(ctx, ttnpb.PeerInfo_GATEWAY_SERVER, id)
		if p == nil {
			return nil, errGatewayServerNotFound
		}
		return ttnpb.NewNsGsClient(p.Conn()), nil
	}
}

// Config represents the NetworkServer configuration.
type Config struct {
	Devices             DeviceRegistry     `name:"-"`
	JoinServers         []ttnpb.NsJsClient `name:"-"`
	DeduplicationWindow time.Duration      `name:"deduplication-window" description:"Time window during which, duplicate messages are collected for metadata"`
	CooldownWindow      time.Duration      `name:"cooldown-window" description:"Time window starting right after deduplication window, during which, duplicate messages are discarded"`
}

// NetworkServer implements the Network Server component.
//
// The Network Server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component

	devices DeviceRegistry

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
		devices:                 conf.Devices,
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

// matchDevice tries to match the uplink message with a device.
// If the uplink message matches a fallback session, that fallback session is recovered.
// If successful, the FCnt in the uplink message is set to the full FCnt. The NextFCntUp in the session is updated accordingly.
func (ns *NetworkServer) matchDevice(ctx context.Context, up *ttnpb.UplinkMessage) (*ttnpb.EndDevice, error) {
	pld := up.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithField("dev_addr", pld.DevAddr)

	var devs []*ttnpb.EndDevice
	if err := ns.devices.RangeByAddr(pld.DevAddr, func(dev *ttnpb.EndDevice) bool {
		if dev.MACState == nil || dev.Session == nil {
			return true
		}
		if dev.SessionFallback != nil && dev.SessionFallback.DevAddr == pld.DevAddr {
			dev.Session = dev.SessionFallback
			dev.SessionFallback = nil
		}
		devs = append(devs, dev)
		return true
	}); err != nil {
		logger.WithError(err).Warn("Failed to find devices in registry by DevAddr")
		return nil, err
	}

	type device struct {
		*ttnpb.EndDevice

		fCnt uint32
		gap  uint32
	}
	matching := make([]device, 0, len(devs))

outer:
	for _, dev := range devs {
		fCnt := pld.FCnt

		switch {
		case !dev.Uses32BitFCnt, fCnt >= dev.Session.NextFCntUp:
		case fCnt > dev.Session.NextFCntUp&0xffff:
			fCnt |= dev.Session.NextFCntUp &^ 0xffff
		case dev.Session.NextFCntUp < 0xffff<<16:
			fCnt |= (dev.Session.NextFCntUp + 1<<16) &^ 0xffff
		}

		gap := uint32(math.MaxUint32)
		if !dev.ResetsFCnt {
			if dev.Session.NextFCntUp > fCnt {
				continue outer
			}

			gap = fCnt - dev.Session.NextFCntUp

			if dev.MACState.LoRaWANVersion.HasMaxFCntGap() {
				fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
				if err != nil {
					logger.WithError(err).WithFields(log.Fields(
						"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
					)).Warn("Failed to get the frequency plan of a device in registry")
					continue
				}

				band, err := band.GetByID(fp.BandID)
				if err != nil {
					logger.WithError(err).WithFields(log.Fields(
						"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
					)).Warn("Failed to get band of a device in registry")
					continue
				}

				if gap > uint32(band.MaxFCntGap) {
					continue outer
				}
			}
		}

		matching = append(matching, device{
			EndDevice: dev,
			gap:       gap,
			fCnt:      fCnt,
		})
		if dev.ResetsFCnt && fCnt != pld.FCnt {
			matching = append(matching, device{
				EndDevice: dev,
				gap:       gap,
				fCnt:      pld.FCnt,
			})
		}
	}

	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	if len(up.RawPayload) < 4 {
		return nil, errRawPayloadTooLong
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

		if dev.Session.FNwkSIntKey == nil || len(dev.Session.FNwkSIntKey.Key) == 0 {
			logger.WithFields(log.Fields(
				"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
			)).Warn("Device missing FNwkSIntKey in registry")
			continue
		}

		var fNwkSIntKey types.AES128Key
		if dev.Session.FNwkSIntKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(fNwkSIntKey[:], dev.Session.FNwkSIntKey.Key[:])

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
			if dev.Session.SNwkSIntKey == nil || len(dev.Session.SNwkSIntKey.Key) == 0 {
				logger.WithFields(log.Fields(
					"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
				)).Warn("Device missing SNwkSIntKey in registry")
				continue
			}

			var sNwkSIntKey types.AES128Key
			if dev.Session.SNwkSIntKey.KEKLabel != "" {
				// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
				panic("Unsupported")
			}
			copy(sNwkSIntKey[:], dev.Session.SNwkSIntKey.Key[:])

			var confFCnt uint32
			if pld.Ack {
				confFCnt = dev.Session.LastConfFCntDown
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
			return nil, errFCntTooHigh
		}
		dev.Session.NextFCntUp = dev.fCnt + 1
		return dev.EndDevice, nil
	}
	return nil, errDeviceNotFound
}

// MACHandler defines the behavior of a MAC command on a device.
type MACHandler func(ctx context.Context, dev *ttnpb.EndDevice, pld []byte, up *ttnpb.UplinkMessage) error

func (ns *NetworkServer) handleUplink(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()

	dev, err := ns.matchDevice(ctx, up)
	if err != nil {
		return errDeviceNotFound.WithCause(err)
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
	))

	pld := up.Payload.GetMACPayload()
	if pld == nil {
		return errNoPayload
	}

	ns.applicationServersMu.RLock()
	asCl, asOk := ns.applicationServers[unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers)]
	ns.applicationServersMu.RUnlock()

	if dev.MACState != nil && dev.MACState.PendingApplicationDownlink != nil {
		asUp := &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			CorrelationIDs:       dev.MACState.PendingApplicationDownlink.CorrelationIDs,
		}

		if pld.Ack {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: dev.MACState.PendingApplicationDownlink,
			}
			asUp.CorrelationIDs = append(asUp.CorrelationIDs, up.CorrelationIDs...)
		} else {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkNack{
				DownlinkNack: dev.MACState.PendingApplicationDownlink,
			}
		}

		if err := asCl.Send(asUp); err != nil {
			return err
		}
	}

	mac := pld.FOpts
	if len(mac) == 0 && pld.FPort == 0 {
		mac = pld.FRMPayload
	}

	if len(mac) > 0 && (len(pld.FOpts) == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || len(dev.Session.NwkSEncKey.Key) == 0 {
			return errUnknownNwkSEncKey
		}

		var key types.AES128Key
		if dev.Session.NwkSEncKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(key[:], dev.Session.NwkSEncKey.Key[:])

		mac, err = crypto.DecryptUplink(key, *dev.EndDeviceIdentifiers.DevAddr, pld.FCnt, mac)
		if err != nil {
			return errDecrypt.WithCause(err)
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

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, dev, up)

	var handleErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID, func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
		if stored.GetSession().GetSessionKeyID() != dev.Session.SessionKeyID {
			logger.Warn("Device changed session during uplink handling, dropping...")
			handleErr = true
			return nil, errOutdatedData
		}
		if stored.GetSession().GetNextFCntUp() > dev.Session.NextFCntUp && !stored.ResetsFCnt {
			logger.WithFields(log.Fields(
				"stored_f_cnt", stored.GetSession().GetNextFCntUp()-1,
				"got_f_cnt", dev.Session.NextFCntUp-1,
			)).Warn("A more recent uplink was received by device during uplink handling, dropping...")
			handleErr = true
			return nil, errOutdatedData
		}
		stored.Session = dev.Session

		stored.RecentUplinks = append(stored.RecentUplinks, up)
		if len(stored.RecentUplinks) >= recentUplinkCount {
			stored.RecentUplinks = stored.RecentUplinks[len(stored.RecentUplinks)-recentUplinkCount+1:]
		}

		if stored.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			// LoRaWAN1.1+ device will send a RekeyInd.
			stored.SessionFallback = nil
		}

		if stored.MACState != nil {
			stored.MACState.PendingApplicationDownlink = nil
		} else if err := resetMACState(ns.Component.FrequencyPlans, stored); err != nil {
			handleErr = true
			return nil, err
		}
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
				err = handleDevStatusAns(ctx, stored, cmd.GetDevStatusAns(), up.ReceivedAt)
			case ttnpb.CID_NEW_CHANNEL:
				err = handleNewChannelAns(ctx, stored, cmd.GetNewChannelAns())
			case ttnpb.CID_RX_TIMING_SETUP:
				err = handleRxTimingSetupAns(ctx, stored)
			case ttnpb.CID_TX_PARAM_SETUP:
				err = handleTxParamSetupAns(ctx, stored)
			case ttnpb.CID_DL_CHANNEL:
				err = handleDLChannelAns(ctx, stored, cmd.GetDlChannelAns())
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
				return nil, err
			}
		}
		return stored, nil
	})
	if err != nil && !handleErr {
		logger.WithError(err).Error("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
	}
	if err != nil {
		return err
	}

	updateTimeout := appQueueUpdateTimeout
	if !asOk {
		updateTimeout = 0
	}

	defer time.AfterFunc(updateTimeout, func() {
		// TODO: Decouple Class A downlink from uplink. (https://github.com/TheThingsIndustries/lorawan-stack/issues/905)

		needsAck := up.Payload.MType == ttnpb.MType_CONFIRMED_UP
		var confFCnt uint32
		if needsAck {
			confFCnt = pld.FHDR.FCnt
		}

		var b []byte
		var genErr bool
		dev, err := ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
			if dev == nil {
				return nil, errOutdatedData
			}

			b, err = generateDownlink(ctx, dev, needsAck, confFCnt)
			if err != nil && !errors.Resemble(err, errNoDownlink) {
				logger.WithError(err).Error("Failed to generate downlink in reception slot")
			}
			if err != nil {
				genErr = true
				return nil, err
			}
			return dev, nil
		})
		if err != nil && !genErr {
			logger.WithError(err).Error("Failed to update device in registry")
			// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
			return
		}

		down, err := ns.scheduleDownlink(ctx, dev, up, acc, b, false)
		if err != nil {
			logger.WithError(err).Error("Failed to schedule downlink in reception slot")
			return
		}

		_, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID, func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
			stored.RecentDownlinks = append(stored.RecentDownlinks, down)
			if len(stored.RecentDownlinks) > recentDownlinkCount {
				stored.RecentDownlinks = append(stored.RecentDownlinks[:0], stored.RecentDownlinks[len(stored.RecentDownlinks)-recentDownlinkCount:]...)
			}
			return stored, nil
		})
		if err != nil {
			logger.WithError(err).Error("Failed to update device in registry")
			// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
		}
	})

	if !asOk {
		return nil
	}

	registerForwardUplink(ctx, dev, up)
	return asCl.Send(&ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		CorrelationIDs:       up.CorrelationIDs,
		Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
			FCnt:         dev.Session.NextFCntUp - 1,
			FPort:        pld.FPort,
			FRMPayload:   pld.FRMPayload,
			RxMetadata:   up.RxMetadata,
			SessionKeyID: dev.Session.SessionKeyID,
		}},
	})
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

	dev, err := ns.devices.GetByEUI(ctx, pld.JoinEUI, pld.DevEUI)
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
		SelectedMacVersion: dev.LoRaWANVersion, // Assume NS version is always higher than the version of the device
		RxDelay:            dev.MACState.DesiredParameters.Rx1Delay,
		CFList:             frequencyplans.CFList(*fp, dev.LoRaWANPHYVersion),
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: dev.MACState.DesiredParameters.Rx1DataRateOffset,
			Rx2DR:       dev.MACState.DesiredParameters.Rx2DataRateIndex,
			OptNeg:      true,
		},
	}

	var errs []error
	for _, js := range ns.joinServers {
		registerForwardUplink(ctx, dev, up)
		resp, err := js.HandleJoin(ctx, req, ns.WithClusterAuth())
		if err != nil {
			errs = append(errs, err)
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ns.deduplicationDone(ctx, up):
		}

		up.RxMetadata = acc.Accumulated()
		registerMergeMetadata(ctx, dev, up)

		var resetErr bool
		dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
			if dev.SessionFallback == nil {
				dev.SessionFallback = dev.Session
			}
			dev.Session = &ttnpb.Session{
				DevAddr:     devAddr,
				SessionKeys: resp.SessionKeys,
				StartedAt:   time.Now(),
			}
			dev.EndDeviceIdentifiers.DevAddr = &devAddr

			if err := resetMACState(ns.Component.FrequencyPlans, dev); err != nil {
				return nil, err
			}
			dev.MACState.CurrentParameters.Rx1Delay = req.RxDelay
			dev.MACState.CurrentParameters.Rx1DataRateOffset = req.DownlinkSettings.Rx1DROffset
			dev.MACState.CurrentParameters.Rx2DataRateIndex = req.DownlinkSettings.Rx2DR
			if req.DownlinkSettings.OptNeg && dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) > 0 {
				// The version will be further negotiated via RekeyInd/RekeyConf
				dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
			}

			dev.MACState.DesiredParameters.Rx1Delay = dev.MACState.CurrentParameters.Rx1Delay
			dev.MACState.DesiredParameters.Rx1DataRateOffset = dev.MACState.CurrentParameters.Rx1DataRateOffset
			dev.MACState.DesiredParameters.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex

			dev.RecentUplinks = append(dev.RecentUplinks, up)
			if len(dev.RecentUplinks) > recentUplinkCount {
				dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
			}
			return dev, nil
		})
		if err != nil && !resetErr {
			logger.WithError(err).Error("Failed to update device in registry")
			// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
		}
		if err != nil {
			return err
		}

		down, err := ns.scheduleDownlink(ctx, dev, up, nil, resp.RawPayload, true)
		if err != nil {
			logger.WithError(err).Debug("Failed to schedule join-accept")
			return err
		}

		dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID, func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
			stored.RecentDownlinks = append(stored.RecentDownlinks, down)
			if len(stored.RecentDownlinks) > recentDownlinkCount {
				stored.RecentDownlinks = append(stored.RecentDownlinks[:0], stored.RecentDownlinks[len(stored.RecentDownlinks)-recentDownlinkCount:]...)
			}
			return dev, nil
		})
		if err != nil {
			logger.WithError(err).Error("Failed to update device in registry")
			// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
			return err
		}

		uid := unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers)
		if uid == "" {
			return errUnknownApplicationID
		}

		go func() {
			// TODO: cluster.GetPeer(ctx, ttnpb.PeerInfo_APPLICATION_SERVER, dev.EndDeviceIdentifiers.ApplicationIdentifiers)
			// If not handled by cluster, try ns.applicationServers[uid].

			ns.applicationServersMu.RLock()
			cl, ok := ns.applicationServers[uid]
			ns.applicationServersMu.RUnlock()

			if !ok {
				return
			}

			if err := cl.Send(&ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       up.CorrelationIDs,
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
	return errJoin
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
func (ns *NetworkServer) scheduleDownlink(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.UplinkMessage, acc *metadataAccumulator, b []byte, isJoinAccept bool) (*ttnpb.DownlinkMessage, error) {
	if dev.MACState == nil {
		return nil, errUnknownMACState
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
	))

	down := &ttnpb.DownlinkMessage{
		RawPayload:   b,
		EndDeviceIDs: &dev.EndDeviceIdentifiers,
	}

	type tx struct {
		ttnpb.TxSettings
		Delay time.Duration
	}
	slots := make([]tx, 0, 2)

	fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return nil, errUnknownFrequencyPlan.WithCause(err)
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, errUnknownBand.WithCause(err)
	}

	var mds []*ttnpb.RxMetadata
	if up == nil {
		// Class C
		if len(dev.RecentUplinks) == 0 {
			return nil, errUplinkNotFound
		}
		mds = dev.RecentUplinks[len(dev.RecentUplinks)-1].RxMetadata
	} else {
		drIdx, err := band.Rx1DataRate(up.Settings.DataRateIndex, dev.MACState.CurrentParameters.Rx1DataRateOffset, dev.MACState.CurrentParameters.DownlinkDwellTime)
		if err != nil {
			return nil, err
		}

		chIdx, err := band.Rx1Channel(up.Settings.ChannelIndex)
		if err != nil {
			return nil, err
		}
		if uint(chIdx) >= uint(len(dev.MACState.CurrentParameters.Channels)) {
			return nil, errChannelIndexTooHigh
		}

		ch := dev.MACState.CurrentParameters.Channels[int(chIdx)]
		if ch == nil || ch.DownlinkFrequency == 0 {
			return nil, errUnknownChannel
		}

		rx1 := tx{
			TxSettings: ttnpb.TxSettings{
				DataRateIndex:      drIdx,
				CodingRate:         "4/5",
				InvertPolarization: true,
				ChannelIndex:       chIdx,
				Frequency:          ch.DownlinkFrequency,
				TxPower:            int32(band.DefaultMaxEIRP),
			},
		}
		if isJoinAccept {
			rx1.Delay = band.JoinAcceptDelay1
		} else {
			rx1.Delay = time.Second * time.Duration(dev.MACState.CurrentParameters.Rx1Delay)
		}

		if err = setDownlinkModulation(&rx1.TxSettings, band.DataRates[drIdx]); err != nil {
			return nil, err
		}

		mds = up.RxMetadata
		slots = append(slots, rx1)
	}

	if uint(dev.MACState.CurrentParameters.Rx2DataRateIndex) > uint(len(band.DataRates)) {
		return nil, errInvalidRx2DataRateIndex
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
	if isJoinAccept {
		rx2.Delay = band.JoinAcceptDelay2
	} else {
		rx2.Delay = time.Second * time.Duration(1+dev.MACState.CurrentParameters.Rx1Delay)
	}

	if err = setDownlinkModulation(&rx2.TxSettings, band.DataRates[dev.MACState.CurrentParameters.Rx2DataRateIndex]); err != nil {
		return nil, err
	}

	slots = append(slots, rx2)

	if acc != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
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
		down.Settings = s.TxSettings

		for _, md := range mds {
			logger := logger.WithField(
				"gateway_uid", unique.ID(ctx, md.GatewayIdentifiers),
			)

			cl, err := ns.gsClient(ctx, md.GatewayIdentifiers)
			if err != nil {
				logger.WithError(err).Debug("Could not get Gateway Server")
				continue
			}

			down.TxMetadata = ttnpb.TxMetadata{
				GatewayIdentifiers: md.GatewayIdentifiers,
				Timestamp:          md.Timestamp + uint64(s.Delay.Nanoseconds()),
			}

			_, err = cl.ScheduleDownlink(ctx, down, ns.WithClusterAuth())
			if err != nil {
				errs = append(errs, err)
				continue
			}

			dev.RecentDownlinks = append(dev.RecentDownlinks, down)
			return down, nil
		}
	}

	for i, err := range errs {
		logger = logger.WithField(fmt.Sprintf("error_%d", i), err)
	}
	logger.Warn("all Gateway Servers failed to schedule the downlink")
	return nil, errSchedule
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
