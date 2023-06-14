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

package loracloudgeolocationv3

import (
	"math/rand"
	"net/url"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// nolint: gosec
func generateRandomizedMD(count int) [][]*RxMetadata {
	md := make([][]*RxMetadata, count)
	for i := range md {
		md[i] = make([]*RxMetadata, rand.Intn(10))
		for j := range md[i] {
			md[i][j] = &RxMetadata{
				GatewayIDs: &GatewayIDs{
					GatewayID: "test-gateway-id",
				},
				AntennaIndex:  rand.Uint32(),
				FineTimestamp: rand.Uint64(),
				RSSI:          rand.Float32(),
				SNR:           rand.Float32(),
				Location: &Location{
					Latitude:  rand.Float64()*180 - 90,
					Longitude: rand.Float64()*360 - 180,
					Altitude:  rand.Int31(),
					Accuracy:  rand.Int31(),
					Source:    rand.Int31n(10),
				},
			}
		}
	}
	return md
}

func TestPackageDataRoundtrip(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	now := time.Now().UTC().Truncate(time.Second)

	rxMDs := generateRandomizedMD(5)
	uplinkMDs := make([]*UplinkMetadata, len(rxMDs))
	for i, rxMD := range rxMDs {
		uplinkMDs[i] = &UplinkMetadata{
			RxMetadata: rxMD,
			ReceivedAt: now.Add(time.Duration(i) * time.Second),
		}
	}

	u, err := url.Parse("https://thethingsindustries.com")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	expectedData := Data{
		Query:                1,
		MultiFrame:           true,
		MultiFrameWindowSize: 10,
		MultiFrameWindowAge:  10 * time.Minute,
		ServerURL:            u,
		Token:                "some_random_token",
		RecentMetadata:       uplinkMDs,
	}
	st, err := expectedData.Struct()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	actualData := Data{}
	if err := actualData.FromStruct(st); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(actualData, should.Resemble, expectedData)
}
