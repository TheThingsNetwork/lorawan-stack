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
	"encoding"
	"encoding/json"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	gcsv2 "go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/cpf"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/semtechudp"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"google.golang.org/grpc"
)

// Server implements the Gateway Configuration Server component.
type Server struct {
	*component.Component
	config *Config
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
			GatewayIds: &gtwID,
			FieldMask: &types.FieldMask{
				Paths: []string{
					"antennas",
					"frequency_plan_id",
					"gateway_server_address",
				},
			},
		}, s.WithClusterAuth())
		if err != nil {
			webhandlers.Error(w, r, err)
			return
		}
		next(w, r, gtw)
	}
}

func (s *Server) makeJSONHandler(f func(context.Context, *ttnpb.Gateway) (interface{}, error)) http.HandlerFunc {
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

// Roles returns the roles that the Gateway Configuration Server fulfills.
func (s *Server) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER}
}

// RegisterServices registers services provided by gcs at s.
func (s *Server) RegisterServices(_ *grpc.Server) {}

// RegisterHandlers registers gRPC handlers.
func (s *Server) RegisterHandlers(_ *runtime.ServeMux, _ *grpc.ClientConn) {}

// RegisterRoutes registers the web frontend routes.
func (s *Server) RegisterRoutes(server *web.Server) {
	router := server.Prefix(ttnpb.HTTPAPIPrefix + "/gcs/gateways/{gateway_id}/").Subrouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Namespace("gatewayconfigurationserver")),
		ratelimit.HTTPMiddleware(s.Component.RateLimiter(), "http:gcs"),
		mux.MiddlewareFunc(webmiddleware.Metadata("Authorization")),
		validateAndFillIDs,
	)
	if s.config.RequireAuth {
		router.Use(mux.MiddlewareFunc(s.requireGatewayRights(ttnpb.Right_RIGHT_GATEWAY_INFO)))
	}

	router.Handle("/semtechudp/global_conf.json",
		s.makeJSONHandler(func(ctx context.Context, gtw *ttnpb.Gateway) (interface{}, error) {
			fps, err := s.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, err
			}
			return semtechudp.Build(gtw, fps)
		}),
	).Methods(http.MethodGet)

	router.Handle("/kerlink-cpf/lorad/lorad.json",
		s.makeJSONHandler(func(ctx context.Context, gtw *ttnpb.Gateway) (interface{}, error) {
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

// New returns new *Server.
func New(c *component.Component, conf *Config) (*Server, error) {
	gcs := &Server{
		Component: c,
		config:    conf,
	}

	bsCUPS := conf.BasicStation.NewServer(c)
	_ = bsCUPS

	v2GCS := gcsv2.New(c, gcsv2.WithTheThingsGatewayConfig(conf.TheThingsGateway))
	_ = v2GCS

	c.RegisterGRPC(gcs)
	c.RegisterWeb(gcs)
	return gcs, nil
}
