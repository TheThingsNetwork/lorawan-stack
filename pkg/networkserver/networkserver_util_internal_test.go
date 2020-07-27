// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

const (
	BeaconPeriod                     = beaconPeriod
	BeaconReserved                   = beaconReserved
	DownlinkProcessTaskName          = downlinkProcessTaskName
	DownlinkRetryInterval            = downlinkRetryInterval
	InfrastructureDelay              = infrastructureDelay
	NetworkInitiatedDownlinkInterval = networkInitiatedDownlinkInterval
	OptimalADRUplinkCount            = optimalADRUplinkCount
	RecentDownlinkCount              = recentDownlinkCount
	RecentUplinkCount                = recentUplinkCount

	AppIDString = "test-app-id"
	DevID       = "test-dev-id"
)

var (
	AdaptDataRate                           = adaptDataRate
	AppendRecentDownlink                    = appendRecentDownlink
	AppendRecentUplink                      = appendRecentUplink
	ApplicationJoinAcceptWithoutAppSKey     = applicationJoinAcceptWithoutAppSKey
	ApplyCFList                             = applyCFList
	DeviceDefaultBeaconFrequency            = deviceDefaultBeaconFrequency
	DeviceDefaultClass                      = deviceDefaultClass
	DeviceDefaultChannels                   = deviceDefaultChannels
	DeviceDefaultLoRaWANVersion             = deviceDefaultLoRaWANVersion
	DeviceDefaultMaxDutyCycle               = deviceDefaultMaxDutyCycle
	DeviceDefaultPingSlotDataRateIndexValue = deviceDefaultPingSlotDataRateIndexValue
	DeviceDefaultPingSlotFrequency          = deviceDefaultPingSlotFrequency
	DeviceDefaultRX1DataRateOffset          = deviceDefaultRX1DataRateOffset
	DeviceDefaultRX1Delay                   = deviceDefaultRX1Delay
	DeviceDefaultRX2DataRateIndex           = deviceDefaultRX2DataRateIndex
	DeviceDefaultRX2Frequency               = deviceDefaultRX2Frequency
	DeviceDesiredADRAckDelayExponent        = deviceDesiredADRAckDelayExponent
	DeviceDesiredADRAckLimitExponent        = deviceDesiredADRAckLimitExponent
	DeviceDesiredBeaconFrequency            = deviceDesiredBeaconFrequency
	DeviceDesiredChannels                   = deviceDesiredChannels
	DeviceDesiredDownlinkDwellTime          = deviceDesiredDownlinkDwellTime
	DeviceDesiredMaxDutyCycle               = deviceDesiredMaxDutyCycle
	DeviceDesiredMaxEIRP                    = deviceDesiredMaxEIRP
	DeviceDesiredPingSlotDataRateIndexValue = deviceDesiredPingSlotDataRateIndexValue
	DeviceDesiredPingSlotFrequency          = deviceDesiredPingSlotFrequency
	DeviceDesiredRX1DataRateOffset          = deviceDesiredRX1DataRateOffset
	DeviceDesiredRX1Delay                   = deviceDesiredRX1Delay
	DeviceDesiredRX2DataRateIndex           = deviceDesiredRX2DataRateIndex
	DeviceDesiredRX2Frequency               = deviceDesiredRX2Frequency
	DeviceDesiredUplinkDwellTime            = deviceDesiredUplinkDwellTime
	DownlinkPathsFromMetadata               = downlinkPathsFromMetadata
	HandleLinkCheckReq                      = handleLinkCheckReq
	JoinResponseWithoutKeys                 = joinResponseWithoutKeys
	LastDownlink                            = lastDownlink
	LastUplink                              = lastUplink
	LoRaWANBands                            = lorawanBands
	LoRaWANVersionPairs                     = lorawanVersionPairs
	NewMACState                             = newMACState
	NextPingSlotAt                          = nextPingSlotAt
	TimePtr                                 = timePtr

	ErrABPJoinRequest             = errABPJoinRequest
	ErrApplicationDownlinkTooLong = errApplicationDownlinkTooLong
	ErrDecodePayload              = errDecodePayload
	ErrDeviceNotFound             = errDeviceNotFound
	ErrDuplicate                  = errDuplicate
	ErrInvalidAbsoluteTime        = errInvalidAbsoluteTime
	ErrInvalidPayload             = errInvalidPayload
	ErrOutdatedData               = errOutdatedData
	ErrRejoinRequest              = errRejoinRequest
	ErrUnsupportedLoRaWANVersion  = errUnsupportedLoRaWANVersion

	EvtBeginApplicationLink          = evtBeginApplicationLink
	EvtClassCSwitch                  = evtClassCSwitch
	EvtClusterJoinAttempt            = evtClusterJoinAttempt
	EvtClusterJoinFail               = evtClusterJoinFail
	EvtClusterJoinSuccess            = evtClusterJoinSuccess
	EvtCreateEndDevice               = evtCreateEndDevice
	EvtDropDataUplink                = evtDropDataUplink
	EvtDropJoinRequest               = evtDropJoinRequest
	EvtEndApplicationLink            = evtEndApplicationLink
	EvtEnqueueDeviceModeConfirmation = evtEnqueueDeviceModeConfirmation
	EvtEnqueueDevStatusRequest       = evtEnqueueDevStatusRequest
	EvtEnqueueLinkADRRequest         = evtEnqueueLinkADRRequest
	EvtEnqueueRekeyConfirmation      = evtEnqueueRekeyConfirmation
	EvtForwardDataUplink             = evtForwardDataUplink
	EvtForwardJoinAccept             = evtForwardJoinAccept
	EvtInteropJoinAttempt            = evtInteropJoinAttempt
	EvtInteropJoinFail               = evtInteropJoinFail
	EvtInteropJoinSuccess            = evtInteropJoinSuccess
	EvtProcessDataUplink             = evtProcessDataUplink
	EvtProcessJoinRequest            = evtProcessJoinRequest
	EvtReceiveDataUplink             = evtReceiveDataUplink
	EvtReceiveDeviceModeIndication   = evtReceiveDeviceModeIndication
	EvtReceiveJoinRequest            = evtReceiveJoinRequest
	EvtReceiveRekeyIndication        = evtReceiveRekeyIndication
	EvtScheduleDataDownlinkAttempt   = evtScheduleDataDownlinkAttempt
	EvtScheduleDataDownlinkFail      = evtScheduleDataDownlinkFail
	EvtScheduleDataDownlinkSuccess   = evtScheduleDataDownlinkSuccess
	EvtScheduleJoinAcceptAttempt     = evtScheduleJoinAcceptAttempt
	EvtScheduleJoinAcceptFail        = evtScheduleJoinAcceptFail
	EvtScheduleJoinAcceptSuccess     = evtScheduleJoinAcceptSuccess
	EvtUpdateEndDevice               = evtUpdateEndDevice

	Timeout = (1 << 10) * test.Delay

	ErrTestInternal = errors.DefineInternal("test_internal", "test error")
	ErrTestNotFound = errors.DefineNotFound("test_not_found", "test error")

	FNwkSIntKey = types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	NwkSEncKey  = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey = types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	AppSKey     = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	JoinEUI = types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	DevEUI  = types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	DevAddr = types.DevAddr{0x42, 0x00, 0x00, 0x00}

	AppID = ttnpb.ApplicationIdentifiers{ApplicationID: AppIDString}

	NetID = test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)

	NewDeviceRegistry         func(t testing.TB) (DeviceRegistry, func())
	NewApplicationUplinkQueue func(t testing.TB) (ApplicationUplinkQueue, func())
	NewDownlinkTaskQueue      func(t testing.TB) (DownlinkTaskQueue, func())
	NewUplinkDeduplicator     func(t testing.TB) (UplinkDeduplicator, func())
)

type DownlinkPath = downlinkPath

var timeMu sync.RWMutex

func SetMockClock(clock *test.MockClock) func() {
	timeMu.Lock()
	oldNow := timeNow
	oldAfter := timeAfter
	timeNow = clock.Now
	timeAfter = clock.After
	return func() {
		timeNow = oldNow
		timeAfter = oldAfter
		timeMu.Unlock()
	}
}

