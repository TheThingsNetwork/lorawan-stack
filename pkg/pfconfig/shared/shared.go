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

// Package shared contains the configuration that is common to various gateway types.
package shared

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var defaultTxLUTConfigs = []*ttnpb.TxLUTConfig{
	{PaGain: 0, MixGain: 8, RfPower: -6},
	{PaGain: 0, MixGain: 10, RfPower: -3},
	{PaGain: 0, MixGain: 12, RfPower: 0},
	{PaGain: 1, MixGain: 8, RfPower: 3},
	{PaGain: 1, MixGain: 10, RfPower: 6},
	{PaGain: 1, MixGain: 12, RfPower: 10},
	{PaGain: 1, MixGain: 13, RfPower: 11},
	{PaGain: 2, MixGain: 9, RfPower: 12},
	{PaGain: 1, MixGain: 15, RfPower: 13},
	{PaGain: 2, MixGain: 10, RfPower: 14},
	{PaGain: 2, MixGain: 11, RfPower: 16},
	{PaGain: 3, MixGain: 9, RfPower: 20},
	{PaGain: 3, MixGain: 10, RfPower: 23},
	{PaGain: 3, MixGain: 11, RfPower: 25},
	{PaGain: 3, MixGain: 12, RfPower: 26},
	{PaGain: 3, MixGain: 14, RfPower: 27},
}

// BuildSX1301Config builds the SX1301 configuration for the given frequency plan.
func BuildSX1301Config(frequencyPlan *frequencyplans.FrequencyPlan) (*ttnpb.SX1301Config, error) {
	phy, err := band.GetLatest(frequencyPlan.BandID)
	if err != nil {
		return nil, err
	}

	conf := new(ttnpb.SX1301Config)

	conf.LorawanPublic = true
	conf.ClockSource = uint32(frequencyPlan.ClockSource)

	if frequencyPlan.LBT != nil {
		lbtConfig := &ttnpb.LBTConfig{
			Enable:     true,
			RssiTarget: frequencyPlan.LBT.RSSITarget,
			RssiOffset: frequencyPlan.LBT.RSSIOffset,
		}
		for i, channel := range frequencyPlan.DownlinkChannels {
			if i > 7 {
				break
			}
			lbtConfig.ChannelConfigs = append(
				lbtConfig.ChannelConfigs,
				&ttnpb.LBTChannelConfig{
					Frequency:            channel.Frequency,
					ScanTimeMicroseconds: uint32(frequencyPlan.LBT.ScanTime / time.Microsecond),
				},
			)
		}
		conf.LbtConfig = lbtConfig
	}

	conf.Radios = make([]*ttnpb.RFConfig, len(frequencyPlan.Radios))
	for i, radio := range frequencyPlan.Radios {
		rfConfig := &ttnpb.RFConfig{
			Enable:     radio.Enable,
			Type:       radio.ChipType,
			Frequency:  radio.Frequency,
			RssiOffset: radio.RSSIOffset,
		}
		if radio.TxConfiguration != nil {
			rfConfig.TxEnable = true
			rfConfig.TxFreqMin = radio.TxConfiguration.MinFrequency
			rfConfig.TxFreqMax = radio.TxConfiguration.MaxFrequency
			if radio.TxConfiguration.NotchFrequency != nil {
				rfConfig.TxNotchFreq = *radio.TxConfiguration.NotchFrequency
			}
		}
		conf.Radios[i] = rfConfig
	}

	numChannels := len(frequencyPlan.UplinkChannels)
	if numChannels < 8 {
		numChannels = 8
	}
	conf.Channels = make([]*ttnpb.IFConfig, numChannels)
	for i, channel := range frequencyPlan.UplinkChannels {
		ifConfig := &ttnpb.IFConfig{
			Enable:  true,
			Radio:   uint32(channel.Radio),
			IfValue: int32(int64(channel.Frequency) - int64(conf.Radios[channel.Radio].Frequency)),
		}
		conf.Channels[i] = ifConfig
	}

	conf.LoraStandardChannel = &ttnpb.IFConfig{Enable: false}
	if channel := frequencyPlan.LoRaStandardChannel; channel != nil {
		dr, ok := phy.DataRates[ttnpb.DataRateIndex(channel.DataRate)]
		if ok {
			if lora := dr.Rate.GetLora(); lora != nil {
				conf.LoraStandardChannel = &ttnpb.IFConfig{
					Enable:       true,
					Radio:        uint32(channel.Radio),
					IfValue:      int32(int64(channel.Frequency) - int64(conf.Radios[channel.Radio].Frequency)),
					Bandwidth:    lora.Bandwidth,
					SpreadFactor: lora.SpreadingFactor,
				}
			}
		}
	}

	conf.FskChannel = &ttnpb.IFConfig{Enable: false}
	if channel := frequencyPlan.FSKChannel; channel != nil {
		dr, ok := phy.DataRates[ttnpb.DataRateIndex(channel.DataRate)]
		if ok {
			if fsk := dr.Rate.GetFsk(); fsk != nil {
				conf.FskChannel = &ttnpb.IFConfig{
					Enable:    true,
					Radio:     uint32(channel.Radio),
					IfValue:   int32(int64(channel.Frequency) - int64(conf.Radios[channel.Radio].Frequency)),
					Bandwidth: 125000,
					Datarate:  fsk.BitRate,
				}
			}
		}
	}

	conf.TxLutConfigs = defaultTxLUTConfigs

	return conf, nil
}

// DefaultGatewayServerUDPPort is the default port used for connecting to Gateway Server.
const DefaultGatewayServerUDPPort = 1700

var (
	errEmptyGatewayServerAddress   = errors.DefineInvalidArgument("empty_gateway_server_address", "gateway server address is empty")
	errInvalidGatewayServerAddress = errors.DefineInvalidArgument("invalid_gateway_server_address", "gateway server address is invalid")
)

// ParseGatewayServerAddress parses gateway server address s into hostname and port,
// port is equal to the port contained in s or DefaultGatewayServerUDPPort otherwise.
func ParseGatewayServerAddress(s string) (string, uint16, error) {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		if host, _, err := net.SplitHostPort(fmt.Sprintf("%s:1700", s)); err == nil && host != "" {
			return host, 1700, nil
		}
		return "", 0, errInvalidGatewayServerAddress.WithCause(err)
	}
	if host == "" {
		return "", 0, errInvalidGatewayServerAddress.WithCause(errEmptyGatewayServerAddress)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, errInvalidGatewayServerAddress.WithCause(err)
	}
	return host, uint16(port), nil
}
