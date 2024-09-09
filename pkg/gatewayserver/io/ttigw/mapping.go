// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package ttigw

import (
	"time"

	lorav1 "go.thethings.industries/pkg/api/gen/tti/gateway/data/lora/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ieee"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	errInvalidBoard                = errors.DefineInvalidArgument("invalid_board", "invalid board `{board}`")
	errInvalidIFChain              = errors.DefineInvalidArgument("invalid_if_chain", "invalid IF chain `{if_chain}`")
	errInvalidModulation           = errors.DefineInvalidArgument("invalid_modulation", "invalid modulation")
	errUnsupportedDownlinkDataRate = errors.DefineInvalidArgument(
		"unsupported_downlink_data_rate", "unsupported downlink data rate index `{data_rate_index}` in channel `{channel}`",
	)
	errDownlinkChannelMixedBandwidths = errors.DefineInvalidArgument(
		"downlink_channel_mixed_bandwidths",
		"downlink channel `{channel}` has mixed bandwidths `{bandwidth_low}` and `{bandwidth_high}` Hz",
	)
	errNotScheduled = errors.DefineInvalidArgument("not_scheduled", "downlink message not scheduled")
)

const eirpDelta = 2.15

func gatewayStatusFromClientHello(clientHello *lorav1.ClientHelloNotification) *ttnpb.GatewayStatus {
	advanced := map[string]*structpb.Value{
		"model": structpb.NewStringValue(clientHello.DeviceModel),
	}
	if manufacturer, ok := ieee.OUI[clientHello.DeviceManufacturer]; ok {
		advanced["manufacturer"] = structpb.NewStringValue(manufacturer)
	}
	res := &ttnpb.GatewayStatus{
		Versions: map[string]string{
			"firmware": clientHello.FirmwareVersion,
			"hardware": clientHello.HardwareVersion,
			"runtime":  clientHello.RuntimeVersion,
		},
		Advanced: &structpb.Struct{
			Fields: advanced,
		},
	}
	if clientHello.Uptime != nil {
		res.BootTime = timestamppb.New(time.Now().Add(-clientHello.Uptime.AsDuration()))
	}
	return res
}

var (
	toBandwidth = map[lorav1.Bandwidth]uint32{
		lorav1.Bandwidth_BANDWIDTH_125_KHZ: 125000,
		lorav1.Bandwidth_BANDWIDTH_250_KHZ: 250000,
		lorav1.Bandwidth_BANDWIDTH_500_KHZ: 500000,
	}
	fromBandwidth = map[uint32]lorav1.Bandwidth{
		125000: lorav1.Bandwidth_BANDWIDTH_125_KHZ,
		250000: lorav1.Bandwidth_BANDWIDTH_250_KHZ,
		500000: lorav1.Bandwidth_BANDWIDTH_500_KHZ,
	}
)

