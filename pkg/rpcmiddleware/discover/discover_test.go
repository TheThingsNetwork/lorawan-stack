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

package discover_test

import (
	"context"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/discover"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc/resolver"
)

type mockResolver struct {
	LookupSRVFunc func(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

func (r *mockResolver) LookupSRV(
	ctx context.Context, service, proto, name string,
) (cname string, addrs []*net.SRV, err error) {
	if r.LookupSRVFunc == nil {
		panic("LookupSRVFunc called, but not set")
	}
	return r.LookupSRVFunc(ctx, service, proto, name)
}

type mockClientConn struct {
	resolver.ClientConn
	UpdateStateFunc func(resolver.State) error
	ReportErrorFunc func(error)
}

func (c *mockClientConn) UpdateState(state resolver.State) error {
	if c.UpdateStateFunc == nil {
		panic("UpdateStateFunc called, but not set")
	}
	return c.UpdateStateFunc(state)
}

func (c *mockClientConn) ReportError(err error) {
	if c.ReportErrorFunc == nil {
		panic("ReportErrorFunc called, but not set")
	}
	c.ReportErrorFunc(err)
}

func TestResolver(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name               string
		LookupResult       []*net.SRV
		LookupError        error
		AddressesAssertion func(*testing.T, []string) bool
		ErrorAssertion     func(*testing.T, error) bool
	}{
		{
			Name: "SRVNotFound",
			LookupError: &net.DNSError{
				Err:        "not found",
				IsNotFound: true,
			},
			AddressesAssertion: func(t *testing.T, addresses []string) bool {
				t.Helper()
				// SRV not set; use default port.
				return assertions.New(t).So(addresses, should.Resemble, []string{"localhost:8884"})
			},
		},
		{
			Name: "LookupSRVFailure",
			LookupError: &net.DNSError{
				Err: "dns failure",
			},
			AddressesAssertion: func(t *testing.T, addresses []string) bool {
				t.Helper()
				return assertions.New(t).So(addresses, should.BeEmpty) // DNS failure; nothing gets dialed.
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.NotBeNil)
			},
		},
		{
			Name: "SkipBogusRecords",
			LookupResult: []*net.SRV{
				{
					Target:   "invalid.",
					Port:     1234,
					Priority: 100,
				},
				{
					Target:   "invalid.",
					Port:     4321,
					Priority: 90,
				},
				{
					Target:   "localhost.",
					Port:     8884,
					Priority: 10,
				},
			},
			AddressesAssertion: func(t *testing.T, addresses []string) bool {
				t.Helper()
				return assertions.New(t).So(addresses, should.Resemble, []string{
					"invalid:1234",
					"invalid:4321",
					"localhost:8884",
				})
			},
		},
		{
			Name: "OnlyBogusRecords",
			LookupResult: []*net.SRV{
				{
					Target:   "invalid.",
					Port:     1234,
					Priority: 100,
				},
				{
					Target:   "invalid.",
					Port:     4321,
					Priority: 90,
				},
			},
			AddressesAssertion: func(t *testing.T, addresses []string) bool {
				t.Helper()
				return assertions.New(t).So(addresses, should.Resemble, []string{
					"invalid:1234",
					"invalid:4321",
				})
			},
		},
		{
			Name: "Multiple",
			LookupResult: []*net.SRV{
				{
					Target:   "localhost.",
					Port:     8884,
					Priority: 100,
					Weight:   1,
				},
				{
					Target:   "localhost.",
					Port:     8885,
					Priority: 100,
					Weight:   2,
				},
			},
			AddressesAssertion: func(t *testing.T, addresses []string) bool {
				t.Helper()
				return assertions.New(t).So(addresses, should.Resemble, []string{"localhost:8884", "localhost:8885"})
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			dns := &mockResolver{
				LookupSRVFunc: func(
					_ context.Context, service, proto, name string,
				) (cname string, addrs []*net.SRV, err error) {
					if tc.LookupError != nil {
						return "", nil, tc.LookupError
					}
					a := assertions.New(t)
					if !a.So(service, should.Equal, "ttn-v3-gs-grpc") ||
						!a.So(proto, should.Equal, "tcp") ||
						!a.So(name, should.Equal, "localhost") {
						return "", nil, &net.DNSError{Err: "invalid request"}
					}
					return "test", tc.LookupResult, nil
				},
			}
			builder := discover.NewBuilder("ttn-v3-gs", discover.WithDNS(dns))

			var (
				resolveState resolver.State
				resolveErr   error
				resolveDone  = make(chan struct{})
			)
			clientConn := &mockClientConn{
				UpdateStateFunc: func(state resolver.State) error {
					resolveState = state
					close(resolveDone)
					return nil
				},
				ReportErrorFunc: func(err error) {
					resolveErr = err
					close(resolveDone)
				},
			}
			res, err := builder.Build(
				resolver.Target{
					URL: url.URL{
						Scheme: "ttn-v3-gs",
						Opaque: "localhost",
					},
				},
				clientConn,
				resolver.BuildOptions{},
			)
			if err != nil {
				t.Fatalf("Failed to build resolver: %v", err)
			}
			defer res.Close()

			select {
			case <-resolveDone:
			case <-time.After(test.Delay << 10):
				t.Fatal("Timeout waiting for resolver to resolve")
			}

			addresses := make([]string, len(resolveState.Addresses))
			for i, a := range resolveState.Addresses {
				addresses[i] = a.Addr
			}

			if resolveErr != nil {
				if tc.ErrorAssertion == nil {
					t.Fatalf("Unexpected error: %v", resolveErr)
				}
				if !tc.ErrorAssertion(t, resolveErr) {
					t.FailNow()
				}
			} else {
				if tc.ErrorAssertion != nil {
					t.Fatal("Expected error but got none")
				}
			}

			if !tc.AddressesAssertion(t, addresses) {
				t.FailNow()
			}
		})
	}
}

