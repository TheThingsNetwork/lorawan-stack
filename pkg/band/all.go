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
var All = map[string]map[ttnpb.PHYVersion]Band{
	AS_923: {
		ttnpb.RP001_V1_0_2:       AS_923_RP1_v1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: AS_923_RP1_v1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: AS_923_RP1_v1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   AS_923_RP1_v1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   AS_923_RP1_v1_1_RevB,
	},
	AU_915_928: {
		ttnpb.TS001_V1_0_1:       AU_915_928_TS1_v1_0_1,
		ttnpb.RP001_V1_0_2:       AU_915_928_RP1_v1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: AU_915_928_RP1_v1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: AU_915_928_RP1_v1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   AU_915_928_RP1_v1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   AU_915_928_RP1_v1_1_RevB,
	},
	CN_470_510: {
		ttnpb.TS001_V1_0_1:       CN_470_510_TS1_v1_0_1,
		ttnpb.RP001_V1_0_2:       CN_470_510_RP1_v1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: CN_470_510_RP1_v1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: CN_470_510_RP1_v1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   CN_470_510_RP1_v1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   CN_470_510_RP1_v1_1_RevB,
	},
	CN_779_787: {
		ttnpb.TS001_V1_0:         CN_779_787_RP1_V1_0,
		ttnpb.TS001_V1_0_1:       CN_779_787_RP1_V1_0_1,
		ttnpb.RP001_V1_0_2:       CN_779_787_RP1_V1_0_2,
		ttnpb.RP001_V1_0_2_REV_B: CN_779_787_RP1_V1_0_2_RevB,
		ttnpb.RP001_V1_0_3_REV_A: CN_779_787_RP1_V1_0_3_RevA,
		ttnpb.RP001_V1_1_REV_A:   CN_779_787_RP1_V1_1_RevA,
		ttnpb.RP001_V1_1_REV_B:   CN_779_787_RP1_V1_1_RevB,
	},
}

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
