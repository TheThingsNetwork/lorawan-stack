// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"encoding"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/cpf"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/semtechudp"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

// RegisterRoutes registers the web frontend routes.
//
// The gateway configuration value returned by the `grpc-gateway` routes are not content formatted, but a stream of
// bytes. This would be a breaking change for the consumers of this API and hence these routes are retained.
func (s *Server) RegisterRoutes(server *web.Server) {
	router := server.Prefix(ttnpb.HTTPAPIPrefix + "/gcs/gateways/{gateway_id}/").Subrouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Namespace("gatewayconfigurationserver")),
		mux.MiddlewareFunc(webmiddleware.Metadata("Authorization")),
		ratelimit.HTTPMiddleware(s.Component.RateLimiter(), "http:gcs"),
		validateAndFillIDs,
	)
	if s.config.RequireAuth {
		router.Use(mux.MiddlewareFunc(s.requireGatewayRights(ttnpb.Right_RIGHT_GATEWAY_INFO)))
	}

	router.Handle("/semtechudp/global_conf.json",
		s.makeJSONHandler(func(ctx context.Context, gtw *ttnpb.Gateway) (any, error) {
			fps, err := s.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, err
			}
			return semtechudp.Build(gtw, fps)
		}),
	).Methods(http.MethodGet)

	router.Handle("/kerlink-cpf/lorad/lorad.json",
		s.makeJSONHandler(func(ctx context.Context, gtw *ttnpb.Gateway) (any, error) {
			fps, err := s.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, err
			}
			return cpf.BuildLorad(gtw, fps)
		}),
	).Methods(http.MethodGet)

	router.Handle("/kerlink-cpf/lorafwd/lorafwd.toml",
		s.makeTextMarshalerHandler("application/toml", func(ctx context.Context, gtw *ttnpb.Gateway) (encoding.TextMarshaler, error) {
			return cpf.BuildLorafwd(gtw)
		}),
	).Methods(http.MethodGet)
}

func (s *Server) withGateway(next func(http.ResponseWriter, *http.Request, *ttnpb.Gateway)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		gtwID := gatewayIDFromContext(ctx)
		cc, err := s.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		client := ttnpb.NewGatewayRegistryClient(cc)
		gtw, err := client.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIds: gtwID,
			FieldMask:  ttnpb.FieldMask("antennas", "frequency_plan_id", "gateway_server_address"),
		}, s.WithClusterAuth())
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		next(w, r, gtw)
	}
}

func (s *Server) makeJSONHandler(f func(context.Context, *ttnpb.Gateway) (any, error)) http.HandlerFunc {
	return s.withGateway(func(w http.ResponseWriter, r *http.Request, gtw *ttnpb.Gateway) {
		msg, err := f(r.Context(), gtw)
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		enc.Encode(msg)
	})
}

func (s *Server) makeTextMarshalerHandler(contentType string, f func(context.Context, *ttnpb.Gateway) (encoding.TextMarshaler, error)) http.HandlerFunc {
	return s.withGateway(func(w http.ResponseWriter, r *http.Request, gtw *ttnpb.Gateway) {
		msg, err := f(r.Context(), gtw)
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		b, err := msg.MarshalText()
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})
}
