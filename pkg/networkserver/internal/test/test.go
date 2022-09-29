// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

// Package test contains testing utilities usable by all subpackages of networkserver including itself.
package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	nstime "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/toa"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	rpclog.ReplaceGrpcLogger(log.Noop)
}

var timeMu sync.RWMutex

func SetMockClock(clock *test.MockClock) func() {
	timeMu.Lock()
	unsetNow := nstime.SetNow(clock.Now)
	unsetAfter := nstime.SetAfter(clock.After)
	return func() {
		unsetNow()
		unsetAfter()
		timeMu.Unlock()
	}
}

// CopyBytes returns a deep copy of []byte.
func CopyBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	return append([]byte{}, b...)
}

// CopyStrings returns a deep copy of []string.
func CopyStrings(ss []string) []string {
	if ss == nil {
		return nil
	}
	return append([]string{}, ss...)
}

func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

func Band(fpID string, phyVer ttnpb.PHYVersion) band.Band {
	return *internal.LoRaWANBands[test.FrequencyPlan(fpID).BandID][phyVer]
}

var (
	DefaultClassBCGatewayIdentifiers = [...]*ttnpb.ClassBCGatewayIdentifiers{
		{
			GatewayIds:   &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-0"},
			AntennaIndex: 3,
		},
		{
			GatewayIds:   &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-1"},
			AntennaIndex: 1,
		},
		{
			GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-2"},
		},
		{
			GatewayIds:   &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-3"},
			AntennaIndex: 2,
		},
		{
			GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-4"},
		},
	}

	DefaultTxSettings = &ttnpb.TxSettings{
		DataRate: &ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Lora{
				Lora: &ttnpb.LoRaDataRate{
					Bandwidth:       125000,
					SpreadingFactor: 7,
					CodingRate:      band.Cr4_5,
				},
			},
		},
	}

	DefaultRxMetadata = [...]*ttnpb.RxMetadata{
		{
			GatewayIds:             &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-1"},
			Snr:                    -9,
			ChannelRssi:            -99,
			UplinkToken:            []byte("token-gtw-1"),
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NONE,
		},
		{
			GatewayIds:             &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-3"},
			Snr:                    -5.3,
			ChannelRssi:            -95.3,
			UplinkToken:            []byte("token-gtw-3"),
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
		},
		{
			GatewayIds:             &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-5"},
			Snr:                    12,
			ChannelRssi:            -22,
			UplinkToken:            []byte("token-gtw-5"),
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER,
		},
		{
			GatewayIds:             &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-0"},
			Snr:                    5.2,
			ChannelRssi:            -15.2,
			UplinkToken:            []byte("token-gtw-0"),
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NONE,
		},
		{
			GatewayIds:             &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-2"},
			Snr:                    6.3,
			ChannelRssi:            -16.3,
			UplinkToken:            []byte("token-gtw-2"),
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
		},
		{
			GatewayIds:             &ttnpb.GatewayIdentifiers{GatewayId: "gateway-test-4"},
			Snr:                    -7,
			ChannelRssi:            -17,
			UplinkToken:            []byte("token-gtw-4"),
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
		},
	}
	PacketBrokerRxMetadata = [...]*ttnpb.RxMetadata{
		{
			GatewayIds:  &ttnpb.GatewayIdentifiers{GatewayId: cluster.PacketBrokerGatewayID.GatewayId},
			Snr:         4.2,
			ChannelRssi: -14.2,
			UplinkToken: []byte("token-pb-1"),
			PacketBroker: &ttnpb.PacketBrokerMetadata{
				ForwarderNetId:     test.DefaultNetID.Bytes(),
				ForwarderClusterId: "test-cluster",
			},
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER,
		},
		{
			GatewayIds:  &ttnpb.GatewayIdentifiers{GatewayId: cluster.PacketBrokerGatewayID.GatewayId},
			Snr:         1.8,
			ChannelRssi: -21.8,
			UplinkToken: []byte("token-pb-2"),
			PacketBroker: &ttnpb.PacketBrokerMetadata{
				ForwarderNetId:     test.DefaultNetID.Bytes(),
				ForwarderClusterId: "other-cluster",
			},
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER,
		},
	}

	DefaultApplicationDownlinkQueue = []*ttnpb.ApplicationDownlink{
		{
			CorrelationIds: []string{"correlation-app-down-1", "correlation-app-down-2"},
			FCnt:           0x22,
			FPort:          0x1,
			FrmPayload:     []byte("testPayload"),
			Priority:       ttnpb.TxSchedulePriority_HIGHEST,
			SessionKeyId:   []byte{0x11, 0x22, 0x33, 0x44},
		},
		{
			CorrelationIds: []string{"correlation-app-down-3", "correlation-app-down-4"},
			FCnt:           0x23,
			FPort:          0x1,
			FrmPayload:     []byte("testPayload"),
			Priority:       ttnpb.TxSchedulePriority_HIGHEST,
			SessionKeyId:   []byte{0x11, 0x22, 0x33, 0x44},
		},
	}
)

