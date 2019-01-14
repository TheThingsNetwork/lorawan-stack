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

package pfconfig

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func formatFrequency(frequency uint64) string {
	freq := float64(frequency) / 1000000
	if freq*10 == math.Floor(freq*10) {
		return fmt.Sprintf("%.1f", freq)
	}
	return fmt.Sprintf("%g", freq)
}

// Build builds a packet forwarder configuration for the given gateway, using the given frequency plan store.
func Build(gateway *ttnpb.Gateway, store *frequencyplans.Store) (*Config, error) {
	var c Config

	host, portStr, err := net.SplitHostPort(gateway.GatewayServerAddress)
	if err != nil {
		host = gateway.GatewayServerAddress
		portStr = "1700"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	c.GatewayConf.ServerAddress, c.GatewayConf.ServerPortUp, c.GatewayConf.ServerPortDown = host, uint32(port), uint32(port)
	server := c.GatewayConf
	server.Enabled = true
	c.GatewayConf.Servers = append(c.GatewayConf.Servers, server)

	frequencyPlan, err := store.GetByID(gateway.FrequencyPlanID)
	if err != nil {
		return nil, err
	}
	band, err := band.GetByID(frequencyPlan.BandID)
	if err != nil {
		return nil, err
	}

	c.SX1301Conf.LoRaWANPublic = true
	c.SX1301Conf.ClockSource = frequencyPlan.ClockSource

	if frequencyPlan.LBT != nil {
		lbtConfig := &LBTConfig{
			Enable:     true,
			RSSITarget: frequencyPlan.LBT.RSSITarget,
			RSSIOffset: frequencyPlan.LBT.RSSIOffset,
		}
		for i, channel := range frequencyPlan.DownlinkChannels {
			if i > 7 {
				break
			}
			lbtConfig.ChannelConfigs = append(lbtConfig.ChannelConfigs, LBTChannelConfig{
				Frequency:            channel.Frequency,
				ScanTimeMicroseconds: uint32(frequencyPlan.LBT.ScanTime / time.Microsecond),
			})
		}
		c.SX1301Conf.LBTConfig = lbtConfig
	}

	c.SX1301Conf.Radios = make([]RFConfig, len(frequencyPlan.Radios))
	for i, radio := range frequencyPlan.Radios {
		rfConfig := RFConfig{
			Enable:     radio.Enable,
			Type:       radio.ChipType,
			Frequency:  radio.Frequency,
			RSSIOffset: radio.RSSIOffset,
		}
		if radio.TxConfiguration != nil {
			rfConfig.TxEnable = true
			rfConfig.TxFreqMin = radio.TxConfiguration.MinFrequency
			rfConfig.TxFreqMax = radio.TxConfiguration.MaxFrequency
			if radio.TxConfiguration.NotchFrequency != nil {
				rfConfig.TxNotchFreq = *radio.TxConfiguration.NotchFrequency
			}
		}
		c.SX1301Conf.Radios[i] = rfConfig
	}

	numChannels := len(frequencyPlan.UplinkChannels)
	if numChannels < 8 {
		numChannels = 8
	}
	c.SX1301Conf.Channels = make([]IFConfig, numChannels)
	for i, channel := range frequencyPlan.UplinkChannels {
		ifConfig := IFConfig{
			Description: fmt.Sprintf("Lora MAC, 125kHz, all SF, %s MHz", formatFrequency(channel.Frequency)),
			Enable:      true,
			Radio:       channel.Radio,
			IFValue:     int32(int64(channel.Frequency) - int64(c.SX1301Conf.Radios[channel.Radio].Frequency)),
		}
		c.SX1301Conf.Channels[i] = ifConfig
	}

	c.SX1301Conf.LoRaStandardChannel = &IFConfig{Enable: false}
	if channel := frequencyPlan.LoRaStandardChannel; channel != nil {
		if lora := band.DataRates[channel.DataRate].Rate.GetLoRa(); lora != nil {
			c.SX1301Conf.LoRaStandardChannel = &IFConfig{
				Description:  fmt.Sprintf("Lora MAC, %dkHz, SF%d, %s MHz", lora.Bandwidth/1000, lora.SpreadingFactor, formatFrequency(channel.Frequency)),
				Enable:       true,
				Radio:        channel.Radio,
				IFValue:      int32(int64(channel.Frequency) - int64(c.SX1301Conf.Radios[channel.Radio].Frequency)),
				Bandwidth:    lora.Bandwidth,
				SpreadFactor: uint8(lora.SpreadingFactor),
			}
		}
	}

	c.SX1301Conf.FSKChannel = &IFConfig{Enable: false}
	if channel := frequencyPlan.FSKChannel; channel != nil {
		if fsk := band.DataRates[channel.DataRate].Rate.GetFSK(); fsk != nil {
			c.SX1301Conf.FSKChannel = &IFConfig{
				Description: fmt.Sprintf("FSK %dkbps, %s MHz", fsk.BitRate/1000, formatFrequency(channel.Frequency)),
				Enable:      true,
				Radio:       channel.Radio,
				IFValue:     int32(int64(channel.Frequency) - int64(c.SX1301Conf.Radios[channel.Radio].Frequency)),
				Bandwidth:   125000,
				Datarate:    fsk.BitRate,
			}
		}
	}

	c.SX1301Conf.TxLUTConfigs = defaultTxLUTConfigs

	return &c, nil
}
