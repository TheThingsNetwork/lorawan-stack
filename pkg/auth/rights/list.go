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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// ListApplication lists the rights for the given application ID in the context.
func ListApplication(ctx context.Context, id ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.ApplicationRights[uid], nil
	}
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err := fetcher.ApplicationRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListClient lists the rights for the given client ID in the context.
func ListClient(ctx context.Context, id ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.ClientRights[uid], nil
	}
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err := fetcher.ClientRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListGateway lists the rights for the given gateway ID in the context.
func ListGateway(ctx context.Context, id ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.GatewayRights[uid], nil
	}
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err := fetcher.GatewayRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListOrganization lists the rights for the given organization ID in the context.
func ListOrganization(ctx context.Context, id ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.OrganizationRights[uid], nil
	}
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err := fetcher.OrganizationRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListUser lists the rights for the given user ID in the context.
func ListUser(ctx context.Context, id ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.UserRights[uid], nil
	}
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err := fetcher.UserRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}
