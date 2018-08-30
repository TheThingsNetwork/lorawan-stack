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

package rights

import (
	"context"
	"fmt"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

var errNoFetcher = errors.DefineInternal("missing_fetcher", "no fetcher found in context")

// HookName denotes the unique name that components should use to register this hook.
//
// Services that need the hook should register it with:
//
//     hooks.RegisterUnaryHook("/ttn.lorawan.v3.SomeService", rights.HookName, rights.Hook)
//
// The hook does not support streaming RPCs.
const HookName = "rights-fetcher"

// Hook for fetching the rights for a request.
//
// This hook requires a Fetcher in the context, which should be inserted by other
// middleware, such as the fillcontext middleware, where you could do:
//
//     ctx = rights.NewContextWithFetcher(ctx, fetcher)
//
// It is recommended to wrap the Identity Server fetcher with a cache:
//
//     fetcher = rights.NewInMemoryCache(fetcher, successTTL, errorTTL)
//
// Also note that all RPCs for which this hook is executed, need to take an
// argument that implements the ttnpb.Identifiers interface.
func Hook(next grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		ids, ok := req.(ttnpb.Identifiers)
		if !ok {
			panic(fmt.Errorf("Could not execute rights hook: %T does not implement ttnpb.Identifiers", req))
		}
		combined := ids.CombinedIdentifiers()
		results := Rights{
			ApplicationRights:  make(map[string]*ttnpb.Rights, len(combined.ApplicationIDs)),
			GatewayRights:      make(map[string]*ttnpb.Rights, len(combined.GatewayIDs)),
			OrganizationRights: make(map[string]*ttnpb.Rights, len(combined.OrganizationIDs)),
		}
		for _, id := range combined.ApplicationIDs {
			uid := unique.ID(ctx, id)
			rights, err := fetcher.ApplicationRights(ctx, *id)
			switch {
			case err == nil:
				results.ApplicationRights[uid] = rights
			case errors.IsPermissionDenied(err):
				results.ApplicationRights[uid] = nil
			default:
				return nil, err
			}
		}
		for _, id := range combined.GatewayIDs {
			uid := unique.ID(ctx, id)
			rights, err := fetcher.GatewayRights(ctx, *id)
			switch {
			case err == nil:
				results.GatewayRights[uid] = rights
			case errors.IsPermissionDenied(err):
				results.GatewayRights[uid] = nil
			default:
				return nil, err
			}
		}
		for _, id := range combined.OrganizationIDs {
			uid := unique.ID(ctx, id)
			rights, err := fetcher.OrganizationRights(ctx, *id)
			switch {
			case err == nil:
				results.OrganizationRights[uid] = rights
			case errors.IsPermissionDenied(err):
				results.OrganizationRights[uid] = nil
			default:
				return nil, err
			}
		}
		ctx = NewContext(ctx, results)
		return next(ctx, req)
	}
}
