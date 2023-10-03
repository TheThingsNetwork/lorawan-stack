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
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"gocloud.dev/blob"
)

// Base represents base component configuration.
type Base struct {
	Config       []string            `name:"config" shorthand:"c" description:"Location of the config files"`
	Log          Log                 `name:"log"`
	Experimental experimental.Config `name:"experimental"`
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
	AllowInsecureForCredentials bool `name:"allow-insecure-for-credentials" description:"Allow transmission of credentials over insecure transport"` //nolint:lll

	Listen    string `name:"listen" description:"Address for the TCP gRPC server to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the TLS gRPC server to listen on"`

	TrustedProxies []string `name:"trusted-proxies" description:"CIDRs of trusted reverse proxies"`

	LogIgnoreMethods []string `name:"log-ignore-methods" description:"List of paths for which successful requests will not be logged"` //nolint:lll
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
	SearchPath []string `name:"search-path" description:"List of paths for finding the directory to serve static assets from"` //nolint:lll
}

// HTTP represents the HTTP and HTTPS server configuration.
type HTTP struct {
	Listen          string           `name:"listen" description:"Address for the HTTP server to listen on"`
	ListenTLS       string           `name:"listen-tls" description:"Address for the HTTPS server to listen on"`
	TrustedProxies  []string         `name:"trusted-proxies" description:"CIDRs of trusted reverse proxies"`
	RedirectToHost  string           `name:"redirect-to-host" description:"Redirect all requests to one host"`
	RedirectToHTTPS bool             `name:"redirect-to-tls" description:"Redirect HTTP requests to HTTPS"`
	LogIgnorePaths  []string         `name:"log-ignore-paths" description:"List of paths for which successful requests will not be logged"` //nolint:lll
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
		Enable              bool          `name:"enable" description:"Enable events store"`
		TTL                 time.Duration `name:"ttl" description:"How long event payloads are retained"`
		EntityCount         int           `name:"entity-count" description:"How many events are indexed for a entity ID"`
		EntityTTL           time.Duration `name:"entity-ttl" description:"How long events are indexed for a entity ID"`
		CorrelationIDCount  int           `name:"correlation-id-count" description:"How many events are indexed for a correlation ID"`     //nolint:lll
		StreamPartitionSize int           `name:"stream-partition-size" description:"How many streams to listen to in a single partition"` //nolint:lll
	} `name:"store"`
	Workers int `name:"workers"`
	Publish struct {
		QueueSize  int `name:"queue-size" description:"The maximum number of events which may be queued for publication"`
		MaxWorkers int `name:"max-workers" description:"The maximum number of workers which may publish events asynchronously"` //nolint:lll
	} `name:"publish"`
}

// BatchEvents represents the configuration for batch event publication.
type BatchEvents struct {
	Enable bool          `name:"enable" description:"Enable events batching (EXPERIMENTAL)"`
	Size   int           `name:"size" description:"How many items to have in a batch (EXPERIMENTAL)"`
	Delay  time.Duration `name:"delay" description:"For how long to delay event submission in order to build a batch (EXPERIMENTAL)"` // nolint:lll
}

// Events represents configuration for the events system.
type Events struct {
	Backend string      `name:"backend" description:"Backend to use for events (internal, redis, cloud)"`
	Redis   RedisEvents `name:"redis"`
	Cloud   CloudEvents `name:"cloud"`
	Batch   BatchEvents `name:"batch"`
}

// Rights represents the configuration to apply when fetching entity rights.
type Rights struct {
	// TTL is the duration that entries will remain in the cache before being
	// garbage collected.
	TTL time.Duration `name:"ttl" description:"Validity of Identity Server responses"`
}

// KeyVaultCache represents the configuration for key vault caching.
type KeyVaultCache struct {
	Size     int           `name:"size" description:"Cache size. Caching is disabled if size is 0"`
	TTL      time.Duration `name:"ttl" description:"Cache elements time to live. No expiration mechanism is used if TTL is 0"` //nolint:lll
	ErrorTTL time.Duration `name:"error-ttl" description:"Cache elements time to live for errors. If 0, the TTL is used"`
}

// KeyVault represents configuration for key vaults.
type KeyVault struct {
	Provider string            `name:"provider" description:"Provider (static)"`
	Cache    KeyVaultCache     `name:"cache"`
	Static   map[string][]byte `name:"static"`
}

