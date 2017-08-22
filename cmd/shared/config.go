// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package shared

import "github.com/TheThingsNetwork/ttn/pkg/config"

// DefaultBaseConfig is the default base component configuration.
var DefaultBaseConfig = config.Base{}

// DefaultTLSConfig is the default TLS config.
var DefaultTLSConfig = config.TLS{
	Certificate: "cert.pem",
	Key:         "key.pem",
}

// DefaultIdentityConfig is the default Identity config
var DefaultIdentityConfig = config.Identity{
	Servers: map[string]string{
		"ttn-account-v2":  "https://account.thethingsnetwork.org",
		"ttn-identity-v3": "https://identity.thethingsnetwork.org",
	},
}

// DefaultGRPCConfig is the default config for GRPC.
var DefaultGRPCConfig = config.GRPC{}

// DefaultRedisConfig is the default config for Redis.
var DefaultRedisConfig = config.Redis{
	Address:  "localhost:3479",
	Database: 0,
}

// DefaultServiceBase is the default base config for a service.
var DefaultServiceBase = config.ServiceBase{
	Base:     DefaultBaseConfig,
	GRPC:     DefaultGRPCConfig,
	TLS:      DefaultTLSConfig,
	Identity: DefaultIdentityConfig,
}
