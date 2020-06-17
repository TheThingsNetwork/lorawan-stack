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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/discover"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type mockResolver struct {
	LookupSRVFunc func(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

func (r *mockResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	if r.LookupSRVFunc == nil {
		panic("LookupSRVFunc called, but not set")
	}
	return r.LookupSRVFunc(ctx, service, proto, name)
}

func TestDialContext(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	serverCert := test.Must(tls.LoadX509KeyPair("testdata/servercert.pem", "testdata/serverkey.pem")).(tls.Certificate)
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      x509.NewCertPool(),
	}
	clientCA := test.Must(ioutil.ReadFile("testdata/clientca.pem")).([]byte)
	serverTLSConfig.RootCAs.AppendCertsFromPEM(clientCA)

	listen := func(addr string) (port int, address string, lis net.Listener) {
		lis = test.Must(tls.Listen("tcp", addr, serverTLSConfig)).(net.Listener)
		go grpc.NewServer().Serve(lis)
		port = lis.Addr().(*net.TCPAddr).Port
		address = fmt.Sprintf("localhost:%d", port)
		return
	}
	lis1Port, lis1Address, lis1 := listen(":0")
	defer lis1.Close()
	lis2Port, lis2Address, lis2 := listen(fmt.Sprintf(":%d", discover.DefaultPorts[true]))
	defer lis2.Close()

	for _, tc := range []struct {
		Name                   string
		LookupResult           []*net.SRV
		LookupError            error
		DialAddressesAssertion func(*testing.T, []string) bool
		ErrorAssertion         func(*testing.T, error) bool
	}{
		{
			Name:        "LookupSRVFailure",
			LookupError: &net.DNSError{Err: "test error"},
			DialAddressesAssertion: func(t *testing.T, addresses []string) bool {
				return assertions.New(t).So(addresses, should.Resemble, []string{lis2Address}) // Lookup failure; use default port
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
					Port:     uint16(lis1Port),
					Priority: 10,
				},
			},
			DialAddressesAssertion: func(t *testing.T, addresses []string) bool {
				return assertions.New(t).So(addresses, should.Resemble, []string{
					"invalid:1234",
					"invalid:4321",
					lis1Address,
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
			DialAddressesAssertion: func(t *testing.T, addresses []string) bool {
				return assertions.New(t).So(addresses, should.Resemble, []string{
					"invalid:1234",
					"invalid:4321",
				})
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.NotBeNil)
			},
		},
		{
			Name: "PickFirst",
			LookupResult: []*net.SRV{
				{
					Target:   "localhost.",
					Port:     uint16(lis1Port),
					Priority: 100,
					Weight:   1,
				},
				{
					Target:   "localhost.",
					Port:     uint16(lis2Port),
					Priority: 100,
					Weight:   2,
				},
			},
			DialAddressesAssertion: func(t *testing.T, addresses []string) bool {
				return assertions.New(t).So(addresses, should.Resemble, []string{lis1Address})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			resolver := &mockResolver{
				LookupSRVFunc: func(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
					if tc.LookupError != nil {
						return "", nil, tc.LookupError
					}
					return "test", tc.LookupResult, nil
				},
			}

			clientCert := test.Must(tls.LoadX509KeyPair("testdata/clientcert.pem", "testdata/clientkey.pem")).(tls.Certificate)
			clientTLSConfig := &tls.Config{
				Certificates: []tls.Certificate{clientCert},
				RootCAs:      x509.NewCertPool(),
			}
			serverCA := test.Must(ioutil.ReadFile("testdata/serverca.pem")).([]byte)
			clientTLSConfig.RootCAs.AppendCertsFromPEM(serverCA)

			var dialAddresses []string
			conn, err := discover.DialContext(
				discover.WithDNSResolver(ctx, resolver),
				ttnpb.ClusterRole_GATEWAY_SERVER,
				"localhost",
				credentials.NewTLS(clientTLSConfig),
				grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
					dialAddresses = append(dialAddresses, address)
					return new(net.Dialer).DialContext(ctx, "tcp", address)
				}),
			)

			if err != nil {
				if tc.ErrorAssertion == nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !tc.ErrorAssertion(t, err) {
					t.FailNow()
				}
			} else {
				defer conn.Close()
				if tc.ErrorAssertion != nil {
					t.Fatal("Expected error but got none")
				}
			}

			if !tc.DialAddressesAssertion(t, dialAddresses) {
				t.FailNow()
			}
		})
	}
}
