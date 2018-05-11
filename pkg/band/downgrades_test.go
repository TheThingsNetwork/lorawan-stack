// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package band_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func compatibleVerifier(t *testing.T) func(version ttnpb.PHYVersion, versionName string, bandIDs ...string) {
	return func(version ttnpb.PHYVersion, versionName string, bandIDs ...string) {
		for _, id := range bandIDs {
			b, err := band.GetByID(id)
			if err != nil {
				t.Fatalf("Could not retrieve band %s: %s\n", id, err)
			}

			if _, err = b.Version(version); err != nil {
				t.Fatalf("Band %s does not support intended LoRaWAN Regional Parameters version %s\n", b.ID, versionName)
			}
		}
	}
}

func TestBands(t *testing.T) {
	verifyCompatibility := compatibleVerifier(t)

	bands := []band.ID{band.EU_863_870, band.US_902_928, band.CN_779_787, band.EU_433}
	verifyCompatibility(ttnpb.PHY_V1_0, "1.0", bands...)

	bands = append(bands, band.AU_915_928, band.CN_470_510)
	verifyCompatibility(ttnpb.PHY_V1_0_1, "1.0.1", bands...)

	bands = append(bands, band.AS_923, band.KR_920_923, band.IN_865_867)
	verifyCompatibility(ttnpb.PHY_V1_0_2, "1.0.2", bands...)

	bands = append(bands, band.RU_864_870)
	verifyCompatibility(ttnpb.PHY_V1_1, "1.1", bands...)
}

func TestUnsupportedBand(t *testing.T) {
	a := assertions.New(t)

	b, err := band.GetByID(band.IN_865_867)
	a.So(err, should.BeNil)

	_, err = b.Version(ttnpb.PHY_V1_0)
	if !a.So(err, should.NotBeNil) {
		t.Log("LoRaWAN Regional Parameters 1.0 is not supported for the Indian band")
	}
}