// ComponentKEKLabeler returns an initialized crypto.ComponentKEKLabeler based on the configuration.
func (v KeyVault) ComponentKEKLabeler() (crypto.ComponentKEKLabeler, error) {
	switch v.Provider { //nolint:revive
	default:
		return &cryptoutil.ComponentPrefixKEKLabeler{
			Separator:     ":",
			ReplaceOldNew: []string{":", "_"},
		}, nil
	}
}

// KeyService returns an initialized crypto.KeyService based on the configuration.
func (v KeyVault) KeyService(ctx context.Context, httpClientProvider httpclient.Provider) (crypto.KeyService, error) {
	var kv crypto.KeyVault
	switch v.Provider {
	case "static":
		kv = cryptoutil.NewMemKeyVault(v.Static)
	default:
		kv = cryptoutil.EmptyKeyVault
	}
	if v.Cache.Size > 0 {
		errTTL := v.Cache.ErrorTTL
		if errTTL == 0 {
			errTTL = v.Cache.TTL
		}
		kv = cryptoutil.NewCacheKeyVault(kv,
			cryptoutil.WithCacheKeyVaultTTL(v.Cache.TTL, errTTL),
			cryptoutil.WithCacheKeyVaultSize(v.Cache.Size),
		)
	}
	ks := crypto.NewKeyService(kv)
	if v.Cache.Size > 0 {
		ks = cryptoutil.NewCacheKeyService(ks, v.Cache.TTL, v.Cache.Size)
	}
	return ks, nil
}

var (
	errUnknownBlobProvider = errors.DefineInvalidArgument(
		"unknown_blob_provider", "unknown blob store provider `{provider}`",
	)
	errMissingBlobConfig = errors.DefineInvalidArgument(
		"missing_blob_config", "missing blob store configuration",
	)
)

// BlobConfigLocal is the blob store configuration for the local filesystem provider.
type BlobConfigLocal struct {
	Directory string `name:"directory" description:"OS filesystem directory, which contains buckets"`
}

// BlobConfigAzure is the blob store configuration for the Azure provider.
type BlobConfigAzure struct {
	AccountName string `name:"account-name" description:"Azure storage account name"`
}

// BlobConfigAWS is the blob store configuration for the AWS provider.
type BlobConfigAWS struct {
	Endpoint         string `name:"endpoint" description:"S3 endpoint"`
	Region           string `name:"region" description:"S3 region"`
	AccessKeyID      string `name:"access-key-id" description:"Access key ID"`
	SecretAccessKey  string `name:"secret-access-key" description:"Secret access key"`
	SessionToken     string `name:"session-token" description:"Session token"`
	S3ForcePathStyle *bool  `name:"s3-force-path-style" description:"Force the AWS SDK to use path-style (s3://) addressing for calls to S3"` //nolint:lll
}

// BlobConfigGCP is the blob store configuration for the GCP provider.
type BlobConfigGCP struct {
	CredentialsFile string `name:"credentials-file" description:"Path to the GCP credentials JSON file"`
	Credentials     string `name:"credentials" description:"JSON data of the GCP credentials, if not using JSON file"`
}

// BlobConfig is the blob store configuration.
type BlobConfig struct {
	Provider string          `name:"provider" description:"Blob store provider (local, aws, gcp, azure)"`
	Local    BlobConfigLocal `name:"local"`
	AWS      BlobConfigAWS   `name:"aws"`
	GCP      BlobConfigGCP   `name:"gcp"`
	Azure    BlobConfigAzure `name:"azure" description:"Azure Storage configuration (EXPERIMENTAL)"`
}

// IsZero returns whether conf is empty.
func (c BlobConfig) IsZero() bool {
	return c.Provider == "" &&
		c.Local == BlobConfigLocal{} &&
		c.AWS == BlobConfigAWS{} &&
		c.GCP == BlobConfigGCP{} &&
		c.Azure == BlobConfigAzure{}
}