const (
	DefaultEU868JoinAcceptDelay = ttnpb.RxDelay_RX_DELAY_5
	DefaultEU868RX1Delay        = ttnpb.RxDelay_RX_DELAY_1
	DefaultEU868RX2Frequency    = 869525000
)

var DefaultEU868Channels = [...]*ttnpb.MACParameters_Channel{
	{
		UplinkFrequency:   868100000,
		DownlinkFrequency: 868100000,
		MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
		EnableUplink:      true,
	},
	{
		UplinkFrequency:   868300000,
		DownlinkFrequency: 868300000,
		MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
		EnableUplink:      true,
	},
	{
		UplinkFrequency:   868500000,
		DownlinkFrequency: 868500000,
		MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
		EnableUplink:      true,
	},
}

func MakeDefaultEU868CurrentChannels() []*ttnpb.MACParameters_Channel {
	return ttnpb.CloneSlice(DefaultEU868Channels[:])
}

func MakeDefaultEU868CurrentMACParameters(phyVersion ttnpb.PHYVersion) *ttnpb.MACParameters {
	return &ttnpb.MACParameters{
		AdrAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32},
		AdrAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64},
		AdrNbTrans:                 1,
		MaxDutyCycle:               ttnpb.AggregatedDutyCycle_DUTY_CYCLE_1,
		MaxEirp:                    16,
		PingSlotDataRateIndexValue: &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex_DATA_RATE_3},
		PingSlotFrequency:          869525000,
		BeaconFrequency:            869525000,
		RejoinCountPeriodicity:     ttnpb.RejoinCountExponent_REJOIN_COUNT_16,
		RejoinTimePeriodicity:      ttnpb.RejoinTimeExponent_REJOIN_TIME_0,
		Rx1Delay:                   DefaultEU868RX1Delay,
		Rx2DataRateIndex:           ttnpb.DataRateIndex_DATA_RATE_0,
		Rx2Frequency:               DefaultEU868RX2Frequency,
		Channels:                   MakeDefaultEU868CurrentChannels(),
	}
}

func MakeDefaultEU868DesiredChannels() []*ttnpb.MACParameters_Channel {
	return append(MakeDefaultEU868CurrentChannels(),
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867100000,
			DownlinkFrequency: 867100000,
			MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867300000,
			DownlinkFrequency: 867300000,
			MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867500000,
			DownlinkFrequency: 867500000,
			MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867700000,
			DownlinkFrequency: 867700000,
			MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867900000,
			DownlinkFrequency: 867900000,
			MinDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DataRateIndex_DATA_RATE_5,
			EnableUplink:      true,
		},
	)
}

func MakeDefaultEU868DesiredMACParameters(phyVersion ttnpb.PHYVersion) *ttnpb.MACParameters {
	params := MakeDefaultEU868CurrentMACParameters(phyVersion)
	params.Channels = MakeDefaultEU868DesiredChannels()
	return params
}

func MakeDefaultEU868MACState(class ttnpb.Class, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) *ttnpb.MACState {
	return &ttnpb.MACState{
		DeviceClass:       class,
		LorawanVersion:    macVersion,
		CurrentParameters: MakeDefaultEU868CurrentMACParameters(phyVersion),
		DesiredParameters: MakeDefaultEU868DesiredMACParameters(phyVersion),
	}
}

