// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package web_test

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/metadata"
)

const (
	registeredApplicationID  = "foo-app"
	registeredApplicationKey = "secret"
)

func newContextWithRightsFetcher(ctx context.Context) context.Context {
	return rights.NewContextWithFetcher(
		ctx,
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (set *ttnpb.Rights, err error) {
			uid := unique.ID(ctx, ids)
			if uid != unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}) {
				return
			}
			md := rpcmetadata.FromIncomingContext(ctx)
			if md.AuthType != "Key" || md.AuthValue != registeredApplicationKey {
				return
			}
			set = ttnpb.RightsFrom(
				ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
				ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
			)
			return
		}),
	)
}

func contextWithKey(ctx context.Context, key string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": fmt.Sprintf("Key %v", key),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}
