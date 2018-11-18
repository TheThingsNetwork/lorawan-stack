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

package applicationserver

import (
	"context"
	"net/http"
	"time"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// LinkMode defines how applications are linked to their Network Server.
type LinkMode int

const (
	// LinkAll links all applications in the link registry to their Network Server automatically.
	LinkAll LinkMode = iota
	// LinkExplicit links applications on request.
	LinkExplicit
)

// Config represents the ApplicationServer configuration.
type Config struct {
	LinkMode         string                 `name:"link-mode" description:"Mode to link applications to their Network Server (all, explicit)"`
	Devices          DeviceRegistry         `name:"-"`
	Links            LinkRegistry           `name:"-"`
	DeviceRepository DeviceRepositoryConfig `name:"device-repository" description:"Source of the device repository"`
	Webhooks         WebhooksConfig         `name:"webhooks" description:"Webhooks configuration"`
}

var errLinkMode = errors.DefineInvalidArgument("link_mode", "invalid link mode `{value}`")

// GetLinkMode returns the converted configuration's link mode to LinkMode.
func (c Config) GetLinkMode() (LinkMode, error) {
	switch c.LinkMode {
	case "all":
		return LinkAll, nil
	case "explicit":
		return LinkExplicit, nil
	default:
		return LinkMode(0), errLinkMode.WithAttributes("value", c.LinkMode)
	}
}

// DeviceRepositoryConfig defines the source of the device repository.
type DeviceRepositoryConfig struct {
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"Retrieve the device repository from the filesystem"`
	URL       string            `name:"url" description:"Retrieve the device repository from a web server"`
}

// Client instantiates a new devicerepository.Client with a fetcher based on the configuration.
// The order of precedence is Static, Directory and URL.
// If neither Static, Directory nor a URL is set, this method returns nil.
func (c DeviceRepositoryConfig) Client() *devicerepository.Client {
	var fetcher fetch.Interface
	switch {
	case c.Static != nil:
		fetcher = fetch.NewMemFetcher(c.Static)
	case c.Directory != "":
		fetcher = fetch.FromFilesystem(c.Directory)
	case c.URL != "":
		fetcher = fetch.FromHTTP(c.URL, true)
	default:
		return nil
	}
	return &devicerepository.Client{
		Fetcher: fetcher,
	}
}

var (
	errWebhooksRegistry = errors.DefineInvalidArgument("webhooks_registry", "invalid webhooks registry")
	errWebhooksTarget   = errors.DefineInvalidArgument("webhooks_target", "invalid webhooks target `{target}`")
)

// WebhooksConfig defines the configuration of the webhooks integration.
type WebhooksConfig struct {
	Registry   web.WebhookRegistry `name:"-"`
	Target     string              `name:"target" description:"Target of the integration (direct)"`
	Timeout    time.Duration       `name:"timeout" description:"Wait timeout of the target to process the request"`
	BufferSize int                 `name:"buffer-size" description:"Number of requests to buffer"`
	Workers    int                 `name:"workers" description:"Number of workers to process requests"`
}

// NewSubscription returns a new *io.Subscription based on the configuration.
// If Target is empty, this method returns nil.
func (c WebhooksConfig) NewSubscription(ctx context.Context) (*io.Subscription, error) {
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
		return nil, errWebhooksRegistry
	}
	if c.BufferSize > 0 || c.Workers > 0 {
		target = &web.BufferedSink{
			Target:  target,
			Buffer:  make(chan *http.Request, c.BufferSize),
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
	w := web.Webhooks{
		Registry: c.Registry,
		Target:   target,
	}
	return w.NewSubscription(ctx), nil
}
