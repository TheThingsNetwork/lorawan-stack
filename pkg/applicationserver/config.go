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

package applicationserver

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	alcsyncv1 "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/alcsync/v1"
	loraclouddevicemanagementv1 "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1"
	loracloudgeolocationv3 "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/lastseen"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// InteropClient is a client, which Application Server can use for interoperability.
type InteropClient interface {
	GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error)
}

// InteropConfig represents interoperability client configuration.
type InteropConfig struct {
	config.InteropClient `name:",squash"`
	ID                   string `name:"id" description:"AS-ID of this Application Server"`
}

// EndDeviceFetcherConfig represents configuration for the end device fetcher in Application Server.
type EndDeviceFetcherConfig struct {
	Timeout        time.Duration                        `name:"timeout" description:"Timeout of the end device retrival operation"`
	Cache          EndDeviceFetcherCacheConfig          `name:"cache" description:"Cache configuration options for the end device fetcher"`
	CircuitBreaker EndDeviceFetcherCircuitBreakerConfig `name:"circuit-breaker" description:"Circuit breaker options for the end device fetcher"`
}

// EndDeviceFetcherCacheConfig represents configuration for device information caching in Application Server.
type EndDeviceFetcherCacheConfig struct {
	Enable bool          `name:"enable" description:"Cache fetched end devices"`
	TTL    time.Duration `name:"ttl" description:"TTL for cached end devices"`
	Size   int           `name:"size" description:"Cache size"`
}

// EndDeviceFetcherCircuitBreakerConfig represents the configuration for the circuit breaker of the end device fetcher.
type EndDeviceFetcherCircuitBreakerConfig struct {
	Enable    bool          `name:"enable" description:"Enable circuit breaker behavior on burst errors"`
	Timeout   time.Duration `name:"timeout" description:"Timeout after which the circuit breaker closes"`
	Threshold int           `name:"threshold" description:"Number of failed fetching attempts after which the circuit breaker opens"`
}

// EndDeviceMetadataStorageConfig represents the configuration of end device metadata operations.
type EndDeviceMetadataStorageConfig struct {
	Location EndDeviceLocationStorageConfig `name:"location"`
}

// EndDeviceLocationStorageConfig represents the configuration of end device locations storage.
type EndDeviceLocationStorageConfig struct {
	Registry metadata.EndDeviceLocationRegistry  `name:"-"`
	Timeout  time.Duration                       `name:"timeout" description:"Timeout of the end device retrival operation"`
	Cache    EndDeviceLocationStorageCacheConfig `name:"cache"`
}

// EndDeviceLocationStorageCacheConfig represents the configuration of end device location registry caching.
type EndDeviceLocationStorageCacheConfig struct {
	Cache              metadata.EndDeviceLocationCache `name:"-"`
	Enable             bool                            `name:"enable" description:"Enable caching of end device locations"`
	MinRefreshInterval time.Duration                   `name:"min-refresh-interval" description:"Minimum time interval between two asynchronous refreshes"`
	MaxRefreshInterval time.Duration                   `name:"max-refresh-interval" description:"Maximum time interval between two asynchronous refreshes"`
	TTL                time.Duration                   `name:"eviction-ttl" description:"Time to live of cached locations"`
}

// FormattersConfig represents the configuration for payload formatters.
type FormattersConfig struct {
	MaxParameterLength int `name:"max-parameter-length" description:"Maximum allowed size for length of formatter parameters (payload formatter scripts)"`
}

// ConfirmationConfig represents the configuration for confirmed downlink.
type ConfirmationConfig struct {
	DefaultRetryAttempts uint32 `name:"default-retry-attempts" description:"Default number of retry attempts for confirmed downlink"` // nolint:lll
	MaxRetryAttempts     uint32 `name:"max-retry-attempts" description:"Maximum number of retry attempts for confirmed downlink"`     // nolint:lll
}

// DownlinksConfig represents the configuration for downlinks.
type DownlinksConfig struct {
	ConfirmationConfig ConfirmationConfig `name:"confirmation" description:"Configuration for confirmed downlink"`
}

