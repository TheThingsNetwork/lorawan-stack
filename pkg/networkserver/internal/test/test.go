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

	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	nstime "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
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

// CopySession returns a deep copy of *ttnpb.Session pb.
func CopySession(pb *ttnpb.Session) *ttnpb.Session {
	return deepcopy.Copy(pb).(*ttnpb.Session)
}

// CopyMessage returns a deep copy of *ttnpb.Message pb.
func CopyMessage(pb *ttnpb.Message) *ttnpb.Message {
	return deepcopy.Copy(pb).(*ttnpb.Message)
}

// CopyUplinkMessages returns a deep copy of ...*ttnpb.UplinkMessage pbs.
func CopyUplinkMessages(pbs ...*ttnpb.UplinkMessage) []*ttnpb.UplinkMessage {
	return deepcopy.Copy(pbs).([]*ttnpb.UplinkMessage)
}

// CopyDownlinkMessage returns a deep copy of ttnpb.DownlinkMessage pb.
func CopyDownlinkMessage(pb *ttnpb.DownlinkMessage) *ttnpb.DownlinkMessage {
	return deepcopy.Copy(pb).(*ttnpb.DownlinkMessage)
}

// CopyDownlinkMessages returns a deep copy of ...*ttnpb.DownlinkMessage pbs.
func CopyDownlinkMessages(pbs ...*ttnpb.DownlinkMessage) []*ttnpb.DownlinkMessage {
	return deepcopy.Copy(pbs).([]*ttnpb.DownlinkMessage)
}

func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

func AES128KeyPtr(key types.AES128Key) *types.AES128Key {
	return &key
}

var (
	DefaultGatewayAntennaIdentifiers = [...]ttnpb.GatewayAntennaIdentifiers{
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-0"},
			AntennaIndex:       3,
		},
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-1"},
			AntennaIndex:       1,
		},
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-2"},
		},
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-3"},
			AntennaIndex:       2,
		},
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-4"},
		},
	}

	DefaultRxMetadata = [...]*ttnpb.RxMetadata{
		{
			GatewayIdentifiers:     ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-1"},
			SNR:                    -9,
			UplinkToken:            []byte("token-gtw-1"),
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
		},
		{
			GatewayIdentifiers:     ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-3"},
			SNR:                    -5.3,
			UplinkToken:            []byte("token-gtw-3"),
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
		},
		{
			GatewayIdentifiers:     ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-5"},
			SNR:                    12,
			UplinkToken:            []byte("token-gtw-5"),
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER,
		},
		{
			GatewayIdentifiers:     ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-0"},
			SNR:                    5.2,
			UplinkToken:            []byte("token-gtw-0"),
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
		},
		{
			GatewayIdentifiers:     ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-2"},
			SNR:                    6.3,
			UplinkToken:            []byte("token-gtw-2"),
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
		},
		{
			GatewayIdentifiers:     ttnpb.GatewayIdentifiers{GatewayID: "gateway-test-4"},
			SNR:                    -7,
			UplinkToken:            []byte("token-gtw-4"),
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER,
		},
	}

	DefaultApplicationDownlinkQueue = []*ttnpb.ApplicationDownlink{
		{
			CorrelationIDs: []string{"correlation-app-down-1", "correlation-app-down-2"},
			FCnt:           0x22,
			FPort:          0x1,
			FRMPayload:     []byte("testPayload"),
			Priority:       ttnpb.TxSchedulePriority_HIGHEST,
			SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
		},
		{
			CorrelationIDs: []string{"correlation-app-down-3", "correlation-app-down-4"},
			FCnt:           0x23,
			FPort:          0x1,
			FRMPayload:     []byte("testPayload"),
			Priority:       ttnpb.TxSchedulePriority_HIGHEST,
			SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44},
		},
	}
)

const (
	DefaultEU868JoinAcceptDelay = ttnpb.RX_DELAY_5
	DefaultEU868RX1Delay        = ttnpb.RX_DELAY_1
	DefaultEU868RX2Frequency    = 869525000
)

var DefaultEU868Channels = [...]*ttnpb.MACParameters_Channel{
	{
		UplinkFrequency:   868100000,
		DownlinkFrequency: 868100000,
		MinDataRateIndex:  ttnpb.DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DATA_RATE_5,
		EnableUplink:      true,
	},
	{
		UplinkFrequency:   868300000,
		DownlinkFrequency: 868300000,
		MinDataRateIndex:  ttnpb.DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DATA_RATE_5,
		EnableUplink:      true,
	},
	{
		UplinkFrequency:   868500000,
		DownlinkFrequency: 868500000,
		MinDataRateIndex:  ttnpb.DATA_RATE_0,
		MaxDataRateIndex:  ttnpb.DATA_RATE_5,
		EnableUplink:      true,
	},
}

