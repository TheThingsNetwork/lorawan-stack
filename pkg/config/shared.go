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

package config

import (
	"context"
	"crypto/tls"
	"time"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/devicerepository"
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
	CryptoServer      string   `name:"crypto-server" description:"Address for the Crypto Server"`
	TLS               bool     `name:"tls" description:"Do cluster gRPC over TLS"`
	Keys              []string `name:"keys" description:"Keys used to communicate between components of the cluster. The first one will be used by the cluster to identify itself"`
}

// GRPC represents gRPC listener configuration.
type GRPC struct {
	AllowInsecureForCredentials bool `name:"allow-insecure-for-credentials" description:"Allow transmission of credentials over insecure transport"`

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

// PProf represents the pprof endpoint configuration.
type PProf struct {
	Enable   bool   `name:"enable" description:"Enable pprof endpoint on HTTP server"`
	Password string `name:"password" description:"Password to protect pprof endpoint (username is pprof)"`
}

// Metrics represents the metrics endpoint configuration.
type Metrics struct {
	Enable   bool   `name:"enable" description:"Enable metrics endpoint on HTTP server"`
	Password string `name:"password" description:"Password to protect metrics endpoint (username is metrics)"`
}

// Health represents the health checks configuration.
type Health struct {
	Enable   bool   `name:"enable" description:"Enable health check endpoint on HTTP server"`
	Password string `name:"password" description:"Password to protect health endpoint (username is health)"`
}

// HTTPStaticConfig represents the HTTP static file server configuration.
type HTTPStaticConfig struct {
	Mount      string   `name:"mount" description:"Path on the server where static assets will be served"`
	SearchPath []string `name:"search-path" description:"List of paths for finding the directory to serve static assets from"`
}

// HTTP represents the HTTP and HTTPS server configuration.
type HTTP struct {
	Listen          string           `name:"listen" description:"Address for the HTTP server to listen on"`
	ListenTLS       string           `name:"listen-tls" description:"Address for the HTTPS server to listen on"`
	RedirectToHost  string           `name:"redirect-to-host" description:"Redirect all requests to one host"`
	RedirectToHTTPS bool             `name:"redirect-to-tls" description:"Redirect HTTP requests to HTTPS"`
	Static          HTTPStaticConfig `name:"static"`
	Cookie          Cookie           `name:"cookie"`
	PProf           PProf            `name:"pprof"`
	Metrics         Metrics          `name:"metrics"`
	Health          Health           `name:"health"`
}

// InteropServer represents the server-side interoperability through LoRaWAN Backend Interfaces configuration.
type InteropServer struct {
	ListenTLS       string            `name:"listen-tls" description:"Address for the interop server to listen on"`
	SenderClientCAs map[string]string `name:"sender-client-cas" description:"Path to PEM encoded file with client CAs of sender IDs to trust"`
}

// Redis represents Redis configuration.
type Redis struct {
	Address   string   `name:"address" description:"Address of the Redis server"`
	Password  string   `name:"password" description:"Password of the Redis server"`
	Database  int      `name:"database" description:"Redis database to use"`
	Namespace []string `name:"namespace" description:"Namespace for Redis keys"`
}

// IsZero returns whether the Redis configuration is empty.
func (r Redis) IsZero() bool { return r.Address == "" && r.Database == 0 && len(r.Namespace) == 0 }

// CloudEvents represents configuration for the cloud events backend.
type CloudEvents struct {
	PublishURL   string `name:"publish-url" description:"URL for the topic to send events"`
	SubscribeURL string `name:"subscribe-url" description:"URL for the subscription to receiving events"`
}

// Events represents configuration for the events system.
type Events struct {
	Backend string      `name:"backend" description:"Backend to use for events (internal, redis, cloud)"`
	Redis   Redis       `name:"redis"`
	Cloud   CloudEvents `name:"cloud"`
}

// Rights represents the configuration to apply when fetching entity rights.
type Rights struct {
	// TTL is the duration that entries will remain in the cache before being
	// garbage collected.
	TTL time.Duration `name:"ttl" description:"Validity of Identity Server responses"`
}

// KeyVault represents configuration for key vaults.
type KeyVault struct {
	Static map[string][]byte `name:"static" description:"Static labeled key encryption keys"`
}

// KeyVault returns an initialized crypto.KeyVault based on the configuration.
// The order of precedence is Static.
func (v KeyVault) KeyVault() crypto.KeyVault {
	switch {
	case v.Static != nil:
		return cryptoutil.NewMemKeyVault(v.Static)
	default:
		return cryptoutil.NewMemKeyVault(map[string][]byte{})
	}
}

// BlobConfigLocal is the blob store configuration for the local filesystem provider.
type BlobConfigLocal struct {
	Directory string `name:"directory" description:"Local directory that holds the buckets"`
}

// BlobConfigAWS is the blob store configuration for the AWS provider.
type BlobConfigAWS struct {
	Endpoint        string `name:"endpoint" description:"S3 endpoint"`
	Region          string `name:"region" description:"S3 region"`
	AccessKeyID     string `name:"access-key-id" description:"Access key ID"`
	SecretAccessKey string `name:"secret-access-key" description:"Secret access key"`
	SessionToken    string `name:"session-token" description:"Session token"`
}

// BlobConfigGCP is the blob store configuration for the GCP provider.
type BlobConfigGCP struct {
	CredentialsFile string `name:"credentials-file" description:"Path to the GCP credentials JSON file"`
	Credentials     string `name:"credentials" description:"JSON data of the GCP credentials, if not using JSON file"`
}

// Blob store configuration.
type Blob struct {
	Provider string          `name:"provider" description:"Blob store provider (local|aws|gcp)"`
	Local    BlobConfigLocal `name:"local"`
	AWS      BlobConfigAWS   `name:"aws"`
	GCP      BlobConfigGCP   `name:"gcp"`
}

// FrequencyPlansConfig contains the source of the frequency plans.
type FrequencyPlansConfig struct {
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"Retrieve the frequency plans from the filesystem"`
	URL       string            `name:"url" description:"Retrieve the frequency plans from a web server"`
}

// Store returns a frequencyplan.Store fwith a fetcher based on the configuration.
// The order of precedence is Static, Directory and URL.
// If neither Static, Directory nor a URL is set, this method returns nil, nil.
func (c FrequencyPlansConfig) Store() (*frequencyplans.Store, error) {
	var fetcher fetch.Interface
	switch {
	case c.Static != nil:
		fetcher = fetch.NewMemFetcher(c.Static)
	case c.Directory != "":
		fetcher = fetch.FromFilesystem(c.Directory)
	case c.URL != "":
		var err error
		fetcher, err = fetch.FromHTTP(c.URL, true)
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	return frequencyplans.NewStore(fetcher), nil
}

// DeviceRepositoryConfig defines the source of the device repository.
type DeviceRepositoryConfig struct {
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"Retrieve the device repository from the filesystem"`
	URL       string            `name:"url" description:"Retrieve the device repository from a web server"`
}

// Client instantiates a new devicerepository.Client with a fetcher based on the configuration.
// The order of precedence is Static, Directory and URL.
// If neither Static, Directory nor a URL is set, this method returns nil, nil.
func (c DeviceRepositoryConfig) Client() (*devicerepository.Client, error) {
	var fetcher fetch.Interface
	switch {
	case c.Static != nil:
		fetcher = fetch.NewMemFetcher(c.Static)
	case c.Directory != "":
		fetcher = fetch.FromFilesystem(c.Directory)
	case c.URL != "":
		var err error
		fetcher, err = fetch.FromHTTP(c.URL, true)
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	return &devicerepository.Client{
		Fetcher: fetcher,
	}, nil
}

// InteropClient represents the client-side interoperability through LoRaWAN Backend Interfaces configuration.
type InteropClient struct {
	Directory   string      `name:"directory" description:"Retrieve the interoperability client configuration from the filesystem"`
	URL         string      `name:"url" description:"Retrieve the interoperability client configuration from a web server"`
	FallbackTLS *tls.Config `name:"-"`
}

// IsZero returns whether conf is empty.
func (c InteropClient) IsZero() bool {
	return c == (InteropClient{})
}

// Fetcher returns fetch.Interface defined by conf.
func (c InteropClient) Fetcher() (fetch.Interface, error) {
	switch {
	case c.Directory != "":
		return fetch.FromFilesystem(c.Directory), nil
	case c.URL != "":
		return fetch.FromHTTP(c.URL, true)
	default:
		return nil, nil
	}
}

// ServiceBase represents base service configuration.
type ServiceBase struct {
	Base             `name:",squash"`
	Cluster          Cluster                `name:"cluster"`
	Redis            Redis                  `name:"redis"`
	Events           Events                 `name:"events"`
	GRPC             GRPC                   `name:"grpc"`
	HTTP             HTTP                   `name:"http"`
	Interop          InteropServer          `name:"interop"`
	TLS              TLS                    `name:"tls"`
	Sentry           Sentry                 `name:"sentry"`
	Blob             Blob                   `name:"blob"`
	FrequencyPlans   FrequencyPlansConfig   `name:"frequency-plans" description:"Source of the frequency plans"`
	DeviceRepository DeviceRepositoryConfig `name:"device-repository" description:"Source of the device repository"`
	Rights           Rights                 `name:"rights"`
	KeyVault         KeyVault               `name:"key-vault"`
}

// MQTT contains the listen and public addresses of an MQTT frontend.
type MQTT struct {
	Listen           string `name:"listen" description:"Address for the MQTT frontend to listen on"`
	ListenTLS        string `name:"listen-tls" description:"Address for the MQTTS frontend to listen on"`
	PublicAddress    string `name:"public-address" description:"Public address of the MQTT frontend"`
	PublicTLSAddress string `name:"public-tls-address" description:"Public address of the MQTTs frontend"`
}

// MQTTConfigProvider provides contextual access to MQTT configuration.
type MQTTConfigProvider interface {
	GetMQTTConfig(context.Context) (*MQTT, error)
}

// MQTTConfigProviderFunc is an functional MQTTConfigProvider.
type MQTTConfigProviderFunc func(context.Context) (*MQTT, error)

// GetMQTTConfig implements MQTTConfigProvider.
func (f MQTTConfigProviderFunc) GetMQTTConfig(ctx context.Context) (*MQTT, error) {
	return f(ctx)
}
