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

package api

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpcretry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	withInsecure        bool
	tlsConfig           *tls.Config
	auth                *rpcmetadata.MD
	withDump            bool
	retryMax            uint
	retryDefaultTimeout time.Duration
	retryEnableMetadata bool
	retryJitter         float64
)

// SetLogger sets the default API logger
func SetLogger(logger log.Interface) {
	rpclog.ReplaceGrpcLogger(logger)
}

// SetInsecure configures the API to use insecure connections.
func SetInsecure(insecure bool) {
	withInsecure = insecure
}

// SetDumpRequests configures the API options to dump gRPC requests.
func SetDumpRequests(dump bool) {
	withDump = dump
}

// SetRetryMax configures the amount of time the client will retry the request.
func SetRetryMax(rm uint) {
	retryMax = rm
}

// SetRetryDefaultTimeout configures the default timeout before making a retry in a failed request.
func SetRetryDefaultTimeout(t time.Duration) {
	retryDefaultTimeout = t
}

// SetRetryEnableMetadata configures if the retry procedure will read the request's metadata or not.
func SetRetryEnableMetadata(b bool) {
	retryEnableMetadata = b
}

// SetRetryJitter configures the fraction to be used in the deviation procedure of the rpcretry timeout.
func SetRetryJitter(f float64) {
	retryJitter = f
}

// SetAuth sets the authentication information.
func SetAuth(authType, authValue string) {
	auth = &rpcmetadata.MD{
		AuthType:  authType,
		AuthValue: authValue,
	}
}

// requestInterceptor is a gRPC interceptor logging the request payload
func requestInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logger := log.FromContext(ctx)
	if b, err := jsonpb.TTN().Marshal(req); err == nil {
		logger.WithFields(log.Fields("grpc_payload", string(b))).Debug("Request payload")
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}

// GetDialOptions gets the dial options for a gRPC connection.
func GetDialOptions() []grpc.DialOption {
	var opts []grpc.DialOption
	if withInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if auth != nil {
			md := *auth
			md.AllowInsecure = true
			opts = append(opts, grpc.WithPerRPCCredentials(md))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		if auth != nil {
			md := *auth
			opts = append(opts, grpc.WithPerRPCCredentials(md))
		}
	}
	if withDump {
		opts = append(opts, grpc.WithChainUnaryInterceptor(requestInterceptor))
	}

	return append(opts, grpc.WithChainUnaryInterceptor(rpcretry.UnaryClientInterceptor(
		rpcretry.WithMax(retryMax),
		rpcretry.WithDefaultTimeout(retryDefaultTimeout),
		rpcretry.UseMetadata(retryEnableMetadata),
		rpcretry.WithJitter(retryJitter),
	)))
}

// GetCallOptions returns the gRPC call options.
func GetCallOptions() (opts []grpc.CallOption) {
	return nil
}

var (
	connMu sync.RWMutex
	conns  = make(map[string]*grpc.ClientConn)
)

// Dial dials a gRPC connection to the target.
func Dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	connMu.Lock()
	defer connMu.Unlock()
	logger := log.FromContext(ctx).WithField("target", target)
	if conn, ok := conns[target]; ok {
		logger.Debug("Using existing gRPC connection")
		return conn, nil
	}
	logger.Debug("Connecting to gRPC server...")
	conn, err := newClient(ctx, target)
	if err != nil {
		return nil, err
	}
	conns[target] = conn
	return conn, nil
}

func newClient(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(append(rpcclient.DefaultDialOptions(ctx), GetDialOptions()...), opts...)
	return grpc.NewClient(target, opts...)
}

// CloseAll closes all remaining gRPC connections.
func CloseAll() {
	connMu.Lock()
	defer connMu.Unlock()
	for target, conn := range conns {
		delete(conns, target)
		if conn == nil {
			continue
		}
		conn.Close()
	}
}