func MakeDefaultEU868CurrentChannels() []*ttnpb.MACParameters_Channel {
	return deepcopy.Copy(DefaultEU868Channels[:]).([]*ttnpb.MACParameters_Channel)
}

func MakeDefaultEU868CurrentMACParameters(phyVersion ttnpb.PHYVersion) ttnpb.MACParameters {
	return ttnpb.MACParameters{
		ADRAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADR_ACK_DELAY_32},
		ADRAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADR_ACK_LIMIT_64},
		ADRNbTrans:                 1,
		MaxDutyCycle:               ttnpb.DUTY_CYCLE_1,
		MaxEIRP:                    16,
		PingSlotDataRateIndexValue: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_3},
		PingSlotFrequency:          869525000,
		RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
		RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
		Rx1Delay:                   DefaultEU868RX1Delay,
		Rx2DataRateIndex:           ttnpb.DATA_RATE_0,
		Rx2Frequency:               DefaultEU868RX2Frequency,
		Channels:                   MakeDefaultEU868CurrentChannels(),
	}
}

func MakeDefaultEU868DesiredChannels() []*ttnpb.MACParameters_Channel {
	return append(MakeDefaultEU868CurrentChannels(),
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867100000,
			DownlinkFrequency: 867100000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867300000,
			DownlinkFrequency: 867300000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867500000,
			DownlinkFrequency: 867500000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867700000,
			DownlinkFrequency: 867700000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
			EnableUplink:      true,
		},
		&ttnpb.MACParameters_Channel{
			UplinkFrequency:   867900000,
			DownlinkFrequency: 867900000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
			EnableUplink:      true,
		},
	)
}

func MakeDefaultEU868DesiredMACParameters(phyVersion ttnpb.PHYVersion) ttnpb.MACParameters {
	params := MakeDefaultEU868CurrentMACParameters(phyVersion)
	params.Channels = MakeDefaultEU868DesiredChannels()
	return params
}

func MakeDefaultEU868MACState(class ttnpb.Class, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) *ttnpb.MACState {
	return &ttnpb.MACState{
		DeviceClass:       class,
		LoRaWANVersion:    macVersion,
		CurrentParameters: MakeDefaultEU868CurrentMACParameters(phyVersion),
		DesiredParameters: MakeDefaultEU868DesiredMACParameters(phyVersion),
	}
}

var DefaultUS915Channels = func() []*ttnpb.MACParameters_Channel {
	var chs []*ttnpb.MACParameters_Channel
	for i := 0; i < 64; i++ {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			UplinkFrequency:  uint64(902300000 + 200000*i),
			MinDataRateIndex: ttnpb.DATA_RATE_0,
			MaxDataRateIndex: ttnpb.DATA_RATE_3,
			EnableUplink:     true,
		})
	}
	for i := 0; i < 8; i++ {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			UplinkFrequency:  uint64(903000000 + 1600000*i),
			MinDataRateIndex: ttnpb.DATA_RATE_4,
			MaxDataRateIndex: ttnpb.DATA_RATE_4,
			EnableUplink:     true,
		})
	}
	for i := 0; i < 72; i++ {
		chs[i].DownlinkFrequency = uint64(923300000 + 600000*(i%8))
	}
	return chs
}()

func MakeDefaultUS915CurrentChannels() []*ttnpb.MACParameters_Channel {
	return deepcopy.Copy(DefaultUS915Channels[:]).([]*ttnpb.MACParameters_Channel)
}

func MakeDefaultUS915CurrentMACParameters(ver ttnpb.PHYVersion) ttnpb.MACParameters {
	return ttnpb.MACParameters{
		ADRAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADR_ACK_DELAY_32},
		ADRAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADR_ACK_LIMIT_64},
		ADRNbTrans:                 1,
		MaxDutyCycle:               ttnpb.DUTY_CYCLE_1,
		MaxEIRP:                    30,
		PingSlotDataRateIndexValue: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_8},
		RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
		RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
		Rx1Delay:                   ttnpb.RX_DELAY_1,
		Rx2DataRateIndex:           ttnpb.DATA_RATE_8,
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