func NSScheduleWindow() time.Duration {
	return nsScheduleWindow()
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

// CopyEndDevice returns a deep copy of ttnpb.EndDevice pb.
func CopyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

// CopyUplinkMessage returns a deep copy of ttnpb.UplinkMessage pb.
func CopyUplinkMessage(pb *ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return deepcopy.Copy(pb).(*ttnpb.UplinkMessage)
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

// CopySessionKeys returns a deep copy of ttnpb.SessionKeys pb.
func CopySessionKeys(pb *ttnpb.SessionKeys) *ttnpb.SessionKeys {
	return deepcopy.Copy(pb).(*ttnpb.SessionKeys)
}

func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

func AES128KeyPtr(key types.AES128Key) *types.AES128Key {
	return &key
}

func FrequencyPlan(id string) *frequencyplans.FrequencyPlan {
	return test.Must(frequencyplans.NewStore(test.FrequencyPlansFetcher).GetByID(id)).(*frequencyplans.FrequencyPlan)
}

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

func MakeDefaultUS915CurrentMACParameters(ver ttnpb.PHYVersion) ttnpb.MACParameters {
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
		Channels:                   chs,
	}
}

func MakeDefaultUS915FSB2DesiredMACParameters(ver ttnpb.PHYVersion) ttnpb.MACParameters {
	params := MakeDefaultUS915CurrentMACParameters(ver)
	for _, ch := range params.Channels {
		switch ch.UplinkFrequency {
		case 903900000,
			904100000,
			904300000,
			904500000,
			904700000,
			904900000,
			905100000,
			905300000:
			continue
		}
		ch.EnableUplink = false
	}
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

func MakeOTAAIdentifiers(devAddr *types.DevAddr) *ttnpb.EndDeviceIdentifiers {
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: AppID,
		DeviceID:               DevID,

		DevEUI:  DevEUI.Copy(&types.EUI64{}),
		JoinEUI: JoinEUI.Copy(&types.EUI64{}),
	}
	if devAddr != nil {
		ids.DevAddr = devAddr.Copy(&types.DevAddr{})
	}
	return ids
}

func MakeABPIdentifiers(withDevEUI bool) *ttnpb.EndDeviceIdentifiers {
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: AppID,
		DeviceID:               DevID,
		DevAddr:                DevAddr.Copy(&types.DevAddr{}),
	}
	if withDevEUI {
		ids.DevEUI = DevEUI.Copy(&types.EUI64{})
	}
	return ids
}

func MakeSessionKeys(macVersion ttnpb.MACVersion, withAppSKey bool) *ttnpb.SessionKeys {
	sk := &ttnpb.SessionKeys{
		FNwkSIntKey: &ttnpb.KeyEnvelope{
			Key: &FNwkSIntKey,
		},
		SessionKeyID: []byte("test-session-key-id"),
	}
	if withAppSKey {
		sk.AppSKey = &ttnpb.KeyEnvelope{
			Key: &AppSKey,
		}
	}
	switch {
	case macVersion.Compare(ttnpb.MAC_V1_1) < 0:
		sk.NwkSEncKey = sk.FNwkSIntKey
		sk.SNwkSIntKey = sk.FNwkSIntKey
	default:
		sk.NwkSEncKey = &ttnpb.KeyEnvelope{
			Key: &NwkSEncKey,
		}
		sk.SNwkSIntKey = &ttnpb.KeyEnvelope{
			Key: &SNwkSIntKey,
		}
	}
	return CopySessionKeys(sk)
}

var RxMetadata = [...]*ttnpb.RxMetadata{
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

var GatewayAntennaIdentifiers = [...]ttnpb.GatewayAntennaIdentifiers{
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

func MakeUplinkSettings(dr ttnpb.DataRate, freq uint64) ttnpb.TxSettings {
	return ttnpb.TxSettings{
		DataRate:  *deepcopy.Copy(&dr).(*ttnpb.DataRate),
		EnableCRC: true,
		Frequency: freq,
		Timestamp: 42,
	}
}

func MakeJoinRequestDevNonce() types.DevNonce {
	return types.DevNonce{0x00, 0x01}
}

func MakeJoinRequestMIC() [4]byte {
	return [...]byte{0x03, 0x02, 0x01, 0x00}
}

func MakeJoinRequestPHYPayload() [23]byte {
	devNonce := MakeJoinRequestDevNonce()
	mic := MakeJoinRequestMIC()
	return [...]byte{
		/* MHDR */
		0b000_000_00,
		JoinEUI[7], JoinEUI[6], JoinEUI[5], JoinEUI[4], JoinEUI[3], JoinEUI[2], JoinEUI[1], JoinEUI[0],
		DevEUI[7], DevEUI[6], DevEUI[5], DevEUI[4], DevEUI[3], DevEUI[2], DevEUI[1], DevEUI[0],
		/* DevNonce */
		devNonce[1], devNonce[0],
		/* MIC */
		mic[0], mic[1], mic[2], mic[3],
	}
}

func MakeJoinRequestDecodedPayload() *ttnpb.Message {
	mic := MakeJoinRequestMIC()
	return &ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: ttnpb.MType_JOIN_REQUEST,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		MIC: mic[:],
		Payload: &ttnpb.Message_JoinRequestPayload{
			JoinRequestPayload: &ttnpb.JoinRequestPayload{
				JoinEUI:  *JoinEUI.Copy(&types.EUI64{}),
				DevEUI:   *DevEUI.Copy(&types.EUI64{}),
				DevNonce: MakeJoinRequestDevNonce(),
			},
		},
	}
}

var JoinRequestCorrelationIDs = [...]string{
	"join-request-correlation-id-1",
	"join-request-correlation-id-2",
	"join-request-correlation-id-3",
}

func MakeJoinRequest(decodePayload bool, dr ttnpb.DataRate, freq uint64, mds ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage {
	phyPayload := MakeJoinRequestPHYPayload()
	msg := &ttnpb.UplinkMessage{
		CorrelationIDs: append([]string{}, JoinRequestCorrelationIDs[:]...),
		RawPayload:     phyPayload[:],
		RxMetadata:     mds,
		ReceivedAt:     timeNow(),
		Settings:       MakeUplinkSettings(dr, freq),
	}
	if decodePayload {
		msg.Payload = MakeJoinRequestDecodedPayload()
	}
	return msg
}

func MakeNsJsJoinRequest(macVersion ttnpb.MACVersion, phyVersion ttnpb.PHYVersion, fp *frequencyplans.FrequencyPlan, devAddr *types.DevAddr, rxDelay ttnpb.RxDelay, rx1DROffset uint8, rx2DR ttnpb.DataRateIndex, correlationIDs ...string) *ttnpb.JoinRequest {
	phyPayload := MakeJoinRequestPHYPayload()
	return &ttnpb.JoinRequest{
		CFList:         frequencyplans.CFList(*fp, phyVersion),
		CorrelationIDs: correlationIDs,
		DevAddr: func() types.DevAddr {
			if devAddr != nil {
				return *devAddr.Copy(&types.DevAddr{})
			} else {
				return types.DevAddr{}
			}
		}(),
		NetID:              *NetID.Copy(&types.NetID{}),
		RawPayload:         phyPayload[:],
		Payload:            MakeJoinRequestDecodedPayload(),
		RxDelay:            rxDelay,
		SelectedMACVersion: macVersion,
		DownlinkSettings: ttnpb.DLSettings{
			OptNeg:      macVersion.Compare(ttnpb.MAC_V1_1) >= 0,
			Rx1DROffset: uint32(rx1DROffset),
			Rx2DR:       rx2DR,
		},
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

func MustEncryptUplink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, isFOpts bool, b ...byte) []byte {
	return test.Must(crypto.EncryptUplink(key, devAddr, fCnt, b, isFOpts)).([]byte)
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
}

func WithDeviceDataUplinkConfig(dev *ttnpb.EndDevice, pending bool, drIdx ttnpb.DataRateIndex, chIdx uint8, lostFrames uint32) func(DataUplinkConfig) DataUplinkConfig {
	session := dev.Session
	macState := dev.MACState
	if pending {
		session = dev.PendingSession
		macState = dev.PendingMACState
	}
	return func(conf DataUplinkConfig) DataUplinkConfig {
		conf.MACVersion = macState.LoRaWANVersion
		conf.DevAddr = session.DevAddr
		conf.FCnt = session.LastFCntUp + 1 + lostFrames
		conf.DataRate = LoRaWANBands[FrequencyPlan(dev.FrequencyPlanID).BandID][dev.LoRaWANPHYVersion].DataRates[drIdx].Rate
		conf.Frequency = macState.CurrentParameters.Channels[chIdx].UplinkFrequency
		return conf
	}
}

func WithMatchedUplinkSettings(msg *ttnpb.UplinkMessage, chIdx uint8, drIdx ttnpb.DataRateIndex) *ttnpb.UplinkMessage {
	msg = CopyUplinkMessage(msg)
	msg.Settings.DataRateIndex = drIdx
	msg.DeviceChannelIndex = uint32(chIdx)
	return msg
}

func MakeDataUplink(conf DataUplinkConfig) *ttnpb.UplinkMessage {
	if !conf.FCtrl.Ack && conf.ConfFCntDown > 0 {
		panic("ConfFCntDown must be zero for uplink frames with ACK bit unset")
	}
	devAddr := *conf.DevAddr.Copy(&types.DevAddr{})
	keys := MakeSessionKeys(conf.MACVersion, false)
	if len(conf.FRMPayload) > 0 && conf.FPort == 0 {
		conf.FRMPayload = MustEncryptUplink(*keys.NwkSEncKey.Key, devAddr, conf.FCnt, false, conf.FRMPayload...)
	} else if len(conf.FOpts) > 0 && conf.MACVersion.EncryptFOpts() {
		conf.FOpts = MustEncryptUplink(*keys.NwkSEncKey.Key, devAddr, conf.FCnt, true, conf.FOpts...)
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
		FOpts:   CopyBytes(conf.FOpts),
	}
	phyPayload := test.Must(lorawan.MarshalMessage(ttnpb.Message{
		MHDR: mhdr,
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: &ttnpb.MACPayload{
				FHDR:       fhdr,
				FPort:      uint32(conf.FPort),
				FRMPayload: conf.FRMPayload,
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
	msg := &ttnpb.UplinkMessage{
		ReceivedAt: conf.ReceivedAt,
		RawPayload: phyPayload,
		RxMetadata: deepcopy.Copy(conf.RxMetadata).([]*ttnpb.RxMetadata),
		Settings:   MakeUplinkSettings(conf.DataRate, conf.Frequency),
		CorrelationIDs: CopyStrings(func() []string {
			if len(conf.CorrelationIDs) == 0 {
				return DataUplinkCorrelationIDs[:]
			}
			return conf.CorrelationIDs
		}()),
	}
	if conf.DecodePayload {
		msg.Payload = &ttnpb.Message{
			MHDR: mhdr,
			MIC:  phyPayload[len(phyPayload)-4:],
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FHDR:       fhdr,
					FPort:      uint32(conf.FPort),
					FRMPayload: CopyBytes(conf.FRMPayload),
					FullFCnt:   conf.FCnt,
				},
			},
		}
	}
	if conf.Matched {
		return WithMatchedUplinkSettings(msg, conf.ChannelIndex, conf.DataRateIndex)
	}
	return msg
}

func MustEncryptDownlink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, isFOpts bool, b ...byte) []byte {
	return test.Must(crypto.EncryptDownlink(key, devAddr, fCnt, b, isFOpts)).([]byte)
}

func MakeDataDownlink(macVersion ttnpb.MACVersion, confirmed bool, devAddr types.DevAddr, fCtrl ttnpb.FCtrl, fCnt, confFCntUp uint32, fPort uint8, frmPayload, fOpts []byte, txReq *ttnpb.TxRequest, cids ...string) *ttnpb.DownlinkMessage {
	if !fCtrl.Ack && confFCntUp > 0 {
		panic("ConfFCntDown must be zero for uplink frames with ACK bit unset")
	}
	devAddr = *devAddr.Copy(&types.DevAddr{})
	keys := MakeSessionKeys(macVersion, false)
	if len(frmPayload) > 0 && fPort == 0 {
		frmPayload = MustEncryptDownlink(*keys.NwkSEncKey.Key, devAddr, fCnt, false, frmPayload...)
	} else if len(fOpts) > 0 && macVersion.EncryptFOpts() {
		fOpts = MustEncryptDownlink(*keys.NwkSEncKey.Key, devAddr, fCnt, true, fOpts...)
	}
	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if confirmed {
		mType = ttnpb.MType_CONFIRMED_DOWN
	}
	phyPayload := test.Must(lorawan.MarshalMessage(ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: &ttnpb.MACPayload{
				FHDR: ttnpb.FHDR{
					DevAddr: devAddr,
					FCtrl:   fCtrl,
					FCnt:    fCnt & 0xffff,
					FOpts:   fOpts,
				},
				FPort:      uint32(fPort),
				FRMPayload: frmPayload,
			},
		},
	})).([]byte)
	var mic [4]byte
	switch {
	case macVersion.Compare(ttnpb.MAC_V1_1) < 0:
		mic = test.Must(crypto.ComputeLegacyDownlinkMIC(*keys.FNwkSIntKey.Key, devAddr, fCnt, phyPayload)).([4]byte)
	default:
		mic = test.Must(crypto.ComputeDownlinkMIC(*keys.SNwkSIntKey.Key, devAddr, confFCntUp, fCnt, phyPayload)).([4]byte)
	}
	return &ttnpb.DownlinkMessage{
		CorrelationIDs: append([]string{}, cids...),
		RawPayload:     append(phyPayload, mic[:]...),
		Settings: &ttnpb.DownlinkMessage_Request{
			Request: txReq,
		},
	}
}

func NewISPeer(ctx context.Context, is interface {
	ttnpb.ApplicationAccessServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, is, ttnpb.RegisterApplicationAccessServer)).(cluster.Peer)
}

func NewGSPeer(ctx context.Context, gs interface {
	ttnpb.NsGsServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, gs, ttnpb.RegisterNsGsServer)).(cluster.Peer)
}

