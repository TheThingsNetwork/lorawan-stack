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

// Package discover implements a gRPC discovery middleware.
package discover

import (
	"context"
	"fmt"
	"net"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var errAddress = errors.DefineInvalidArgument("address", "invalid address")

func defaultPort(target string, port int) (string, error) {
	i := strings.LastIndexByte(target, ':')
	if i < 0 {
		return fmt.Sprintf("%s:%d", target, port), nil
	}
	// Check if target is an IPv6 host, i.e. [::1]:8886.
	if target[0] == '[' {
		end := strings.IndexByte(target, ']')
		if end < 0 || end+1 != i {
			return "", errAddress
		}
		return target, nil
	}
	// No IPv6 hostport, so target with colon must be a hostport or IPv6.
	ip := net.ParseIP(target)
	if len(ip) == net.IPv6len {
		return fmt.Sprintf("[%s]:%d", ip.String(), port), nil
	}
	return target, nil
}

var defaultPorts = map[bool]int{
	false: 1884,
	true:  8884,
}

func resolver(tls bool) func(ctx context.Context, target string) (net.Conn, error) {
	return func(ctx context.Context, target string) (net.Conn, error) {
		// TODO: If no port is specified, discover through SRV records (https://github.com/TheThingsNetwork/lorawan-stack/issues/138)
		target, err := defaultPort(target, defaultPorts[tls])
		if err != nil {
			return nil, err
		}
		return new(net.Dialer).DialContext(ctx, "tcp", target)
	}
}

// WithTransportCredentials returns gRPC dial options which configures connection level security credentials (e.g.,
// TLS/SSL) and discover the TLS/SSL listen port if not specified in the dial target.
func WithTransportCredentials(creds credentials.TransportCredentials) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(resolver(true)),
	}
}

// WithInsecure returns gRPC dial options which disable transport security and discover the default insecure listen
// port if not specified in the dial target.
func WithInsecure() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(resolver(false)),
	}
}

type tlsFallbackKeyType struct{}

var tlsFallbackKey tlsFallbackKeyType

// WithTLSFallback returns a derived context which is configured to fall back to the given TLS setting if discovery
// fails.
func WithTLSFallback(parent context.Context, tls bool) context.Context {
	return context.WithValue(parent, tlsFallbackKey, tls)
}

// DialOptions discovers gRPC dial options based on the given target. This includes whether or not transport level
// security is enabled and service port discovery.
func DialOptions(ctx context.Context, target string, creds credentials.TransportCredentials) ([]grpc.DialOption, error) {
	// TODO: Discover through SRV records and cache result (https://github.com/TheThingsNetwork/lorawan-stack/issues/138)
	if val, ok := ctx.Value(tlsFallbackKey).(bool); ok && !val {
		return WithInsecure(), nil
	}
	return WithTransportCredentials(creds), nil
}

// DialContext creates a client connection to the given target. It uses DialOptions to discover gRPC dial options for
// the target.
func DialContext(ctx context.Context, target string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	discoveredOpts, err := DialOptions(ctx, target, creds)
	if err != nil {
		return nil, err
	}
	allOpts := make([]grpc.DialOption, 0, len(discoveredOpts)+len(opts))
	allOpts = append(allOpts, discoveredOpts...)
	allOpts = append(allOpts, opts...)
	return grpc.DialContext(ctx, target, allOpts...)
}
