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
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

const (
	RecentUplinkCount     = recentUplinkCount
	RecentDownlinkCount   = recentDownlinkCount
	OptimalADRUplinkCount = optimalADRUplinkCount

	AppIDString = "test-app-id"
	DevID       = "test-dev-id"
)

var (
	AdaptDataRate                       = adaptDataRate
	AppendRecentUplink                  = appendRecentUplink
	ApplicationJoinAcceptWithoutAppSKey = applicationJoinAcceptWithoutAppSKey
	DownlinkPathsFromMetadata           = downlinkPathsFromMetadata
	FrequencyPlanChannels               = frequencyPlanChannels
	HandleLinkCheckReq                  = handleLinkCheckReq
	JoinResponseWithoutKeys             = joinResponseWithoutKeys
	NewMACState                         = newMACState
	TimePtr                             = timePtr

	ErrABPJoinRequest            = errABPJoinRequest
	ErrDecodePayload             = errDecodePayload
	ErrDeviceNotFound            = errDeviceNotFound
	ErrDuplicate                 = errDuplicate
	ErrInvalidPayload            = errInvalidPayload
	ErrOutdatedData              = errOutdatedData
	ErrRejoinRequest             = errRejoinRequest
	ErrUnsupportedLoRaWANVersion = errUnsupportedLoRaWANVersion

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

const InfrastructureDelay = infrastructureDelay

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

func Band(id string, phyVersion ttnpb.PHYVersion) band.Band {
	return test.Must(test.Must(band.GetByID(id)).(band.Band).Version(phyVersion)).(band.Band)
}

func MakeDefaultEU868CurrentChannels() []*ttnpb.MACParameters_Channel {
	return []*ttnpb.MACParameters_Channel{
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
		Rx1Delay:                   ttnpb.RX_DELAY_1,
		Rx2DataRateIndex:           ttnpb.DATA_RATE_0,
		Rx2Frequency:               869525000,
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

func MakeUplinkMACBuffer(phy band.Band, cmds ...MACCommander) []byte {
	var b []byte
	for _, cmd := range cmds {
		b = test.Must(lorawan.DefaultMACCommands.AppendUplink(phy, b, *cmd.MACCommand())).([]byte)
	}
	return b
}

func MakeDownlinkMACBuffer(phy band.Band, cmds ...MACCommander) []byte {
	var b []byte
	for _, cmd := range cmds {
		b = test.Must(lorawan.DefaultMACCommands.AppendDownlink(phy, b, *cmd.MACCommand())).([]byte)
	}
	return b
}

func MustEncryptUplink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, isFOpts bool, b ...byte) []byte {
	return test.Must(crypto.EncryptUplink(key, devAddr, fCnt, b, isFOpts)).([]byte)
}

func MakeDataUplink(macVersion ttnpb.MACVersion, decodePayload, confirmed bool, devAddr types.DevAddr, fCtrl ttnpb.FCtrl, fCnt, confFCntDown uint32, fPort uint8, frmPayload, fOpts []byte, dr ttnpb.DataRate, drIdx ttnpb.DataRateIndex, freq uint64, chIdx uint8, mds ...*ttnpb.RxMetadata) *ttnpb.UplinkMessage {
	if len(fOpts) > 0 && fPort == 0 {
		panic("FOpts must not be set for FPort == 0")
	}
	devAddr = *devAddr.Copy(&types.DevAddr{})
	mType := ttnpb.MType_UNCONFIRMED_UP
	if confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}
	mhdr := ttnpb.MHDR{
		MType: mType,
		Major: ttnpb.Major_LORAWAN_R1,
	}
	keys := MakeSessionKeys(macVersion, false)
	if fPort == 0 {
		frmPayload = MustEncryptUplink(*keys.NwkSEncKey.Key, devAddr, fCnt, false, frmPayload...)
	} else if len(fOpts) > 0 && macVersion.EncryptFOpts() {
		fOpts = MustEncryptUplink(*keys.NwkSEncKey.Key, devAddr, fCnt, true, fOpts...)
	}
	fhdr := ttnpb.FHDR{
		DevAddr: devAddr,
		FCtrl:   fCtrl,
		FCnt:    fCnt & 0xffff,
		FOpts:   fOpts,
	}
	phyPayload := append(
		append(
			test.Must(lorawan.AppendFHDR(
				test.Must(lorawan.AppendMHDR(nil, mhdr)).([]byte), fhdr, true),
			).([]byte),
			fPort),
		frmPayload...)
	var mic [4]byte
	switch {
	case macVersion.Compare(ttnpb.MAC_V1_1) < 0:
		mic = test.Must(crypto.ComputeLegacyUplinkMIC(*keys.FNwkSIntKey.Key, devAddr, fCnt, phyPayload)).([4]byte)
	default:
		if !fCtrl.Ack {
			confFCntDown = 0
		}
		mic = test.Must(crypto.ComputeUplinkMIC(*keys.SNwkSIntKey.Key, *keys.FNwkSIntKey.Key, confFCntDown, uint8(drIdx), chIdx, devAddr, fCnt, phyPayload)).([4]byte)
	}
	phyPayload = append(phyPayload, mic[:]...)
	msg := &ttnpb.UplinkMessage{
		CorrelationIDs: append([]string{}, DataUplinkCorrelationIDs[:]...),
		RawPayload:     phyPayload,
		RxMetadata:     deepcopy.Copy(mds).([]*ttnpb.RxMetadata),
		Settings:       MakeUplinkSettings(dr, freq),
	}
	if decodePayload {
		if frmPayload == nil {
			frmPayload = []byte{}
		}
		msg.Payload = &ttnpb.Message{
			MHDR: mhdr,
			MIC:  phyPayload[len(phyPayload)-4:],
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FHDR:       fhdr,
					FPort:      uint32(fPort),
					FRMPayload: frmPayload,
				},
			},
		}
	}
	return msg
}

