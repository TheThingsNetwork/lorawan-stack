// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package shared

import (
	"github.com/TheThingsNetwork/ttn/pkg/applicationserver"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
	"github.com/TheThingsNetwork/ttn/pkg/joinserver"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/networkserver"
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

// DefaultClusterConfig is the default cluster configuration.
var DefaultClusterConfig = config.Cluster{}

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
var DefaultGRPCConfig = config.GRPC{
	Listen: ":8088",
}

// DefaultRedisConfig is the default config for Redis.
var DefaultRedisConfig = config.Redis{
	Address:  "localhost:6379",
	Database: 0,
	Prefix:   "ttn",
}

// DefaultServiceBase is the default base config for a service.
var DefaultServiceBase = config.ServiceBase{
	Base:     DefaultBaseConfig,
	Cluster:  DefaultClusterConfig,
	Redis:    DefaultRedisConfig,
	GRPC:     DefaultGRPCConfig,
	HTTP:     DefaultHTTPConfig,
	TLS:      DefaultTLSConfig,
	Identity: DefaultIdentityConfig,
}

// DefaultIdentityServerConfig is the default configuration for the IdentityServer
var DefaultIdentityServerConfig = identityserver.Config{
	DatabaseURI:      "postgres://root@localhost:26257/is_development?sslmode=disable",
	Hostname:         "localhost",
	OrganizationName: "The Things Network",
}

// DefaultGatewayServerConfig is the default configuration for the GatewayServer
var DefaultGatewayServerConfig = gatewayserver.Config{}

// DefaultNetworkServerConfig is the default configuration for the NetworkServer
var DefaultNetworkServerConfig = networkserver.Config{}

// DefaultApplicationServerConfig is the default configuration for the ApplicationServer
var DefaultApplicationServerConfig = applicationserver.Config{}

// DefaultJoinServerConfig is the default configuration for the JoinServer
var DefaultJoinServerConfig = joinserver.Config{}
