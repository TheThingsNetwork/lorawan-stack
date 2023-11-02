// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package events

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errUnknownCaller = errors.DefineInternal("unknown_caller", "unknown caller type `{type}`")
	errRateExceeded  = errors.DefineResourceExhausted("rate_exceeded", "request rate exceeded")
)

func makeRateLimiter(ctx context.Context, limiter ratelimit.Interface) (func() error, error) {
	authInfo, err := rights.AuthInfo(ctx)
	if err != nil {
		return nil, err
	}
	resourceID := ""
	switch method := authInfo.AccessMethod.(type) {
	case *ttnpb.AuthInfoResponse_ApiKey:
		resourceID = fmt.Sprintf("api-key:%s", method.ApiKey.ApiKey.Id)
	case *ttnpb.AuthInfoResponse_OauthAccessToken:
		resourceID = fmt.Sprintf("access-token:%s", method.OauthAccessToken.Id)
	case *ttnpb.AuthInfoResponse_UserSession:
		resourceID = fmt.Sprintf("session-id:%s", method.UserSession.SessionId)
	// NOTE: *ttnpb.AuthInfoResponse_GatewayToken_ is intentionally left out.
	default:
		return nil, errUnknownCaller.WithAttributes("type", fmt.Sprintf("%T", authInfo.AccessMethod))
	}
	resource := ratelimit.ConsoleEventsRequestResource(resourceID)
	return func() error {
		if limit, _ := limiter.RateLimit(resource); limit {
			return errRateExceeded.New()
		}
		return nil
	}, nil
}