func NewJSPeer(ctx context.Context, js interface {
	ttnpb.NsJsServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, js, ttnpb.RegisterNsJsServer)).(cluster.Peer)
}

var _ InteropClient = MockInteropClient{}

// MockInteropClient is a mock InteropClient used for testing.
type MockInteropClient struct {
	HandleJoinRequestFunc func(context.Context, types.NetID, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
}

// HandleJoinRequest calls HandleJoinRequestFunc if set and panics otherwise.
func (m MockInteropClient) HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if m.HandleJoinRequestFunc == nil {
		panic("HandleJoinRequest called, but not set")
	}
	return m.HandleJoinRequestFunc(ctx, netID, req)
}

type InteropClientHandleJoinRequestResponse struct {
	Response *ttnpb.JoinResponse
	Error    error
}

type InteropClientHandleJoinRequestRequest struct {
	Context  context.Context
	NetID    types.NetID
	Request  *ttnpb.JoinRequest
	Response chan<- InteropClientHandleJoinRequestResponse
}

func MakeInteropClientHandleJoinRequestChFunc(reqCh chan<- InteropClientHandleJoinRequestRequest) func(context.Context, types.NetID, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	return func(ctx context.Context, netID types.NetID, msg *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
		respCh := make(chan InteropClientHandleJoinRequestResponse)
		reqCh <- InteropClientHandleJoinRequestRequest{
			Context:  ctx,
			NetID:    netID,
			Request:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

var _ ttnpb.NsJsServer = &MockNsJsServer{}

type MockNsJsServer struct {
	HandleJoinFunc  func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(context.Context, *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error)
}

// HandleJoin calls HandleJoinFunc if set and panics otherwise.
func (m MockNsJsServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if m.HandleJoinFunc == nil {
		panic("HandleJoin called, but not set")
	}
	return m.HandleJoinFunc(ctx, req)
}

// GetNwkSKeys calls GetNwkSKeysFunc if set and panics otherwise.
func (m MockNsJsServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if m.GetNwkSKeysFunc == nil {
		panic("GetNwkSKeys called, but not set")
	}
	return m.GetNwkSKeysFunc(ctx, req)
}

type NsJsHandleJoinResponse struct {
	Response *ttnpb.JoinResponse
	Error    error
}

type NsJsHandleJoinRequest struct {
	Context  context.Context
	Message  *ttnpb.JoinRequest
	Response chan<- NsJsHandleJoinResponse
}

func MakeNsJsHandleJoinChFunc(reqCh chan<- NsJsHandleJoinRequest) func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	return func(ctx context.Context, msg *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
		respCh := make(chan NsJsHandleJoinResponse)
		reqCh <- NsJsHandleJoinRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

var _ ttnpb.NsJsClient = &MockNsJsClient{}

type MockNsJsClient struct {
	*test.MockClientStream
	HandleJoinFunc  func(context.Context, *ttnpb.JoinRequest, ...grpc.CallOption) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(context.Context, *ttnpb.SessionKeyRequest, ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error)
}

// HandleJoin calls HandleJoinFunc if set and panics otherwise.
func (m MockNsJsClient) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
	if m.HandleJoinFunc == nil {
		panic("HandleJoin called, but not set")
	}
	return m.HandleJoinFunc(ctx, req, opts...)
}

// GetNwkSKeys calls GetNwkSKeysFunc if set and panics otherwise.
func (m MockNsJsClient) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
	if m.GetNwkSKeysFunc == nil {
		panic("GetNwkSKeys called, but not set")
	}
	return m.GetNwkSKeysFunc(ctx, req, opts...)
}

var _ ttnpb.NsGsServer = &MockNsGsServer{}

type MockNsGsServer struct {
	ScheduleDownlinkFunc func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error)
}

// ScheduleDownlink calls ScheduleDownlinkFunc if set and panics otherwise.
func (m MockNsGsServer) ScheduleDownlink(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	if m.ScheduleDownlinkFunc == nil {
		panic("ScheduleDownlink called, but not set")
	}
	return m.ScheduleDownlinkFunc(ctx, msg)
}

type NsGsScheduleDownlinkResponse struct {
	Response *ttnpb.ScheduleDownlinkResponse
	Error    error
}

type NsGsScheduleDownlinkRequest struct {
	Context  context.Context
	Message  *ttnpb.DownlinkMessage
	Response chan<- NsGsScheduleDownlinkResponse
}

func MakeNsGsScheduleDownlinkChFunc(reqCh chan<- NsGsScheduleDownlinkRequest) func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	return func(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
		respCh := make(chan NsGsScheduleDownlinkResponse)
		reqCh <- NsGsScheduleDownlinkRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

func AssertInteropClientHandleJoinRequestRequest(ctx context.Context, reqCh <-chan InteropClientHandleJoinRequestRequest, assert func(context.Context, types.NetID, *ttnpb.JoinRequest) bool, resp InteropClientHandleJoinRequestResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for InteropClient.HandleJoinRequest to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.NetID, req.Request) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for InteropClient.HandleJoinRequest response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertNsJsHandleJoinRequest(ctx context.Context, reqCh <-chan NsJsHandleJoinRequest, assert func(ctx context.Context, msg *ttnpb.JoinRequest) bool, resp NsJsHandleJoinResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for NsJs.HandleJoin to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Message) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for NsJs.HandleJoin response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertAuthNsJsHandleJoinRequest(ctx context.Context, authReqCh <-chan test.ClusterAuthRequest, joinReqCh <-chan NsJsHandleJoinRequest, joinAssert func(ctx context.Context, msg *ttnpb.JoinRequest) bool, authResp grpc.CallOption, joinResp NsJsHandleJoinResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	if !test.AssertClusterAuthRequest(ctx, authReqCh, authResp) {
		return false
	}
	return AssertNsJsHandleJoinRequest(ctx, joinReqCh, joinAssert, joinResp)
}

func AssertNsJsPeerHandleAuthJoinRequest(ctx context.Context, peerReqCh <-chan test.ClusterGetPeerRequest, authReqCh <-chan test.ClusterAuthRequest, idsAssert func(ctx context.Context, ids ttnpb.Identifiers) bool, joinAssert func(ctx context.Context, msg *ttnpb.JoinRequest) bool, authResp grpc.CallOption, joinResp NsJsHandleJoinResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	joinReqCh := make(chan NsJsHandleJoinRequest)
	if !test.AssertClusterGetPeerRequest(ctx, peerReqCh, func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
		return assertions.New(t).So(role, should.Equal, ttnpb.ClusterRole_JOIN_SERVER) && idsAssert(ctx, ids)
	},
		test.ClusterGetPeerResponse{
			Peer: NewJSPeer(ctx, &MockNsJsServer{
				HandleJoinFunc: MakeNsJsHandleJoinChFunc(joinReqCh),
			}),
		},
	) {
		return false
	}
	return AssertAuthNsJsHandleJoinRequest(ctx, authReqCh, joinReqCh, joinAssert, authResp, joinResp)
}

func AssertNsGsScheduleDownlinkRequest(ctx context.Context, reqCh <-chan NsGsScheduleDownlinkRequest, assert func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool, resp NsGsScheduleDownlinkResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for NsGs.ScheduleDownlink to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Message) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for NsGs.ScheduleDownlink response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertAuthNsGsScheduleDownlinkRequest(ctx context.Context, authReqCh <-chan test.ClusterAuthRequest, scheduleReqCh <-chan NsGsScheduleDownlinkRequest, scheduleAssert func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool, authResp grpc.CallOption, scheduleResp NsGsScheduleDownlinkResponse) bool {
	if !test.AssertClusterAuthRequest(ctx, authReqCh, authResp) {
		return false
	}
	return AssertNsGsScheduleDownlinkRequest(ctx, scheduleReqCh, scheduleAssert, scheduleResp)
}

func AssertNetworkServerClose(ctx context.Context, ns *NetworkServer) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	if !test.WaitContext(ctx, ns.Close) {
		t.Error("Timed out while waiting for Network Server to gracefully close")
		return false
	}
	return true
}

