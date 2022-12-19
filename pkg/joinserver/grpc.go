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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type jsServer struct {
	ttnpb.UnimplementedJsServer

	JS *JoinServer
}

// GetJoinEUIPrefixes returns the JoinEUIPrefixes associated with the join server.
func (srv jsServer) GetJoinEUIPrefixes(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.JoinEUIPrefixes, error) {
	prefixes := make([]*ttnpb.JoinEUIPrefix, 0, len(srv.JS.euiPrefixes))
	for _, p := range srv.JS.euiPrefixes {
		prefixes = append(prefixes, &ttnpb.JoinEUIPrefix{
			JoinEui: p.EUI64.Bytes(),
			Length:  uint32(p.Length),
		})
	}
	return &ttnpb.JoinEUIPrefixes{
		Prefixes: prefixes,
	}, nil
}

// GetDefaultJoinEUI returns the default JoinEUI that is configured for this Join Server.
func (srv jsServer) GetDefaultJoinEUI(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.GetDefaultJoinEUIResponse, error) {
	return &ttnpb.GetDefaultJoinEUIResponse{
		JoinEui: srv.JS.defaultJoinEUI.Bytes(),
	}, nil
}
