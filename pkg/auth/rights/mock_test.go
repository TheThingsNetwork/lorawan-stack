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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type mockHandler struct {
	call func(ctx context.Context, req interface{})

	ctx context.Context
	req interface{}

	res interface{}
	err error
}

func (h *mockHandler) Handler(ctx context.Context, req interface{}) (interface{}, error) {
	h.ctx, h.req = ctx, req
	if h.call != nil {
		h.call(ctx, req)
	}
	return h.res, h.err
}

type mockFetcher struct {
	// Request vars
	applicationCtx  context.Context
	applicationIDs  ttnpb.ApplicationIdentifiers
	clientCtx       context.Context
	clientIDs       ttnpb.ClientIdentifiers
	gatewayCtx      context.Context
	gatewayIDs      ttnpb.GatewayIdentifiers
	organizationCtx context.Context
	organizationIDs ttnpb.OrganizationIdentifiers
	userCtx         context.Context
	userIDs         ttnpb.UserIdentifiers

	// Response vars
	applicationRights  *ttnpb.Rights
	applicationError   error
	clientRights       *ttnpb.Rights
	clientError        error
	gatewayRights      *ttnpb.Rights
	gatewayError       error
	organizationRights *ttnpb.Rights
	organizationError  error
	userRights         *ttnpb.Rights
	userError          error
}

func (f *mockFetcher) ApplicationRights(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	f.applicationCtx, f.applicationIDs = ctx, ids
	return f.applicationRights, f.applicationError
}
func (f *mockFetcher) ClientRights(ctx context.Context, ids ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	f.clientCtx, f.clientIDs = ctx, ids
	return f.clientRights, f.clientError
}
func (f *mockFetcher) GatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	f.gatewayCtx, f.gatewayIDs = ctx, ids
	return f.gatewayRights, f.gatewayError
}
func (f *mockFetcher) OrganizationRights(ctx context.Context, ids ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	f.organizationCtx, f.organizationIDs = ctx, ids
	return f.organizationRights, f.organizationError
}
func (f *mockFetcher) UserRights(ctx context.Context, ids ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	f.userCtx, f.userIDs = ctx, ids
	return f.userRights, f.userError
}
