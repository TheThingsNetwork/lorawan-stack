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
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	ttnblob "go.thethings.network/lorawan-stack/v3/pkg/blob"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"gocloud.dev/blob"
)

// Base represents base component configuration.
type Base struct {
	Config []string `name:"config" shorthand:"c" description:"Location of the config files"`
	Log    Log      `name:"log"`
}

// Log represents configuration for the logger.
type Log struct {
	Format string    `name:"format" description:"Log format to write (console, json)"`
	Level  log.Level `name:"level" description:"The minimum level log messages must have to be shown"`
}

// Sentry represents configuration for error tracking using Sentry.
type Sentry struct {
	DSN         string `name:"dsn" description:"Sentry Data Source Name"`
	Environment string `name:"environment" description:"Environment to report to Sentry"`
}

// GRPC represents gRPC listener configuration.
type GRPC struct {
	AllowInsecureForCredentials bool `name:"allow-insecure-for-credentials" description:"Allow transmission of credentials over insecure transport"`

	Listen    string `name:"listen" description:"Address for the TCP gRPC server to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the TLS gRPC server to listen on"`

	TrustedProxies []string `name:"trusted-proxies" description:"CIDRs of trusted reverse proxies"`

	LogIgnoreMethods []string `name:"log-ignore-methods" description:"List of paths for which successful requests will not be logged"`
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
	TrustedProxies  []string         `name:"trusted-proxies" description:"CIDRs of trusted reverse proxies"`
	RedirectToHost  string           `name:"redirect-to-host" description:"Redirect all requests to one host"`
	RedirectToHTTPS bool             `name:"redirect-to-tls" description:"Redirect HTTP requests to HTTPS"`
	LogIgnorePaths  []string         `name:"log-ignore-paths" description:"List of paths for which successful requests will not be logged"`
	Static          HTTPStaticConfig `name:"static"`
	Cookie          Cookie           `name:"cookie"`
	PProf           PProf            `name:"pprof"`
	Metrics         Metrics          `name:"metrics"`
	Health          Health           `name:"health"`
}

// CloudEvents represents configuration for the cloud events backend.
type CloudEvents struct {
	PublishURL   string `name:"publish-url" description:"URL for the topic to send events"`
	SubscribeURL string `name:"subscribe-url" description:"URL for the subscription to receiving events"`
}

// Cache represents configuration for a caching system.
type Cache struct {
	Service string       `name:"service" description:"Service used for caching (redis)"`
	Redis   redis.Config `name:"redis"`
}

// RedisEvents represents configuration for the Redis events backend.
type RedisEvents struct {
	redis.Config `name:",squash"`
	Store        struct {
		Enable             bool          `name:"enable" description:"Enable events store"`
		TTL                time.Duration `name:"ttl" description:"How long event payloads are retained"`
		EntityCount        int           `name:"entity-count" description:"How many events are indexed for a entity ID"`
		EntityTTL          time.Duration `name:"entity-ttl" description:"How long events are indexed for a entity ID"`
		CorrelationIDCount int           `name:"correlation-id-count" description:"How many events are indexed for a correlation ID"`
	} `name:"store"`
	Workers int `name:"workers"`
}

// Events represents configuration for the events system.
type Events struct {
	Backend string      `name:"backend" description:"Backend to use for events (internal, redis, cloud)"`
	Redis   RedisEvents `name:"redis"`
	Cloud   CloudEvents `name:"cloud"`
}

// Rights represents the configuration to apply when fetching entity rights.
type Rights struct {
	// TTL is the duration that entries will remain in the cache before being
	// garbage collected.
	TTL time.Duration `name:"ttl" description:"Validity of Identity Server responses"`
}

// KeyVaultCache represents the configuration for key vault caching.
type KeyVaultCache struct {
	Size int           `name:"size" description:"Cache size. Caching is disabled if size is 0"`
	TTL  time.Duration `name:"ttl" description:"Cache elements time to live. No expiration mechanism is used if TTL is 0"`
}

// KeyVault represents configuration for key vaults.
type KeyVault struct {
	Provider string            `name:"provider" description:"Provider (static)"`
	Cache    KeyVaultCache     `name:"cache"`
	Static   map[string][]byte `name:"static"`

	HTTPClient *http.Client `name:"-"`
}

