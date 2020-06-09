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

package gatewayconfigurationserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

// rewriteAuthorization rewrites the Authorization header from The Things Network Stack V2 style to The Things Stack.
// Packet Forwarders designed for The Things Stack Network V2 pass the gateway access key via the Authorization header
// prepended by `key`. If the authentication value is a The Things Stack auth token or API key, this function rewrites
// the authentication type to `bearer`, otherwise, the authentication type stays `key`.
func rewriteAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("Authorization")
		parts := strings.SplitN(value, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "key" {
			authType, authValue := parts[0], parts[1]
			tokenType, _, _, err := auth.SplitToken(authValue)
			if err == nil && (tokenType == auth.APIKey || tokenType == auth.AccessToken) {
				authType = "bearer"
			}
			r.Header.Set("Authorization", fmt.Sprintf("%v %v", authType, authValue))
		}
		next.ServeHTTP(w, r)
	})
}

type gatewayIDKeyType struct{}

var gatewayIDKey gatewayIDKeyType

func withGatewayID(ctx context.Context, id ttnpb.GatewayIdentifiers) context.Context {
	return context.WithValue(ctx, gatewayIDKey, id)
}

func gatewayIDFromContext(ctx context.Context) ttnpb.GatewayIdentifiers {
	id, ok := ctx.Value(gatewayIDKey).(ttnpb.GatewayIdentifiers)
	if !ok {
		panic("no gateway identifiers found in context")
	}
	return id
}

func validateAndFillIDs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := ttnpb.GatewayIdentifiers{
			GatewayID: mux.Vars(r)["gateway_id"],
		}
		if err := id.ValidateContext(ctx); err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		ctx = withGatewayID(ctx, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