func buildLoRaGatewayConfig(fp *frequencyplans.FrequencyPlan) (*lorav1.GatewayConfig, error) {
	phy, err := band.GetLatest(fp.BandID)
	if err != nil {
		return nil, err
	}

	var (
		board = &lorav1.Board{
			Ifs: &lorav1.Board_IntermediateFrequencies{},
		}
		tx = make([]*lorav1.TransmitChannel, 0, 16)
	)

	for i, rfChain := range []**lorav1.Board_RFChain{&board.RfChain0, &board.RfChain1} {
		if i >= len(fp.Radios) {
			break
		}
		radio := fp.Radios[i]
		if !radio.Enable {
			continue
		}
		*rfChain = &lorav1.Board_RFChain{
			Frequency: radio.Frequency,
		}
	}

	for i, multiSF := range []**lorav1.Board_IntermediateFrequencies_MultipleSF{
		&board.Ifs.MultipleSf0,
		&board.Ifs.MultipleSf1,
		&board.Ifs.MultipleSf2,
		&board.Ifs.MultipleSf3,
		&board.Ifs.MultipleSf4,
		&board.Ifs.MultipleSf5,
		&board.Ifs.MultipleSf6,
		&board.Ifs.MultipleSf7,
	} {
		if i >= len(fp.UplinkChannels) {
			break
		}
		*multiSF = &lorav1.Board_IntermediateFrequencies_MultipleSF{
			RfChain:   uint32(fp.UplinkChannels[i].Radio),
			Frequency: int32(int64(fp.UplinkChannels[i].Frequency) - int64(fp.Radios[fp.UplinkChannels[i].Radio].Frequency)),
		}
	}
	if fp.FSKChannel != nil {
		if dataRate := phy.DataRates[ttnpb.DataRateIndex(fp.FSKChannel.DataRate)].Rate.GetFsk(); dataRate != nil {
			board.Ifs.Fsk = &lorav1.Board_IntermediateFrequencies_FSK{
				RfChain:   uint32(fp.FSKChannel.Radio),
				Frequency: int32(int64(fp.FSKChannel.Frequency) - int64(fp.Radios[fp.FSKChannel.Radio].Frequency)),
				Bitrate:   dataRate.BitRate,
				Bandwidth: lorav1.Bandwidth_BANDWIDTH_125_KHZ,
			}
		}
	}
	if fp.LoRaStandardChannel != nil {
		if dataRate := phy.DataRates[ttnpb.DataRateIndex(fp.LoRaStandardChannel.DataRate)].Rate.GetLora(); dataRate != nil {
			board.Ifs.LoraServiceChannel = &lorav1.Board_IntermediateFrequencies_LoraServiceChannel{
				RfChain: uint32(fp.LoRaStandardChannel.Radio),
				Frequency: int32(
					int64(fp.LoRaStandardChannel.Frequency) - int64(fp.Radios[fp.LoRaStandardChannel.Radio].Frequency),
				),
				SpreadingFactor: dataRate.SpreadingFactor,
				Bandwidth:       fromBandwidth[dataRate.Bandwidth],
			}
		}
	}

	for i, ch := range fp.DownlinkChannels {
		if i == 16 {
			break
		}
		// To minimize protocol messages, it is assumed that a transmission channel uses a single bandwidth.
		drLow := phy.DataRates[ttnpb.DataRateIndex(ch.MinDataRate)].Rate.GetLora()
		if drLow == nil {
			return nil, errUnsupportedDownlinkDataRate.WithAttributes(
				"data_rate_index", ch.MinDataRate,
				"channel", i,
			)
		}
		drHigh := phy.DataRates[ttnpb.DataRateIndex(ch.MaxDataRate)].Rate.GetLora()
		if drHigh == nil {
			return nil, errUnsupportedDownlinkDataRate.WithAttributes(
				"data_rate_index", ch.MaxDataRate,
				"channel", i,
			)
		}
		if drLow.Bandwidth != drHigh.Bandwidth {
			return nil, errDownlinkChannelMixedBandwidths.WithAttributes(
				"channel", i,
				"bandwidth_low", drLow.Bandwidth,
				"bandwidth_high", drHigh.Bandwidth,
			)
		}
		tx = append(tx, &lorav1.TransmitChannel{
			Frequency: ch.Frequency,
			Bandwidth: fromBandwidth[drLow.Bandwidth],
		})
	}
	// Add the default RX2 frequency and bandwidth if it fits.
	// The frequency plan's overrides are only relevant to end devices, so only the default RX2 parameters are considered.
	if rx2DR := phy.DataRates[phy.DefaultRx2Parameters.DataRateIndex].Rate.GetLora(); len(tx) < 16 && rx2DR != nil {
		tx = append(tx, &lorav1.TransmitChannel{
			Frequency: phy.DefaultRx2Parameters.Frequency,
			Bandwidth: fromBandwidth[rx2DR.Bandwidth],
		})
	}

	return &lorav1.GatewayConfig{
		Boards: []*lorav1.Board{board},
		Tx:     tx,
	}, nil
}

var (
	toCodingRate = map[lorav1.CodeRate]string{
		lorav1.CodeRate_CODE_RATE_4_5: "4/5",
		lorav1.CodeRate_CODE_RATE_4_6: "4/6",
		lorav1.CodeRate_CODE_RATE_4_7: "4/7",
		lorav1.CodeRate_CODE_RATE_4_8: "4/8",
	}
	fromCodingRate = map[string]lorav1.CodeRate{
		"4/5": lorav1.CodeRate_CODE_RATE_4_5,
		"4/6": lorav1.CodeRate_CODE_RATE_4_6,
		"4/7": lorav1.CodeRate_CODE_RATE_4_7,
		"4/8": lorav1.CodeRate_CODE_RATE_4_8,
	}
)

