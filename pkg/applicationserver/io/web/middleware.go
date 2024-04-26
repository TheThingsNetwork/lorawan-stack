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

package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

func (*webhooks) validateAndFillIDs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		appID := ttnpb.ApplicationIdentifiers{
			ApplicationId: vars["application_id"],
		}

		devID := &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &appID,
			DeviceId:       vars["device_id"],
		}
		if err := devID.ValidateContext(ctx); err != nil {
			webhandlers.Error(w, r, err)
			return
		}

		hookID := &ttnpb.ApplicationWebhookIdentifiers{
			ApplicationIds: &appID,
			WebhookId:      vars["webhook_id"],
		}
		if err := hookID.ValidateContext(ctx); err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		ctx = internal.WithWebhookData(ctx, &internal.WebhookData{
			EndDeviceIDs: devID,
			WebhookIDs:   hookID,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (*webhooks) requireApplicationRights(required ...ttnpb.Right) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			appID := internal.DeviceIDFromContext(ctx).ApplicationIds
			if err := rights.RequireApplication(ctx, appID, required...); err != nil {
				webhandlers.Error(res, req, err)
				return
			}
			next.ServeHTTP(res, req)
		})
	}
}

var errRateLimitExceeded = errors.DefineResourceExhausted("rate_limit_exceeded", "rate limit exceeded")

func (w *webhooks) requireRateLimits() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			authTokenID := ""
			token := req.Header.Get("Authorization")
			if _, v, _, err := auth.SplitToken(token); err == nil && v != "" {
				authTokenID = v
			}

			resource := ratelimit.ApplicationWebhooksDownResource(ctx, internal.DeviceIDFromContext(ctx), authTokenID)
			limit, result := w.server.RateLimiter().RateLimit(resource)
			result.SetHTTPHeaders(res.Header())
			if limit {
				webhandlers.Error(res, req, errRateLimitExceeded.New())
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}