type InteropClientEnvironment struct {
	HandleJoinRequest <-chan InteropClientHandleJoinRequestRequest
}

func newMockInteropClient(t *testing.T) (InteropClient, InteropClientEnvironment, func()) {
	t.Helper()

	handleJoinCh := make(chan InteropClientHandleJoinRequestRequest)
	return &MockInteropClient{
			HandleJoinRequestFunc: MakeInteropClientHandleJoinRequestChFunc(handleJoinCh),
		}, InteropClientEnvironment{
			HandleJoinRequest: handleJoinCh,
		},
		func() {
			select {
			case <-handleJoinCh:
				t.Error("InteropClient.HandleJoin call missed")
			default:
				close(handleJoinCh)
			}
		}
}

type TestClusterEnvironment struct {
	Auth    <-chan test.ClusterAuthRequest
	GetPeer <-chan test.ClusterGetPeerRequest
}

type TestEnvironment struct {
	Config

	Cluster       TestClusterEnvironment
	Events        <-chan test.EventPubSubPublishRequest
	InteropClient *InteropClientEnvironment

	*grpc.ClientConn
}

func (env TestEnvironment) AssertLinkApplication(ctx context.Context, appID ttnpb.ApplicationIdentifiers, replaceEvents ...events.Event) (ttnpb.AsNs_LinkApplicationClient, func(error) events.Event, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var link ttnpb.AsNs_LinkApplicationClient
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		link, err = ttnpb.NewAsNsClient(env.ClientConn).LinkApplication(
			(rpcmetadata.MD{
				ID: appID.ApplicationID,
			}).ToOutgoingContext(ctx),
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "link-application-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	var reqCIDs []string
	if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
		func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
			reqCIDs = events.CorrelationIDsFromContext(ctx)
			return a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS) && a.So(ids, should.BeNil)
		},
		test.ClusterGetPeerResponse{
			Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
				ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
			}),
		},
	), should.BeTrue) {
		return nil, nil, false
	}
	a.So(reqCIDs, should.HaveLength, 1)

	if !a.So(test.AssertListRightsRequest(ctx, listRightsCh,
		func(ctx context.Context, ids ttnpb.Identifiers) bool {
			md := rpcmetadata.FromIncomingContext(ctx)
			cids := events.CorrelationIDsFromContext(ctx)
			return a.So(cids, should.NotResemble, reqCIDs) &&
				a.So(cids, should.HaveLength, 1) &&
				a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "link-application-key") &&
				a.So(ids, should.Resemble, &appID)
		}, ttnpb.RIGHT_APPLICATION_LINK,
	), should.BeTrue) {
		return nil, nil, false
	}

	if !a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 1+len(replaceEvents), func(evs ...events.Event) bool {
		return a.So(evs, should.HaveSameElementsEvent, append(
			[]events.Event{EvtBeginApplicationLink.NewWithIdentifiersAndData(events.ContextWithCorrelationID(test.Context(), reqCIDs...), appID, nil)},
			replaceEvents...,
		))
	}), should.BeTrue) {
		t.Error("AS link events assertion failed")
		return nil, nil, false
	}

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for AS link to open")
		return nil, nil, false
	}
	if !a.So(err, should.BeNil) {
		t.Errorf("Link failed with: %s", err)
		return nil, nil, false
	}
	return link, func(err error) events.Event {
		return EvtEndApplicationLink.NewWithIdentifiersAndData(events.ContextWithCorrelationID(test.Context(), reqCIDs...), appID, err)
	}, a.So(err, should.BeNil)
}

func (env TestEnvironment) AssertWithApplicationLink(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, ttnpb.AsNs_LinkApplicationClient) bool, replaceEvents ...events.Event) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	ctx, cancel := context.WithCancel(ctx)

	var once sync.Once
	defer once.Do(cancel)

	link, linkEndEvent, ok := env.AssertLinkApplication(ctx, appID, replaceEvents...)
	if !test.AllTrue(
		a.So(ok, should.BeTrue),
		f(ctx, link),
	) {
		return false
	}
	once.Do(cancel)
	return a.So(env.Events, should.ReceiveEventResembling,
		linkEndEvent(context.Canceled),
	)
}

func (env TestEnvironment) AssertScheduleDownlink(ctx context.Context, assert func(context.Context, *ttnpb.DownlinkMessage) bool, paths []DownlinkPath) bool {
	return test.MustTFromContext(ctx).Run("Schedule downlink", func(t *testing.T) {
		a, ctx := test.NewWithContext(ctx, t)
		t.Helper()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		scheduleDownlinkCh := make(chan NsGsScheduleDownlinkRequest)
		gsPeer := NewGSPeer(ctx, &MockNsGsServer{
			ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlinkCh),
		})
		for _, path := range paths {
			if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(test.AllTrue(
					a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER),
					a.So(ids, should.Resemble, *path.GatewayIdentifiers),
				), should.BeTrue)
			},
				test.ClusterGetPeerResponse{
					Peer: gsPeer,
				},
			), should.BeTrue) {
				t.Error("Gateway Server peer look-up assertion failed")
				return
			}
		}

		if !a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, env.Cluster.Auth, scheduleDownlinkCh, assert,
			&grpc.EmptyCallOption{},
			NsGsScheduleDownlinkResponse{
				Response: &ttnpb.ScheduleDownlinkResponse{},
			},
		), should.BeTrue) {
			t.Error("Gateway Server scheduling assertion failed")
			return
		}
	})
}

