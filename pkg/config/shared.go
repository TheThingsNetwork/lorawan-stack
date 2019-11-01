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
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	ttnblob "go.thethings.network/lorawan-stack/pkg/blob"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/log"
	"gocloud.dev/blob"
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

// Cache represents configuration for a caching system.
type Cache struct {
	Service string `name:"service" description:"Service used for caching (redis)"`
	Redis   Redis  `name:"redis"`
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
	Provider string            `name:"provider" description:"Provider (static)"`
	Static   map[string][]byte `name:"static"`
}

// KeyVault returns an initialized crypto.KeyVault based on the configuration.
func (v KeyVault) KeyVault() (crypto.KeyVault, error) {
	switch v.Provider {
	case "static":
		kv := cryptoutil.NewMemKeyVault(v.Static)
		kv.Separator = ":"
		kv.ReplaceOldNew = []string{":", "_"}
		return kv, nil
	default:
		return cryptoutil.EmptyKeyVault, nil
	}
}

var (
	errUnknownBlobProvider = errors.DefineInvalidArgument("unknown_blob_provider", "unknown blob store provider `{provider}`")
	errMissingBlobConfig   = errors.DefineInvalidArgument("missing_blob_config", "missing blob store configuration")
)

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

type blobConfigAWSCredentials BlobConfigAWS

func (c blobConfigAWSCredentials) Retrieve() (credentials.Value, error) {
	if c.AccessKeyID == "" || c.SecretAccessKey == "" {
		return credentials.Value{}, errMissingBlobConfig
	}
	return credentials.Value{
		ProviderName:    "TTNConfigProvider",
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
	}, nil
}

func (c blobConfigAWSCredentials) IsExpired() bool { return false }

// BlobConfigGCP is the blob store configuration for the GCP provider.
type BlobConfigGCP struct {
	CredentialsFile string `name:"credentials-file" description:"Path to the GCP credentials JSON file"`
	Credentials     string `name:"credentials" description:"JSON data of the GCP credentials, if not using JSON file"`
}

// BlobConfig is the blob store configuration.
type BlobConfig struct {
	Provider string          `name:"provider" description:"Blob store provider (local, aws, gcp)"`
	Local    BlobConfigLocal `name:"local"`
	AWS      BlobConfigAWS   `name:"aws"`
	GCP      BlobConfigGCP   `name:"gcp"`
}

// Bucket returns the requested blob bucket using the config.
func (c BlobConfig) Bucket(ctx context.Context, bucket string) (*blob.Bucket, error) {
	switch c.Provider {
	case "local":
		return ttnblob.Local(ctx, bucket, c.Local.Directory)
	case "aws":
		return ttnblob.AWS(ctx, bucket, &aws.Config{
			Endpoint:    &c.AWS.Endpoint,
			Region:      &c.AWS.Region,
			Credentials: credentials.NewCredentials(blobConfigAWSCredentials(c.AWS)),
		})
	case "gcp":
		var jsonCreds []byte
		if c.GCP.Credentials != "" {
			jsonCreds = []byte(c.GCP.Credentials)
		} else if c.GCP.CredentialsFile != "" {
			var err error
			jsonCreds, err = ioutil.ReadFile(c.GCP.CredentialsFile)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errMissingBlobConfig
		}
		return ttnblob.GCP(ctx, bucket, jsonCreds)
	default:
		return nil, errUnknownBlobProvider.WithAttributes("provider", c.Provider)
	}
}

type BlobPathConfig struct {
	Bucket string `name:"bucket" description:"Bucket to use"`
	Path   string `name:"path" description:"Path to use"`
}

func (c BlobPathConfig) IsZero() bool {
	return c == BlobPathConfig{}
}

// FrequencyPlansConfig contains the source of the frequency plans.
type FrequencyPlansConfig struct {
	ConfigSource string            `name:"config-source" description:"Source of the frequency plans (static, directory, url, blob)"`
	Static       map[string][]byte `name:"-"`
	Directory    string            `name:"directory"`
	URL          string            `name:"url"`
	Blob         BlobPathConfig    `name:"blob"`
}

