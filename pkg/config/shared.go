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

package config

import (
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// Base represents base component configuration.
type Base struct {
	Config []string `name:"config" shorthand:"c" description:"Location of the config files"`
	Log    Log      `name:"log"`
}

// Log represents configuration for the logger.
type Log struct {
	Level log.Level `name:"level" description:"The minimum level log messages must have to be shown"`
}

// Sentry represents configuration for error tracking using Sentry.
type Sentry struct {
	DSN string `name:"dsn" description:"Sentry Data Source Name"`
}

// TLS represents TLS configuration.
type TLS struct {
	Certificate string `name:"certificate" description:"Location of TLS certificate"`
	Key         string `name:"key" description:"Location of TLS private key"`
}

// Cluster represents clustering configuration.
type Cluster struct {
	Join              []string `name:"join" description:"Addresses of cluster peers to join"`
	Name              string   `name:"name" description:"Name of the current cluster peer (default: $HOSTNAME)"`
	Address           string   `name:"address" description:"Address to use for cluster communication"`
	IdentityServer    string   `name:"identity-server" description:"Address for the Identity Server"`
	GatewayServer     string   `name:"gateway-server" description:"Address for the Gateway Server"`
	NetworkServer     string   `name:"network-server" description:"Address for the Network Server"`
	ApplicationServer string   `name:"application-server" description:"Address for the Application Server"`
	JoinServer        string   `name:"join-server" description:"Address for the Join Server"`
	TLS               bool     `name:"tls" description:"Do cluster gRPC over TLS"`
	Keys              []string `name:"keys" description:"Keys used to communicate between components of the cluster. The first one will be used by the cluster to identify itself"`
}

// GRPC represents gRPC listener configuration.
type GRPC struct {
	Listen    string `name:"listen" description:"Address for the TCP gRPC server to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the TLS gRPC server to listen on"`
}

// Cookie represents cookie configuration.
// These 128, 192 or 256 bit keys are used to verify and encrypt cookies set by the web server.
// Make sure that all instances of a cluster use the same keys.
// Changing these keys will result in all sessions being invalidated.
type Cookie struct {
	HashKey  []byte `name:"hash-key" description:"Key for cookie contents verification (32 or 64 bytes)"`
	BlockKey []byte `name:"block-key" description:"Key for cookie contents encryption (16, 24 or 32 bytes)"`
}

// HTTP represents the HTTP and HTTPS server configuration.
type HTTP struct {
	Listen    string `name:"listen" description:"Address for the HTTP server to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the HTTPS server to listen on"`
	Cookie    Cookie `name:"cookie"`
	PProf     bool   `name:"pprof" description:"Expose pprof over HTTP"`
}

// Redis represents Redis configuration.
type Redis struct {
	Address   string   `name:"address" description:"Address of the Redis server"`
	Database  int      `name:"database" description:"Redis database to use"`
	Namespace []string `name:"namespace" description:"Namespace for Redis keys"`
}

// IsZero returns whether the Redis configuration is empty.
func (r Redis) IsZero() bool { return r.Address == "" && r.Database == 0 && len(r.Namespace) == 0 }

// Events represents configuration for the events system.
type Events struct {
	Backend string `name:"backend" description:"Backend to use for events (internal, redis)"`
	Redis   Redis  `name:"redis"`
}

// ServiceBase represents base service configuration.
type ServiceBase struct {
	Base           `name:",squash"`
	Cluster        Cluster        `name:"cluster"`
	Redis          Redis          `name:"redis"`
	Events         Events         `name:"events"`
	GRPC           GRPC           `name:"grpc"`
	HTTP           HTTP           `name:"http"`
	TLS            TLS            `name:"tls"`
	Sentry         Sentry         `name:"sentry"`
	FrequencyPlans FrequencyPlans `name:"frequency-plans"`
	Rights         rights.Config  `name:"rights"`
}

// FrequencyPlans represents frequency plans fetching configuration.
type FrequencyPlans struct {
	StoreDirectory string `name:"directory" description:"Retrieve the frequency plans from the filesystem"`
	StoreURL       string `name:"url" description:"Retrieve the frequency plans from a web server"`
}

type unconfiguredFetcher struct{}

func (u unconfiguredFetcher) File(path ...string) ([]byte, error) {
	return nil, errors.New("Fetcher not configured")
}

// Store returns the abstraction to retrieve frequency plans from the given configuration.
func (f FrequencyPlans) Store() *frequencyplans.Store {
	switch {
	case f.StoreDirectory != "":
		return frequencyplans.NewStore(fetch.FromFilesystem(f.StoreDirectory))
	case f.StoreURL != "":
		return frequencyplans.NewStore(fetch.FromHTTP(f.StoreURL, true))
	}

	return frequencyplans.NewStore(unconfiguredFetcher{})
}
