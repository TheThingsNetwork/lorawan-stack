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
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

type (
	deviceIDKeyType  struct{}
	webhookIDKeyType struct{}
)

var (
	deviceIDKey  deviceIDKeyType
	webhookIDKey webhookIDKeyType
)

func withDeviceID(ctx context.Context, id ttnpb.EndDeviceIdentifiers) context.Context {
	return context.WithValue(ctx, deviceIDKey, id)
}

func deviceIDFromContext(ctx context.Context) ttnpb.EndDeviceIdentifiers {
	id, ok := ctx.Value(deviceIDKey).(ttnpb.EndDeviceIdentifiers)
	if !ok {
		panic("no end device identifiers found in context")
	}
	return id
}

func withWebhookID(ctx context.Context, id ttnpb.ApplicationWebhookIdentifiers) context.Context {
	return context.WithValue(ctx, webhookIDKey, id)
}

func webhookIDFromContext(ctx context.Context) ttnpb.ApplicationWebhookIdentifiers {
	id, ok := ctx.Value(webhookIDKey).(ttnpb.ApplicationWebhookIdentifiers)
	if !ok {
		panic("no webhook identifiers found in context")
	}
	return id
}

func (w *webhooks) validateAndFillIDs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		appID := ttnpb.ApplicationIdentifiers{
			ApplicationID: vars["application_id"],
		}

		devID := ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: appID,
			DeviceID:               vars["device_id"],
		}
		if err := devID.ValidateContext(ctx); err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		ctx = withDeviceID(ctx, devID)

		hookID := ttnpb.ApplicationWebhookIdentifiers{
			ApplicationIdentifiers: appID,
			WebhookID:              vars["webhook_id"],
		}
		if err := hookID.ValidateContext(ctx); err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		ctx = withWebhookID(ctx, hookID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (w *webhooks) requireApplicationRights(required ...ttnpb.Right) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			appID := deviceIDFromContext(ctx).ApplicationIdentifiers
			if err := rights.RequireApplication(ctx, appID, required...); err != nil {
				webhandlers.Error(res, req, err)
				return
			}
			next.ServeHTTP(res, req)
		})
	}
}
