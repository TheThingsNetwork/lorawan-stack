// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func validateRelaySecondChannel(secondCh *ttnpb.RelaySecondChannel, phy *band.Band, path ...string) error {
	if secondCh == nil {
		return nil
	}
	if _, ok := phy.DataRates[secondCh.DataRateIndex]; !ok {
		return newInvalidFieldValueError(strings.Join(append(path, "data_rate_index"), "."))
	}
	inSubBand := false
	for _, sb := range phy.SubBands {
		if sb.MinFrequency >= secondCh.Frequency && secondCh.Frequency <= sb.MaxFrequency {
			inSubBand = true
			break
		}
	}
	if !inSubBand {
		return newInvalidFieldValueError(strings.Join(append(path, "frequency"), "."))
	}
	return nil
}

func validateRelayConfigurationServed(served *ttnpb.ServedRelayParameters, phy *band.Band, path ...string) error {
	if served == nil {
		return nil
	}
	if err := validateRelaySecondChannel(served.SecondChannel, phy, append(path, "second_channel")...); err != nil {
		return err
	}
	return nil
}

func validateDefaultChannelIndex(index uint32, phy *band.Band, path ...string) error {
	if index >= uint32(len(phy.Relay.WORChannels)) {
		return newInvalidFieldValueError(strings.Join(append(path, "default_channel_index"), "."))
	}
	return nil
}

func validateRelayConfigurationServing(serving *ttnpb.ServingRelayParameters, phy *band.Band, path ...string) error {
	if serving == nil {
		return nil
	}
	if err := validateRelaySecondChannel(serving.SecondChannel, phy, append(path, "second_channel")...); err != nil {
		return err
	}
	if err := validateDefaultChannelIndex(
		serving.DefaultChannelIndex, phy, append(path, "default_channel_index")...,
	); err != nil {
		return err
	}
	return nil
}

func validateRelayConfiguration(conf *ttnpb.RelayParameters, phy *band.Band, path ...string) error {
	if conf == nil {
		return nil
	}
	switch mode := conf.Mode.(type) {
	case *ttnpb.RelayParameters_Served:
		return validateRelayConfigurationServed(mode.Served, phy, append(path, "mode", "served")...)
	case *ttnpb.RelayParameters_Serving:
		return validateRelayConfigurationServing(mode.Serving, phy, append(path, "mode", "serving")...)
	case nil:
		return nil
	default:
		panic(fmt.Sprintf("unknown mode %T", mode))
	}
}
