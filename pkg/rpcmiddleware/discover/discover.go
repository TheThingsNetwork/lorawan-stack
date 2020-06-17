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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	ttnpb.ClusterRole_TENANT_BILLING_SERVER:        "tbs",
	ttnpb.ClusterRole_EVENT_SERVER:                 "es",
}

// ServiceName returns th service name of the given role. This service name is typically used in SRV records.
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

type tlsFallbackKeyType struct{}

var tlsFallbackKey tlsFallbackKeyType

// WithTLSFallback returns a derived context which is configured to fall back to the given TLS setting if discovery fails.
func WithTLSFallback(parent context.Context, tls bool) context.Context {
	return context.WithValue(parent, tlsFallbackKey, tls)
}

// DNSResolver provides DNS lookup for discovery.
type DNSResolver interface {
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

type resolverKeyType struct{}

var resolverKey resolverKeyType

// WithDNSResolver returns a derived context which is configured to use the DNS resolver for discovery.
func WithDNSResolver(parent context.Context, resolver DNSResolver) context.Context {
	return context.WithValue(parent, resolverKey, resolver)
}

// DialContext creates a client connection to the given host using service discovery.
//
// This function performs a DNS SRV lookup if the target does not contain a port and the given role is discoverable.
// All discovered targets are assumed to use TLS.
//
// If the given role does not support discovery or if the DNS SRV lookup fails, this function falls back to the default port.
func DialContext(ctx context.Context, role ttnpb.ClusterRole, target string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var (
		addresses []string
		cred      grpc.DialOption
	)
	if _, _, err := net.SplitHostPort(target); err == nil {
		addresses = []string{target}
	} else {
		if service, ok := ServiceName(role); ok {
			resolver, ok := ctx.Value(resolverKey).(DNSResolver)
			if !ok {
				resolver = net.DefaultResolver
			}
			_, addrs, err := resolver.LookupSRV(ctx, service, "tcp", target)
			if err == nil {
				addresses = make([]string, len(addrs))
				for i, addr := range addrs {
					addresses[i] = fmt.Sprintf("%s:%d", strings.TrimSuffix(addr.Target, "."), addr.Port)
				}
				cred = grpc.WithTransportCredentials(creds)
			}
		}
		if len(addresses) == 0 {
			var defaultPort int
			if val, ok := ctx.Value(tlsFallbackKey).(bool); ok && !val {
				defaultPort = DefaultPorts[false]
			} else {
				defaultPort = DefaultPorts[true]
			}
			target, err := DefaultPort(target, defaultPort)
			if err != nil {
				return nil, err
			}
			addresses = []string{target}
		}
	}
	if cred == nil {
		if val, ok := ctx.Value(tlsFallbackKey).(bool); ok && !val {
			cred = grpc.WithInsecure()
		} else {
			cred = grpc.WithTransportCredentials(creds)
		}
	}

	logger := log.FromContext(ctx).WithField("target", target)
	var err error
	var conn *grpc.ClientConn
	for _, address := range addresses {
		logger := logger.WithError(err).WithField("address", address)
		logger.Debug("Dial target address")
		conn, err = grpc.DialContext(ctx, address, append(opts, cred, grpc.WithBlock(), grpc.FailOnNonTempDialError(true))...)
		if err == nil {
			return conn, nil
		}
		logger.Debug("Failed to dial target address")
	}
	return nil, err
}
