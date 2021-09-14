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

// GetPhyVersions returns the list of supported phy versions for the given band.
func GetPhyVersions(ctx context.Context, req *ttnpb.GetPhyVersionsRequest) (*ttnpb.GetPhyVersionsResponse, error) {
	var res *ttnpb.GetPhyVersionsResponse
	if req.BandId != "" {
		versions, ok := All[req.BandId]
		if !ok {
			return nil, errBandNotFound.WithAttributes("id", req.BandId)
		}
		vs := make([]ttnpb.PHYVersion, 0, len(versions))
		for version := range versions {
			vs = append(vs, version)
		}
		res = &ttnpb.GetPhyVersionsResponse{
			VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
				{
					BandId:      req.BandId,
					PhyVersions: vs,
				},
			},
		}
	} else {
		versionInfo := make([]*ttnpb.GetPhyVersionsResponse_VersionInfo, 0, len(All))
		for bandID, versions := range All {
			vs := make([]ttnpb.PHYVersion, 0, len(versions))
			for version := range versions {
				vs = append(vs, version)
			}
			versionInfo = append(versionInfo, &ttnpb.GetPhyVersionsResponse_VersionInfo{
				BandId:      bandID,
				PhyVersions: vs,
			})
		}
		res = &ttnpb.GetPhyVersionsResponse{
			VersionInfo: versionInfo,
		}
	}
	return res, nil
}