func (env TestEnvironment) AssertSendDeviceUplink(ctx context.Context, expectedEvs []events.Event, eventEqual func(x, y events.Event) bool, ups ...*ttnpb.UplinkMessage) (<-chan error, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	errCh := make(chan error, len(ups))
	wg := &sync.WaitGroup{}
	wg.Add(len(ups) - 1)
	go func() {
		_, err := ttnpb.NewGsNsClient(env.ClientConn).HandleUplink(ctx, ups[0])
		t.Logf("First HandleUplink returned %v", err)
		errCh <- err
		wg.Wait()
		close(errCh)
	}()
	for _, up := range ups[1:] {
		up := up
		time.AfterFunc((1<<3)*test.Delay, func() {
			_, err := ttnpb.NewGsNsClient(env.ClientConn).HandleUplink(ctx, up)
			t.Logf("Duplicate HandleUplink returned %v", err)
			errCh <- err
			wg.Done()
		})
	}
	if !a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, len(expectedEvs), func(evs ...events.Event) bool {
		if !a.So(evs, should.HaveSameElementsFunc, eventEqual, expectedEvs) {
			actualEvs := map[events.Event]struct{}{}
			for _, ev := range evs {
				actualEvs[ev] = struct{}{}
			}
		outer:
			for _, expected := range expectedEvs {
				for actual := range actualEvs {
					if eventEqual(expected, actual) {
						delete(actualEvs, actual)
						continue outer
					}
				}
				t.Logf("Failed to match expected event '%s' with payload '%+v'", expected.Name(), expected.Data())
			}
			for actual := range actualEvs {
				t.Logf("Failed to match actual event '%s' with payload '%+v'", actual.Name(), actual.Data())
			}
			return false
		}
		return true
	}), should.BeTrue) {
		t.Error("Data uplink duplicate event assertion failed")
		return nil, false
	}
	for range ups[1:] {
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for duplicate HandleUplink to return")
			return nil, false

		case err := <-errCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle duplicate uplink: %s", err)
				return nil, false
			}
		}
	}
	return errCh, true
}

func AssertProcessApplicationUp(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, assert func(context.Context, *ttnpb.ApplicationUp) bool) bool {
	return test.MustTFromContext(ctx).Run("Application uplink", func(t *testing.T) {
		a, ctx := test.NewWithContext(ctx, t)
		t.Helper()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var asUp *ttnpb.ApplicationUp
		var err error
		if !a.So(test.WaitContext(ctx, func() {
			asUp, err = link.Recv()
		}), should.BeTrue) {
			t.Error("Timed out while waiting for application uplink to be sent to Application Server")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to receive Application Server uplink: %s", err)
			return
		}
		if !a.So(assert(ctx, asUp), should.BeTrue) {
			t.Errorf("Application uplink assertion failed")
			return
		}
		if !a.So(test.WaitContext(ctx, func() {
			err = link.Send(ttnpb.Empty)
		}), should.BeTrue) {
			t.Error("Timed out while waiting for Network Server to process Application Server response")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to send Application Server uplink response: %s", err)
			return
		}
	})
}

func AssertProcessApplicationUps(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, asserts ...func(context.Context, *ttnpb.ApplicationUp) bool) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	for _, assert := range asserts {
		if !a.So(AssertProcessApplicationUp(ctx, link, assert), should.BeTrue) {
			return false
		}
	}
	return true
}

func DownlinkProtoPaths(paths ...DownlinkPath) (pbs []*ttnpb.DownlinkPath) {
	for _, p := range paths {
		pbs = append(pbs, p.DownlinkPath)
	}
	return pbs
}

type JoinAssertionConfig struct {
	Context         context.Context
	Link            ttnpb.AsNs_LinkApplicationClient
	LinkContext     context.Context
	Identifiers     ttnpb.EndDeviceIdentifiers
	FrequencyPlanID string
	MACVersion      ttnpb.MACVersion
	PHYVersion      ttnpb.PHYVersion
	ChannelIndex    uint8
	DataRateIndex   ttnpb.DataRateIndex
	EventEqual      func(x, y events.Event) bool

	Response *ttnpb.JoinResponse

	DesiredRX1Delay         ttnpb.RxDelay
	RX1DROffset             uint8
	DesiredRX1DROffset      uint8
	RX2DataRateIndex        ttnpb.DataRateIndex
	DesiredRX2DataRateIndex ttnpb.DataRateIndex
	RX2Frequency            uint64
}

func (env TestEnvironment) AssertJoin(conf JoinAssertionConfig) (*ttnpb.JoinRequest, bool) {
	t := test.MustTFromContext(conf.Context)

	fp := FrequencyPlan(conf.FrequencyPlanID)
	phy := LoRaWANBands[fp.BandID][conf.PHYVersion]
	upCh := phy.UplinkChannels[conf.ChannelIndex]
	upDR := phy.DataRates[conf.DataRateIndex].Rate

	var expectedEvs []events.Event
	var joinReq *ttnpb.JoinRequest
	var joinCIDs []string
	if !t.Run("Join-request", func(t *testing.T) {
		a, ctx := test.NewWithContext(conf.Context, t)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handleJoinCh := make(chan NsJsHandleJoinRequest)
		jsPeer := NewJSPeer(ctx, &MockNsJsServer{
			HandleJoinFunc: MakeNsJsHandleJoinChFunc(handleJoinCh),
		})

		makeUplink := func(matched bool, rxMetadata ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage {
			msg := MakeJoinRequest(matched, upDR, upCh.Frequency, rxMetadata...)
			if matched {
				return WithMatchedUplinkSettings(msg, conf.ChannelIndex, conf.DataRateIndex)
			}
			return msg
		}
		var ups []*ttnpb.UplinkMessage
		var firstUp *ttnpb.UplinkMessage
		var preSendEvs []events.Event
		for i, upMD := range [][]*ttnpb.RxMetadata{
			RxMetadata[:2],
			nil,
			RxMetadata[2:],
		} {
			ups = append(ups, makeUplink(false, upMD...))
			if i == 0 {
				firstUp = makeUplink(true, upMD...)
			} else {
				preSendEvs = append(preSendEvs,
					EvtReceiveJoinRequest.NewWithIdentifiersAndData(ctx, conf.Identifiers, makeUplink(true, upMD...)),
					EvtDropJoinRequest.NewWithIdentifiersAndData(ctx, conf.Identifiers, ErrDuplicate),
				)
			}
		}

		start := time.Now()
		handleUplinkErrCh, ok := env.AssertSendDeviceUplink(ctx, preSendEvs, conf.EventEqual, ups...)
		if !a.So(ok, should.BeTrue) {
			t.Error("Uplink send assertion failed")
			return
		}

		var getPeerCtx context.Context
		if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
			func(ctx context.Context, role ttnpb.ClusterRole, peerIDs ttnpb.Identifiers) bool {
				for _, cid := range ups[0].CorrelationIDs {
					a.So(events.CorrelationIDsFromContext(ctx), should.Contain, cid)
				}
				getPeerCtx = ctx
				return test.AllTrue(
					a.So(role, should.Equal, ttnpb.ClusterRole_JOIN_SERVER),
					a.So(peerIDs, should.Resemble, conf.Identifiers),
				)
			},
			test.ClusterGetPeerResponse{Peer: jsPeer},
		), should.BeTrue) {
			t.Error("Join Server peer look-up assertion failed")
			return
		}
		joinReq = MakeNsJsJoinRequest(conf.MACVersion, conf.PHYVersion, fp, nil, conf.DesiredRX1Delay, uint8(conf.DesiredRX1DROffset), conf.DesiredRX2DataRateIndex, events.CorrelationIDsFromContext(getPeerCtx)...)
		if !a.So(AssertAuthNsJsHandleJoinRequest(ctx, env.Cluster.Auth, handleJoinCh, func(ctx context.Context, req *ttnpb.JoinRequest) bool {
			joinReq.DevAddr = req.DevAddr
			return test.AllTrue(
				a.So(req.DevAddr, should.NotBeEmpty),
				a.So(req.DevAddr.NwkID(), should.Resemble, env.Config.NetID.ID()),
				a.So(req.DevAddr.NetIDType(), should.Equal, env.Config.NetID.Type()),
				a.So(req, should.Resemble, joinReq),
			)
		},
			&grpc.EmptyCallOption{},
			NsJsHandleJoinResponse{
				Response: conf.Response,
			},
		), should.BeTrue) {
			t.Error("Join-request send assertion failed")
			return
		}
		joinCIDs = append(events.CorrelationIDsFromContext(getPeerCtx), conf.Response.CorrelationIDs...)

		a.So(env.Events, should.ReceiveEventFunc, conf.EventEqual,
			EvtReceiveJoinRequest.NewWithIdentifiersAndData(events.ContextWithCorrelationID(test.Context(), ups[0].CorrelationIDs...), conf.Identifiers, firstUp),
		)
		a.So(env.Events, should.ReceiveEventsResembling,
			EvtClusterJoinAttempt.NewWithIdentifiersAndData(getPeerCtx, conf.Identifiers, joinReq),
			EvtClusterJoinSuccess.NewWithIdentifiersAndData(getPeerCtx, conf.Identifiers, &ttnpb.JoinResponse{
				RawPayload: conf.Response.RawPayload,
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: conf.Response.SessionKeys.SessionKeyID,
				},
				Lifetime:       conf.Response.Lifetime,
				CorrelationIDs: conf.Response.CorrelationIDs,
			}),
		)
		a.So(env.Events, should.ReceiveEventFunc, conf.EventEqual,
			EvtProcessJoinRequest.NewWithIdentifiersAndData(getPeerCtx, conf.Identifiers, makeUplink(true, RxMetadata[:]...)),
		)
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for HandleUplink to return")
			return
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle uplink: %s", err)
				return
			}
		}

		if !a.So(AssertProcessApplicationUp(ctx, conf.Link, func(ctx context.Context, up *ttnpb.ApplicationUp) bool {
			expectedEvs = append(expectedEvs, EvtForwardJoinAccept.NewWithIdentifiersAndData(conf.LinkContext, up.EndDeviceIdentifiers, up))

			a := assertions.New(test.MustTFromContext(ctx))
			return a.So(test.AllTrue(
				a.So(up.CorrelationIDs, should.HaveSameElementsDeep, joinCIDs),
				a.So([]time.Time{start, up.GetJoinAccept().GetReceivedAt(), time.Now()}, should.BeChronological),
				a.So(up, should.Resemble, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: *MakeOTAAIdentifiers(&joinReq.DevAddr),
					CorrelationIDs:       up.CorrelationIDs,
					Up: &ttnpb.ApplicationUp_JoinAccept{
						JoinAccept: &ttnpb.ApplicationJoinAccept{
							AppSKey:      conf.Response.AppSKey,
							SessionKeyID: conf.Response.SessionKeyID,
							ReceivedAt:   up.GetJoinAccept().GetReceivedAt(),
						},
					},
				}),
			), should.BeTrue)
		}), should.BeTrue) {
			t.Error("Failed to send join-accept to Application Server")
			return
		}
	}) {
		t.Error("Join-request assertion failed")
		return nil, false
	}
	if joinReq == nil {
		panic("JOIN REQ NOT TRUE")
	}
	return joinReq, t.Run("Join-accept", func(t *testing.T) {
		a, ctx := test.NewWithContext(conf.Context, t)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		paths := DownlinkPathsFromMetadata(RxMetadata[:]...)
		txReq := &ttnpb.TxRequest{
			Class:            ttnpb.CLASS_A,
			DownlinkPaths:    DownlinkProtoPaths(paths...),
			Rx1Delay:         ttnpb.RxDelay(phy.JoinAcceptDelay1.Seconds()),
			Rx1DataRateIndex: test.Must(phy.Rx1DataRate(conf.DataRateIndex, uint32(conf.RX1DROffset), fp.DwellTime.GetUplinks())).(ttnpb.DataRateIndex),
			Rx1Frequency:     phy.DownlinkChannels[test.Must(phy.Rx1Channel(conf.ChannelIndex)).(uint8)].Frequency,
			Rx2DataRateIndex: conf.RX2DataRateIndex,
			Rx2Frequency:     conf.RX2Frequency,
			Priority:         ttnpb.TxSchedulePriority_HIGHEST,
			FrequencyPlanID:  conf.FrequencyPlanID,
		}
		if !a.So(env.AssertScheduleDownlink(ctx, func(ctx context.Context, down *ttnpb.DownlinkMessage) bool {
			return test.AllTrue(
				a.So(events.CorrelationIDsFromContext(ctx), should.NotBeEmpty),
				a.So(down.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, joinCIDs),
				a.So(down, should.Resemble, &ttnpb.DownlinkMessage{
					CorrelationIDs: down.CorrelationIDs,
					RawPayload:     conf.Response.RawPayload,
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: txReq,
					},
				}),
			)
		}, paths), should.BeTrue) {
			t.Error("Join-accept assertion failed")
			return
		}
		a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 2+len(expectedEvs), func(evs ...events.Event) bool {
			return a.So(evs, should.HaveSameElementsFunc, conf.EventEqual, append(expectedEvs,
				EvtScheduleJoinAcceptAttempt.NewWithIdentifiersAndData(ctx, conf.Identifiers, txReq),
				EvtScheduleJoinAcceptSuccess.NewWithIdentifiersAndData(ctx, conf.Identifiers, &ttnpb.ScheduleDownlinkResponse{}),
			))
		}), should.BeTrue)
	})
}