var DefaultUS915Channels = func() []*ttnpb.MACParameters_Channel {
	var chs []*ttnpb.MACParameters_Channel
	for i := 0; i < 64; i++ {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			UplinkFrequency:  uint64(902300000 + 200000*i),
			MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
			EnableUplink:     true,
		})
	}
	for i := 0; i < 8; i++ {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			UplinkFrequency:  uint64(903000000 + 1600000*i),
			MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
			MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
			EnableUplink:     true,
		})
	}
	for i := 0; i < 72; i++ {
		chs[i].DownlinkFrequency = uint64(923300000 + 600000*(i%8))
	}
	return chs
}()

func MakeDefaultUS915CurrentChannels() []*ttnpb.MACParameters_Channel {
	return ttnpb.CloneSlice(DefaultUS915Channels[:])
}

func MakeDefaultUS915CurrentMACParameters(ver ttnpb.PHYVersion) *ttnpb.MACParameters {
	return &ttnpb.MACParameters{
		AdrAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32},
		AdrAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64},
		AdrNbTrans:                 1,
		MaxDutyCycle:               ttnpb.AggregatedDutyCycle_DUTY_CYCLE_1,
		MaxEirp:                    30,
		PingSlotDataRateIndexValue: &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex_DATA_RATE_8},
		RejoinCountPeriodicity:     ttnpb.RejoinCountExponent_REJOIN_COUNT_16,
		RejoinTimePeriodicity:      ttnpb.RejoinTimeExponent_REJOIN_TIME_0,
		Rx1Delay:                   ttnpb.RxDelay_RX_DELAY_1,
		Rx2DataRateIndex:           ttnpb.DataRateIndex_DATA_RATE_8,
		Rx2Frequency:               923300000,
		Channels:                   MakeDefaultUS915CurrentChannels(),
	}
}

func MakeDefaultUS915FSB2DesiredChannels() []*ttnpb.MACParameters_Channel {
	chs := MakeDefaultUS915CurrentChannels()
	for _, ch := range chs {
		switch ch.UplinkFrequency {
		case 903900000,
			904100000,
			904300000,
			904500000,
			904700000,
			904900000,
			905100000,
			905300000:
		default:
			ch.EnableUplink = false
		}
	}
	return chs
}

func MakeDefaultUS915FSB2DesiredMACParameters(ver ttnpb.PHYVersion) *ttnpb.MACParameters {
	params := MakeDefaultUS915CurrentMACParameters(ver)
	params.Channels = MakeDefaultUS915FSB2DesiredChannels()
	return params
}

func MakeDefaultUS915FSB2MACState(class ttnpb.Class, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) *ttnpb.MACState {
	return &ttnpb.MACState{
		DeviceClass:       class,
		LorawanVersion:    macVersion,
		CurrentParameters: MakeDefaultUS915CurrentMACParameters(phyVersion),
		DesiredParameters: MakeDefaultUS915FSB2DesiredMACParameters(phyVersion),
	}
}

// MakeUplinkSettings builds the ttnpb.TxSettings for an uplink.
func MakeUplinkSettings(dr *ttnpb.DataRate, _ ttnpb.DataRateIndex, freq uint64) *ttnpb.TxSettings {
	return &ttnpb.TxSettings{
		DataRate:  ttnpb.Clone(dr),
		EnableCrc: true,
		Frequency: freq,
		Timestamp: 42,
	}
}

type UplinkMessageConfig struct {
	RawPayload     []byte
	Payload        *ttnpb.Message
	DataRate       *ttnpb.DataRate
	DataRateIndex  ttnpb.DataRateIndex
	Frequency      uint64
	ChannelIndex   uint8
	ReceivedAt     time.Time
	RxMetadata     []*ttnpb.RxMetadata
	CorrelationIDs []string
}