func WithMatchedUplinkSettings(msg *ttnpb.UplinkMessage, chIdx uint8, drIdx ttnpb.DataRateIndex) *ttnpb.UplinkMessage {
	msg = CopyUplinkMessage(msg)
	msg.Settings.DataRateIndex = drIdx
	msg.DeviceChannelIndex = uint32(chIdx)
	return msg
}

func MustEncryptDownlink(key types.AES128Key, devAddr types.DevAddr, fCnt uint32, isFOpts bool, b ...byte) []byte {
	return test.Must(crypto.EncryptDownlink(key, devAddr, fCnt, b, isFOpts)).([]byte)
}

func MakeDataDownlink(macVersion ttnpb.MACVersion, confirmed bool, devAddr types.DevAddr, fCtrl ttnpb.FCtrl, fCnt, confFCntDown uint32, fPort uint8, frmPayload, fOpts []byte, txReq *ttnpb.TxRequest, cids ...string) *ttnpb.DownlinkMessage {
	if len(fOpts) > 0 && fPort == 0 {
		panic("FOpts must not be set for FPort == 0")
	}
	devAddr = *devAddr.Copy(&types.DevAddr{})
	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if confirmed {
		mType = ttnpb.MType_CONFIRMED_DOWN
	}
	mhdr := ttnpb.MHDR{
		MType: mType,
		Major: ttnpb.Major_LORAWAN_R1,
	}
	keys := MakeSessionKeys(macVersion, false)
	if fPort == 0 {
		frmPayload = MustEncryptDownlink(*keys.NwkSEncKey.Key, devAddr, fCnt, false, frmPayload...)
	} else if len(fOpts) > 0 && macVersion.EncryptFOpts() {
		fOpts = MustEncryptDownlink(*keys.NwkSEncKey.Key, devAddr, fCnt, true, fOpts...)
	}
	fhdr := ttnpb.FHDR{
		DevAddr: devAddr,
		FCtrl:   fCtrl,
		FCnt:    fCnt & 0xffff,
		FOpts:   fOpts,
	}
	phyPayload := append(
		append(
			test.Must(lorawan.AppendFHDR(
				test.Must(lorawan.AppendMHDR(nil, mhdr)).([]byte), fhdr, false),
			).([]byte),
			fPort),
		frmPayload...)
	var mic [4]byte
	switch {
	case macVersion.Compare(ttnpb.MAC_V1_1) < 0:
		mic = test.Must(crypto.ComputeLegacyDownlinkMIC(*keys.FNwkSIntKey.Key, devAddr, fCnt, phyPayload)).([4]byte)
	default:
		if !fCtrl.Ack {
			confFCntDown = 0
		}
		mic = test.Must(crypto.ComputeDownlinkMIC(*keys.SNwkSIntKey.Key, devAddr, confFCntDown, fCnt, phyPayload)).([4]byte)
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

func AssertLinkApplication(ctx context.Context, conn *grpc.ClientConn, getPeerCh <-chan test.ClusterGetPeerRequest, eventsPublishCh <-chan test.EventPubSubPublishRequest, appID ttnpb.ApplicationIdentifiers, replaceEvents ...events.Event) (ttnpb.AsNs_LinkApplicationClient, func(error) events.Event, bool) {
	t := test.MustTFromContext(ctx)
	t.Helper()

	a := assertions.New(t)

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var link ttnpb.AsNs_LinkApplicationClient
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		link, err = ttnpb.NewAsNsClient(conn).LinkApplication(
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
	if !a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
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

	if !a.So(test.AssertEventPubSubPublishRequests(ctx, eventsPublishCh, 1+len(replaceEvents), func(evs ...events.Event) bool {
		return a.So(evs, should.HaveSameElementsEvent, append(
			[]events.Event{EvtBeginApplicationLink(events.ContextWithCorrelationID(test.Context(), reqCIDs...), appID, nil)},
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
	return link, func(err error) events.Event {
		return EvtEndApplicationLink(events.ContextWithCorrelationID(test.Context(), reqCIDs...), appID, err)
	}, a.So(err, should.BeNil)
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

type ApplicationUplinkQueueEnvironment struct {
	Add       <-chan ApplicationUplinkQueueAddRequest
	Subscribe <-chan ApplicationUplinkQueueSubscribeRequest
}

func newMockApplicationUplinkQueue(t *testing.T) (ApplicationUplinkQueue, ApplicationUplinkQueueEnvironment, func()) {
	t.Helper()

	addCh := make(chan ApplicationUplinkQueueAddRequest)
	subscribeCh := make(chan ApplicationUplinkQueueSubscribeRequest)
	return &MockApplicationUplinkQueue{
			AddFunc:       MakeApplicationUplinkQueueAddChFunc(addCh),
			SubscribeFunc: MakeApplicationUplinkQueueSubscribeChFunc(subscribeCh),
		}, ApplicationUplinkQueueEnvironment{
			Add:       addCh,
			Subscribe: subscribeCh,
		},
		func() {
			select {
			case <-addCh:
				t.Error("ApplicationUplinkQueue.Add call missed")
			default:
				close(addCh)
			}
			select {
			case <-subscribeCh:
				t.Error("ApplicationUplinkQueue.Subscribe call missed")
			default:
				close(subscribeCh)
			}
		}
}

type DeviceRegistryEnvironment struct {
	GetByID     <-chan DeviceRegistryGetByIDRequest
	GetByEUI    <-chan DeviceRegistryGetByEUIRequest
	RangeByAddr <-chan DeviceRegistryRangeByAddrRequest
	SetByID     <-chan DeviceRegistrySetByIDRequest
}

func newMockDeviceRegistry(t *testing.T) (DeviceRegistry, DeviceRegistryEnvironment, func()) {
	t.Helper()

	getByEUICh := make(chan DeviceRegistryGetByEUIRequest)
	getByIDCh := make(chan DeviceRegistryGetByIDRequest)
	rangeByAddrCh := make(chan DeviceRegistryRangeByAddrRequest)
	setByIDCh := make(chan DeviceRegistrySetByIDRequest)
	return &MockDeviceRegistry{
			GetByEUIFunc:    MakeDeviceRegistryGetByEUIChFunc(getByEUICh),
			GetByIDFunc:     MakeDeviceRegistryGetByIDChFunc(getByIDCh),
			RangeByAddrFunc: MakeDeviceRegistryRangeByAddrChFunc(rangeByAddrCh),
			SetByIDFunc:     MakeDeviceRegistrySetByIDChFunc(setByIDCh),
		}, DeviceRegistryEnvironment{
			GetByEUI:    getByEUICh,
			RangeByAddr: rangeByAddrCh,
			SetByID:     setByIDCh,
		},
		func() {
			select {
			case <-getByEUICh:
				t.Error("DeviceRegistry.GetByEUI call missed")
			default:
				close(getByEUICh)
			}
			select {
			case <-getByIDCh:
				t.Error("DeviceRegistry.GetByID call missed")
			default:
				close(getByIDCh)
			}
			select {
			case <-rangeByAddrCh:
				t.Error("DeviceRegistry.RangeByAddr call missed")
			default:
				close(rangeByAddrCh)
			}
			select {
			case <-setByIDCh:
				t.Error("DeviceRegistry.SetByID call missed")
			default:
				close(setByIDCh)
			}
		}
}

type DownlinkTaskQueueEnvironment struct {
	Add <-chan DownlinkTaskAddRequest
	Pop <-chan DownlinkTaskPopRequest
}

func newMockDownlinkTaskQueue(t *testing.T) (DownlinkTaskQueue, DownlinkTaskQueueEnvironment, func()) {
	t.Helper()

	addCh := make(chan DownlinkTaskAddRequest)
	popCh := make(chan DownlinkTaskPopRequest)
	return &MockDownlinkTaskQueue{
			AddFunc: MakeDownlinkTaskAddChFunc(addCh),
			PopFunc: MakeDownlinkTaskPopChFunc(popCh),
		}, DownlinkTaskQueueEnvironment{
			Add: addCh,
			Pop: popCh,
		},
		func() {
			select {
			case <-addCh:
				t.Error("DownlinkTaskQueue.Add call missed")
			default:
				close(addCh)
			}
			select {
			case <-popCh:
				t.Error("DownlinkTaskQueue.Pop call missed")
			default:
				close(popCh)
			}
		}
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

type UplinkDeduplicatorEnvironment struct {
	DeduplicateUplink   <-chan UplinkDeduplicatorDeduplicateUplinkRequest
	AccumulatedMetadata <-chan UplinkDeduplicatorAccumulatedMetadataRequest
}

func newMockUplinkDeduplicator(t *testing.T) (UplinkDeduplicator, UplinkDeduplicatorEnvironment, func()) {
	t.Helper()

	deduplicateUplinkCh := make(chan UplinkDeduplicatorDeduplicateUplinkRequest)
	accumulatedMetadataCh := make(chan UplinkDeduplicatorAccumulatedMetadataRequest)
	return &MockUplinkDeduplicator{
			DeduplicateUplinkFunc:   MakeUplinkDeduplicatorDeduplicateUplinkChFunc(deduplicateUplinkCh),
			AccumulatedMetadataFunc: MakeUplinkDeduplicatorAccumulatedMetadataChFunc(accumulatedMetadataCh),
		}, UplinkDeduplicatorEnvironment{
			DeduplicateUplink:   deduplicateUplinkCh,
			AccumulatedMetadata: accumulatedMetadataCh,
		},
		func() {
			select {
			case <-deduplicateUplinkCh:
				t.Error("UplinkDeduplicator.DeduplicateUplink call missed")
			default:
				close(deduplicateUplinkCh)
			}
			select {
			case <-accumulatedMetadataCh:
				t.Error("UplinkDeduplicator.AccumulatedMetadata call missed")
			default:
				close(accumulatedMetadataCh)
			}
		}
}

type TestEnvironment struct {
	Cluster struct {
		Auth    <-chan test.ClusterAuthRequest
		GetPeer <-chan test.ClusterGetPeerRequest
	}
	ApplicationUplinks *ApplicationUplinkQueueEnvironment
	DeviceRegistry     *DeviceRegistryEnvironment
	DownlinkTasks      *DownlinkTaskQueueEnvironment
	Events             <-chan test.EventPubSubPublishRequest
	InteropClient      *InteropClientEnvironment
	UplinkDeduplicator *UplinkDeduplicatorEnvironment
}

func StartTest(t *testing.T, cmpConf component.Config, nsConf Config, timeout time.Duration, opts ...Option) (*NetworkServer, context.Context, TestEnvironment, func()) {
	t.Helper()

	authCh := make(chan test.ClusterAuthRequest)
	getPeerCh := make(chan test.ClusterGetPeerRequest)
	eventsPublishCh := make(chan test.EventPubSubPublishRequest)

	env := TestEnvironment{
		Events: eventsPublishCh,
	}
	env.Cluster.Auth = authCh
	env.Cluster.GetPeer = getPeerCh

	var closeFuncs []func()
	closeFuncs = append(closeFuncs, test.SetDefaultEventsPubSub(&test.MockEventPubSub{
		PublishFunc: test.MakeEventPubSubPublishChFunc(eventsPublishCh),
	}))
	if nsConf.ApplicationUplinks == nil {
		m, mEnv, closeM := newMockApplicationUplinkQueue(t)
		nsConf.ApplicationUplinks = m
		env.ApplicationUplinks = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}
	if nsConf.Devices == nil {
		m, mEnv, closeM := newMockDeviceRegistry(t)
		nsConf.Devices = m
		env.DeviceRegistry = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}
	if nsConf.DownlinkTasks == nil {
		m, mEnv, closeM := newMockDownlinkTaskQueue(t)
		nsConf.DownlinkTasks = m
		env.DownlinkTasks = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}
	if nsConf.UplinkDeduplicator == nil {
		m, mEnv, closeM := newMockUplinkDeduplicator(t)
		nsConf.UplinkDeduplicator = m
		env.UplinkDeduplicator = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}
	if nsConf.DeduplicationWindow == 0 {
		nsConf.DeduplicationWindow = time.Nanosecond
	}
	if nsConf.CooldownWindow == 0 {
		nsConf.CooldownWindow = nsConf.DeduplicationWindow + time.Nanosecond
	}

	ns := test.Must(New(
		componenttest.NewComponent(
			t,
			&cmpConf,
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
		),
		&nsConf,
		opts...,
	)).(*NetworkServer)
	ns.Component.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

	if ns.interopClient == nil {
		m, mEnv, closeM := newMockInteropClient(t)
		ns.interopClient = m
		env.InteropClient = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}

	componenttest.StartComponent(t, ns.Component)

	ctx := test.ContextWithT(test.Context(), t)
	ctx = log.NewContext(ctx, ns.Logger())
	ctx, cancel := context.WithTimeout(ctx, timeout)
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

func ForEachBand(t *testing.T, f func(func(...string) string, band.Band, ttnpb.PHYVersion)) {
	for phyID, phy := range band.All {
		for _, phyVersion := range phy.Versions() {
			phy, err := phy.Version(phyVersion)
			if err != nil {
				t.Errorf("Failed to convert %s band to %s version", phyID, phyVersion)
				continue
			}
			f(func(parts ...string) string {
				return MakeTestCaseName(append(parts, fmt.Sprintf("%s/PHY:%s", phyID, phyVersion.String()))...)
			}, phy, phyVersion)
		}
	}
}

func ForEachPHYVersion(f func(func(...string) string, ttnpb.PHYVersion)) {
	for _, phyVersion := range []ttnpb.PHYVersion{
		ttnpb.PHY_V1_0,
		ttnpb.PHY_V1_0_1,
		ttnpb.PHY_V1_0_2_REV_A,
		ttnpb.PHY_V1_0_2_REV_B,
		ttnpb.PHY_V1_0_3_REV_A,
		ttnpb.PHY_V1_1_REV_A,
		ttnpb.PHY_V1_1_REV_B,
	} {
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("PHY:%s", phyVersion.String()))...)
		}, phyVersion)
	}
}

func ForEachMACVersion(f func(func(...string) string, ttnpb.MACVersion)) {
	for _, macVersion := range []ttnpb.MACVersion{
		ttnpb.MAC_V1_0,
		ttnpb.MAC_V1_0_1,
		ttnpb.MAC_V1_0_2,
		ttnpb.MAC_V1_0_3,
		ttnpb.MAC_V1_0_4,
		ttnpb.MAC_V1_1,
	} {
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("MAC:%s", macVersion.String()))...)
		}, macVersion)
	}
}

func ForEachClass(f func(func(...string) string, ttnpb.Class)) {
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

func ForEachFrequencyPlan(t *testing.T, f func(func(...string) string, string, *frequencyplans.FrequencyPlan)) {
	fps := frequencyplans.NewStore(test.FrequencyPlansFetcher)
	fpIDs, err := fps.GetAllIDs()
	if err != nil {
		t.Errorf("failed to get frequency plans: %w", err)
		return
	}
	for _, fpID := range fpIDs {
		fp, err := fps.GetByID(fpID)
		if err != nil {
			t.Errorf("failed to get frequency plan `%s`: %w", fpID, err)
			continue
		}
		f(func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("FP:%s", fpID))...)
		}, fpID, fp)
	}
}

func ForEachPHYMACVersion(f func(func(...string) string, ttnpb.PHYVersion, ttnpb.MACVersion)) {
	ForEachPHYVersion(func(makePHYName func(...string) string, phyVersion ttnpb.PHYVersion) {
		ForEachMACVersion(func(makeMACName func(...string) string, macVersion ttnpb.MACVersion) {
			f(func(parts ...string) string {
				return makePHYName(makeMACName(parts...))
			}, phyVersion, macVersion)
		})
	})
}

func ForEachClassPHYMACVersion(f func(func(...string) string, ttnpb.Class, ttnpb.PHYVersion, ttnpb.MACVersion)) {
	ForEachClass(func(makeClassName func(...string) string, class ttnpb.Class) {
		ForEachPHYMACVersion(func(makePHYMACName func(parts ...string) string, phyVersion ttnpb.PHYVersion, macVersion ttnpb.MACVersion) {
			f(func(parts ...string) string {
				return makeClassName(makePHYMACName(parts...))
			}, class, phyVersion, macVersion)
		})
	})
}

func ForEachClassMACVersion(f func(func(...string) string, ttnpb.Class, ttnpb.MACVersion)) {
	ForEachClass(func(makeClassName func(...string) string, class ttnpb.Class) {
		ForEachMACVersion(func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
			f(func(parts ...string) string {
				return makeClassName(makeMACName(parts...))
			}, class, macVersion)
		})
	})
}