func toUplinkMessage(
	ids *ttnpb.GatewayIdentifiers, gtwConf *lorav1.GatewayConfig, msg *lorav1.UplinkMessage,
) (*ttnpb.UplinkMessage, error) {
	if msg.Board != 0 {
		return nil, errInvalidBoard.WithAttributes("board", msg.Board)
	}
	board := gtwConf.Boards[0]
	var (
		rfChain         uint32
		frequencyOffset int32
		dataRate        = &ttnpb.DataRate{}
		rxMetadata      = &ttnpb.RxMetadata{
			GatewayIds:  ids,
			Timestamp:   msg.Timestamp,
			Rssi:        -msg.RssiChannelNegated,
			ChannelRssi: -msg.RssiChannelNegated,
		}
	)
	switch {
	case msg.IfChain < 8: // LoRa multi-SF
		multipleSF := []*lorav1.Board_IntermediateFrequencies_MultipleSF{
			board.Ifs.MultipleSf0,
			board.Ifs.MultipleSf1,
			board.Ifs.MultipleSf2,
			board.Ifs.MultipleSf3,
			board.Ifs.MultipleSf4,
			board.Ifs.MultipleSf5,
			board.Ifs.MultipleSf6,
			board.Ifs.MultipleSf7,
		}[msg.IfChain]
		rfChain = multipleSF.RfChain
		frequencyOffset = multipleSF.Frequency
		modulation := msg.GetLora()
		if modulation == nil {
			return nil, errInvalidModulation.New()
		}
		dataRate.Modulation = &ttnpb.DataRate_Lora{
			Lora: &ttnpb.LoRaDataRate{
				SpreadingFactor: modulation.SpreadingFactor,
				Bandwidth:       125000,
				CodingRate:      toCodingRate[modulation.CodeRate],
			},
		}
		rxMetadata.SignalRssi = wrapperspb.Float(-modulation.RssiSignalNegated)
		rxMetadata.FrequencyOffset = int64(modulation.FrequencyOffset)
		switch snrAbs := modulation.Snr.(type) {
		case *lorav1.UplinkMessage_Lora_SnrPositive:
			rxMetadata.Snr = snrAbs.SnrPositive
		case *lorav1.UplinkMessage_Lora_SnrNegative:
			rxMetadata.Snr = -snrAbs.SnrNegative
		}

	case msg.IfChain == 8: // FSK
		rfChain = board.Ifs.Fsk.RfChain
		frequencyOffset = board.Ifs.Fsk.Frequency
		dataRate.Modulation = &ttnpb.DataRate_Fsk{
			Fsk: &ttnpb.FSKDataRate{
				BitRate: board.Ifs.Fsk.Bitrate,
			},
		}

	case msg.IfChain == 9: // LoRa standard channel
		rfChain = board.Ifs.LoraServiceChannel.RfChain
		frequencyOffset = board.Ifs.LoraServiceChannel.Frequency
		modulation := msg.GetLora()
		if modulation == nil {
			return nil, errInvalidModulation.New()
		}
		dataRate.Modulation = &ttnpb.DataRate_Lora{
			Lora: &ttnpb.LoRaDataRate{
				SpreadingFactor: board.Ifs.LoraServiceChannel.SpreadingFactor,
				Bandwidth:       toBandwidth[board.Ifs.LoraServiceChannel.Bandwidth],
				CodingRate:      toCodingRate[modulation.CodeRate],
			},
		}
		rxMetadata.SignalRssi = wrapperspb.Float(-modulation.RssiSignalNegated)
		rxMetadata.FrequencyOffset = int64(modulation.FrequencyOffset)
		switch snrAbs := modulation.Snr.(type) {
		case *lorav1.UplinkMessage_Lora_SnrPositive:
			rxMetadata.Snr = snrAbs.SnrPositive
		case *lorav1.UplinkMessage_Lora_SnrNegative:
			rxMetadata.Snr = -snrAbs.SnrNegative
		}

	default:
		return nil, errInvalidIFChain.WithAttributes("if_chain", msg.IfChain)
	}

	centerFrequency := []*lorav1.Board_RFChain{board.RfChain0, board.RfChain1}[rfChain].Frequency
	frequency := uint64(int64(centerFrequency) + int64(frequencyOffset))
	return &ttnpb.UplinkMessage{
		RawPayload: msg.Payload,
		Settings: &ttnpb.TxSettings{
			DataRate:  dataRate,
			Frequency: frequency,
			Timestamp: msg.Timestamp,
		},
		RxMetadata: []*ttnpb.RxMetadata{rxMetadata},
	}, nil
}

