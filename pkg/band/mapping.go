// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package band

import (
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
)

var errDataRateNotFound = errors.DefineNotFound(
	"data_rate_not_found", "data rate `{data_rate_index}` not found",
)

// MapDataRateIndex maps a data rate index between two bands.
// Note that if multiple data rate indices may map to the provided
// data rate index, the lowest one is returned, except when
// the data rate index is equivalent between the two bands.
// See Band.FindUplinkDataRate for more details on why this matters.
func MapDataRateIndex(
	sourceBand *Band, sourceDataRateIndex ttnpb.DataRateIndex, targetBand *Band,
) (targetDataRateIndex ttnpb.DataRateIndex, err error) {
	sourceDataRate, ok := sourceBand.DataRates[sourceDataRateIndex]
	if !ok {
		return 0, errDataRateNotFound.WithAttributes("data_rate_index", sourceDataRateIndex)
	}
	// Fast path: check if the index is equivalent between the two bands.
	targetDataRate, ok := targetBand.DataRates[sourceDataRateIndex]
	if ok && proto.Equal(sourceDataRate.Rate, targetDataRate.Rate) {
		return sourceDataRateIndex, nil
	}
	// Slow path: scan the target band for the indices.
	for i := ttnpb.DataRateIndex_DATA_RATE_0; i <= ttnpb.DataRateIndex_DATA_RATE_15; i++ {
		targetDataRate, ok := targetBand.DataRates[i]
		if ok && proto.Equal(sourceDataRate.Rate, targetDataRate.Rate) {
			return i, nil
		}
	}
	return 0, errDataRateNotFound.WithAttributes("data_rate_index", sourceDataRateIndex)
}
