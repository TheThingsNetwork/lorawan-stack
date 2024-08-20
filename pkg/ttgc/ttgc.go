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

// Package ttgc implements configuration and a client for The Things Gateway Controller.
package ttgc

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	ttica "go.thethings.industries/pkg/ca"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Component is the component interface required for this package.
type Component interface {
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
}

// Client is a client for The Things Gateway Controller.
type Client struct {
	*grpc.ClientConn
	domain string
}

// NewClient returns a new client for The Things Gateway Controller.
func NewClient(
	ctx context.Context,
	c Component,
	config Config,
	dialOpts ...grpc.DialOption,
) (*Client, error) {
	tlsConfig, err := c.GetTLSClientConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Add The Things Industries Root or Test CA if the Gateway Controller address is the production or staging address.
	// Otherwise, the root CAs must be configured via global TLS client configuration or system roots.
	var knownDeploymentPEM []byte
	switch config.Address {
	case "gc.thethings.industries:443":
		knownDeploymentPEM = ttica.RootCA
	case "gc.thethingslabs.com:443":
		knownDeploymentPEM = ttica.TestCA
	}
	if knownDeploymentPEM != nil {
		if tlsConfig.RootCAs == nil {
			tlsConfig.RootCAs = x509.NewCertPool()
		}
		tlsConfig.RootCAs.AppendCertsFromPEM(knownDeploymentPEM)
	}

	if err := config.TLS.ApplyTo(tlsConfig); err != nil {
		return nil, err
	}
	opts := rpcclient.DefaultDialOptions(ctx)
	opts = append(opts, dialOpts...)
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	cc, err := grpc.NewClient(config.Address, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		ClientConn: cc,
		domain:     config.Domain,
	}, nil
}

// Domain returns the domain of the client.
func (c *Client) Domain(context.Context) string {
	return c.domain
}
