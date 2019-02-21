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

package messages

import (
	"bytes"
	"encoding/binary"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errDataRateIndex = errors.Define("data_rate_index", "data rate index is out of range")

func getInt32AsByteSlice(value int32) ([]byte, error) {
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, value)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func getDataRateFromIndex(bandID string, index int) (ttnpb.DataRate, error) {
	band, err := band.GetByID(bandID)
	if err != nil {
		return ttnpb.DataRate{}, errDataRateIndex.WithCause(err)
	}
	if index >= len(band.DataRates) {
		return ttnpb.DataRate{}, errDataRateIndex
	}
	return band.DataRates[index].Rate, nil
}

// getDataRatesFromBandID parses the available data rates from the band into DataRates.
func getDataRatesFromBandID(id string) (DataRates, error) {
	band, err := band.GetByID(id)
	if err != nil {
		return DataRates{}, err
	}

	// Set the default values
	drs := DataRates{}
	for _, dr := range drs {
		dr[0] = -1
		dr[1] = 0
		dr[2] = 0
	}

	var i = 0
	for _, dr := range band.DataRates {
		if loraDR := dr.Rate.GetLoRa(); loraDR != nil {
			loraDR.GetSpreadingFactor()
			drs[i][0] = int(loraDR.GetSpreadingFactor())
			drs[i][1] = int(loraDR.GetBandwidth() / 1000)
			i++
		} else if fskDR := dr.Rate.GetFSK(); fskDR != nil {
			drs[i][0] = 0 // must be set to 0 for FSK, the BW field is ignored.
			i++
		}
	}
	return drs, nil
}
