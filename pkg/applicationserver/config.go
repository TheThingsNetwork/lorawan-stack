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
	Fetcher EndDeviceFetcher            `name:"-"`
	Cache   EndDeviceFetcherCacheConfig `name:"cache" description:"Cache configuration options for the end device fetcher"`
}

// EndDeviceFetcherCacheConfig represents configuration for device information caching in Application Server.
type EndDeviceFetcherCacheConfig struct {
	Enable bool          `name:"enable" description:"Cache fetched end devices"`
	TTL    time.Duration `name:"ttl" description:"TTL for cached end devices"`
	Size   int           `name:"size" description:"Cache size"`
}

// Config represents the ApplicationServer configuration.
type Config struct {
	LinkMode         string                    `name:"link-mode" description:"Deprecated - mode to link applications to their Network Server (all, explicit)"`
	Devices          DeviceRegistry            `name:"-"`
	Links            LinkRegistry              `name:"-"`
	Distribution     DistributionConfig        `name:"distribution" description:"Distribution configuration"`
	EndDeviceFetcher EndDeviceFetcherConfig    `name:"fetcher" description:"End Device fetcher configuration"`
	MQTT             config.MQTT               `name:"mqtt" description:"MQTT configuration"`
	Webhooks         WebhooksConfig            `name:"webhooks" description:"Webhooks configuration"`
	PubSub           PubSubConfig              `name:"pubsub" description:"Pub/sub messaging configuration"`
	Packages         ApplicationPackagesConfig `name:"packages" description:"Application packages configuration"`
	Interop          InteropConfig             `name:"interop" description:"Interop client configuration"`
	DeviceKEKLabel   string                    `name:"device-kek-label" description:"Label of KEK used to encrypt device keys at rest"`
}

var (
	errWebhooksRegistry = errors.DefineInvalidArgument("webhooks_registry", "invalid webhooks registry")
	errWebhooksTarget   = errors.DefineInvalidArgument("webhooks_target", "invalid webhooks target `{target}`")
)

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
		target = &web.HTTPClientSink{
			Client: &http.Client{
				Timeout: c.Timeout,
			},
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
	return pubsub.New(comp, server, c.Registry)
}

// NewApplicationPackages returns a new applications packages frontend based on the configuration.
// If the registry is nil, it returns nil.
func (c ApplicationPackagesConfig) NewApplicationPackages(ctx context.Context, server io.Server) (packages.Server, error) {
	if c.Registry == nil {
		return nil, nil
	}
	handlers := make(map[string]packages.ApplicationPackageHandler)

	// Initialize LoRa Cloud Device Management v1 package handler
	loradmsHandler := loraclouddevicemanagementv1.New(server, c.Registry)
	handlers[loradmsHandler.Package().Name] = loradmsHandler

	return packages.New(ctx, server, c.Registry, handlers)
}

var (
	errInvalidTTL = errors.DefineInvalidArgument("invalid_ttl", "Invalid TTL `{ttl}`")
)

// NewFetcher creates an EndDeviceFetcher from config.
func (c EndDeviceFetcherConfig) NewFetcher(comp *component.Component) (EndDeviceFetcher, error) {
	fetcher := NewRegistryEndDeviceFetcher(comp)
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

	return fetcher, nil
}
