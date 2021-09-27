// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// All contains all the bands available.
var All = map[string]map[ttnpb.PHYVersion]Band{}

// Get returns the band if it was found, and returns an error otherwise.
func Get(id string, version ttnpb.PHYVersion) (Band, error) {
	versions, ok := All[id]
	if !ok {
		return Band{}, errBandNotFound.WithAttributes("id", id, "version", version)
	}
	band, ok := versions[version]
	if !ok {
		return Band{}, errBandNotFound.WithAttributes("id", id, "version", version)
	}
	return band, nil
}

const latestSupportedVersion = ttnpb.RP001_V1_1_REV_B

// GetLatest returns the latest version of the band if it was found,
// and returns an error otherwise.
func GetLatest(id string) (Band, error) {
	return Get(id, latestSupportedVersion)
}
