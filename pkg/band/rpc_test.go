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
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func TestGetPhyVersions(t *testing.T) {
	a := assertions.New(t)
	ctx := context.Background()
	for _, tc := range []struct {
		Name           string
		BandID         string
		Expected       []ttnpb.PHYVersion
		ErrorAssertion func(err error) bool
	}{
		{
			Name:   "Empty",
			BandID: "",
			ErrorAssertion: func(err error) bool {
				return errors.IsNotFound(err)
			},
		},
		{
			Name:   "Unknown",
			BandID: "AS_925",
			ErrorAssertion: func(err error) bool {
				return errors.IsNotFound(err)
			},
		},
		{
			Name:   "EU868",
			BandID: "EU_863_870",
			Expected: []ttnpb.PHYVersion{
				ttnpb.RP001_V1_1_REV_B,
				ttnpb.RP001_V1_1_REV_A,
				ttnpb.RP001_V1_0_3_REV_A,
				ttnpb.RP001_V1_0_2_REV_B,
				ttnpb.RP001_V1_0_2,
				ttnpb.TS001_V1_0_1,
				ttnpb.TS001_V1_0,
			},
		},
		{
			Name:   "AU915",
			BandID: "AU_915_928",
			Expected: []ttnpb.PHYVersion{
				ttnpb.RP001_V1_1_REV_B,
				ttnpb.RP001_V1_1_REV_A,
				ttnpb.RP001_V1_0_3_REV_A,
				ttnpb.RP001_V1_0_2_REV_B,
				ttnpb.RP001_V1_0_2,
				ttnpb.TS001_V1_0_1,
			},
		},
		{
			Name:   "AS923",
			BandID: "AS_923",
			Expected: []ttnpb.PHYVersion{
				ttnpb.RP001_V1_1_REV_B,
				ttnpb.RP001_V1_1_REV_A,
				ttnpb.RP001_V1_0_3_REV_A,
				ttnpb.RP001_V1_0_2_REV_B,
				ttnpb.RP001_V1_0_2,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			res, err := GetPhyVersions(ctx, &ttnpb.GetPhyVersionsRequest{
				BandId: tc.BandID,
			})
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(err), should.BeTrue)
			} else {
				if !a.So(res, should.NotBeNil) {
					t.Fatalf("Nil value received. Expected :%v", tc.Expected)
				}
				if !a.So(res.PhyVersions, should.Resemble, tc.Expected) {
					t.Fatalf("Unexpected value: %v", res)
				}
			}
		})
	}
}
