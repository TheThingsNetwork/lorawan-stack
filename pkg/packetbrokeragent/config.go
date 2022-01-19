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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"gopkg.in/square/go-jose.v2"
)

// Config configures Packet Broker clients.
type Config struct {
	Registration         RegistrationConfig `name:"registration" description:"Registration with Packet Broker"`
	IAMAddress           string             `name:"iam-address" description:"Address of Packet Broker IAM"`
	ControlPlaneAddress  string             `name:"control-plane-address" description:"Address of Packet Broker Control Plane"`
	DataPlaneAddress     string             `name:"data-plane-address" description:"Address of the Packet Broker Data Plane"`
	MapperAddress        string             `name:"mapper-address" description:"Address of Packet Broker Mapper"`
	Insecure             bool               `name:"insecure" description:"Connect without using TLS"`
	NetID                types.NetID        `name:"net-id" description:"LoRa Alliance NetID"`
	TenantID             string             `name:"tenant-id" description:"Tenant ID within the NetID"`
	ClusterID            string             `name:"cluster-id" description:"Cluster ID uniquely identifying the Forwarder in the NetID and Tenant ID"`
	HomeNetworkClusterID string             `name:"home-network-cluster-id" description:"Home Network Cluster ID, if different from the Cluster ID"`
	AuthenticationMode   string             `name:"authentication-mode" description:"Authentication mode (oauth2)"`
	OAuth2               OAuth2Config       `name:"oauth2" description:"OAuth 2.0 configuration"`
	Forwarder            ForwarderConfig    `name:"forwarder" description:"Forwarder configuration for publishing uplink messages and subscribing to downlink messages"`
	HomeNetwork          HomeNetworkConfig  `name:"home-network" description:"Home Network configuration for subscribing to uplink and publishing downlink messages"`
}

// RegistrationConfig defines the registration configuration for Packet Broker.
type RegistrationConfig struct {
	Name                  string            `name:"name" description:"Friendly name to register with Packet Broker"`
	AdministrativeContact ContactInfoConfig `name:"administrative-contact" description:"Administrative contact to register with Packet Broker"`
	TechnicalContact      ContactInfoConfig `name:"technical-contact" description:"Technical contact to register with Packet Broker"`
	Listed                bool              `name:"listed" description:"List the Home Network in the Packet Broker catalog"`
}

// ContactInfoConfig defines contact info for Packet Broker.
type ContactInfoConfig struct {
	Email string `name:"email" description:"Email address"`
}

// ContactInfo returns a ContactInfo proto for the configuration.
func (c ContactInfoConfig) ContactInfo(contactType ttnpb.ContactType) *ttnpb.ContactInfo {
	if c.Email == "" {
		return nil
	}
	return &ttnpb.ContactInfo{
		ContactType:   contactType,
		ContactMethod: ttnpb.ContactMethod_CONTACT_METHOD_EMAIL,
		Value:         c.Email,
	}
}

// OAuth2Config defines OAuth 2.0 configuration for authentication.
type OAuth2Config struct {
	ClientID     string `name:"client-id" description:"API key ID used as client ID"`
	ClientSecret string `name:"client-secret" description:"Secret API key value used as client secret"`
	TokenURL     string `name:"token-url" description:"Token URL"`
}

// ForwarderConfig defines configuration of the Forwarder role.
type ForwarderConfig struct {
	Enable            bool             `name:"enable" description:"Enable Forwarder role"`
	WorkerPool        WorkerPoolConfig `name:"worker-pool" description:"Workers pool configuration"`
	TokenKey          []byte           `name:"token-key" description:"AES 128 or 256-bit key for encrypting tokens"`
	TokenEncrypter    jose.Encrypter   `name:"-"`
	IncludeGatewayEUI bool             `name:"include-gateway-eui" description:"Include the gateway EUI in forwarded metadata"`
	IncludeGatewayID  bool             `name:"include-gateway-id" description:"Include the gateway ID in forwarded metadata"`
	HashGatewayID     bool             `name:"hash-gateway-id" description:"Hash the gateway ID (if forwarded in the metadata)"`
	GatewayOnlineTTL  time.Duration    `name:"gateway-online-ttl" description:"Time-to-live of online status reported to Packet Broker"`
}

// HomeNetworkConfig defines the configuration of the Home Network role.
type HomeNetworkConfig struct {
	Enable          bool                  `name:"enable" description:"Enable Home Network role"`
	DevAddrPrefixes []types.DevAddrPrefix `name:"dev-addr-prefixes" description:"DevAddr prefixes to subscribe to"`
	WorkerPool      WorkerPoolConfig      `name:"worker-pool" description:"Workers pool configuration"`
	IncludeHops     bool                  `name:"include-hops" description:"Include hops in the metadata"`
}

// WorkerPoolConfig contains the worker pool configuration for a Packet Broker role.
type WorkerPoolConfig struct {
	Limit int `name:"limit" description:"Limit of active workers"`
}
