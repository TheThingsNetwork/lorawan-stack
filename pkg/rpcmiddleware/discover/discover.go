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
	"sync"
	"time"

	stderrors "errors"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
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

// DefaultPorts is a map of the default gRPC ports, with/without TLS.
var DefaultPorts = map[bool]int{
	false: 1884,
	true:  8884,
}

var errRole = errors.DefineFailedPrecondition("role", "service role `{role}` not discoverable")

// Scheme returns the gRPC scheme for service discovery.
func Scheme(role ttnpb.ClusterRole) (string, error) {
	svc, ok := clusterRoleServices[role]
	if !ok {
		return "", errRole.WithAttributes("role", strings.Title(strings.Replace(role.String(), "_", " ", -1)))
	}
	return fmt.Sprintf("ttn-v3-%s", svc), nil
}

// Address returns the host with service discovery gRPC scheme for the given role.
func Address(role ttnpb.ClusterRole, host string) (string, error) {
	scheme, err := Scheme(role)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:///%s", scheme, host), nil
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

// DefaultURL appends protocol and port if target does not already have one.
func DefaultURL(target string, port int, tls bool) (string, error) {
	if port != DefaultHTTPPorts[tls] {
		var err error
		target, err = DefaultPort(target, port)
		if err != nil {
			return "", nil
		}
	}
	return fmt.Sprintf("%s://%s", HTTPScheme[tls], target), nil
}

// DNS provides DNS lookup functionality for service discovery.
type DNS interface {
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

type options struct {
	dns DNS
}

// Option configures the discovery dialing.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(opts *options) {
	f(opts)
}

// WithDNS configures the resolver with a custom DNS resolver.
func WithDNS(dns DNS) Option {
	return optionFunc(func(opts *options) {
		opts.dns = dns
	})
}

// NewBuilder returns a new resolver builder for service discovery.
func NewBuilder(scheme string, opts ...Option) resolver.Builder {
	options := options{
		dns: net.DefaultResolver,
	}
	for _, opt := range opts {
		opt.apply(&options)
	}

	return &clusterBuilder{
		scheme:  scheme,
		options: options,
	}
}

type clusterBuilder struct {
	scheme  string
	options options
}

func (r *clusterBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// Use passthrough when the endpoint is an address.
	if _, _, err := net.SplitHostPort(target.Endpoint); err == nil {
		return resolver.Get("passthrough").Build(target, cc, opts)
	}

	ctx, cancel := context.WithCancel(context.Background())
	service := target.Scheme + "-grpc"
	res := &clusterResolver{
		ctx:     ctx,
		cancel:  cancel,
		dns:     r.options.dns,
		service: service,
		name:    target.Endpoint,
		cc:      cc,
		rn:      make(chan struct{}, 1),
	}
	res.wg.Add(1)
	go res.watch()
	res.ResolveNow(resolver.ResolveNowOptions{})
	return res, nil
}

func (r *clusterBuilder) Scheme() string {
	return r.scheme
}

type clusterResolver struct {
	ctx    context.Context
	cancel context.CancelFunc
	dns    DNS
	service,
	name string
	cc resolver.ClientConn
	rn chan struct{}
	wg sync.WaitGroup
}

func (r *clusterResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	select {
	case r.rn <- struct{}{}:
	default:
	}
}

func (r *clusterResolver) Close() {
	r.cancel()
	r.wg.Wait()
}

func (r *clusterResolver) watch() {
	defer r.wg.Done()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.rn:
		}

		state, err := r.lookup()
		if err != nil {
			r.cc.ReportError(err)
		} else {
			r.cc.UpdateState(*state)
		}

		debounceTimer := time.NewTimer(30 * time.Second)
		select {
		case <-r.ctx.Done():
			debounceTimer.Stop()
			return
		case <-debounceTimer.C:
		}
	}
}

func (r *clusterResolver) lookup() (*resolver.State, error) {
	_, addrs, err := r.dns.LookupSRV(r.ctx, r.service, "tcp", r.name)
	var dnsErr *net.DNSError
	if err != nil && stderrors.As(err, &dnsErr) && !dnsErr.IsNotFound {
		return nil, err
	}
	state := new(resolver.State)
	if len(addrs) == 0 {
		addr, err := DefaultPort(r.name, DefaultPorts[true])
		if err != nil {
			return nil, err
		}
		state.Addresses = []resolver.Address{
			{
				Addr:       addr,
				ServerName: r.name,
			},
		}
	} else {
		state.Addresses = make([]resolver.Address, len(addrs))
		for i, addr := range addrs {
			name := strings.TrimSuffix(addr.Target, ".")
			state.Addresses[i] = resolver.Address{
				Addr:       fmt.Sprintf("%s:%d", name, addr.Port),
				ServerName: name,
				Attributes: attributes.New(
					"priority", int(addr.Priority),
					"weight", int(addr.Weight),
				),
			}
		}
	}
	return state, nil
}

func init() {
	m := make(map[string]struct{}, len(clusterRoleServices))
	for role := range clusterRoleServices {
		scheme, err := Scheme(role)
		if err != nil {
			panic(err)
		}
		if _, ok := m[scheme]; ok {
			continue
		}
		resolver.Register(NewBuilder(scheme))
		m[scheme] = struct{}{}
	}
}
