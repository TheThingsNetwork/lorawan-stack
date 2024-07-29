// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package ttgc provides functions to use The Things Gateway Controller.
package ttgc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	errDialGatewayServer = errors.DefineAborted("dial_gateway_server", "dial Gateway Gerver")
	errGatewayServerTLS  = errors.DefineAborted(
		"gateway_server_tls", "establish TLS connection with Gateway Server",
	)
)

func (u *Upstream) getRootCA(ctx context.Context, address string) (*x509.Certificate, error) {
	d := new(net.Dialer)
	netConn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, errDialGatewayServer.WithCause(err)
	}
	defer netConn.Close()

	tlsConfig, err := u.GetTLSClientConfig(ctx)
	if err != nil {
		return nil, err
	}
	tlsConn := tls.Client(netConn, tlsConfig)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, errGatewayServerTLS.WithCause(err)
	}

	state := tlsConn.ConnectionState()
	verifiedChain := state.VerifiedChains[0]
	return verifiedChain[len(verifiedChain)-1], nil
}
