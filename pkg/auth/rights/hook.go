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

package rights

import (
	"context"
	"runtime/trace"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

var errNoFetcher = errors.DefineInternal("no_fetcher", "no fetcher found in context")

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
		defer trace.StartRegion(ctx, "fetch rights").End()
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		ids, ok := req.(ttnpb.Identifiers)
		if !ok {
			return next(ctx, req)
		}
		fetchRights := func(entityIDs *ttnpb.EntityIdentifiers) (*ttnpb.Rights, error) {
			switch ids := entityIDs.Identifiers().(type) {
			case *ttnpb.ApplicationIdentifiers:
				return fetcher.ApplicationRights(ctx, *ids)
			case *ttnpb.ClientIdentifiers:
				return fetcher.ClientRights(ctx, *ids)
			case *ttnpb.GatewayIdentifiers:
				return fetcher.GatewayRights(ctx, *ids)
			case *ttnpb.OrganizationIdentifiers:
				return fetcher.OrganizationRights(ctx, *ids)
			case *ttnpb.UserIdentifiers:
				return fetcher.UserRights(ctx, *ids)
			}
			return nil, nil
		}
		combined := ids.CombinedIdentifiers()
		results := Rights{
			ApplicationRights:  make(map[string]*ttnpb.Rights),
			ClientRights:       make(map[string]*ttnpb.Rights),
			GatewayRights:      make(map[string]*ttnpb.Rights),
			OrganizationRights: make(map[string]*ttnpb.Rights),
			UserRights:         make(map[string]*ttnpb.Rights),
		}
		setRights := func(entityIDs *ttnpb.EntityIdentifiers, rights *ttnpb.Rights) {
			switch ids := entityIDs.Identifiers().(type) {
			case *ttnpb.ApplicationIdentifiers:
				results.ApplicationRights[unique.ID(ctx, ids)] = rights
			case *ttnpb.ClientIdentifiers:
				results.ClientRights[unique.ID(ctx, ids)] = rights
			case *ttnpb.GatewayIdentifiers:
				results.GatewayRights[unique.ID(ctx, ids)] = rights
			case *ttnpb.OrganizationIdentifiers:
				results.OrganizationRights[unique.ID(ctx, ids)] = rights
			case *ttnpb.UserIdentifiers:
				results.UserRights[unique.ID(ctx, ids)] = rights
			}
		}
		for _, ids := range combined.GetEntityIdentifiers() {
			if devIDs := ids.GetDeviceIDs(); devIDs != nil {
				ids = devIDs.ApplicationIdentifiers.EntityIdentifiers()
			}
			if ids.IDString() == "" {
				continue
			}
			rights, err := fetchRights(ids)
			if err == nil {
				setRights(ids, rights)
				continue
			}
			if !errors.IsPermissionDenied(err) {
				return nil, err
			}
		}
		ctx = NewContext(ctx, results)
		return next(ctx, req)
	}
}
