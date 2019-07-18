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
func ListApplication(ctx context.Context, id ttnpb.ApplicationIdentifiers) (rights *ttnpb.Rights, err error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.ApplicationRights[uid], nil
	}
	if inCtx, ok := cacheFromContext(ctx); ok {
		if rights, ok := inCtx.ApplicationRights[uid]; ok {
			return rights, nil
		}
	}
	defer func() {
		if err == nil {
			cacheInContext(ctx, func(r *Rights) { r.setApplicationRights(uid, rights) })
		}
	}()
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err = fetcher.ApplicationRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListClient lists the rights for the given client ID in the context.
func ListClient(ctx context.Context, id ttnpb.ClientIdentifiers) (rights *ttnpb.Rights, err error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.ClientRights[uid], nil
	}
	if inCtx, ok := cacheFromContext(ctx); ok {
		if rights, ok := inCtx.ClientRights[uid]; ok {
			return rights, nil
		}
	}
	defer func() {
		if err == nil {
			cacheInContext(ctx, func(r *Rights) { r.setClientRights(uid, rights) })
		}
	}()
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err = fetcher.ClientRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListGateway lists the rights for the given gateway ID in the context.
func ListGateway(ctx context.Context, id ttnpb.GatewayIdentifiers) (rights *ttnpb.Rights, err error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.GatewayRights[uid], nil
	}
	if inCtx, ok := cacheFromContext(ctx); ok {
		if rights, ok := inCtx.GatewayRights[uid]; ok {
			return rights, nil
		}
	}
	defer func() {
		if err == nil {
			cacheInContext(ctx, func(r *Rights) { r.setGatewayRights(uid, rights) })
		}
	}()
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err = fetcher.GatewayRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListOrganization lists the rights for the given organization ID in the context.
func ListOrganization(ctx context.Context, id ttnpb.OrganizationIdentifiers) (rights *ttnpb.Rights, err error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.OrganizationRights[uid], nil
	}
	if inCtx, ok := cacheFromContext(ctx); ok {
		if rights, ok := inCtx.OrganizationRights[uid]; ok {
			return rights, nil
		}
	}
	defer func() {
		if err == nil {
			cacheInContext(ctx, func(r *Rights) { r.setOrganizationRights(uid, rights) })
		}
	}()
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err = fetcher.OrganizationRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}

// ListUser lists the rights for the given user ID in the context.
func ListUser(ctx context.Context, id ttnpb.UserIdentifiers) (rights *ttnpb.Rights, err error) {
	uid := unique.ID(ctx, id)
	if inCtx, ok := fromContext(ctx); ok {
		return inCtx.UserRights[uid], nil
	}
	if inCtx, ok := cacheFromContext(ctx); ok {
		if rights, ok := inCtx.UserRights[uid]; ok {
			return rights, nil
		}
	}
	defer func() {
		if err == nil {
			cacheInContext(ctx, func(r *Rights) { r.setUserRights(uid, rights) })
		}
	}()
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	rights, err = fetcher.UserRights(ctx, id)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return &ttnpb.Rights{}, nil
		}
		return nil, err
	}
	return rights, nil
}