func fromDownlinkMessage(gtwConf *lorav1.GatewayConfig, msg *ttnpb.DownlinkMessage) (*lorav1.DownlinkMessage, error) {
	scheduled := msg.GetScheduled()
	if scheduled == nil || scheduled.Downlink == nil {
		return nil, errNotScheduled.New()
	}
	res := &lorav1.DownlinkMessage{
		TxPower:   uint32(scheduled.Downlink.TxPower - eirpDelta),
		Timestamp: scheduled.Timestamp,
		Payload:   msg.RawPayload,
	}
	if dr := scheduled.DataRate.GetLora(); dr != nil {
		// Only LoRa transmission channels are configured via gateway config so that TxChannelIndex can be used.
		// FSK is always configured via TxChannelConfig, see below.
		for i, ch := range gtwConf.Tx {
			if ch.Frequency == scheduled.Frequency && ch.Bandwidth == fromBandwidth[dr.Bandwidth] {
				res.TxChannel = &lorav1.DownlinkMessage_TxChannelIndex{
					TxChannelIndex: uint32(i),
				}
				break
			}
		}
	}
	switch mod := scheduled.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_Lora:
		if res.TxChannel == nil {
			res.TxChannel = &lorav1.DownlinkMessage_TxChannelConfig{
				TxChannelConfig: &lorav1.TransmitChannel{
					Frequency: scheduled.Frequency,
					Bandwidth: fromBandwidth[mod.Lora.Bandwidth],
				},
			}
		}
		res.DataRate = &lorav1.DownlinkMessage_Lora_{
			Lora: &lorav1.DownlinkMessage_Lora{
				SpreadingFactor: mod.Lora.SpreadingFactor,
				CodeRate:        fromCodingRate[mod.Lora.CodingRate],
				LorawanUplink:   scheduled.EnableCrc && !scheduled.Downlink.InvertPolarization,
			},
		}
	case *ttnpb.DataRate_Fsk:
		if res.TxChannel == nil {
			res.TxChannel = &lorav1.DownlinkMessage_TxChannelConfig{
				TxChannelConfig: &lorav1.TransmitChannel{
					Frequency: scheduled.Frequency,
					Bandwidth: lorav1.Bandwidth_BANDWIDTH_125_KHZ,
				},
			}
		}
		res.DataRate = &lorav1.DownlinkMessage_Fsk{
			Fsk: &lorav1.DownlinkMessage_FSK{
				Bitrate:            mod.Fsk.BitRate,
				FrequencyDeviation: mod.Fsk.BitRate / 2 / 1000,
			},
		}
	default:
		return nil, errInvalidModulation.New()
	}
	return res, nil
}

var toTxAcknowledgmentResult = map[lorav1.ErrorCode]ttnpb.TxAcknowledgment_Result{
	lorav1.ErrorCode_ERROR_CODE_TX_TOO_LATE:  ttnpb.TxAcknowledgment_TOO_LATE,
	lorav1.ErrorCode_ERROR_CODE_TX_TOO_EARLY: ttnpb.TxAcknowledgment_TOO_EARLY,
	lorav1.ErrorCode_ERROR_CODE_TX_FREQUENCY: ttnpb.TxAcknowledgment_TX_FREQ,
	lorav1.ErrorCode_ERROR_CODE_TX_POWER:     ttnpb.TxAcknowledgment_TX_POWER,
}