// PaginationConfig represents the configuration for pagination.
type PaginationConfig struct {
	DefaultLimit int64 `name:"default-limit" description:"Default limit for pagination"`
}

// Config represents the ApplicationServer configuration.
type Config struct {
	LinkMode                 string                         `name:"link-mode" description:"Deprecated - mode to link applications to their Network Server (all, explicit)"`
	Devices                  DeviceRegistry                 `name:"-"`
	Links                    LinkRegistry                   `name:"-"`
	UplinkStorage            UplinkStorageConfig            `name:"uplink-storage" description:"Application uplinks storage configuration"`
	Formatters               FormattersConfig               `name:"formatters" description:"Payload formatters configuration"`
	Distribution             DistributionConfig             `name:"distribution" description:"Distribution configuration"`
	EndDeviceFetcher         EndDeviceFetcherConfig         `name:"fetcher" description:"Deprecated - End Device fetcher configuration"`
	EndDeviceMetadataStorage EndDeviceMetadataStorageConfig `name:"end-device-metadata-storage" description:"End device metadata storage configuration"`
	MQTT                     config.MQTT                    `name:"mqtt" description:"MQTT configuration"`
	Webhooks                 WebhooksConfig                 `name:"webhooks" description:"Webhooks configuration"`
	PubSub                   PubSubConfig                   `name:"pubsub" description:"Pub/sub messaging configuration"`
	Packages                 ApplicationPackagesConfig      `name:"packages" description:"Application packages configuration"`
	Interop                  InteropConfig                  `name:"interop" description:"Interop client configuration"`
	DeviceKEKLabel           string                         `name:"device-kek-label" description:"Label of KEK used to encrypt device keys at rest"`
	DeviceLastSeen           LastSeenConfig                 `name:"device-last-seen" description:"End Device last seen batch update configuration"`
	Downlinks                DownlinksConfig                `name:"downlinks" description:"Downlink configuration"`
	Pagination               PaginationConfig               `name:"pagination" description:"Pagination configuration"`
}

func (c Config) toProto() *ttnpb.AsConfiguration {
	return &ttnpb.AsConfiguration{
		Pubsub:   c.PubSub.toProto(),
		Webhooks: c.Webhooks.toProto(),
	}
}

var (
	errWebhooksRegistry = errors.DefineInvalidArgument("webhooks_registry", "invalid webhooks registry")
	errWebhooksTarget   = errors.DefineInvalidArgument("webhooks_target", "invalid webhooks target `{target}`")
)

// UplinkStorageConfig defines the configuration of the application uplinks storage used by integrations.
type UplinkStorageConfig struct {
	Registry ApplicationUplinkRegistry `name:"-"`
	Limit    int64                     `name:"limit" description:"DEPRECATED"`
}

// WebhooksConfig defines the configuration of the webhooks integration.
type WebhooksConfig struct {
	Registry                   web.WebhookRegistry `name:"-"`
	Target                     string              `name:"target" description:"Target of the integration (direct)"`
	Timeout                    time.Duration       `name:"timeout" description:"Wait timeout of the target to process the request"`
	QueueSize                  int                 `name:"queue-size" description:"Number of requests to queue"`
	Workers                    int                 `name:"workers" description:"Number of workers to process requests"`
	UnhealthyAttemptsThreshold int                 `name:"unhealthy-attempts-threshold" description:"Number of failed webhook attempts before the webhook is disabled"`
	UnhealthyRetryInterval     time.Duration       `name:"unhealthy-retry-interval" description:"Time interval after which disabled webhooks may execute again"`
	Templates                  web.TemplatesConfig `name:"templates" description:"The store of the webhook templates"`
	Downlinks                  web.DownlinksConfig `name:"downlink" description:"The downlink queue operations configuration"`
}

func (c WebhooksConfig) toProto() *ttnpb.AsConfiguration_Webhooks {
	return &ttnpb.AsConfiguration_Webhooks{
		UnhealthyAttemptsThreshold: int64(c.UnhealthyAttemptsThreshold),
		UnhealthyRetryInterval:     durationpb.New(c.UnhealthyRetryInterval),
	}
}

