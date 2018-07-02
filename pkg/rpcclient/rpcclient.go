// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package rpcclient contains the default options for TTN gRPC clients.
package rpcclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/version"
	"google.golang.org/grpc"
)

// DefaultDialOptions for gRPC clients
func DefaultDialOptions(ctx context.Context) []grpc.DialOption {
	streamInterceptors := []grpc.StreamClientInterceptor{
		errors.StreamClientInterceptor(),
		grpc_prometheus.StreamClientInterceptor,
		rpclog.StreamClientInterceptor(ctx), // Gets logger from global context
	}

	unaryInterceptors := []grpc.UnaryClientInterceptor{
		errors.UnaryClientInterceptor(),
		grpc_prometheus.UnaryClientInterceptor,
		rpclog.UnaryClientInterceptor(ctx), // Gets logger from global context
	}

	ttnVersion := strings.TrimPrefix(version.TTN, "v")
	if version.GitBranch != "" && version.GitCommit != "" && version.BuildDate != "" {
		ttnVersion += fmt.Sprintf("(%s@%s, %s)", version.GitBranch, version.GitCommit, version.BuildDate)
	}

	return []grpc.DialOption{
		grpc.WithStatsHandler(new(ocgrpc.ClientHandler)),
		grpc.WithUserAgent(fmt.Sprintf(
			"%s go/%s ttn/%s",
			filepath.Base(os.Args[0]),
			strings.TrimPrefix(runtime.Version(), "go"),
			ttnVersion,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptors...)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
	}
}
