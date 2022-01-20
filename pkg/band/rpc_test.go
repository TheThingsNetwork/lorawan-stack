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
	"sort"
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
		Expected       ttnpb.GetPhyVersionsResponse
		ErrorAssertion func(err error) bool
	}{
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
			Expected: ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "EU_863_870",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
				},
			},
		},
		{
			Name:   "AU915",
			BandID: "AU_915_928",
			Expected: ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "AU_915_928",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
						},
					},
				},
			},
		},
		{
			Name:   "AS923",
			BandID: "AS_923",
			Expected: ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "AS_923",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
						},
					},
				},
			},
		},
		{
			Name: "All",
			Expected: ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "AS_923",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
						},
					},
					{
						BandId: "AS_923_2",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
						},
					},
					{
						BandId: "AS_923_3",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
						},
					},
					{
						BandId: "AU_915_928",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
						},
					},
					{
						BandId: "CN_470_510",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
						},
					},
					{
						BandId: "CN_470_510_20_A",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_470_510_20_B",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_470_510_26_A",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_470_510_26_B",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_779_787",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "EU_433",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "EU_863_870",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "IN_865_867",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
						},
					},
					{
						BandId: "ISM_2400",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "KR_920_923",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
						},
					},
					{
						BandId: "RU_864_870",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
						},
					},
					{
						BandId: "US_902_928",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
				},
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
				sort.Slice(res.VersionInfo, func(i, j int) bool { return res.VersionInfo[i].BandId <= res.VersionInfo[j].BandId })
				for _, vi := range res.VersionInfo {
					sort.Slice(vi.PhyVersions, func(i, j int) bool { return vi.PhyVersions[i] >= vi.PhyVersions[j] })
				}
				if !a.So(*res, should.Resemble, tc.Expected) {
					t.Fatalf("Unexpected value: %v", res)
				}
			}
		})
	}
}