func MakeDefaultUS915FSB2DesiredMACParameters(ver ttnpb.PHYVersion) ttnpb.MACParameters {
	params := MakeDefaultUS915CurrentMACParameters(ver)
	params.Channels = MakeDefaultUS915FSB2DesiredChannels()
	return params
}

func MakeDefaultUS915FSB2MACState(class ttnpb.Class, macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion) *ttnpb.MACState {
	return &ttnpb.MACState{
		DeviceClass:       class,
		LoRaWANVersion:    macVersion,
		CurrentParameters: MakeDefaultUS915CurrentMACParameters(phyVersion),
		DesiredParameters: MakeDefaultUS915FSB2DesiredMACParameters(phyVersion),
	}
}

func MakeUplinkSettings(dr ttnpb.DataRate, drIdx ttnpb.DataRateIndex, freq uint64) ttnpb.TxSettings {
	return ttnpb.TxSettings{
		DataRate:      *deepcopy.Copy(&dr).(*ttnpb.DataRate),
		DataRateIndex: drIdx,
		EnableCRC:     true,
		Frequency:     freq,
		Timestamp:     42,
	}
}

type UplinkMessageConfig struct {
	RawPayload     []byte
	Payload        *ttnpb.Message
	DataRate       ttnpb.DataRate
	DataRateIndex  ttnpb.DataRateIndex
	Frequency      uint64
	ChannelIndex   uint8
	ReceivedAt     time.Time
	RxMetadata     []*ttnpb.RxMetadata
	CorrelationIDs []string
}

func MakeUplinkMessage(conf UplinkMessageConfig) *ttnpb.UplinkMessage {
	return &ttnpb.UplinkMessage{
		RawPayload:         conf.RawPayload,
		Payload:            conf.Payload,
		Settings:           MakeUplinkSettings(conf.DataRate, conf.DataRateIndex, conf.Frequency),
		RxMetadata:         deepcopy.Copy(conf.RxMetadata).([]*ttnpb.RxMetadata),
		ReceivedAt:         conf.ReceivedAt,
		CorrelationIDs:     CopyStrings(conf.CorrelationIDs),
		DeviceChannelIndex: uint32(conf.ChannelIndex),
	}
}

func WithMatchedUplinkSettings(msg *ttnpb.UplinkMessage, chIdx uint8, drIdx ttnpb.DataRateIndex) *ttnpb.UplinkMessage {
	msg = CopyUplinkMessage(msg)
	msg.Settings.DataRateIndex = drIdx
	msg.DeviceChannelIndex = uint32(chIdx)
	return msg
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
		b = test.Must(lorawan.DefaultMACCommands.AppendUplink(*phy, b, *cmd.MACCommand())).([]byte)
	}
	return b
}

func MakeDownlinkMACBuffer(phy *band.Band, cmds ...MACCommander) []byte {
	var b []byte
	for _, cmd := range cmds {
		b = test.Must(lorawan.DefaultMACCommands.AppendDownlink(*phy, b, *cmd.MACCommand())).([]byte)
	}
	return b
}

var SessionKeysOptions = test.SessionKeysOptions

func MakeSessionKeys(macVersion ttnpb.MACVersion, wrapKeys bool, opts ...test.SessionKeysOption) *ttnpb.SessionKeys {
	defaultKeyOpt := SessionKeysOptions.WithDefaultNwkKeys
	if wrapKeys {
		defaultKeyOpt = SessionKeysOptions.WithDefaultNwkKeysWrapped
	}
	return test.MakeSessionKeys(
		defaultKeyOpt(macVersion),
		SessionKeysOptions.Compose(opts...),
	)
}