// DistributionConfig contains the upstream traffic distribution configuration of the Application Server.
type DistributionConfig struct {
	Timeout time.Duration           `name:"timeout" description:"Wait timeout of an empty subscription set"`
	Local   LocalDistributorConfig  `name:"local" description:"Local distributor configuration"`
	Global  GlobalDistributorConfig `name:"global" description:"Global distributor configuration"`
}

// DistributorConfig contains the configuration of a traffic distributor of the Application Server.
type DistributorConfig struct {
	SubscriptionQueueSize int  `name:"subscription-queue-size" description:"Number of uplinks to queue for each subscriber"`
	SubscriptionBlocks    bool `name:"subscription-blocks" description:"Controls if traffic should be dropped if the queue of a subscriber is full"`
}

// SubscriptionOptions generates the subscription options based on the configuration.
func (c DistributorConfig) SubscriptionOptions() []io.SubscriptionOption {
	if c.SubscriptionQueueSize == 0 {
		c.SubscriptionQueueSize = io.DefaultBufferSize
	}
	if c.SubscriptionQueueSize < 0 {
		c.SubscriptionQueueSize = 0
	}
	return []io.SubscriptionOption{
		io.WithBlocking(c.SubscriptionBlocks),
		io.WithBufferSize(c.SubscriptionQueueSize),
	}
}

// LocalDistributorConfig contains the configuration of the local traffic distributor of the Application Server.
type LocalDistributorConfig struct {
	Broadcast  DistributorConfig `name:"broadcast" description:"Broadcast distributor configuration"`
	Individual DistributorConfig `name:"individual" description:"Individual distributor configuration"`
}

// GlobalDistributorConfig contains the configuration of the global traffic distributor of the Application Server.
type GlobalDistributorConfig struct {
	PubSub     distribution.PubSub `name:"-"`
	Individual DistributorConfig   `name:"individual" description:"Individual distributor configuration"`
}

// PubSubConfig contains go-cloud pub/sub configuration of the Application Server.
type PubSubConfig struct {
	Registry pubsub.Registry `name:"-"`

	Providers map[string]string `name:"providers" description:"Controls the status of each provider (enabled, disabled, warning)"`
}

func (c PubSubConfig) toProto() *ttnpb.AsConfiguration_PubSub {
	toStatus := func(s string) ttnpb.AsConfiguration_PubSub_Providers_Status {
		switch s {
		case "enabled":
			return ttnpb.AsConfiguration_PubSub_Providers_ENABLED
		case "warning":
			return ttnpb.AsConfiguration_PubSub_Providers_WARNING
		case "disabled":
			return ttnpb.AsConfiguration_PubSub_Providers_DISABLED
		default:
			panic("unknown provider status")
		}
	}
	providers := &ttnpb.AsConfiguration_PubSub_Providers{}
	if status, ok := c.Providers["mqtt"]; ok {
		providers.Mqtt = toStatus(status)
	}
	if status, ok := c.Providers["nats"]; ok {
		providers.Nats = toStatus(status)
	}
	return &ttnpb.AsConfiguration_PubSub{
		Providers: providers,
	}
}

// ApplicationPackagesConfig contains application packages associations configuration.
type ApplicationPackagesConfig struct {
	packages.Config `name:",squash"`
	Registry        packages.Registry `name:"-"`
}

// NewWebhooks returns a new web.Webhooks based on the configuration.
// If Target is empty, this method returns nil.
func (c WebhooksConfig) NewWebhooks(ctx context.Context, server io.Server) (web.Webhooks, error) {
	var target sink.Sink
	switch c.Target {
	case "":
		return nil, nil
	case "direct":
		client, err := server.HTTPClient(ctx)
		if err != nil {
			return nil, err
		}
		client.Timeout = c.Timeout
		target = sink.NewHTTPClientSink(client)
	default:
		return nil, errWebhooksTarget.WithAttributes("target", c.Target)
	}
	if c.Registry == nil {
		return nil, errWebhooksRegistry.New()
	}
	if c.UnhealthyAttemptsThreshold > 0 || c.UnhealthyRetryInterval > 0 {
		registry := web.NewHealthStatusRegistry(c.Registry)
		registry = web.NewCachedHealthStatusRegistry(registry)
		target = sink.NewHealthCheckSink(target, registry, c.UnhealthyAttemptsThreshold, c.UnhealthyRetryInterval)
	}
	if c.QueueSize > 0 || c.Workers > 0 {
		target = sink.NewPooledSink(ctx, server, target, c.Workers, c.QueueSize)
	}
	return web.NewWebhooks(ctx, server, c.Registry, target, c.Downlinks)
}

