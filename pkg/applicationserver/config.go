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
	"net/http"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	loraclouddevicemanagementv1 "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1"
	loracloudgeolocationv3 "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// InteropClient is a client, which Application Server can use for interoperability.
type InteropClient interface {
	GetAppSKey(ctx context.Context, asID string, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error)
}

// InteropConfig represents interoperability client configuration.
type InteropConfig struct {
	config.InteropClient `name:",squash"`
	ID                   string `name:"id" description:"AS-ID used for interoperability"`
}

// EndDeviceFetcherConfig represents configuration for the end device fetcher in Application Server.
type EndDeviceFetcherConfig struct {
	Fetcher        EndDeviceFetcher                     `name:"-"`
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

type EndDeviceFetcherCircuitBreakerConfig struct {
	Enable    bool          `name:"enable" description:"Enable circuit breaker behavior on burst errors"`
	Timeout   time.Duration `name:"timeout" description:"Timeout after which the circuit breaker closes"`
	Threshold int           `name:"threshold" description:"Number of failed fetching attempts after which the circuit breaker opens"`
}

type FormattersConfig struct {
	MaxParameterLength int `name:"max-parameter-length" description:"Maximum allowed size for length of formatter parameters (payload formatter scripts)"`
}

// Config represents the ApplicationServer configuration.
type Config struct {
	LinkMode         string                    `name:"link-mode" description:"Deprecated - mode to link applications to their Network Server (all, explicit)"`
	Devices          DeviceRegistry            `name:"-"`
	Links            LinkRegistry              `name:"-"`
	UplinkStorage    UplinkStorageConfig       `name:"uplink-storage" description:"Application uplinks storage configuration"`
	Formatters       FormattersConfig          `name:"formatters" description:"Payload formatters configuration"`
	Distribution     DistributionConfig        `name:"distribution" description:"Distribution configuration"`
	EndDeviceFetcher EndDeviceFetcherConfig    `name:"fetcher" description:"End Device fetcher configuration"`
	MQTT             config.MQTT               `name:"mqtt" description:"MQTT configuration"`
	Webhooks         WebhooksConfig            `name:"webhooks" description:"Webhooks configuration"`
	PubSub           PubSubConfig              `name:"pubsub" description:"Pub/sub messaging configuration"`
	Packages         ApplicationPackagesConfig `name:"packages" description:"Application packages configuration"`
	Interop          InteropConfig             `name:"interop" description:"Interop client configuration"`
	DeviceKEKLabel   string                    `name:"device-kek-label" description:"Label of KEK used to encrypt device keys at rest"`
}

func (c Config) toProto() *ttnpb.AsConfiguration {
	return &ttnpb.AsConfiguration{
		Pubsub: c.PubSub.toProto(),
	}
}

var (
	errWebhooksRegistry = errors.DefineInvalidArgument("webhooks_registry", "invalid webhooks registry")
	errWebhooksTarget   = errors.DefineInvalidArgument("webhooks_target", "invalid webhooks target `{target}`")
)

// UplinkStorageConfig defines the configuration of the application uplinks storage used by integrations.
type UplinkStorageConfig struct {
	Registry ApplicationUplinkRegistry `name:"-"`
	Limit    int64                     `name:"limit" description:"Number of application uplinks to be stored"`
}

// WebhooksConfig defines the configuration of the webhooks integration.
type WebhooksConfig struct {
	Registry  web.WebhookRegistry `name:"-"`
	Target    string              `name:"target" description:"Target of the integration (direct)"`
	Timeout   time.Duration       `name:"timeout" description:"Wait timeout of the target to process the request"`
	QueueSize int                 `name:"queue-size" description:"Number of requests to queue"`
	Workers   int                 `name:"workers" description:"Number of workers to process requests"`
	Templates web.TemplatesConfig `name:"templates" description:"The store of the webhook templates"`
	Downlinks web.DownlinksConfig `name:"downlink" description:"The downlink queue operations configuration"`
}

// DistributionConfig contains the upstream traffic distribution configuration of the Application Server.
type DistributionConfig struct {
	PubSub  distribution.PubSub `name:"-"`
	Timeout time.Duration       `name:"timeout" description:"Wait timeout of an empty subscription set"`
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
	var target web.Sink
	switch c.Target {
	case "":
		return nil, nil
	case "direct":
		client, err := server.HTTPClient(ctx)
		if err != nil {
			return nil, err
		}
		client.Timeout = c.Timeout
		target = &web.HTTPClientSink{
			Client: client,
		}
	default:
		return nil, errWebhooksTarget.WithAttributes("target", c.Target)
	}
	if c.Registry == nil {
		return nil, errWebhooksRegistry.New()
	}
	if c.QueueSize > 0 || c.Workers > 0 {
		target = &web.QueuedSink{
			Target:  target,
			Queue:   make(chan *http.Request, c.QueueSize),
			Workers: c.Workers,
		}
	}
	if controllable, ok := target.(web.ControllableSink); ok {
		go func() {
			if err := controllable.Run(ctx); err != nil && !errors.IsCanceled(err) {
				log.FromContext(ctx).WithError(err).Error("Webhooks target sink failed")
			}
		}()
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

	return packages.New(ctx, server, c.Registry, handlers, c.Workers, c.Timeout)
}

var (
	errInvalidTTL       = errors.DefineInvalidArgument("invalid_ttl", "invalid TTL `{ttl}`")
	errInvalidThreshold = errors.DefineInvalidArgument("invalid_threshold", "invalid threshold `{threshold}`")
)

// NewFetcher creates an EndDeviceFetcher from config.
func (c EndDeviceFetcherConfig) NewFetcher(comp *component.Component) (EndDeviceFetcher, error) {
	fetcher := NewRegistryEndDeviceFetcher(comp)
	if c.Timeout != 0 {
		fetcher = NewTimeoutEndDeviceFetcher(fetcher, c.Timeout)
	}
	if c.CircuitBreaker.Enable {
		if c.CircuitBreaker.Threshold <= 0 {
			return nil, errInvalidThreshold.WithAttributes("threshold", c.CircuitBreaker.Threshold)
		}
		fetcher = NewCircuitBreakerEndDeviceFetcher(fetcher, uint64(c.CircuitBreaker.Threshold), c.CircuitBreaker.Timeout)
	}
	if c.Cache.Enable {
		if c.Cache.TTL <= 0 {
			return nil, errInvalidTTL.WithAttributes("ttl", c.Cache.TTL)
		}
		var builder *gcache.CacheBuilder
		if c.Cache.Size > 0 {
			builder = gcache.New(c.Cache.Size).LFU()
		} else {
			builder = gcache.New(-1)
		}
		builder = builder.Expiration(c.Cache.TTL)
		fetcher = NewCachedEndDeviceFetcher(fetcher, builder.Build())
	}
	fetcher = NewSingleFlightEndDeviceFetcher(fetcher)

	return fetcher, nil
}
