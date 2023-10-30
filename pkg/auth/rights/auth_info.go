// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// AuthInfo lists the authentication info with universal rights, whether the caller is admin and the authentication method.
func AuthInfo(ctx context.Context) (authInfo *ttnpb.AuthInfoResponse, err error) {
	if inCtx, ok := authInfoFromContext(ctx); ok {
		return inCtx, nil
	}
	if inCtx, ok := cacheAuthInfoFromContext(ctx); ok {
		return inCtx, nil
	}
	defer func() {
		if err == nil {
			cacheAuthInfoInContext(ctx, authInfo)
		}
	}()
	fetcher, ok := fetcherFromContext(ctx)
	if !ok {
		panic(errNoFetcher)
	}
	authInfo, err = fetcher.AuthInfo(ctx)
	if err != nil {
		if errors.IsPermissionDenied(err) {
			return nil, nil
		}
		return nil, err
	}
	return authInfo, nil
}

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")

// RequireAuthenticated confirms if the authentication information within a context contains any rights, if so,
// the request is considered to be authenticated.
func RequireAuthenticated(ctx context.Context) error {
	authInfo, err := AuthInfo(ctx)
	if err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to validate authentication information")
		return errUnauthenticated.WithCause(err)
	}
	if authInfo.GetAccessMethod() == nil && len(authInfo.GetUniversalRights().GetRights()) == 0 {
		return errUnauthenticated.New()
	}
	return nil
}
