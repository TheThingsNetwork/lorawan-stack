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

package metrics

import (
	"context"
	"net"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

var serverMetrics = grpc_prometheus.NewServerMetrics()

func init() {
	MustRegister(serverMetrics)
}

// InitializeServerMetrics initializes server metrics for the given gRPC server.
func InitializeServerMetrics(s *grpc.Server) {
	serverMetrics.InitializeMetrics(s)
}

// Server interceptors.
var (
	StreamServerInterceptor = serverMetrics.StreamServerInterceptor()
	UnaryServerInterceptor  = serverMetrics.UnaryServerInterceptor()
)

var clientMetrics = grpc_prometheus.NewClientMetrics()

func init() {
	MustRegister(clientMetrics)
}

// Client interceptors.
var (
	StreamClientInterceptor = clientMetrics.StreamClientInterceptor()
	UnaryClientInterceptor  = clientMetrics.UnaryClientInterceptor()
)

var gRPCStats = statsHandler{
	openedClientConns: NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: "grpc",
			Name:      "client_conns_opened_total",
			Help:      "Opened client connections",
		},
		[]string{"server_address"},
	),
	closedClientConns: NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: "grpc",
			Name:      "client_conns_closed_total",
			Help:      "Closed client connections",
		},
		[]string{"server_address"},
	),
	openedServerConns: NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: "grpc",
			Name:      "server_conns_opened_total",
			Help:      "Opened server connections",
		},
		[]string{"server_address"},
	),
	closedServerConns: NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: "grpc",
			Name:      "server_conns_closed_total",
			Help:      "Closed server connections",
		},
		[]string{"server_address"},
	),
}

// StatsHandler for gRPC.
var StatsHandler stats.Handler = gRPCStats

func init() {
	MustRegister(gRPCStats)
}

type statsHandler struct {
	openedClientConns *ContextualCounterVec
	openedServerConns *ContextualCounterVec
	closedClientConns *ContextualCounterVec
	closedServerConns *ContextualCounterVec
}

func (hdl statsHandler) Describe(ch chan<- *prometheus.Desc) {
	hdl.openedClientConns.Describe(ch)
	hdl.openedServerConns.Describe(ch)
	hdl.closedClientConns.Describe(ch)
	hdl.closedServerConns.Describe(ch)
}

func (hdl statsHandler) Collect(ch chan<- prometheus.Metric) {
	hdl.openedClientConns.Collect(ch)
	hdl.openedServerConns.Collect(ch)
	hdl.closedClientConns.Collect(ch)
	hdl.closedServerConns.Collect(ch)
}

func (hdl statsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

func (hdl statsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {}

func (hdl statsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return newContextWithConnInfo(ctx, info)
}

func (hdl statsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	info := connInfoFromContext(ctx)
	peer := info.RemoteAddr.String()
	if !s.IsClient() {
		peer, _, _ = net.SplitHostPort(peer) // Remove the port number.
	}
	switch s.(type) {
	case *stats.ConnBegin:
		if s.IsClient() {
			hdl.openedClientConns.WithLabelValues(ctx, peer).Inc()
		} else {
			hdl.openedServerConns.WithLabelValues(ctx, peer).Inc()
		}
	case *stats.ConnEnd:
		if s.IsClient() {
			hdl.closedClientConns.WithLabelValues(ctx, peer).Inc()
		} else {
			hdl.closedServerConns.WithLabelValues(ctx, peer).Inc()
		}
	}
}

type connStatsKeyType struct{}

var connStatsKey connStatsKeyType

func newContextWithConnInfo(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, connStatsKey, info)
}

func connInfoFromContext(ctx context.Context) *stats.ConnTagInfo {
	if r, ok := ctx.Value(connStatsKey).(*stats.ConnTagInfo); ok {
		return r
	}
	return nil
}
