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

package grpc_test

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc/metadata"
)

type mockGtwGsLinkServerStream struct {
	*test.MockServerStream
	SendFunc func(*ttnpb.GatewayDown) error
	RecvFunc func() (*ttnpb.GatewayUp, error)
}

func (s *mockGtwGsLinkServerStream) Send(down *ttnpb.GatewayDown) error {
	if s.SendFunc == nil {
		panic("Send called, but not set")
	}
	return s.SendFunc(down)
}

func (s *mockGtwGsLinkServerStream) Recv() (*ttnpb.GatewayUp, error) {
	if s.RecvFunc == nil {
		panic("Recv called, but not set")
	}
	return s.RecvFunc()
}

func newContextWithRightsFetcher(ctx context.Context) context.Context {
	return rights.NewContextWithFetcher(
		ctx,
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (set *ttnpb.Rights, err error) {
			uid := unique.ID(ctx, ids)
			if uid != registeredGatewayUID {
				return
			}
			md := rpcmetadata.FromIncomingContext(ctx)
			if md.AuthType != "Bearer" || md.AuthValue != registeredGatewayKey {
				return
			}
			set = ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_LINK)
			return
		}),
	)
}

func contextWithKey(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string) func() context.Context {
	return func() context.Context {
		md := metadata.New(map[string]string{
			"id":            ids.GatewayID,
			"authorization": fmt.Sprintf("Bearer %v", key),
		})
		if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
			md = metadata.Join(ctxMd, md)
		}
		return metadata.NewIncomingContext(ctx, md)
	}
}
