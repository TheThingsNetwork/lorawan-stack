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

	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
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
	Certificate: "cert.pem",
	Key:         "key.pem",
	ACME: config.ACME{
		Endpoint: "https://acme-v01.api.letsencrypt.org/directory",
	},
}

// DefaultClusterConfig is the default cluster configuration.
var DefaultClusterConfig = config.Cluster{}

// DefaultHTTPConfig is the default HTTP config.
var DefaultHTTPConfig = config.HTTP{
	Listen:    ":1885",
	ListenTLS: ":8885",
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
	ListenTLS: ":8886",
}

// DefaultGRPCConfig is the default config for GRPC.
var DefaultGRPCConfig = config.GRPC{
	Listen:    ":1884",
	ListenTLS: ":8884",
}

// DefaultRedisConfig is the default config for Redis.
var DefaultRedisConfig = config.Redis{
	Address:   "localhost:6379",
	Database:  0,
	Namespace: []string{"ttn", "v3"},
}

// DefaultEventsConfig is the default config for Events.
var DefaultEventsConfig = config.Events{
	Backend: "internal",
}

// DefaultBlobConfig is the default config for the blob store.
var DefaultBlobConfig = config.BlobConfig{
	Provider: "local",
	Local: config.BlobConfigLocal{
		Directory: "./public/blob",
	},
}

// DefaultFrequencyPlansConfig is the default config to retrieve frequency plans.
var DefaultFrequencyPlansConfig = config.FrequencyPlansConfig{
	URL: "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master",
}

// DefaultDeviceRepositoryConfig is the default config to retrieve device blueprints.
var DefaultDeviceRepositoryConfig = config.DeviceRepositoryConfig{}

// DefaultRightsConfig is the default config to fetch rights from the Identity Server.
var DefaultRightsConfig = config.Rights{
	TTL: 2 * time.Minute,
}

// DefaultKeyVaultConfig is the default config for key vaults.
var DefaultKeyVaultConfig = config.KeyVault{
	Provider: "static",
}

// DefaultServiceBase is the default base config for a service.
var DefaultServiceBase = config.ServiceBase{
	Base:             DefaultBaseConfig,
	Cluster:          DefaultClusterConfig,
	Redis:            DefaultRedisConfig,
	Events:           DefaultEventsConfig,
	GRPC:             DefaultGRPCConfig,
	HTTP:             DefaultHTTPConfig,
	Interop:          DefaultInteropServerConfig,
	TLS:              DefaultTLSConfig,
	Blob:             DefaultBlobConfig,
	FrequencyPlans:   DefaultFrequencyPlansConfig,
	DeviceRepository: DefaultDeviceRepositoryConfig,
	Rights:           DefaultRightsConfig,
	KeyVault:         DefaultKeyVaultConfig,
}

// DefaultPublicHost is the default public host where The Things Stack is served.
var DefaultPublicHost = "localhost"

// DefaultPublicURL is the default public URL where The Things Stack is served.
var DefaultPublicURL = "http://" + DefaultPublicHost + ":1885"

// DefaultAssetsBaseURL is the default public URL where the assets are served.
var DefaultAssetsBaseURL = DefaultHTTPConfig.Static.Mount

// DefaultOAuthPublicURL is the default public URL where OAuth is served.
var DefaultOAuthPublicURL = DefaultPublicURL + "/oauth"

// DefaultConsolePublicURL is the default public URL where the Console is served.
var DefaultConsolePublicURL = DefaultPublicURL + "/console"
