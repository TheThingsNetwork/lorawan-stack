// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package shared

import (
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/middleware/sentry"
	raven "github.com/getsentry/raven-go"
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

// SentryMiddleware generates a log.Middleware sending errors logs to Sentry from a config.
//
// If no Sentry config was found, the function returns nil.
func SentryMiddleware(c config.ServiceBase) (log.Middleware, error) {
	if c.Sentry.DSN == "" {
		return nil, nil
	}

	s, err := raven.New(c.Sentry.DSN)
	if err != nil {
		return nil, err
	}

	return sentry.New(s), nil
}