func (env TestEnvironment) AssertSendDataUplink(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, linkCtx context.Context, ids ttnpb.EndDeviceIdentifiers, makeUplink func(matched bool, rxMetadata ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage, eventEqual func(x, y events.Event) bool, processEvBuilders ...events.Builder) bool {
	return test.MustTFromContext(ctx).Run("Data uplink", func(t *testing.T) {
		a, ctx := test.NewWithContext(ctx, t)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		expectedEvBuilders := processEvBuilders
		var ups []*ttnpb.UplinkMessage
		for i, upMD := range [][]*ttnpb.RxMetadata{
			nil,
			RxMetadata[3:],
			RxMetadata[:3],
		} {
			ups = append(ups, makeUplink(false, upMD...))
			expectedEvBuilders = append(expectedEvBuilders, EvtReceiveDataUplink.With(events.WithData(makeUplink(true, upMD...))))
			if i > 0 {
				expectedEvBuilders = append(expectedEvBuilders, EvtDropDataUplink.With(events.WithData(ErrDuplicate)))
			}
		}

		handleUplinkErrCh, ok := env.AssertSendDeviceUplink(ctx, events.Builders(expectedEvBuilders).New(ctx, events.WithIdentifiers(ids)), eventEqual, ups...)
		if !a.So(ok, should.BeTrue) {
			t.Error("Uplink send assertion failed")
			return
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for HandleUplink to return")
			return
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle uplink: %s", err)
				return
			}
		}
	})
}

func (env TestEnvironment) AssertSetDevice(ctx context.Context, create bool, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var dev *ttnpb.EndDevice
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		dev, err = ttnpb.NewNsEndDeviceRegistryClient(env.ClientConn).Set(
			ctx,
			req,
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "set-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	var reqCIDs []string
	if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
		func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
			reqCIDs = events.CorrelationIDsFromContext(ctx)
			return a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS) && a.So(ids, should.BeNil)
		},
		test.ClusterGetPeerResponse{
			Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
				ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
			}),
		},
	), should.BeTrue) {
		return nil, false
	}

	a.So(reqCIDs, should.HaveLength, 1)

	if !a.So(test.AssertListRightsRequest(ctx, listRightsCh,
		func(ctx context.Context, ids ttnpb.Identifiers) bool {
			md := rpcmetadata.FromIncomingContext(ctx)
			return a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "set-key") &&
				a.So(ids, should.Resemble, &req.EndDevice.ApplicationIdentifiers)
		}, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	), should.BeTrue) {
		return nil, false
	}

	ev := EvtCreateEndDevice.BindData(nil)
	if !create {
		ev = EvtUpdateEndDevice.BindData(nil)
	}
	if !a.So(env.Events, should.ReceiveEventResembling, ev.New(events.ContextWithCorrelationID(test.Context(), reqCIDs...), events.WithIdentifiers(req.EndDevice.EndDeviceIdentifiers))) {
		if create {
			t.Error("Failed to assert end device create event")
		} else {
			t.Error("Failed to assert end device update event")
		}
		return nil, false
	}

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for device to be set")
		return nil, false
	}
	return dev, a.So(err, should.BeNil)
}

type TestConfig struct {
	NetworkServer        Config
	NetworkServerOptions []Option
	Component            component.Config
	TaskStarter          component.TaskStarter
	Timeout              time.Duration
}

