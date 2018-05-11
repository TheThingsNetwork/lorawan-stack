// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
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
}

// DefaultClusterConfig is the default cluster configuration.
var DefaultClusterConfig = config.Cluster{}

// DefaultHTTPConfig is the default HTTP config.
var DefaultHTTPConfig = config.HTTP{
	Listen:    ":1885",
	ListenTLS: ":8885",
	PProf:     true,
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

// DefaultServiceBase is the default base config for a service.
var DefaultServiceBase = config.ServiceBase{
	Base:           DefaultBaseConfig,
	Cluster:        DefaultClusterConfig,
	Redis:          DefaultRedisConfig,
	Events:         DefaultEventsConfig,
	GRPC:           DefaultGRPCConfig,
	HTTP:           DefaultHTTPConfig,
	TLS:            DefaultTLSConfig,
	Identity:       DefaultIdentityConfig,
	FrequencyPlans: DefaultFrequencyPlansConfig,
}

// DefaultFrequencyPlansConfig is the default config to retrieve frequency plans.
var DefaultFrequencyPlansConfig = config.FrequencyPlans{
	StoreURL: "https://raw.githubusercontent.com/TheThingsNetwork/gateway-conf/yaml-master",
}

// DefaultISRightsFetchingConfig is the default config to fetch rights from the Identity Server.
var DefaultISRightsFetchingConfig = rights.Config{
	TTL: 2 * time.Minute,
}