func ForEachFrequencyPlanBandMACVersion(t *testing.T, f func(func(...string) string, string, *frequencyplans.FrequencyPlan, band.Band, ttnpb.PHYVersion, ttnpb.MACVersion)) {
	ForEachFrequencyPlan(t, func(makeFPName func(...string) string, fpID string, fp *frequencyplans.FrequencyPlan) {
		phy, err := band.GetByID(fp.BandID)
		if err != nil {
			t.Errorf("failed to get PHY by id `%s` associated with frequency plan `%s`: %s", fp.BandID, fpID, err)
			return
		}
		for _, phyVersion := range phy.Versions() {
			phy, err := phy.Version(phyVersion)
			if err != nil {
				t.Errorf("Failed to convert band `%s` to version `%s`: %s", fp.BandID, phyVersion, err)
				continue
			}
			ForEachMACVersion(func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
				f(func(parts ...string) string {
					return makeFPName(makeMACName(append(parts, fmt.Sprintf("PHY:%s", phyVersion))...))
				}, fpID, fp, phy, phyVersion, macVersion)
			})
		}
	})
}

func ForEachBandMACVersion(t *testing.T, f func(func(...string) string, band.Band, ttnpb.PHYVersion, ttnpb.MACVersion)) {
	ForEachBand(t, func(makeBandName func(...string) string, phy band.Band, phyVersion ttnpb.PHYVersion) {
		ForEachMACVersion(func(makeMACName func(...string) string, macVersion ttnpb.MACVersion) {
			f(func(parts ...string) string {
				return makeBandName(makeMACName(parts...))
			}, phy, phyVersion, macVersion)
		})
	})
}

