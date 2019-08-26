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

package joinserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type jsServer struct {
	JS *JoinServer
}

// GetJoinEUIPrefixes returns the JoinEUIPrefixes associated with the join server.
func (srv jsServer) GetJoinEUIPrefixes(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.JoinEUIPrefixes, error) {
	prefixes := make([]ttnpb.JoinEUIPrefix, 0, len(srv.JS.euiPrefixes))
	for _, p := range srv.JS.euiPrefixes {
		prefixes = append(prefixes, ttnpb.JoinEUIPrefix{
			JoinEUI: p.EUI64.Copy(&types.EUI64{}),
			Length:  uint32(p.Length),
		})
	}
	return &ttnpb.JoinEUIPrefixes{
		Prefixes: prefixes,
	}, nil
}