func messageGenerationKeys(sk *ttnpb.SessionKeys, macVersion ttnpb.MACVersion) ttnpb.SessionKeys {
	if sk == nil {
		return *MakeSessionKeys(macVersion, false)
	}
	decrypt := func(ke *ttnpb.KeyEnvelope) *types.AES128Key {
		switch {
		case ke == nil:
			return nil
		case ke.Key != nil:
			return ke.Key
		case len(ke.EncryptedKey) > 0:
			k := &types.AES128Key{}
			test.Must(nil, k.UnmarshalBinary(ke.EncryptedKey))
			return k
		default:
			return nil
		}
	}
	return ttnpb.SessionKeys{
		SessionKeyID: sk.SessionKeyID,
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

func MustEncryptUplink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, isFOpts bool, b ...byte) []byte {
	return test.Must(crypto.EncryptUplink(key, devAddr, fCnt, b, isFOpts)).([]byte)
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
	FCtrl          ttnpb.FCtrl
	FCnt           uint32
	ConfFCntDown   uint32
	FPort          uint8
	FRMPayload     []byte
	FOpts          []byte
	DataRate       ttnpb.DataRate
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
	macState := dev.MACState
	if pending {
		session = dev.PendingSession
		macState = dev.PendingMACState
	}
	return func(conf DataUplinkConfig) DataUplinkConfig {
		conf.MACVersion = macState.LoRaWANVersion
		conf.DevAddr = session.DevAddr
		conf.FCnt = session.LastFCntUp + fCntDelta
		conf.DataRate = LoRaWANBands[test.FrequencyPlan(dev.FrequencyPlanID).BandID][dev.LoRaWANPHYVersion].DataRates[drIdx].Rate
		conf.DataRateIndex = drIdx
		conf.Frequency = macState.CurrentParameters.Channels[chIdx].UplinkFrequency
		conf.ChannelIndex = chIdx
		conf.SessionKeys = &session.SessionKeys
		return conf
	}
}

func MakeDataUplink(conf DataUplinkConfig) *ttnpb.UplinkMessage {
	if !conf.FCtrl.Ack && conf.ConfFCntDown > 0 {
		panic("ConfFCntDown must be zero for uplink frames with ACK bit unset")
	}
	devAddr := *conf.DevAddr.Copy(&types.DevAddr{})
	keys := messageGenerationKeys(conf.SessionKeys, conf.MACVersion)
	frmPayload := conf.FRMPayload
	fOpts := conf.FOpts
	if len(conf.FRMPayload) > 0 && conf.FPort == 0 {
		frmPayload = MustEncryptUplink(*keys.NwkSEncKey.Key, devAddr, conf.FCnt, false, frmPayload...)
	} else if len(conf.FOpts) > 0 && conf.MACVersion.EncryptFOpts() {
		fOpts = MustEncryptUplink(*keys.NwkSEncKey.Key, devAddr, conf.FCnt, true, fOpts...)
	}
	mType := ttnpb.MType_UNCONFIRMED_UP
	if conf.Confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}
	mhdr := ttnpb.MHDR{
		MType: mType,
		Major: ttnpb.Major_LORAWAN_R1,
	}
	fhdr := ttnpb.FHDR{
		DevAddr: devAddr,
		FCtrl:   conf.FCtrl,
		FCnt:    conf.FCnt & 0xffff,
		FOpts:   CopyBytes(fOpts),
	}
	phyPayload := test.Must(lorawan.MarshalMessage(ttnpb.Message{
		MHDR: mhdr,
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: &ttnpb.MACPayload{
				FHDR:       fhdr,
				FPort:      uint32(conf.FPort),
				FRMPayload: frmPayload,
			},
		},
	})).([]byte)
	var mic [4]byte
	switch {
	case conf.MACVersion.Compare(ttnpb.MAC_V1_1) < 0:
		mic = test.Must(crypto.ComputeLegacyUplinkMIC(*keys.FNwkSIntKey.Key, devAddr, conf.FCnt, phyPayload)).([4]byte)
	default:
		mic = test.Must(crypto.ComputeUplinkMIC(*keys.SNwkSIntKey.Key, *keys.FNwkSIntKey.Key, conf.ConfFCntDown, uint8(conf.DataRateIndex), conf.ChannelIndex, devAddr, conf.FCnt, phyPayload)).([4]byte)
	}

	phyPayload = append(phyPayload, mic[:]...)
	return MakeUplinkMessage(UplinkMessageConfig{
		RawPayload: phyPayload,
		Payload: func() *ttnpb.Message {
			if conf.DecodePayload {
				return &ttnpb.Message{
					MHDR: mhdr,
					MIC:  phyPayload[len(phyPayload)-4:],
					Payload: &ttnpb.Message_MACPayload{
						MACPayload: &ttnpb.MACPayload{
							FHDR:       fhdr,
							FPort:      uint32(conf.FPort),
							FRMPayload: CopyBytes(frmPayload),
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

func MustEncryptDownlink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, isFOpts bool, b ...byte) []byte {
	return test.Must(crypto.EncryptDownlink(key, devAddr, fCnt, b, isFOpts)).([]byte)
}

type DataDownlinkConfig struct {
	DecodePayload bool

	Confirmed  bool
	MACVersion ttnpb.MACVersion
	DevAddr    types.DevAddr
	FCtrl      ttnpb.FCtrl
	FCnt       uint32
	ConfFCntUp uint32
	FPort      uint8
	FRMPayload []byte
	FOpts      []byte

	Request ttnpb.TxRequest

	SessionKeys *ttnpb.SessionKeys
}

func MakeDataDownlink(conf DataDownlinkConfig) *ttnpb.DownlinkMessage {
	if !conf.FCtrl.Ack && conf.ConfFCntUp > 0 {
		panic("ConfFCntDown must be zero for uplink frames with ACK bit unset")
	}
	devAddr := *conf.DevAddr.Copy(&types.DevAddr{})
	keys := messageGenerationKeys(conf.SessionKeys, conf.MACVersion)
	frmPayload := conf.FRMPayload
	fOpts := conf.FOpts
	if len(frmPayload) > 0 && conf.FPort == 0 {
		frmPayload = MustEncryptDownlink(*keys.NwkSEncKey.Key, devAddr, conf.FCnt, false, frmPayload...)
	} else if len(fOpts) > 0 && conf.MACVersion.EncryptFOpts() {
		fOpts = MustEncryptDownlink(*keys.NwkSEncKey.Key, devAddr, conf.FCnt, true, fOpts...)
	}
	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if conf.Confirmed {
		mType = ttnpb.MType_CONFIRMED_DOWN
	}
	msg := &ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: &ttnpb.MACPayload{
				FHDR: ttnpb.FHDR{
					DevAddr: devAddr,
					FCtrl:   conf.FCtrl,
					FCnt:    conf.FCnt & 0xffff,
					FOpts:   fOpts,
				},
				FullFCnt:   conf.FCnt,
				FPort:      uint32(conf.FPort),
				FRMPayload: frmPayload,
			},
		},
	}
	phyPayload := test.Must(lorawan.MarshalMessage(*msg)).([]byte)
	var mic [4]byte
	switch {
	case conf.MACVersion.Compare(ttnpb.MAC_V1_1) < 0:
		mic = test.Must(crypto.ComputeLegacyDownlinkMIC(*keys.FNwkSIntKey.Key, devAddr, conf.FCnt, phyPayload)).([4]byte)
	default:
		mic = test.Must(crypto.ComputeDownlinkMIC(*keys.SNwkSIntKey.Key, devAddr, conf.ConfFCntUp, conf.FCnt, phyPayload)).([4]byte)
	}
	msg.MIC = mic[:]
	return &ttnpb.DownlinkMessage{
		Settings: &ttnpb.DownlinkMessage_Request{
			Request: deepcopy.Copy(&conf.Request).(*ttnpb.TxRequest),
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
	for phyID, phyVersions := range LoRaWANBands {
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
			case ttnpb.PHY_V1_0_3_REV_A, ttnpb.PHY_V1_1_REV_B:
			case ttnpb.PHY_V1_0_2_REV_B:
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
		ttnpb.MAC_V1_0,
		ttnpb.MAC_V1_0_1,
		ttnpb.MAC_V1_0_2,
		ttnpb.MAC_V1_0_3,
		ttnpb.MAC_V1_0_4,
		ttnpb.MAC_V1_1,
	} {
		switch macVersion {
		case ttnpb.MAC_V1_0_4, ttnpb.MAC_V1_1:
		case ttnpb.MAC_V1_0_3:
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
		ttnpb.CLASS_A,
		ttnpb.CLASS_B,
		ttnpb.CLASS_C,
	} {
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("Class:%s", class.String()))...)
		}, class)
	}
}

func ForEachFrequencyPlan(tb testing.TB, f func(func(...string) string, string, *frequencyplans.FrequencyPlan)) {
	fpIDs, err := frequencyplans.NewStore(test.FrequencyPlansFetcher).GetAllIDs()
	if err != nil {
		tb.Errorf("failed to get frequency plans: %w", err)
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
	for macVersion, phyVersions := range LoRaWANVersionPairs {
		switch macVersion {
		case ttnpb.MAC_V1_0_3, ttnpb.MAC_V1_1:
		case ttnpb.MAC_V1_0_2:
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
			b, ok := LoRaWANBands[fp.BandID][phyVersion]
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
			if _, ok := LoRaWANVersionPairs[macVersion][phyVersion]; !ok {
				return
			}
			f(func(parts ...string) string {
				return makeBandName(makeMACName(parts...))
			}, phy, phyVersion, macVersion)
		})
	})
}