// Bucket returns the requested blob bucket using the config.
func (c BlobConfig) Bucket(
	ctx context.Context, bucket string, httpClientProvider httpclient.Provider,
) (*blob.Bucket, error) {
	switch c.Provider {
	case "local":
		return ttnblob.Local(ctx, bucket, c.Local.Directory)
	case "aws":
		httpClient, err := httpClientProvider.HTTPClient(ctx)
		if err != nil {
			return nil, err
		}
		conf := aws.NewConfig().WithHTTPClient(httpClient)
		if c.AWS.Endpoint != "" {
			conf = conf.WithEndpoint(c.AWS.Endpoint)
		}
		if c.AWS.S3ForcePathStyle != nil {
			conf.WithS3ForcePathStyle(*c.AWS.S3ForcePathStyle)
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
	case "azure":
		if c.Azure.AccountName == "" {
			return nil, errMissingBlobConfig.New()
		}
		return ttnblob.Azure(ctx, c.Azure.AccountName, bucket)
	case "gcp":
		if c.GCP.Credentials != "" {
			return ttnblob.GCP(ctx, bucket, []byte(c.GCP.Credentials))
		}
		if c.GCP.CredentialsFile != "" {
			jsonCreds, err := os.ReadFile(c.GCP.CredentialsFile)
			if err != nil {
				return nil, err
			}
			return ttnblob.GCP(ctx, bucket, jsonCreds)
		}
		return nil, errMissingBlobConfig.New()
	default:
		return nil, errUnknownBlobProvider.WithAttributes("provider", c.Provider)
	}
}

// BlobPathConfig configures the blob bucket and path.
type BlobPathConfig struct {
	Bucket string `name:"bucket" description:"Bucket to use"`
	Path   string `name:"path" description:"Path to use"`
}

// IsZero returns whether conf is empty.
func (c BlobPathConfig) IsZero() bool {
	return c == BlobPathConfig{}
}

// FrequencyPlansConfig contains the source of the frequency plans.
type FrequencyPlansConfig struct {
	ConfigSource string            `name:"config-source" description:"Source of the frequency plans (static, directory, url, blob)"` //nolint:lll
	Static       map[string][]byte `name:"-"`
	Directory    string            `name:"directory" description:"OS filesystem directory, which contains frequency plans"` //nolint:lll
	URL          string            `name:"url" description:"URL, which contains frequency plans"`
	Blob         BlobPathConfig    `name:"blob"`
}

// Fetcher returns a fetch.Interface based on the configuration.
// If no configuration source is set, this method returns nil, nil.
func (c FrequencyPlansConfig) Fetcher(
	ctx context.Context, blobConf BlobConfig, httpClientProvider httpclient.Provider,
) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	configSource := c.ConfigSource
	if configSource == "" {
		switch {
		case c.Static != nil:
			configSource = "static"
		case c.Directory != "":
			if stat, err := os.Stat(c.Directory); err == nil && stat.IsDir() {
				configSource = "directory"
				break
			}
			fallthrough
		case c.URL != "":
			configSource = "url"
		case !c.Blob.IsZero():
			configSource = "blob"
		}
	}
	switch configSource {
	case "static":
		return fetch.NewMemFetcher(c.Static), nil
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		httpClient, err := httpClientProvider.HTTPClient(ctx, httpclient.WithCache(true))
		if err != nil {
			return nil, err
		}
		return fetch.FromHTTP(httpClient, c.URL)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket, httpClientProvider)
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
	ConfigSource string         `name:"config-source" description:"Source of the interoperability client configuration (directory, url, blob)"` //nolint:lll
	Directory    string         `name:"directory" description:"OS filesystem directory, which contains interoperability client configuration"`  //nolint:lll
	URL          string         `name:"url" description:"URL, which contains interoperability client configuration"`
	Blob         BlobPathConfig `name:"blob"`

	BlobConfig BlobConfig `name:"-"`
}