func StartTest(t *testing.T, conf TestConfig) (*NetworkServer, context.Context, TestEnvironment, func()) {
	t.Helper()

	authCh := make(chan test.ClusterAuthRequest)
	getPeerCh := make(chan test.ClusterGetPeerRequest)
	eventsPublishCh := make(chan test.EventPubSubPublishRequest)

	var closeFuncs []func()
	closeFuncs = append(closeFuncs, test.SetDefaultEventsPubSub(&test.MockEventPubSub{
		PublishFunc: test.MakeEventPubSubPublishChFunc(eventsPublishCh),
	}))
	if conf.NetworkServer.DeduplicationWindow == 0 {
		conf.NetworkServer.DeduplicationWindow = time.Nanosecond
	}
	if conf.NetworkServer.CooldownWindow == 0 {
		conf.NetworkServer.CooldownWindow = conf.NetworkServer.DeduplicationWindow + time.Nanosecond
	}

	cmpOpts := []component.Option{
		component.WithClusterNew(func(context.Context, *cluster.Config, ...cluster.Option) (cluster.Cluster, error) {
			return &test.MockCluster{
				AuthFunc:    test.MakeClusterAuthChFunc(authCh),
				GetPeerFunc: test.MakeClusterGetPeerChFunc(getPeerCh),
				JoinFunc:    test.ClusterJoinNilFunc,
				WithVerifiedSourceFunc: func(ctx context.Context) context.Context {
					return clusterauth.NewContext(ctx, nil)
				},
			}, nil
		}),
	}
	if conf.TaskStarter != nil {
		cmpOpts = append(cmpOpts, component.WithTaskStarter(conf.TaskStarter))
	}

	if conf.NetworkServer.Devices == nil {
		v, closeFn := NewDeviceRegistry(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.Devices = v
	}
	if conf.NetworkServer.ApplicationUplinkQueue.Queue == nil {
		v, closeFn := NewApplicationUplinkQueue(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.ApplicationUplinkQueue.Queue = v
	}
	if conf.NetworkServer.DownlinkTasks == nil {
		v, closeFn := NewDownlinkTaskQueue(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.DownlinkTasks = v
	}
	if conf.NetworkServer.UplinkDeduplicator == nil {
		v, closeFn := NewUplinkDeduplicator(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.UplinkDeduplicator = v
	}

	ns := test.Must(New(
		componenttest.NewComponent(t, &conf.Component, cmpOpts...),
		&conf.NetworkServer,
		conf.NetworkServerOptions...,
	)).(*NetworkServer)
	ns.Component.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

	env := TestEnvironment{
		Config: conf.NetworkServer,

		Cluster: TestClusterEnvironment{
			Auth:    authCh,
			GetPeer: getPeerCh,
		},
		Events: eventsPublishCh,
	}
	if ns.interopClient == nil {
		m, mEnv, closeM := newMockInteropClient(t)
		ns.interopClient = m
		env.InteropClient = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}

	componenttest.StartComponent(t, ns.Component)
	env.ClientConn = ns.LoopbackConn()

	_, ctx := test.New(t)
	ctx, cancel := context.WithTimeout(ctx, conf.Timeout)
	return ns, ctx, env, func() {
		cancel()
		for _, f := range closeFuncs {
			f()
		}
		select {
		case <-authCh:
			t.Error("Cluster.Auth call missed")
		default:
			close(authCh)
		}
		select {
		case <-getPeerCh:
			t.Error("Cluster.GetPeer call missed")
		default:
			close(getPeerCh)
		}
		select {
		case <-eventsPublishCh:
			t.Error("events.Publish call missed")
		default:
			close(eventsPublishCh)
		}
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
		}, fpID, FrequencyPlan(fpID))
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

func LogEvents(t *testing.T, ch <-chan test.EventPubSubPublishRequest) {
	for ev := range ch {
		t.Logf("Event %s published with data %v", ev.Event.Name(), ev.Event.Data())
		ev.Response <- struct{}{}
	}
}

func MustCreateDevice(ctx context.Context, r DeviceRegistry, dev *ttnpb.EndDevice, paths ...string) (*ttnpb.EndDevice, context.Context) {
	dev, ctx, err := CreateDevice(ctx, r, dev, paths...)
	test.Must(nil, err)
	return dev, ctx
}

func MustGetDeviceByID(ctx context.Context, r DeviceRegistry, appID ttnpb.ApplicationIdentifiers, devID string, paths ...string) (*ttnpb.EndDevice, context.Context) {
	if len(paths) == 0 {
		paths = ttnpb.EndDeviceFieldPathsTopLevel
	}
	dev, ctx, err := r.GetByID(ctx, appID, devID, paths)
	test.Must(nil, err)
	return dev, ctx
}

type SetDeviceRequest struct {
	*ttnpb.EndDevice
	Paths []string
}

type ContextualEndDevice struct {
	context.Context
	*ttnpb.EndDevice
}

func MustCreateDevices(ctx context.Context, r DeviceRegistry, devs ...SetDeviceRequest) []*ContextualEndDevice {
	var setDevices []*ContextualEndDevice
	for _, dev := range devs {
		set, ctx := MustCreateDevice(ctx, r, dev.EndDevice, dev.Paths...)
		setDevices = append(setDevices, &ContextualEndDevice{
			Context:   ctx,
			EndDevice: set,
		})
	}
	return setDevices
}

// TODO: Remove mocks below

var _ ApplicationUplinkQueue = MockApplicationUplinkQueue{}

// MockApplicationUplinkQueue is a mock ApplicationUplinkQueue used for testing.
type MockApplicationUplinkQueue struct {
	AddFunc       func(ctx context.Context, ups ...*ttnpb.ApplicationUp) error
	SubscribeFunc func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error
}

// Add calls AddFunc if set and panics otherwise.
func (m MockApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	if m.AddFunc == nil {
		panic("Add called, but not set")
	}
	return m.AddFunc(ctx, ups...)
}

// Subscribe calls SubscribeFunc if set and panics otherwise.
func (m MockApplicationUplinkQueue) Subscribe(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error {
	if m.SubscribeFunc == nil {
		panic("Subscribe called, but not set")
	}
	return m.SubscribeFunc(ctx, appID, f)
}

type ApplicationUplinkQueueAddRequest struct {
	Context  context.Context
	Uplinks  []*ttnpb.ApplicationUp
	Response chan<- error
}

func MakeApplicationUplinkQueueAddChFunc(reqCh chan<- ApplicationUplinkQueueAddRequest) func(context.Context, ...*ttnpb.ApplicationUp) error {
	return func(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
		respCh := make(chan error)
		reqCh <- ApplicationUplinkQueueAddRequest{
			Context:  ctx,
			Uplinks:  ups,
			Response: respCh,
		}
		return <-respCh
	}
}

type ApplicationUplinkQueueSubscribeRequest struct {
	Context     context.Context
	Identifiers ttnpb.ApplicationIdentifiers
	Func        func(context.Context, *ttnpb.ApplicationUp) error
	Response    chan<- error
}

func MakeApplicationUplinkQueueSubscribeChFunc(reqCh chan<- ApplicationUplinkQueueSubscribeRequest) func(context.Context, ttnpb.ApplicationIdentifiers, func(context.Context, *ttnpb.ApplicationUp) error) error {
	return func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error {
		respCh := make(chan error)
		reqCh <- ApplicationUplinkQueueSubscribeRequest{
			Context:     ctx,
			Identifiers: appID,
			Func:        f,
			Response:    respCh,
		}
		return <-respCh
	}
}

var _ DownlinkTaskQueue = MockDownlinkTaskQueue{}

// MockDownlinkTaskQueue is a mock DownlinkTaskQueue used for testing.
type MockDownlinkTaskQueue struct {
	AddFunc func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error
	PopFunc func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
}

// Add calls AddFunc if set and panics otherwise.
func (m MockDownlinkTaskQueue) Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error {
	if m.AddFunc == nil {
		panic("Add called, but not set")
	}
	return m.AddFunc(ctx, devID, t, replace)
}

// Pop calls PopFunc if set and panics otherwise.
func (m MockDownlinkTaskQueue) Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	if m.PopFunc == nil {
		panic("Pop called, but not set")
	}
	return m.PopFunc(ctx, f)
}

type DownlinkTaskAddRequest struct {
	Context     context.Context
	Identifiers ttnpb.EndDeviceIdentifiers
	Time        time.Time
	Replace     bool
	Response    chan<- error
}

func MakeDownlinkTaskAddChFunc(reqCh chan<- DownlinkTaskAddRequest) func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error {
	return func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error {
		respCh := make(chan error)
		reqCh <- DownlinkTaskAddRequest{
			Context:     ctx,
			Identifiers: devID,
			Time:        t,
			Replace:     replace,
			Response:    respCh,
		}
		return <-respCh
	}
}

type DownlinkTaskPopRequest struct {
	Context  context.Context
	Func     func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error
	Response chan<- error
}

func MakeDownlinkTaskPopChFunc(reqCh chan<- DownlinkTaskPopRequest) func(context.Context, func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	return func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
		respCh := make(chan error)
		reqCh <- DownlinkTaskPopRequest{
			Context:  ctx,
			Func:     f,
			Response: respCh,
		}
		return <-respCh
	}
}

// DownlinkTaskPopBlockFunc is DownlinkTasks.Pop function, which blocks until context is done and returns nil.
func DownlinkTaskPopBlockFunc(ctx context.Context, _ func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	<-ctx.Done()
	return nil
}

var _ DeviceRegistry = MockDeviceRegistry{}

// MockDeviceRegistry is a mock DeviceRegistry used for testing.
type MockDeviceRegistry struct {
	GetByEUIFunc    func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error)
	GetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error)
	RangeByAddrFunc func(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error
	SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
}

// GetByEUI calls GetByEUIFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	if m.GetByEUIFunc == nil {
		panic("GetByEUI called, but not set")
	}
	return m.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

// GetByID calls GetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	if m.GetByIDFunc == nil {
		panic("GetByID called, but not set")
	}
	return m.GetByIDFunc(ctx, appID, devID, paths)
}

// RangeByAddr calls RangeByAddrFunc if set and panics otherwise.
func (m MockDeviceRegistry) RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error {
	if m.RangeByAddrFunc == nil {
		panic("RangeByAddr called, but not set")
	}
	return m.RangeByAddrFunc(ctx, devAddr, paths, f)
}

// SetByID calls SetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	if m.SetByIDFunc == nil {
		panic("SetByID called, but not set")
	}
	return m.SetByIDFunc(ctx, appID, devID, paths, f)
}

type contextualDeviceAndError struct {
	Device  *ttnpb.EndDevice
	Context context.Context
	Error   error
}

type DeviceRegistryGetByEUIResponse contextualDeviceAndError