func MakeUplinkMessage(conf UplinkMessageConfig) *ttnpb.UplinkMessage {
	settings := MakeUplinkSettings(conf.DataRate, conf.DataRateIndex, conf.Frequency)
	return &ttnpb.UplinkMessage{
		RawPayload:         conf.RawPayload,
		Payload:            conf.Payload,
		Settings:           settings,
		RxMetadata:         ttnpb.CloneSlice(conf.RxMetadata),
		ReceivedAt:         timestamppb.New(conf.ReceivedAt),
		CorrelationIds:     CopyStrings(conf.CorrelationIDs),
		DeviceChannelIndex: uint32(conf.ChannelIndex),
		ConsumedAirtime: durationpb.New(
			test.Must(toa.Compute(len(conf.RawPayload), settings)).(time.Duration),
		),
	}
}

var DataUplinkCorrelationIDs = [...]string{
	"data-uplink-correlation-id-1",
	"data-uplink-correlation-id-2",
	"data-uplink-correlation-id-3",
}

type MACCommander interface {
	MACCommand() *ttnpb.MACCommand
}

func AppendMACCommanders(queue []*ttnpb.MACCommand, cmds ...MACCommander) []*ttnpb.MACCommand {
	for _, cmd := range cmds {
		queue = append(queue, cmd.MACCommand())
	}
	return queue
}

func MakeUplinkMACBuffer(phy *band.Band, cmds ...MACCommander) []byte {
	var b []byte
	for _, cmd := range cmds {
		b = test.Must(lorawan.DefaultMACCommands.AppendUplink(*phy, b, cmd.MACCommand())).([]byte)
	}
	return b
}

func MakeDownlinkMACBuffer(phy *band.Band, cmds ...MACCommander) []byte {
	var b []byte
	for _, cmd := range cmds {
		b = test.Must(lorawan.DefaultMACCommands.AppendDownlink(*phy, b, cmd.MACCommand())).([]byte)
	}
	return b
}

var SessionKeysOptions = test.SessionKeysOptions

func MakeSessionKeys(macVersion ttnpb.MACVersion, wrapKeys, withID bool, opts ...test.SessionKeysOption) *ttnpb.SessionKeys {
	defaultKeyOpt := SessionKeysOptions.WithDefaultNwkKeys
	if wrapKeys {
		defaultKeyOpt = SessionKeysOptions.WithDefaultNwkKeysWrapped
	}
	var id []byte
	if withID {
		id = test.DefaultSessionKeyID
	}
	return test.MakeSessionKeys(
		defaultKeyOpt(macVersion),
		SessionKeysOptions.WithSessionKeyId(id),
		SessionKeysOptions.Compose(opts...),
	)
}

func messageGenerationKeys(sk *ttnpb.SessionKeys, macVersion ttnpb.MACVersion) ttnpb.SessionKeys {
	if sk == nil {
		return *MakeSessionKeys(macVersion, false, false)
	}
	decrypt := func(ke *ttnpb.KeyEnvelope) []byte {
		switch {
		case ke == nil:
			return nil
		case len(ke.Key) > 0:
			return types.MustAES128Key(ke.Key).Bytes()
		case len(ke.EncryptedKey) > 0:
			k := &types.AES128Key{}
			test.Must(nil, k.UnmarshalBinary(ke.EncryptedKey))
			return k.Bytes()
		default:
			return nil
		}
	}
	return ttnpb.SessionKeys{
		SessionKeyId: sk.SessionKeyId,
		FNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: decrypt(sk.FNwkSIntKey),
		},
		SNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: decrypt(sk.SNwkSIntKey),
		},
		NwkSEncKey: &ttnpb.KeyEnvelope{
			Key: decrypt(sk.NwkSEncKey),
		},
	}
}

func MustEncryptUplink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, encOpts []crypto.EncryptionOption, b ...byte) []byte {
	return test.Must(crypto.EncryptUplink(key, devAddr, fCnt, b, encOpts...)).([]byte)
}

func MustComputeUplinkCMACF(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, b ...byte) [4]byte {
	return test.Must(crypto.ComputeLegacyUplinkMIC(key, devAddr, fCnt, b)).([4]byte)
}

