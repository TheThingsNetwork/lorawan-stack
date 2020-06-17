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

	stderrors "errors"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

var errAddress = errors.DefineInvalidArgument("address", "invalid address")

// DefaultPort appends port if target does not already have one.
func DefaultPort(target string, port int) (string, error) {
	i := strings.LastIndexByte(target, ':')
	if i < 0 {
		return fmt.Sprintf("%s:%d", target, port), nil
	}
	// Check if target is an IPv6 host, i.e. [::1]:8886.
	if target[0] == '[' {
		end := strings.IndexByte(target, ']')
		if end < 0 || end+1 != i {
			return "", errAddress.New()
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

// DefaultPorts is a map of the default gRPC ports, with/without TLS.
var DefaultPorts = map[bool]int{
	false: 1884,
	true:  8884,
}

// DefaultHTTPPorts is a map of the default HTTP ports, with/without TLS.
var DefaultHTTPPorts = map[bool]int{
	false: 80,
	true:  443,
}

// HTTPScheme is a map of the HTTP schemes, with/without TLS.
var HTTPScheme = map[bool]string{
	false: "http",
	true:  "https",
}

var clusterRoleServices = map[ttnpb.ClusterRole]string{
	ttnpb.ClusterRole_ENTITY_REGISTRY:              "is",
	ttnpb.ClusterRole_ACCESS:                       "is",
	ttnpb.ClusterRole_GATEWAY_SERVER:               "gs",
	ttnpb.ClusterRole_NETWORK_SERVER:               "ns",
	ttnpb.ClusterRole_APPLICATION_SERVER:           "as",
	ttnpb.ClusterRole_JOIN_SERVER:                  "js",
	ttnpb.ClusterRole_DEVICE_TEMPLATE_CONVERTER:    "dtc",
	ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER:       "dcs",
	ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER: "gcs",
	ttnpb.ClusterRole_QR_CODE_GENERATOR:            "qrg",
}

// ServiceName returns the service name of the given role. This service name is typically used in SRV records.
func ServiceName(role ttnpb.ClusterRole) (string, bool) {
	service, ok := clusterRoleServices[role]
	if !ok {
		return "", false
	}
	return fmt.Sprintf("ttn-v3-%s-grpc", service), true
}

// DefaultURL appends protocol and port if target does not already have one.
func DefaultURL(target string, port int, tls bool) (string, error) {
	target, err := DefaultPort(target, port)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s://%s", HTTPScheme[tls], target), nil
}

// DNSResolver provides DNS lookup for discovery.
type DNSResolver interface {
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

type options struct {
	insecureFallback bool
	dnsResolver      DNSResolver
	dialer           func(context.Context, string) (net.Conn, error)
}

// Option configures the discovery dialing.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(opts *options) {
	f(opts)
}

// WithInsecureFallback configures the dialer to use the default insecure port when discovery fails.
// Use this option with gRPC's WithInsecure dial option.
func WithInsecureFallback() Option {
	return optionFunc(func(opts *options) {
		opts.insecureFallback = true
	})
}

// WithDNSResolver configures the dialer with a custom DNS resolver.
func WithDNSResolver(resolver DNSResolver) Option {
	return optionFunc(func(opts *options) {
		opts.dnsResolver = resolver
	})
}

// WithAddressDialer configures the discovery dialer with a custom address dialer.
// This dialer is called for each discovered address until a nil error is returned.
func WithAddressDialer(dialer func(context.Context, string) (net.Conn, error)) Option {
	return optionFunc(func(opts *options) {
		opts.dialer = dialer
	})
}

// WithDialer returns a gRPC dialer that uses service discovery.
//
// Service discovery performs a DNS SRV lookup on the given target if it does not contain a port and the given role is
// discoverable.
//
// If the given role does not support discovery or if the DNS SRV lookup fails, this function falls back to the default port.
func WithDialer(role ttnpb.ClusterRole, opts ...Option) grpc.DialOption {
	options := &options{
		insecureFallback: false,
		dnsResolver:      net.DefaultResolver,
		dialer: func(ctx context.Context, address string) (net.Conn, error) {
			return new(net.Dialer).DialContext(ctx, "tcp", address)
		},
	}
	for _, opt := range opts {
		opt.apply(options)
	}

	return grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
		var addresses []string
		if _, _, err := net.SplitHostPort(target); err == nil {
			addresses = []string{target}
		} else {
			if service, ok := ServiceName(role); ok {
				_, addrs, err := options.dnsResolver.LookupSRV(ctx, service, "tcp", target)
				var dnsErr *net.DNSError
				if err != nil && stderrors.As(err, &dnsErr) && !dnsErr.IsNotFound {
					return nil, err
				}
				if err == nil {
					addresses = make([]string, len(addrs))
					for i, addr := range addrs {
						addresses[i] = fmt.Sprintf("%s:%d", strings.TrimSuffix(addr.Target, "."), addr.Port)
					}
				}
			}
			if len(addresses) == 0 {
				target, err := DefaultPort(target, DefaultPorts[!options.insecureFallback])
				if err != nil {
					return nil, err
				}
				addresses = []string{target}
			}
		}

		var err error
		var conn net.Conn
		for _, address := range addresses {
			conn, err = options.dialer(ctx, address)
			if err == nil {
				return conn, nil
			}
		}
		return nil, err
	})
}

// DialContext creates a client connection to the given host using service discovery. See WithDialer for more information.
// To configure a DNS resolver, an address dialer or fallback to the default insecure port, use WithDialer with options.
func DialContext(ctx context.Context, role ttnpb.ClusterRole, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target, append(opts, WithDialer(role))...)
}