// NewPubSub returns a new pubsub.PubSub based on the configuration.
// If the registry is nil, it returns nil.
func (c PubSubConfig) NewPubSub(comp *component.Component, server io.Server) (*pubsub.PubSub, error) {
	if c.Registry == nil {
		return nil, nil
	}
	statuses, err := pubsub.ProviderStatusesFromMap(comp.Context(), c.Providers)
	if err != nil {
		return nil, err
	}
	return pubsub.New(comp, server, c.Registry, statuses)
}

// NewApplicationPackages returns a new applications packages frontend based on the configuration.
// If the registry is nil, it returns nil.
func (c ApplicationPackagesConfig) NewApplicationPackages(ctx context.Context, server io.Server) (packages.Server, error) {
	if c.Registry == nil {
		return nil, nil
	}
	handlers := make(map[string]packages.ApplicationPackageHandler)

	// Initialize LoRa Cloud Device Management v1 package handler
	handlers[loraclouddevicemanagementv1.PackageName] = loraclouddevicemanagementv1.New(server, c.Registry)

	// Initialize LoRa Cloud Geolocation v3 package handler
	handlers[loracloudgeolocationv3.PackageName] = loracloudgeolocationv3.New(server, c.Registry)

	// Initialize LoRa Application Layer Clock Synchronization v1 package handler.
	handlers[alcsyncv1.PackageName] = alcsyncv1.New(server, c.Registry)

	return packages.New(ctx, server, c.Registry, handlers, c.Workers, c.Timeout)
}

var (
	errInvalidTimeout = errors.DefineInvalidArgument("invalid_timeout", "invalid timeout `{timeout}`")
	errInvalidTTL     = errors.DefineInvalidArgument("invalid_ttl", "invalid TTL `{ttl}`")
)

// NewRegistry returns a new end device location registry based on the configuration.
func (c EndDeviceLocationStorageConfig) NewRegistry(ctx context.Context, comp *component.Component) (metadata.EndDeviceLocationRegistry, error) {
	if c.Timeout <= 0 {
		return nil, errInvalidTimeout.WithAttributes("timeout", c.Timeout)
	}
	registry := metadata.NewClusterEndDeviceLocationRegistry(comp, c.Timeout)
	registry = metadata.NewMetricsEndDeviceLocationRegistry(registry)
	if c.Cache.Enable {
		for _, ttl := range []time.Duration{c.Cache.MinRefreshInterval, c.Cache.MaxRefreshInterval, c.Cache.TTL} {
			if ttl <= 0 {
				return nil, errInvalidTTL.WithAttributes("ttl", ttl)
			}
		}
		cache := metadata.NewMetricsEndDeviceLocationCache(c.Cache.Cache)
		registry = metadata.NewCachedEndDeviceLocationRegistry(ctx, comp, registry, cache, c.Cache.MinRefreshInterval, c.Cache.MaxRefreshInterval, c.Cache.TTL)
	}
	return registry, nil
}

// LastSeenConfig defines configuration for the device last seen map which stores timestamps for batch updates.
type LastSeenConfig struct {
	BatchSize     int           `name:"batch-size" description:"Maximum number of end device last seen timestamps to store for batch update"`
	FlushInterval time.Duration `name:"flush-interval" description:"Interval at which last seen timestamps are updated in batches"`
}

// NewLastSeen defines a new batch update map.
func (c LastSeenConfig) NewLastSeen(ctx context.Context, comp *component.Component) (lastseen.LastSeenProvider, error) {
	if c.FlushInterval <= 0 {
		return lastseen.NewNoopLastSeenProvider()
	}
	return lastseen.NewBatchLastSeen(ctx, c.BatchSize, time.NewTicker(c.FlushInterval).C, comp)
}