type DataUplinkConfig struct {
	DecodePayload bool
	Matched       bool

	Confirmed      bool
	MACVersion     ttnpb.MACVersion
	DevAddr        types.DevAddr
	FCtrl          *ttnpb.FCtrl
	FCnt           uint32
	ConfFCntDown   uint32
	FPort          uint8
	FRMPayload     []byte
	FOpts          []byte
	DataRate       *ttnpb.DataRate
	DataRateIndex  ttnpb.DataRateIndex
	Frequency      uint64
	ChannelIndex   uint8
	RxMetadata     []*ttnpb.RxMetadata
	CorrelationIDs []string
	ReceivedAt     time.Time

	SessionKeys *ttnpb.SessionKeys
}

func WithDeviceDataUplinkConfig(dev *ttnpb.EndDevice, pending bool, drIdx ttnpb.DataRateIndex, chIdx uint8, fCntDelta uint32) func(DataUplinkConfig) DataUplinkConfig {
	session := dev.Session
	macState := dev.MacState
	if pending {
		session = dev.PendingSession
		macState = dev.PendingMacState
	}
	return func(conf DataUplinkConfig) DataUplinkConfig {
		conf.MACVersion = macState.LorawanVersion
		conf.DevAddr = types.MustDevAddr(session.DevAddr).OrZero()
		conf.FCnt = session.LastFCntUp + fCntDelta
		conf.DataRate = internal.LoRaWANBands[test.FrequencyPlan(dev.FrequencyPlanId).BandID][dev.LorawanPhyVersion].DataRates[drIdx].Rate
		conf.DataRateIndex = drIdx
		conf.Frequency = macState.CurrentParameters.Channels[chIdx].UplinkFrequency
		conf.ChannelIndex = chIdx
		conf.SessionKeys = session.Keys
		return conf
	}
}

func MakeDataUplink(conf DataUplinkConfig) *ttnpb.UplinkMessage {
	if conf.FCtrl == nil {
		conf.FCtrl = &ttnpb.FCtrl{}
	}
	if !conf.FCtrl.Ack && conf.ConfFCntDown > 0 {
		panic("ConfFCntDown must be zero for uplink frames with ACK bit unset")
	}
	devAddr := *conf.DevAddr.Copy(&types.DevAddr{})
	keys := messageGenerationKeys(conf.SessionKeys, conf.MACVersion)
	frmPayload := conf.FRMPayload
	fOpts := conf.FOpts
	if len(frmPayload) > 0 && conf.FPort == 0 {
		frmPayload = MustEncryptUplink(*types.MustAES128Key(keys.NwkSEncKey.Key), devAddr, conf.FCnt, nil, frmPayload...)
	} else if len(fOpts) > 0 && macspec.EncryptFOpts(conf.MACVersion) {
		encOpts := macspec.EncryptionOptions(conf.MACVersion, macspec.UplinkFrame, uint32(conf.FPort), true)
		fOpts = MustEncryptUplink(*types.MustAES128Key(keys.NwkSEncKey.Key), devAddr, conf.FCnt, encOpts, fOpts...)
	}
	mType := ttnpb.MType_UNCONFIRMED_UP
	if conf.Confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}
	mhdr := &ttnpb.MHDR{
		MType: mType,
		Major: ttnpb.Major_LORAWAN_R1,
	}
	fhdr := &ttnpb.FHDR{
		DevAddr: devAddr.Bytes(),
		FCtrl:   conf.FCtrl,
		FCnt:    conf.FCnt & 0xffff,
		FOpts:   CopyBytes(fOpts),
	}
	phyPayload := test.Must(lorawan.MarshalMessage(&ttnpb.Message{
		MHdr: mhdr,
		Payload: &ttnpb.Message_MacPayload{
			MacPayload: &ttnpb.MACPayload{
				FHdr:       fhdr,
				FPort:      uint32(conf.FPort),
				FrmPayload: frmPayload,
			},
		},
	})).([]byte)
	var mic [4]byte
	switch {
	case macspec.UseLegacyMIC(conf.MACVersion):
		mic = test.Must(
			crypto.ComputeLegacyUplinkMIC(*types.MustAES128Key(keys.FNwkSIntKey.Key), devAddr, conf.FCnt, phyPayload),
		).([4]byte)
	default:
		mic = test.Must(
			crypto.ComputeUplinkMIC(*types.MustAES128Key(keys.SNwkSIntKey.Key), *types.MustAES128Key(keys.FNwkSIntKey.Key),
				conf.ConfFCntDown, uint8(conf.DataRateIndex), conf.ChannelIndex, devAddr, conf.FCnt, phyPayload),
		).([4]byte)
	}

	phyPayload = append(phyPayload, mic[:]...)
	return MakeUplinkMessage(UplinkMessageConfig{
		RawPayload: phyPayload,
		Payload: func() *ttnpb.Message {
			if conf.DecodePayload {
				return &ttnpb.Message{
					MHdr: mhdr,
					Mic:  phyPayload[len(phyPayload)-4:],
					Payload: &ttnpb.Message_MacPayload{
						MacPayload: &ttnpb.MACPayload{
							FHdr:       fhdr,
							FPort:      uint32(conf.FPort),
							FrmPayload: CopyBytes(frmPayload),
							FullFCnt:   conf.FCnt,
						},
					},
				}
			}
			return nil
		}(),
		DataRate: conf.DataRate,
		DataRateIndex: func() ttnpb.DataRateIndex {
			if conf.Matched {
				return conf.DataRateIndex
			}
			return 0
		}(),
		Frequency: conf.Frequency,
		ChannelIndex: func() uint8 {
			if conf.Matched {
				return conf.ChannelIndex
			}
			return 0
		}(),
		ReceivedAt: conf.ReceivedAt,
		RxMetadata: conf.RxMetadata,
		CorrelationIDs: func() []string {
			if len(conf.CorrelationIDs) == 0 {
				return DataUplinkCorrelationIDs[:]
			}
			return conf.CorrelationIDs
		}(),
	})
}