type DeviceRegistryGetByEUIRequest struct {
	Context  context.Context
	JoinEUI  types.EUI64
	DevEUI   types.EUI64
	Paths    []string
	Response chan<- DeviceRegistryGetByEUIResponse
}

func MakeDeviceRegistryGetByEUIChFunc(reqCh chan<- DeviceRegistryGetByEUIRequest) func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.EndDevice, context.Context, error) {
	return func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error) {
		respCh := make(chan DeviceRegistryGetByEUIResponse)
		reqCh <- DeviceRegistryGetByEUIRequest{
			Context:  ctx,
			JoinEUI:  joinEUI,
			DevEUI:   devEUI,
			Paths:    paths,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Device, resp.Context, resp.Error
	}
}

type DeviceRegistryGetByIDResponse contextualDeviceAndError

type DeviceRegistryGetByIDRequest struct {
	Context                context.Context
	ApplicationIdentifiers ttnpb.ApplicationIdentifiers
	DeviceID               string
	Paths                  []string
	Response               chan<- DeviceRegistryGetByIDResponse
}

func MakeDeviceRegistryGetByIDChFunc(reqCh chan<- DeviceRegistryGetByIDRequest) func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, context.Context, error) {
	return func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
		respCh := make(chan DeviceRegistryGetByIDResponse)
		reqCh <- DeviceRegistryGetByIDRequest{
			Context:                ctx,
			ApplicationIdentifiers: appID,
			DeviceID:               devID,
			Paths:                  paths,
			Response:               respCh,
		}
		resp := <-respCh
		return resp.Device, resp.Context, resp.Error
	}
}

type DeviceRegistryRangeByAddrRequest struct {
	Context  context.Context
	DevAddr  types.DevAddr
	Paths    []string
	Func     func(context.Context, *ttnpb.EndDevice) bool
	Response chan<- error
}

func MakeDeviceRegistryRangeByAddrChFunc(reqCh chan<- DeviceRegistryRangeByAddrRequest) func(context.Context, types.DevAddr, []string, func(context.Context, *ttnpb.EndDevice) bool) error {
	return func(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error {
		respCh := make(chan error)
		reqCh <- DeviceRegistryRangeByAddrRequest{
			Context:  ctx,
			DevAddr:  devAddr,
			Paths:    paths,
			Func:     f,
			Response: respCh,
		}
		return <-respCh
	}
}

type DeviceRegistrySetByIDResponse contextualDeviceAndError

type DeviceRegistrySetByIDRequest struct {
	Context                context.Context
	ApplicationIdentifiers ttnpb.ApplicationIdentifiers
	DeviceID               string
	Paths                  []string
	Func                   func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)
	Response               chan<- DeviceRegistrySetByIDResponse
}

func MakeDeviceRegistrySetByIDChFunc(reqCh chan<- DeviceRegistrySetByIDRequest) func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	return func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
		respCh := make(chan DeviceRegistrySetByIDResponse)
		reqCh <- DeviceRegistrySetByIDRequest{
			Context:                ctx,
			ApplicationIdentifiers: appID,
			DeviceID:               devID,
			Paths:                  paths,
			Func:                   f,
			Response:               respCh,
		}
		resp := <-respCh
		return resp.Device, resp.Context, resp.Error
	}
}

type DeviceRegistrySetByIDRequestFuncResponse struct {
	Device *ttnpb.EndDevice
	Paths  []string
	Error  error
}

var _ UplinkDeduplicator = &MockUplinkDeduplicator{}

type MockUplinkDeduplicator struct {
	DeduplicateUplinkFunc   func(context.Context, *ttnpb.UplinkMessage, time.Duration) (bool, error)
	AccumulatedMetadataFunc func(context.Context, *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error)
}

// DeduplicateUplink calls DeduplicateUplinkFunc if set and panics otherwise.
func (m MockUplinkDeduplicator) DeduplicateUplink(ctx context.Context, up *ttnpb.UplinkMessage, d time.Duration) (bool, error) {
	if m.DeduplicateUplinkFunc == nil {
		panic("DeduplicateUplink called, but not set")
	}
	return m.DeduplicateUplinkFunc(ctx, up, d)
}

// AccumulatedMetadata calls AccumulatedMetadataFunc if set and panics otherwise.
func (m MockUplinkDeduplicator) AccumulatedMetadata(ctx context.Context, up *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error) {
	if m.AccumulatedMetadataFunc == nil {
		panic("AccumulatedMetadata called, but not set")
	}
	return m.AccumulatedMetadataFunc(ctx, up)
}

type UplinkDeduplicatorDeduplicateUplinkResponse struct {
	Ok    bool
	Error error
}

type UplinkDeduplicatorDeduplicateUplinkRequest struct {
	Context  context.Context
	Uplink   *ttnpb.UplinkMessage
	Window   time.Duration
	Response chan<- UplinkDeduplicatorDeduplicateUplinkResponse
}

func MakeUplinkDeduplicatorDeduplicateUplinkChFunc(reqCh chan<- UplinkDeduplicatorDeduplicateUplinkRequest) func(context.Context, *ttnpb.UplinkMessage, time.Duration) (bool, error) {
	return func(ctx context.Context, up *ttnpb.UplinkMessage, window time.Duration) (bool, error) {
		respCh := make(chan UplinkDeduplicatorDeduplicateUplinkResponse)
		reqCh <- UplinkDeduplicatorDeduplicateUplinkRequest{
			Context:  ctx,
			Uplink:   up,
			Window:   window,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Ok, resp.Error
	}
}

type UplinkDeduplicatorAccumulatedMetadataResponse struct {
	Metadata []*ttnpb.RxMetadata
	Error    error
}

type UplinkDeduplicatorAccumulatedMetadataRequest struct {
	Context  context.Context
	Uplink   *ttnpb.UplinkMessage
	Response chan<- UplinkDeduplicatorAccumulatedMetadataResponse
}

func MakeUplinkDeduplicatorAccumulatedMetadataChFunc(reqCh chan<- UplinkDeduplicatorAccumulatedMetadataRequest) func(context.Context, *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error) {
	return func(ctx context.Context, up *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error) {
		respCh := make(chan UplinkDeduplicatorAccumulatedMetadataResponse)
		reqCh <- UplinkDeduplicatorAccumulatedMetadataRequest{
			Context:  ctx,
			Uplink:   up,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Metadata, resp.Error
	}
}

func AssertDownlinkTaskAddRequest(ctx context.Context, reqCh <-chan DownlinkTaskAddRequest, assert func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) bool, resp error) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for DownlinkTasks.Add to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Identifiers, req.Time, req.Replace) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for DownlinkTasks.Add response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertDeduplicateUplink(ctx context.Context, reqCh <-chan UplinkDeduplicatorDeduplicateUplinkRequest, assert func(context.Context, *ttnpb.UplinkMessage, time.Duration) bool, resp UplinkDeduplicatorDeduplicateUplinkResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for UplinkDeduplicator.DeduplicateUplink to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Uplink, req.Window) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for UplinkDeduplicator.DeduplicateUplink response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertAccumulatedMetadata(ctx context.Context, reqCh <-chan UplinkDeduplicatorAccumulatedMetadataRequest, assert func(context.Context, *ttnpb.UplinkMessage) bool, resp UplinkDeduplicatorAccumulatedMetadataResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for UplinkDeduplicator.AccumulatedMetadata to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Uplink) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for UplinkDeduplicator.AccumulatedMetadata response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertDeviceRegistryGetByEUI(ctx context.Context, reqCh <-chan DeviceRegistryGetByEUIRequest, assert func(context.Context, types.EUI64, types.EUI64, []string) bool, respFunc func(context.Context) DeviceRegistryGetByEUIResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.JoinEUI, req.DevEUI, req.Paths) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for DeviceRegistry.GetByEUI response to be processed")
			return false

		case req.Response <- respFunc(req.Context):
			return true
		}
	}
}

func AssertDeviceRegistrySetByID(ctx context.Context, reqCh <-chan DeviceRegistrySetByIDRequest, assert func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) bool, respFunc func(context.Context) DeviceRegistrySetByIDResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.ApplicationIdentifiers, req.DeviceID, req.Paths, req.Func) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for DeviceRegistry.SetByID response to be processed")
			return false

		case req.Response <- respFunc(req.Context):
			return true
		}
	}
}

func AssertDeviceRegistryRangeByAddr(ctx context.Context, reqCh <-chan DeviceRegistryRangeByAddrRequest, assert func(context.Context, types.DevAddr, []string, func(context.Context, *ttnpb.EndDevice) bool) bool, resp error) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for DeviceRegistry.RangeByAddr to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.DevAddr, req.Paths, req.Func) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for DeviceRegistry.RangeByAddr response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertApplicationUplinkQueueAddRequest(ctx context.Context, reqCh <-chan ApplicationUplinkQueueAddRequest, assert func(context.Context, ...*ttnpb.ApplicationUp) bool, resp error) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for ApplicationUplinkQueue.Add to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Uplinks...) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for ApplicationUplinkQueue.Add response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}