// Fetcher returns a fetch.Interface based on the configuration.
// If no configuration source is set, this method returns nil, nil.
func (c FrequencyPlansConfig) Fetcher(ctx context.Context, blobConf BlobConfig) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	if c.ConfigSource == "" {
		switch {
		case c.Static != nil:
			c.ConfigSource = "static"
		case c.Directory != "":
			c.ConfigSource = "directory"
		case c.URL != "":
			c.ConfigSource = "url"
		case !c.Blob.IsZero():
			c.ConfigSource = "blob"
		}
	}
	switch c.ConfigSource {
	case "static":
		return fetch.NewMemFetcher(c.Static), nil
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		return fetch.FromHTTP(c.URL, true)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

// DeviceRepositoryConfig defines the source of the device repository.
type DeviceRepositoryConfig struct {
	ConfigSource string            `name:"config-source" description:"Source of the device repository (static, directory, url, blob)"`
	Static       map[string][]byte `name:"-"`
	Directory    string            `name:"directory"`
	URL          string            `name:"url"`
	Blob         BlobPathConfig    `name:"blob"`
}

// Fetcher returns a fetch.Interface based on the configuration.
// If no configuration source is set, this method returns nil, nil.
func (c DeviceRepositoryConfig) Fetcher(ctx context.Context, blobConf BlobConfig) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	if c.ConfigSource == "" {
		switch {
		case c.Static != nil:
			c.ConfigSource = "static"
		case c.Directory != "":
			c.ConfigSource = "directory"
		case c.URL != "":
			c.ConfigSource = "url"
		case !c.Blob.IsZero():
			c.ConfigSource = "blob"
		}
	}
	switch c.ConfigSource {
	case "static":
		return fetch.NewMemFetcher(c.Static), nil
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		return fetch.FromHTTP(c.URL, true)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

// InteropClient represents the client-side interoperability through LoRaWAN Backend Interfaces configuration.
type InteropClient struct {
	ConfigSource string         `name:"config-source" description:"Source of the interoperability client configuration (directory, url, blob)"`
	Directory    string         `name:"directory"`
	URL          string         `name:"url"`
	Blob         BlobPathConfig `name:"blob"`

	GetFallbackTLSConfig func(ctx context.Context) (*tls.Config, error) `name:"-"`
	BlobConfig           BlobConfig                                     `name:"-"`
}

// IsZero returns whether conf is empty.
func (c InteropClient) IsZero() bool {
	return c.ConfigSource == "" &&
		c.Directory == "" &&
		c.URL == "" &&
		c.Blob.IsZero() &&
		c.GetFallbackTLSConfig == nil &&
		c.BlobConfig == BlobConfig{}
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c InteropClient) Fetcher(ctx context.Context, blobConf BlobConfig) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	if c.ConfigSource == "" {
		switch {
		case c.Directory != "":
			c.ConfigSource = "directory"
		case c.URL != "":
			c.ConfigSource = "url"
		case !c.Blob.IsZero():
			c.ConfigSource = "blob"
		}
	}
	switch c.ConfigSource {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		return fetch.FromHTTP(c.URL, true)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

// ServiceBase represents base service configuration.
type ServiceBase struct {
	Base             `name:",squash"`
	Cluster          Cluster                `name:"cluster"`
	Cache            Cache                  `name:"cache"`
	Redis            Redis                  `name:"redis"`
	Events           Events                 `name:"events"`
	GRPC             GRPC                   `name:"grpc"`
	HTTP             HTTP                   `name:"http"`
	Interop          InteropServer          `name:"interop"`
	TLS              TLS                    `name:"tls"`
	Sentry           Sentry                 `name:"sentry"`
	Blob             BlobConfig             `name:"blob"`
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

// MQTTConfigProviderFunc is a functional MQTTConfigProvider.
type MQTTConfigProviderFunc func(context.Context) (*MQTT, error)

// GetMQTTConfig implements MQTTConfigProvider.
func (f MQTTConfigProviderFunc) GetMQTTConfig(ctx context.Context) (*MQTT, error) {
	return f(ctx)
}