func MustEncryptDownlink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, encOpts []crypto.EncryptionOption, b ...byte) []byte {
	return test.Must(crypto.EncryptDownlink(key, devAddr, fCnt, b, encOpts...)).([]byte)
}

type DataDownlinkConfig struct {
	DecodePayload bool

	Confirmed  bool
	MACVersion ttnpb.MACVersion
	DevAddr    types.DevAddr
	FCtrl      *ttnpb.FCtrl
	FCnt       uint32
	ConfFCntUp uint32
	FPort      uint8
	FRMPayload []byte
	FOpts      []byte

	Request *ttnpb.TxRequest

	SessionKeys *ttnpb.SessionKeys
}

func MakeDataDownlink(conf *DataDownlinkConfig) *ttnpb.DownlinkMessage {
	if conf.FCtrl == nil {
		conf.FCtrl = &ttnpb.FCtrl{}
	}
	if !conf.FCtrl.Ack && conf.ConfFCntUp > 0 {
		panic("ConfFCntDown must be zero for uplink frames with ACK bit unset")
	}
	devAddr := *conf.DevAddr.Copy(&types.DevAddr{})
	keys := messageGenerationKeys(conf.SessionKeys, conf.MACVersion)
	frmPayload := conf.FRMPayload
	fOpts := conf.FOpts
	if len(frmPayload) > 0 && conf.FPort == 0 {
		frmPayload = MustEncryptDownlink(*types.MustAES128Key(keys.NwkSEncKey.Key), devAddr, conf.FCnt, nil, frmPayload...)
	} else if len(fOpts) > 0 && macspec.EncryptFOpts(conf.MACVersion) {
		encOpts := macspec.EncryptionOptions(conf.MACVersion, macspec.DownlinkFrame, uint32(conf.FPort), true)
		fOpts = MustEncryptDownlink(*types.MustAES128Key(keys.NwkSEncKey.Key), devAddr, conf.FCnt, encOpts, fOpts...)
	}
	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if conf.Confirmed {
		mType = ttnpb.MType_CONFIRMED_DOWN
	}
	msg := &ttnpb.Message{
		MHdr: &ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MacPayload{
			MacPayload: &ttnpb.MACPayload{
				FHdr: &ttnpb.FHDR{
					DevAddr: devAddr.Bytes(),
					FCtrl:   conf.FCtrl,
					FCnt:    conf.FCnt & 0xffff,
					FOpts:   fOpts,
				},
				FullFCnt:   conf.FCnt,
				FPort:      uint32(conf.FPort),
				FrmPayload: frmPayload,
			},
		},
	}
	phyPayload := test.Must(lorawan.MarshalMessage(msg)).([]byte)
	var mic [4]byte
	switch {
	case macspec.UseLegacyMIC(conf.MACVersion):
		mic = test.Must(
			crypto.ComputeLegacyDownlinkMIC(*types.MustAES128Key(keys.FNwkSIntKey.Key), devAddr, conf.FCnt, phyPayload),
		).([4]byte)
	default:
		mic = test.Must(
			crypto.ComputeDownlinkMIC(*types.MustAES128Key(keys.SNwkSIntKey.Key),
				devAddr, conf.ConfFCntUp, conf.FCnt, phyPayload),
		).([4]byte)
	}
	msg.Mic = mic[:]
	return &ttnpb.DownlinkMessage{
		Settings: &ttnpb.DownlinkMessage_Request{
			Request: ttnpb.Clone(conf.Request),
		},
		RawPayload: append(phyPayload, mic[:]...),
		Payload: func() *ttnpb.Message {
			if !conf.DecodePayload {
				return nil
			}
			return msg
		}(),
	}
}

