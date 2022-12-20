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

package mock

import (
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	mappingpb "go.packetbroker.org/api/mapping/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

// PBMapper is a mock Packet Broker Mapper.
type PBMapper struct {
	*grpc.Server
	UpdateGatewayHandler func(ctx context.Context, in *mappingpb.UpdateGatewayRequest, opts ...grpc.CallOption) (*pbtypes.Empty, error)
}

// NewPBMapper instantiates a new mock Packet Broker Data Plane.
func NewPBMapper(tb testing.TB) *PBMapper {
	mp := &PBMapper{
		Server: grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctx = test.ContextWithTB(ctx, tb)
				return handler(ctx, req)
			}),
		),
	}
	mappingpb.RegisterMapperServer(mp.Server, &pbMapper{PBMapper: mp})
	return mp
}

type pbMapper struct {
	mappingpb.UnimplementedMapperServer

	*PBMapper
}

func (s *pbMapper) UpdateGateway(ctx context.Context, req *mappingpb.UpdateGatewayRequest) (*pbtypes.Empty, error) {
	if s.UpdateGatewayHandler == nil {
		panic("UpdateGateway called but not set")
	}
	return s.UpdateGatewayHandler(ctx, req)
}
