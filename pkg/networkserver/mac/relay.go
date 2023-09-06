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

package mac

import (
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func secondChFields(secondCh *ttnpb.RelaySecondChannel) []any {
	if secondCh == nil {
		return nil
	}
	return []any{
		"relay_second_ch_ack_offset", secondCh.AckOffset,
		"relay_second_ch_data_rate_index", secondCh.DataRateIndex,
		"relay_second_ch_frequency", secondCh.Frequency,
	}
}

func servingRelayFields(serving *ttnpb.ServingRelayParameters) log.Fielder {
	if serving == nil {
		return log.Fields()
	}
	return log.Fields(
		append(
			secondChFields(serving.SecondChannel),
			"relay_default_ch_index", serving.DefaultChannelIndex,
			"relay_cad_periodicity", serving.CadPeriodicity,
		)...,
	)
}