func TestDefaultPort(t *testing.T) {
	t.Parallel()
	for input, expected := range map[string]string{
		"localhost:http": "localhost:http",
		"localhost:80":   "localhost:80",
		"localhost":      "localhost:8884",
		"[::1]:80":       "[::1]:80",
		"::1":            "[::1]:8884",
		"192.168.1.1:80": "192.168.1.1:80",
		"192.168.1.1":    "192.168.1.1:8884",
		":80":            ":80",
		"":               ":8884",
		"[::]:80":        "[::]:80",
		"::":             "[::]:8884",
		"[::]":           "", // Invalid address
		"[::":            "", // Invalid address
	} {
		input, expected := input, expected
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			target, err := discover.DefaultPort(input, 8884)
			if err != nil {
				target = ""
			}
			assertions.New(t).So(target, should.Equal, expected)
		})
	}
}

func TestDefaultURL(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		target   string
		port     int
		tls      bool
		expected string
	}{
		{
			target:   "localhost",
			port:     80,
			tls:      false,
			expected: "http://localhost",
		},
		{
			target:   "localhost",
			port:     8080,
			tls:      false,
			expected: "http://localhost:8080",
		},
		{
			target:   "host.with.port:http",
			port:     8000,
			tls:      false,
			expected: "http://host.with.port:http",
		},
		{
			target:   "hostname:433",
			port:     4000,
			tls:      true,
			expected: "https://hostname:433",
		},
		{
			target:   "hostname",
			port:     443,
			tls:      true,
			expected: "https://hostname",
		},
		{
			target:   "hostname",
			port:     8443,
			tls:      true,
			expected: "https://hostname:8443",
		},
	} {
		tc := tc
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			target, err := discover.DefaultURL(tc.target, tc.port, tc.tls)
			if err != nil {
				target = ""
			}
			assertions.New(t).So(target, should.Equal, tc.expected)
		})
	}
}
