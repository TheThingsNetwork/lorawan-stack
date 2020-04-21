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

package packetbrokeragent_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"time"

	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/packetbrokeragent/mock"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

func mustServePBDataPlane(ctx context.Context) (*mock.PBDataPlane, net.Addr) {
	cert, err := tls.LoadX509KeyPair("testdata/servercert.pem", "testdata/serverkey.pem")
	if err != nil {
		panic(err)
	}
	clientCA, err := ioutil.ReadFile("testdata/clientca.pem")
	if err != nil {
		panic(err)
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(clientCA) {
		panic("failed to append client CA from PEM")
	}
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	dp := mock.NewPBDataPlane(cert, clientCAs)
	go dp.Serve(lis)
	go func() {
		<-ctx.Done()
		dp.GracefulStop()
	}()
	return dp, lis.Addr()
}
