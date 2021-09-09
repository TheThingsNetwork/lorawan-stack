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
	band, ok := All[req.BandId]
	if !ok {
		return nil, errBandNotFound.WithAttributes("id", req.BandId)
	}
	return &ttnpb.GetPhyVersionsResponse{
		PhyVersions: band.Versions(),
	}, nil
}
