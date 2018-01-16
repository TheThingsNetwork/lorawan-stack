// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package shared

import (
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// DefaultBaseConfig is the default base component configuration.
var DefaultBaseConfig = config.Base{
	Log: DefaultLogConfig,
}

// DefaultLogConfig is the default log configuration.
var DefaultLogConfig = config.Log{
	Level: log.InfoLevel,
}

// DefaultTLSConfig is the default TLS config.
var DefaultTLSConfig = config.TLS{
	Certificate: "",
	Key:         "",
}

// DefaultHTTPConfig is the default HTTP config.
var DefaultHTTPConfig = config.HTTP{
	Listen: ":8080",
	PProf:  true,
}

// DefaultIdentityConfig is the default Identity config.
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
	Address:  "localhost:6379",
	Database: 0,
}

// DefaultServiceBase is the default base config for a service.
var DefaultServiceBase = config.ServiceBase{
	Base:     DefaultBaseConfig,
	Redis:    DefaultRedisConfig,
	GRPC:     DefaultGRPCConfig,
	HTTP:     DefaultHTTPConfig,
	TLS:      DefaultTLSConfig,
	Identity: DefaultIdentityConfig,
}
