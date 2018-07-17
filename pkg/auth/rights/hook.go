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
	"google.golang.org/grpc"
)

var errNoFetcher = errors.DefineInternal("missing_fetcher", "no fetcher found in context")

// HookName denotes the unique name that components should use to register this hook.
const HookName = "rights-fetcher"

// Hook for fetching the rights for a request.
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
			ApplicationRights:  make(map[ttnpb.ApplicationIdentifiers][]ttnpb.Right, len(combined.ApplicationIDs)),
			GatewayRights:      make(map[ttnpb.GatewayIdentifiers][]ttnpb.Right, len(combined.GatewayIDs)),
			OrganizationRights: make(map[ttnpb.OrganizationIdentifiers][]ttnpb.Right, len(combined.OrganizationIDs)),
		}
		for _, id := range combined.ApplicationIDs {
			rights, err := fetcher.ApplicationRights(ctx, *id)
			switch {
			case err == nil:
				results.ApplicationRights[*id] = rights
			case errors.IsPermissionDenied(err):
				results.ApplicationRights[*id] = nil
			default:
				return nil, err
			}
		}
		for _, id := range combined.GatewayIDs {
			rights, err := fetcher.GatewayRights(ctx, *id)
			switch {
			case err == nil:
				results.GatewayRights[*id] = rights
			case errors.IsPermissionDenied(err):
				results.GatewayRights[*id] = nil
			default:
				return nil, err
			}
		}
		for _, id := range combined.OrganizationIDs {
			rights, err := fetcher.OrganizationRights(ctx, *id)
			switch {
			case err == nil:
				results.OrganizationRights[*id] = rights
			case errors.IsPermissionDenied(err):
				results.OrganizationRights[*id] = nil
			default:
				return nil, err
			}
		}
		ctx = newContext(ctx, results)
		return next(ctx, req)
	}
}