// IsZero returns whether conf is empty.
func (c InteropClient) IsZero() bool {
	return c.ConfigSource == "" &&
		c.Directory == "" &&
		c.URL == "" &&
		c.Blob.IsZero() &&
		c.BlobConfig.IsZero()
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c InteropClient) Fetcher(ctx context.Context, httpClientProvider httpclient.Provider) (fetch.Interface, error) {
	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	configSource := c.ConfigSource
	if configSource == "" {
		switch {
		case c.Directory != "":
			if stat, err := os.Stat(c.Directory); err == nil && stat.IsDir() {
				configSource = "directory"
				break
			}
			fallthrough
		case c.URL != "":
			configSource = "url"
		case !c.Blob.IsZero():
			configSource = "blob"
		}
	}
	switch configSource {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		httpClient, err := httpClientProvider.HTTPClient(ctx, httpclient.WithCache(true))
		if err != nil {
			return nil, err
		}
		return fetch.FromHTTP(httpClient, c.URL)
	case "blob":
		b, err := c.BlobConfig.Bucket(ctx, c.Blob.Bucket, httpClientProvider)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

// SenderClientCA is the sender client CA configuration.
type SenderClientCA struct {
	Source    string            `name:"source" description:"Source of the sender client CA configuration (static, directory, url, blob)"` //nolint:lll
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"OS filesystem directory, which contains sender client CA configuration"` //nolint:lll
	URL       string            `name:"url" description:"URL, which contains sender client CA configuration"`
	Blob      BlobPathConfig    `name:"blob"`

	BlobConfig BlobConfig `name:"-"`
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c SenderClientCA) Fetcher(ctx context.Context, httpClientProvider httpclient.Provider) (fetch.Interface, error) {
	switch c.Source {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		httpClient, err := httpClientProvider.HTTPClient(ctx, httpclient.WithCache(true))
		if err != nil {
			return nil, err
		}
		return fetch.FromHTTP(httpClient, c.URL)
	case "blob":
		b, err := c.BlobConfig.Bucket(ctx, c.Blob.Bucket, httpClientProvider)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}

// PacketBrokerInteropAuth respresents Packet Broker authentication configuration.
type PacketBrokerInteropAuth struct {
	Enabled     bool   `name:"enabled" description:"Enable Packet Broker to authenticate"`
	TokenIssuer string `name:"token-issuer" description:"Required issuer of Packet Broker tokens"`
}

// InteropServer represents the server-side interoperability through LoRaWAN Backend Interfaces configuration.
type InteropServer struct {
	Listen           string `name:"listen" description:"Address for the interop server for LoRaWAN Backend Interfaces to listen on"`         //nolint:lll
	ListenTLS        string `name:"listen-tls" description:"TLS address for the interop server for LoRaWAN Backend Interfaces to listen on"` //nolint:lll
	PublicTLSAddress string `name:"public-tls-address" description:"Public address of the interop server for LoRaWAN Backend Interfaces"`    //nolint:lll

	SenderClientCA           SenderClientCA    `name:"sender-client-ca" description:"Client CAs for sender IDs to trust (DEPRECATED)"`                               //nolint:lll
	SenderClientCADeprecated map[string]string `name:"sender-client-cas" description:"Path to PEM encoded file with client CAs of sender IDs to trust (DEPRECATED)"` //nolint:lll

	PacketBroker PacketBrokerInteropAuth `name:"packet-broker"`
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
	Tracing          tracing.Config       `name:"tracing" yaml:"tracing" description:"Tracing configuration"`
	SkipVersionCheck bool                 `name:"skip-version-check" yaml:"skip-version-check" description:"Skip version checks"` //nolint:lll
	Telemetry        telemetry.Config     `name:"telemetry" yaml:"telemetry" description:"Telemetry configuration"`
}

// FrequencyPlansFetcher returns a fetch.Interface based on the frequency plans configuration.
// If no configuration source is set, this method returns nil, nil.
func (c ServiceBase) FrequencyPlansFetcher(
	ctx context.Context, httpClientProvider httpclient.Provider,
) (fetch.Interface, error) {
	return c.FrequencyPlans.Fetcher(ctx, c.Blob, httpClientProvider)
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
	Directory    string         `name:"directory" description:"OS filesystem directory, which contains rate limiting configuration"` //nolint:lll
	URL          string         `name:"url" description:"URL, which contains rate limiting configuration"`
	Blob         BlobPathConfig `name:"blob"`

	Memory   RateLimitingMemory    `name:"memory" description:"In-memory rate limiting store configuration"`
	Profiles []RateLimitingProfile `name:"profiles" description:"Rate limiting profiles"`
}

// Fetcher returns fetch.Interface defined by conf.
// If no configuration source is set, this method returns nil, nil.
func (c RateLimiting) Fetcher(
	ctx context.Context, blobConf BlobConfig, httpClientProvider httpclient.Provider,
) (fetch.Interface, error) {
	switch c.ConfigSource {
	case "directory":
		return fetch.FromFilesystem(c.Directory), nil
	case "url":
		httpClient, err := httpClientProvider.HTTPClient(ctx, httpclient.WithCache(true))
		if err != nil {
			return nil, err
		}
		return fetch.FromHTTP(httpClient, c.URL)
	case "blob":
		b, err := blobConf.Bucket(ctx, c.Blob.Bucket, httpClientProvider)
		if err != nil {
			return nil, err
		}
		return fetch.FromBucket(ctx, b, c.Blob.Path), nil
	default:
		return nil, nil
	}
}
