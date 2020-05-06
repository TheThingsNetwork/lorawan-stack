// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"context"
	"crypto/tls"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"gopkg.in/square/go-jose.v2"
)

// Config configures Packet Broker clients.
type Config struct {
	DataPlaneAddress string            `name:"data-plane-address" description:"Address of the Packet Broker Data Plane"`
	NetID            types.NetID       `name:"net-id" description:"LoRa Alliance NetID"`
	TenantID         string            `name:"tenant-id" description:"Tenant ID within the NetID"`
	ClusterID        string            `name:"cluster-id" description:"Cluster ID uniquely identifying this cluster within a NetID and tenant"`
	TLS              TLSConfig         `name:"tls"`
	Forwarder        ForwarderConfig   `name:"forwarder" description:"Forwarder configuration for publishing uplink messages and subscribing to downlink messages"`
	HomeNetwork      HomeNetworkConfig `name:"home-network" description:"Home Network configuration for subscribing to uplink and publishing downlink messages"`
}

// ForwarderConfig defines configuration of the Forwarder role.
type ForwarderConfig struct {
	Enable         bool             `name:"enable" description:"Enable Forwarder role"`
	WorkerPool     WorkerPoolConfig `name:"worker-pool" description:"Workers pool configuration"`
	TokenKey       []byte           `name:"token-key" description:"AES 128 or 256-bit key for encrypting tokens"`
	TokenEncrypter jose.Encrypter   `name:"-"`
}

// HomeNetworkConfig defines the configuration of the Home Network role.
type HomeNetworkConfig struct {
	Enable             bool                  `name:"enable" description:"Enable Home Network role"`
	DevAddrPrefixes    []types.DevAddrPrefix `name:"dev-addr-prefixes" description:"DevAddr prefixes to subscribe to"`
	WorkerPool         WorkerPoolConfig      `name:"worker-pool" description:"Workers pool configuration"`
	BlacklistForwarder bool                  `name:"blacklist-forwarder" description:"Blacklist traffic from Forwarder to avoid traffic loops"`
}

// WorkerPoolConfig contains the worker pool configuration for a Packet Broker role.
type WorkerPoolConfig struct {
	Limit int `name:"limit" description:"Limit of active workers"`
}

// TLSConfig contains TLS configuration for connecting to Packet Broker.
type TLSConfig struct {
	Source      string             `name:"source" description:"Source of the TLS certificate (file, key-vault)"`
	Certificate string             `name:"certificate" description:"Location of TLS certificate"`
	Key         string             `name:"key" description:"Location of TLS private key"`
	KeyVault    config.TLSKeyVault `name:"key-vault"`
}

var errNoTLSCertificate = errors.DefineFailedPrecondition("no_tls_certificate", "no TLS certificate configured")

func (c TLSConfig) loadCertificate(ctx context.Context, keyVault crypto.KeyVault) (tls.Certificate, error) {
	switch c.Source {
	case "file":
		return tls.LoadX509KeyPair(c.Certificate, c.Key)
	case "key-vault":
		cert, err := keyVault.ExportCertificate(ctx, c.KeyVault.ID)
		if err != nil {
			return tls.Certificate{}, err
		}
		return *cert, nil
	default:
		return tls.Certificate{}, errNoTLSCertificate.New()
	}
}