func MakeTestCaseName(parts ...string) string {
	return strings.Join(parts, "/")
}

func ForEachBand(tb testing.TB, f func(func(...string) string, *band.Band, ttnpb.PHYVersion)) {
	for phyID, phyVersions := range internal.LoRaWANBands {
		switch phyID {
		case band.EU_863_870, band.US_902_928:
		case band.AS_923:
			if !testing.Short() {
				break
			}
			fallthrough
		default:
			tb.Logf("Skip %s band", phyID)
			continue
		}
		for phyVersion, b := range phyVersions {
			switch phyVersion {
			case ttnpb.PHYVersion_RP001_V1_0_3_REV_A, ttnpb.PHYVersion_RP001_V1_1_REV_B:
			case ttnpb.PHYVersion_RP001_V1_0_2_REV_B:
				if !testing.Short() {
					break
				}
				fallthrough
			default:
				tb.Logf("Skip %s version of %s band", phyVersion, phyID)
				continue
			}
			f(func(parts ...string) string {
				return MakeTestCaseName(append(parts, phyID, fmt.Sprintf("PHY:%s", phyVersion.String()))...)
			}, b, phyVersion)
		}
	}
}

func ForEachMACVersion(tb testing.TB, f func(func(...string) string, ttnpb.MACVersion)) {
	for _, macVersion := range []ttnpb.MACVersion{
		ttnpb.MACVersion_MAC_V1_0,
		ttnpb.MACVersion_MAC_V1_0_1,
		ttnpb.MACVersion_MAC_V1_0_2,
		ttnpb.MACVersion_MAC_V1_0_3,
		ttnpb.MACVersion_MAC_V1_0_4,
		ttnpb.MACVersion_MAC_V1_1,
	} {
		switch macVersion {
		case ttnpb.MACVersion_MAC_V1_0_4, ttnpb.MACVersion_MAC_V1_1:
		case ttnpb.MACVersion_MAC_V1_0_3:
			if !testing.Short() {
				break
			}
			fallthrough
		default:
			tb.Logf("Skip MAC version %s", macVersion)
			continue
		}
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("MAC:%s", macVersion.String()))...)
		}, macVersion)
	}
}

func ForEachClass(tb testing.TB, f func(func(...string) string, ttnpb.Class)) {
	for _, class := range []ttnpb.Class{
		ttnpb.Class_CLASS_A,
		ttnpb.Class_CLASS_B,
		ttnpb.Class_CLASS_C,
	} {
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("Class:%s", class.String()))...)
		}, class)
	}
}

