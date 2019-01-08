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

package joinserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config represents the JoinServer configuration.
type Config struct {
	Devices                  DeviceRegistry      `name:"-"`
	Keys                     KeyRegistry         `name:"-"`
	JoinEUIPrefixes          []types.EUI64Prefix `name:"join-eui-prefix" description:"JoinEUI prefixes handled by this JS"`
	NetworkCryptoService     CryptoServiceConfig `name:"network-crypto-service" description:"Crypto service for network layer operations"`
	ApplicationCryptoService CryptoServiceConfig `name:"application-crypto-service" description:"Crypto service for application layer operations"`
}

// CryptoServiceConfig defines configuration of a crypto service.
type CryptoServiceConfig struct {
	Enabled bool        `name:"enabled" description:"Enable the crypto service"`
	Address string      `name:"address" description:"Address of the crypto service"`
	TLS     *config.TLS `name:"tls" description:"TLS client authentication configuration for the crypto service"`
}

func (c CryptoServiceConfig) dial(ctx context.Context) (*grpc.ClientConn, error) {
	opts := rpcclient.DefaultDialOptions(ctx)
	if c.TLS != nil {
		tlsConfig, err := c.TLS.Config(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.DialContext(ctx, c.Address, opts...)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	return conn, nil
}

// DialNetworkCryptoService dials the crypto service for network layer operations.
func (c Config) DialNetworkCryptoService(ctx context.Context, keyVault crypto.KeyVault) (NetworkCryptoService, error) {
	conn, err := c.NetworkCryptoService.dial(ctx)
	if err != nil {
		return nil, err
	}
	client := ttnpb.NewNetworkCryptoServiceClient(conn)
	return &NetworkCryptoServiceRPCClient{
		Client:   client,
		KeyVault: keyVault,
	}, nil
}

// DialApplicationCryptoService dials the crypto service for application layer operations.
func (c Config) DialApplicationCryptoService(ctx context.Context, keyVault crypto.KeyVault) (ApplicationCryptoService, error) {
	conn, err := c.ApplicationCryptoService.dial(ctx)
	if err != nil {
		return nil, err
	}
	client := ttnpb.NewApplicationCryptoServiceClient(conn)
	return &ApplicationCryptoServiceRPCClient{
		Client:   client,
		KeyVault: keyVault,
	}, nil
}
