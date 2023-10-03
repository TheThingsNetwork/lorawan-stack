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

package shared

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbroker"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"golang.org/x/crypto/acme"
)

// DefaultBaseConfig is the default base component configuration.
var DefaultBaseConfig = config.Base{
	Log: DefaultLogConfig,
}

// DefaultLogConfig is the default log configuration.
var DefaultLogConfig = config.Log{
	Format: "console",
	Level:  log.InfoLevel,
}

// DefaultTLSConfig is the default TLS config.
var DefaultTLSConfig = tlsconfig.Config{
	ServerAuth: tlsconfig.ServerAuth{
		Certificate: "cert.pem",
		Key:         "key.pem",
		ACME: tlsconfig.ACME{
			Endpoint: acme.LetsEncryptURL,
		},
	},
}

// DefaultClusterConfig is the default cluster configuration.
var DefaultClusterConfig = cluster.Config{}

// DefaultHTTPConfig is the default HTTP config.
var DefaultHTTPConfig = config.HTTP{
	Listen:         ":1885",
	ListenTLS:      ":8885",
	TrustedProxies: []string{"127.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "172.16.0.0/12", "192.168.0.0/16"},
	Static: config.HTTPStaticConfig{
		Mount:      "/assets",
		SearchPath: []string{"public", "/srv/ttn-lorawan/public"},
	},
	PProf: config.PProf{
		Enable: true,
	},
	Metrics: config.Metrics{
		Enable: true,
	},
	Health: config.Health{
		Enable: true,
	},
}

// DefaultInteropServerConfig is the default interop server config.
var DefaultInteropServerConfig = config.InteropServer{
	Listen:    ":1886",
	ListenTLS: ":8886",
	PacketBroker: config.PacketBrokerInteropAuth{
		Enabled:     false,
		TokenIssuer: packetbroker.DefaultTokenIssuer,
	},
}

// DefaultGRPCConfig is the default config for GRPC.
var DefaultGRPCConfig = config.GRPC{
	Listen:         ":1884",
	ListenTLS:      ":8884",
	TrustedProxies: []string{"127.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "172.16.0.0/12", "192.168.0.0/16"},
}

// DefaultRedisConfig is the default config for Redis.
var DefaultRedisConfig = redis.Config{
	Address:       "localhost:6379",
	Database:      0,
	RootNamespace: []string{"ttn", "v3"},
}

// DefaultCacheConfig is the default cache configuration.
var DefaultCacheConfig = config.Cache{
	Redis: DefaultRedisConfig,
}

// DefaultEventsConfig is the default config for Events.
var DefaultEventsConfig = func() config.Events {
	c := config.Events{
		Backend: "internal",
	}
	c.Redis.Config = DefaultRedisConfig
	c.Redis.Store.TTL = 10 * time.Minute
	c.Redis.Store.EntityTTL = 24 * time.Hour
	c.Redis.Store.EntityCount = 100
	c.Redis.Store.CorrelationIDCount = 100
	c.Redis.Store.StreamPartitionSize = 64
	c.Redis.Workers = 16
	c.Redis.Publish.QueueSize = 8192
	c.Redis.Publish.MaxWorkers = 1024
	c.Batch.TargetSize = 64
	c.Batch.Delay = 32 * time.Millisecond
	return c
}()

// DefaultBlobConfig is the default config for the blob store.
var DefaultBlobConfig = config.BlobConfig{
	Provider: "local",
	Local: config.BlobConfigLocal{
		Directory: "./public/blob",
	},
}

// DefaultFrequencyPlansConfig is the default config to retrieve frequency plans.
var DefaultFrequencyPlansConfig = config.FrequencyPlansConfig{
	Directory: "/srv/ttn-lorawan/lorawan-frequency-plans",
	URL:       "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master",
}

// DefaultRightsConfig is the default config to fetch rights from the Identity Server.
var DefaultRightsConfig = config.Rights{
	TTL: 2 * time.Minute,
}

// DefaultKeyVaultConfig is the default config for key vaults.
var DefaultKeyVaultConfig = config.KeyVault{
	Provider: "static",
}

// DefaultTracingConfig is the default config for telemetry tracing.
var DefaultTracingConfig = tracing.Config{
	Enable:   false,
	Exporter: "otlp",
	WriterConfig: tracing.WriterConfig{
		Destination: "stderr",
	},
	CollectorConfig: tracing.CollectorConfig{
		EndpointURL: "localhost:4317",
	},
	SampleProbability: 1.,
}

// DefaultTelemetryConfig  is the default config for telemetry.
var DefaultTelemetryConfig = telemetry.Config{
	Enable:       true,
	Target:       "https://telemetry.thethingsstack.io/collect",
	NumConsumers: 1,
	EntityCountTelemetry: telemetry.EntityCountTelemetry{
		Enable:   true,
		Interval: 24 * time.Hour,
	},
}

// DefaultServiceBase is the default base config for a service.
var DefaultServiceBase = config.ServiceBase{
	Base:           DefaultBaseConfig,
	Cluster:        DefaultClusterConfig,
	Cache:          DefaultCacheConfig,
	Redis:          DefaultRedisConfig,
	Events:         DefaultEventsConfig,
	GRPC:           DefaultGRPCConfig,
	HTTP:           DefaultHTTPConfig,
	Interop:        DefaultInteropServerConfig,
	TLS:            DefaultTLSConfig,
	Blob:           DefaultBlobConfig,
	FrequencyPlans: DefaultFrequencyPlansConfig,
	Rights:         DefaultRightsConfig,
	KeyVault:       DefaultKeyVaultConfig,
	Tracing:        DefaultTracingConfig,
	Telemetry:      DefaultTelemetryConfig,
}

// DefaultPublicHost is the default public host where The Things Stack is served.
var DefaultPublicHost = "localhost"

// DefaultPublicURL is the default public URL where The Things Stack is served.
var DefaultPublicURL = "http://" + DefaultPublicHost + ":1885"

// DefaultAssetsBaseURL is the default public URL where the assets are served.
var DefaultAssetsBaseURL = DefaultHTTPConfig.Static.Mount

// DefaultOAuthPublicURL is the default URL where the OAuth API as well as
// OAuth and Account application frontend is served.
var DefaultOAuthPublicURL = DefaultPublicURL + "/oauth"

// DefaultConsolePublicURL is the default public URL where the Console is served.
var DefaultConsolePublicURL = DefaultPublicURL + "/console"
