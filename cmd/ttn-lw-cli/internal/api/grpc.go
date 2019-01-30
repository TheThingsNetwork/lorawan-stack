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
	"crypto/x509"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

var (
	withInsecure bool
	tlsConfig    *tls.Config
	auth         *rpcmetadata.MD
)

// SetLogger sets the default API logger
func SetLogger(logger log.Interface) {
	rpclog.ReplaceGrpcLogger(logger)
}

// SetInsecure configures the API to use insecure connections.
func SetInsecure(insecure bool) {
	withInsecure = insecure
}

// AddCA adds the CA certificate file.
func AddCA(pemBytes []byte) (err error) {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	rootCAs := tlsConfig.RootCAs
	if rootCAs == nil {
		if rootCAs, err = x509.SystemCertPool(); err != nil {
			rootCAs = x509.NewCertPool()
		}
	}
	rootCAs.AppendCertsFromPEM(pemBytes)
	tlsConfig.RootCAs = rootCAs
	return nil
}

// SetAuth sets the authentication information.
func SetAuth(authType, authValue string) {
	auth = &rpcmetadata.MD{
		AuthType:  authType,
		AuthValue: authValue,
	}
}

// GetDialOptions gets the dial options for a gRPC connection.
func GetDialOptions() (opts []grpc.DialOption) {
	if withInsecure {
		opts = append(opts, grpc.WithInsecure())
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
	return
}

// GetCallOptions returns the gRPC call options.
func GetCallOptions() (opts []grpc.CallOption) {
	return nil
}

var (
	connMu sync.RWMutex
	conns  = make(map[string]*grpc.ClientConn)
)

func Dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	connMu.Lock()
	defer connMu.Unlock()
	if conn, ok := conns[target]; ok {
		return conn, nil
	}
	conn, err := dialContext(ctx, target)
	if err != nil {
		return nil, err
	}
	conns[target] = conn
	return conn, nil
}

func dialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(append(rpcclient.DefaultDialOptions(ctx), GetDialOptions()...), opts...)
	return grpc.DialContext(ctx, target, opts...)
}

// CloseAll closes all remaining gRPC connections.
func CloseAll() {
	connMu.Lock()
	defer connMu.Unlock()
	for target, conn := range conns {
		delete(conns, target)
		if conn == nil || conn.GetState() == connectivity.Shutdown {
			continue
		}
		conn.Close()
	}
}