func ForEachFrequencyPlan(tb testing.TB, f func(func(...string) string, string, *frequencyplans.FrequencyPlan)) {
	fpIDs, err := frequencyplans.NewStore(test.FrequencyPlansFetcher).GetAllIDs()
	if err != nil {
		tb.Errorf("failed to get frequency plans: %s", err)
		return
	}
	for _, fpID := range fpIDs {
		switch fpID {
		case test.EUFrequencyPlanID, test.USFrequencyPlanID:
		case test.ASAUFrequencyPlanID:
			if !testing.Short() {
				break
			}
			fallthrough
		default:
			tb.Logf("Skip frequency plan %s", fpID)
			continue
		}
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("FP:%s", fpID))...)
		}, fpID, test.FrequencyPlan(fpID))
	}
}

func ForEachLoRaWANVersionPair(tb testing.TB, f func(func(...string) string, ttnpb.MACVersion, ttnpb.PHYVersion)) {
	for macVersion, phyVersions := range internal.LoRaWANVersionPairs {
		switch macVersion {
		case ttnpb.MACVersion_MAC_V1_0_3, ttnpb.MACVersion_MAC_V1_1:
		case ttnpb.MACVersion_MAC_V1_0_2:
			if !testing.Short() {
				break
			}
			fallthrough
		default:
			tb.Logf("Skip MAC version %s", macVersion)
			continue
		}
		for phyVersion := range phyVersions {
			f(func(parts ...string) string {
				return MakeTestCaseName(append(parts, fmt.Sprintf("MAC:%s", macVersion.String()), fmt.Sprintf("PHY:%s", phyVersion.String()))...)
			}, macVersion, phyVersion)
		}
	}
}

func ForEachClassLoRaWANVersionPair(tb testing.TB, f func(func(...string) string, ttnpb.Class, ttnpb.MACVersion, ttnpb.PHYVersion)) {
	ForEachClass(tb, func(makeClassName func(...string) string, class ttnpb.Class) {
		ForEachLoRaWANVersionPair(tb, func(makeLoRaWANName func(parts ...string) string, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) {
			f(func(parts ...string) string {
				return makeClassName(makeLoRaWANName(parts...))
			}, class, macVersion, phyVersion)
		})
	})
}

func ForEachClassMACVersion(tb testing.TB, f func(func(...string) string, ttnpb.Class, ttnpb.MACVersion)) {
	ForEachClass(tb, func(makeClassName func(...string) string, class ttnpb.Class) {
		ForEachMACVersion(tb, func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
			f(func(parts ...string) string {
				return makeClassName(makeMACName(parts...))
			}, class, macVersion)
		})
	})
}

func ForEachFrequencyPlanLoRaWANVersionPair(tb testing.TB, f func(func(...string) string, string, *frequencyplans.FrequencyPlan, *band.Band, ttnpb.MACVersion, ttnpb.PHYVersion)) {
	ForEachFrequencyPlan(tb, func(makeFPName func(...string) string, fpID string, fp *frequencyplans.FrequencyPlan) {
		ForEachLoRaWANVersionPair(tb, func(makeLoRaWANName func(parts ...string) string, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) {
			b, ok := internal.LoRaWANBands[fp.BandID][phyVersion]
			if !ok || b == nil {
				return
			}
			f(func(parts ...string) string {
				return makeFPName(makeLoRaWANName(parts...))
			}, fpID, fp, b, macVersion, phyVersion)
		})
	})
}

func ForEachBandMACVersion(tb testing.TB, f func(func(...string) string, *band.Band, ttnpb.PHYVersion, ttnpb.MACVersion)) {
	ForEachBand(tb, func(makeBandName func(...string) string, phy *band.Band, phyVersion ttnpb.PHYVersion) {
		ForEachMACVersion(tb, func(makeMACName func(...string) string, macVersion ttnpb.MACVersion) {
			if _, ok := internal.LoRaWANVersionPairs[macVersion][phyVersion]; !ok {
				return
			}
			f(func(parts ...string) string {
				return makeBandName(makeMACName(parts...))
			}, phy, phyVersion, macVersion)
		})
	})
}
