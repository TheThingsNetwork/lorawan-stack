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

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var bandIDs = []string{AS_923, AU_915_928, CN_470_510, CN_779_787, EU_433, EU_863_870, IN_865_867, ISM_2400, KR_920_923, RU_864_870, US_902_928}

// GetPhyVersions returns the list of supported phy versions for the given band.
func GetPhyVersions(ctx context.Context, req *ttnpb.GetPhyVersionsRequest) (*ttnpb.GetPhyVersionsResponse, error) {
	var res *ttnpb.GetPhyVersionsResponse
	if req.BandId != "" {
		band, ok := All[req.BandId]
		if !ok {
			return nil, errBandNotFound.WithAttributes("id", req.BandId)
		}
		res = &ttnpb.GetPhyVersionsResponse{
			VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
				{
					BandId:      req.BandId,
					PhyVersions: band.Versions(),
				},
			},
		}
	} else {
		versionInfo := []*ttnpb.GetPhyVersionsResponse_VersionInfo{}
		for _, bandID := range bandIDs {
			band := All[bandID]
			versionInfo = append(versionInfo, &ttnpb.GetPhyVersionsResponse_VersionInfo{
				BandId:      bandID,
				PhyVersions: band.Versions(),
			})
		}
		res = &ttnpb.GetPhyVersionsResponse{
			VersionInfo: versionInfo,
		}
	}
	return res, nil
}
