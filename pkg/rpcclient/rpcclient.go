// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpcclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors/grpcerrors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/rpclog"
	"github.com/TheThingsNetwork/ttn/pkg/version"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// DefaultDialOptions for gRPC clients
func DefaultDialOptions(ctx context.Context) []grpc.DialOption {
	streamInterceptors := []grpc.StreamClientInterceptor{
		grpcerrors.StreamClientInterceptor(),
		grpc_prometheus.StreamClientInterceptor,
		rpclog.StreamClientInterceptor(ctx), // Gets logger from global context
	}

	unaryInterceptors := []grpc.UnaryClientInterceptor{
		grpcerrors.UnaryClientInterceptor(),
		grpc_prometheus.UnaryClientInterceptor,
		rpclog.UnaryClientInterceptor(ctx), // Gets logger from global context
	}

	ttnVersion := strings.TrimPrefix(version.TTN, "v")
	if version.GitBranch != "" && version.GitCommit != "" && version.BuildDate != "" {
		ttnVersion += fmt.Sprintf("(%s@%s, %s)", version.GitBranch, version.GitCommit, version.BuildDate)
	}

	return []grpc.DialOption{
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