var redisNamespace = [...]string{
	"networkserver_test",
}

const (
	redisConsumerGroup = "ns"
	redisConsumerID    = "test"
)

func NewRedisApplicationUplinkQueue(t testing.TB) (ApplicationUplinkQueue, func() error) {
	cl, flush := test.NewRedis(t, append(redisNamespace[:], "application-uplinks")...)
	return redis.NewApplicationUplinkQueue(cl, 100, redisConsumerGroup, redisConsumerID),
		func() error {
			flush()
			return cl.Close()
		}
}

func NewRedisDeviceRegistry(t testing.TB) (DeviceRegistry, func() error) {
	cl, flush := test.NewRedis(t, append(redisNamespace[:], "devices")...)
	return &redis.DeviceRegistry{
			Redis: cl,
		},
		func() error {
			flush()
			return cl.Close()
		}
}

func NewRedisDownlinkTaskQueue(t testing.TB) (DownlinkTaskQueue, func() error) {
	a := assertions.New(t)

	cl, flush := test.NewRedis(t, append(redisNamespace[:], "downlink-tasks")...)
	q := redis.NewDownlinkTaskQueue(cl, 10000, redisConsumerGroup, redisConsumerID)
	err := q.Init()
	a.So(err, should.BeNil)

	ctx, cancel := context.WithCancel(test.Context())
	errCh := make(chan error, 1)
	go func() {
		t.Log("Running Redis downlink task queue...")
		err := q.Run(ctx)
		errCh <- err
		close(errCh)
		t.Logf("Stopped Redis downlink task queue with error: %s", err)
	}()
	return q,
		func() error {
			cancel()
			err := q.Add(ctx, ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test"},
			}, time.Now(), false)
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to add mock device to task queue: %s", err)
				return err
			}

			var runErr error
			select {
			case <-time.After(Timeout):
				t.Error("Timed out waiting for redis.DownlinkTaskQueue.Run to return")
			case runErr = <-errCh:
			}

			flush()
			closeErr := cl.Close()
			if runErr != nil && runErr != context.Canceled {
				return runErr
			}
			return closeErr
		}
}

func NewRedisUplinkDeduplicator(t testing.TB) (UplinkDeduplicator, func() error) {
	cl, flush := test.NewRedis(t, append(redisNamespace[:], "uplink-deduplication")...)
	return &redis.UplinkDeduplicator{
			Redis: cl,
		},
		func() error {
			flush()
			return cl.Close()
		}
}

func AllTrue(vs ...bool) bool {
	for _, v := range vs {
		if !v {
			return false
		}
	}
	return true
}

func LogEvents(t *testing.T, ch <-chan test.EventPubSubPublishRequest) {
	for ev := range ch {
		t.Logf("Event %s published with data %v", ev.Event.Name(), ev.Event.Data())
		ev.Response <- struct{}{}
	}
}