// KeyVault returns an initialized crypto.KeyVault based on the configuration.
func (v KeyVault) KeyVault() (crypto.KeyVault, error) {
	vault := cryptoutil.EmptyKeyVault
	switch v.Provider {
	case "static":
		kv := cryptoutil.NewMemKeyVault(v.Static)
		kv.Separator = ":"
		kv.ReplaceOldNew = []string{":", "_"}
		vault = kv
	}
	if v.Cache.Size > 0 {
		vault = cryptoutil.NewCacheKeyVault(vault, v.Cache.TTL, v.Cache.Size)
	}
	return vault, nil
}

var (
	errUnknownBlobProvider = errors.DefineInvalidArgument("unknown_blob_provider", "unknown blob store provider `{provider}`")
	errMissingBlobConfig   = errors.DefineInvalidArgument("missing_blob_config", "missing blob store configuration")
)

// BlobConfigLocal is the blob store configuration for the local filesystem provider.
type BlobConfigLocal struct {
	Directory string `name:"directory" description:"OS filesystem directory, which contains buckets"`
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

// BlobConfig is the blob store configuration.
type BlobConfig struct {
	Provider string          `name:"provider" description:"Blob store provider (local, aws, gcp)"`
	Local    BlobConfigLocal `name:"local"`
	AWS      BlobConfigAWS   `name:"aws"`
	GCP      BlobConfigGCP   `name:"gcp"`

	HTTPClient *http.Client `name:"-"`
}

// IsZero returns whether conf is empty.
func (c BlobConfig) IsZero() bool {
	return c.Provider == "" &&
		c.Local == BlobConfigLocal{} &&
		c.AWS == BlobConfigAWS{} &&
		c.GCP == BlobConfigGCP{}
}

// Bucket returns the requested blob bucket using the config.
func (c BlobConfig) Bucket(ctx context.Context, bucket string) (*blob.Bucket, error) {
	switch c.Provider {
	case "local":
		return ttnblob.Local(ctx, bucket, c.Local.Directory)
	case "aws":
		conf := aws.NewConfig().WithHTTPClient(c.HTTPClient)
		if c.AWS.Endpoint != "" {
			conf = conf.WithEndpoint(c.AWS.Endpoint)
		}
		if c.AWS.Region != "" {
			conf = conf.WithRegion(c.AWS.Region)
		}
		if c.AWS.AccessKeyID != "" && c.AWS.SecretAccessKey != "" {
			conf = conf.WithCredentials(credentials.NewStaticCredentials(
				c.AWS.AccessKeyID, c.AWS.SecretAccessKey, c.AWS.SessionToken,
			))
		}
		return ttnblob.AWS(ctx, bucket, conf)
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
			return nil, errMissingBlobConfig.New()
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
	Directory    string            `name:"directory" description:"OS filesystem directory, which contains frequency plans"`
	URL          string            `name:"url" description:"URL, which contains frequency plans"`
	Blob         BlobPathConfig    `name:"blob"`

	HTTPClient *http.Client `name:"-"`
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
			if stat, err := os.Stat(c.Directory); err == nil && stat.IsDir() {
				c.ConfigSource = "directory"
				break
			}
			fallthrough
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
		return fetch.FromHTTP(c.HTTPClient, c.URL, true)
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
	Directory    string         `name:"directory" description:"OS filesystem directory, which contains interoperability client configuration"`
	URL          string         `name:"url" description:"URL, which contains interoperability client configuration"`
	Blob         BlobPathConfig `name:"blob"`

	GetFallbackTLSConfig func(ctx context.Context) (*tls.Config, error) `name:"-"`
	BlobConfig           BlobConfig                                     `name:"-"`

	HTTPClient *http.Client `name:"-"`
}

// IsZero returns whether conf is empty.
func (c InteropClient) IsZero() bool {
	return c.ConfigSource == "" &&
		c.Directory == "" &&
		c.URL == "" &&
		c.Blob.IsZero() &&
		c.GetFallbackTLSConfig == nil &&
		c.BlobConfig.IsZero()
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c InteropClient) Fetcher(ctx context.Context) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	if c.ConfigSource == "" {
		switch {
		case c.Directory != "":
			if stat, err := os.Stat(c.Directory); err == nil && stat.IsDir() {
				c.ConfigSource = "directory"
				break
			}
			fallthrough
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
		return fetch.FromHTTP(c.HTTPClient, c.URL, true)
	case "blob":
		b, err := c.BlobConfig.Bucket(ctx, c.Blob.Bucket)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

type SenderClientCA struct {
	Source    string            `name:"source" description:"Source of the sender client CA configuration (static, directory, url, blob)"`
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"OS filesystem directory, which contains sender client CA configuration"`
	URL       string            `name:"url" description:"URL, which contains sender client CA configuration"`
	Blob      BlobPathConfig    `name:"blob"`

	BlobConfig BlobConfig `name:"-"`

	HTTPClient *http.Client `name:"-"`
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c SenderClientCA) Fetcher(ctx context.Context) (fetch.Interface, error) {
	switch c.Source {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		return fetch.FromHTTP(c.HTTPClient, c.URL, true)
	case "blob":
		b, err := c.BlobConfig.Bucket(ctx, c.Blob.Bucket)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

// InteropServer represents the server-side interoperability through LoRaWAN Backend Interfaces configuration.
type InteropServer struct {
	ListenTLS                string            `name:"listen-tls" description:"Address for the interop server to listen on"`
	SenderClientCA           SenderClientCA    `name:"sender-client-ca"`
	SenderClientCADeprecated map[string]string `name:"sender-client-cas" description:"Path to PEM encoded file with client CAs of sender IDs to trust; deprecated - use sender-client-ca instead"`
}

// ServiceBase represents base service configuration.
type ServiceBase struct {
	Base             `name:",squash"`
	Cluster          cluster.Config       `name:"cluster"`
	Cache            Cache                `name:"cache"`
	Redis            redis.Config         `name:"redis"`
	Events           Events               `name:"events"`
	GRPC             GRPC                 `name:"grpc"`
	HTTP             HTTP                 `name:"http"`
	Interop          InteropServer        `name:"interop"`
	TLS              tlsconfig.Config     `name:"tls"`
	Sentry           Sentry               `name:"sentry"`
	Blob             BlobConfig           `name:"blob"`
	FrequencyPlans   FrequencyPlansConfig `name:"frequency-plans" description:"Source of the frequency plans"`
	Rights           Rights               `name:"rights"`
	KeyVault         KeyVault             `name:"key-vault"`
	RateLimiting     RateLimiting         `name:"rate-limiting" description:"Rate limiting configuration"`
	SkipVersionCheck bool                 `name:"skip-version-check" yaml:"skip-version-check" description:"Skip version checks"`
}

// FrequencyPlansFetcher returns a fetch.Interface based on the frequency plans configuration.
// If no configuration source is set, this method returns nil, nil.
func (c ServiceBase) FrequencyPlansFetcher(ctx context.Context) (fetch.Interface, error) {
	return c.FrequencyPlans.Fetcher(ctx, c.Blob)
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

// RateLimitingProfile represents configuration for a rate limiting class.
type RateLimitingProfile struct {
	Name         string   `name:"name" description:"Rate limiting class name"`
	MaxPerMin    uint     `name:"max-per-min" yaml:"max-per-min" description:"Maximum allowed rate (per minute)"`
	MaxBurst     uint     `name:"max-burst" yaml:"max-burst" description:"Maximum rate allowed for short bursts"`
	Associations []string `name:"associations" description:"List of classes to apply this profile on"`
}

// RateLimitingMemory represents configuration for the in-memory rate limiting store.
type RateLimitingMemory struct {
	MaxSize uint `name:"max-size" description:"Maximum store size for the rate limiter"`
}

// RateLimiting represents configuration for rate limiting.
type RateLimiting struct {
	ConfigSource string         `name:"config-source" description:"Source of rate-limiting.yml (directory, url, blob)"`
	Directory    string         `name:"directory" description:"OS filesystem directory, which contains rate limiting configuration"`
	URL          string         `name:"url" description:"URL, which contains rate limiting configuration"`
	Blob         BlobPathConfig `name:"blob"`

	HTTPClient *http.Client `name:"-"`

	Memory   RateLimitingMemory    `name:"memory" description:"In-memory rate limiting store configuration"`
	Profiles []RateLimitingProfile `name:"profiles" description:"Rate limiting profiles"`
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c RateLimiting) Fetcher(ctx context.Context, blobConf BlobConfig) (fetch.Interface, error) {
	switch c.ConfigSource {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		return fetch.FromHTTP(c.HTTPClient, c.URL, true)
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
