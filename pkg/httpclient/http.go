// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/klauspost/compress/gzhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

var transportCompressionFeatureFlag = experimental.DefineFeature("http.client.transport.compression", true)

// defaultHTTPClientTimeout is the default timeout for the HTTP client.
const defaultHTTPClientTimeout = 10 * time.Second

// TLSClientConfigurationProvider provides a *tls.Config to be used by TLS clients.
type TLSClientConfigurationProvider interface {
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
}

// Provider constructs *http.Clients.
type Provider interface {
	HTTPClient(context.Context, ...Option) (*http.Client, error)
}

// Option is an option for HTTP clients.
type Option func(*httpClientOptions)

type httpClientOptions struct {
	cache            bool
	tlsConfig        *tls.Config
	tlsConfigOptions []tlsconfig.Option
}

// WithCache enables caching at transport level.
func WithCache(b bool) Option {
	return Option(func(o *httpClientOptions) {
		o.cache = true
	})
}

// WithTLSConfig configures the TLS configuration to be used by the transport.
func WithTLSConfig(c *tls.Config) Option {
	return Option(func(o *httpClientOptions) {
		o.tlsConfig = c
	})
}

// WithTLSConfigOptions configures the TLS configuration options provided to the TLS configuration provider
// by the transport.
func WithTLSConfigOptions(opts ...tlsconfig.Option) Option {
	return Option(func(o *httpClientOptions) {
		o.tlsConfigOptions = opts
	})
}

type provider struct {
	tlsConfigProvider TLSClientConfigurationProvider
}

// HTTPClient returns a new *http.Client with a default timeout and a configured transport.
func (p *provider) HTTPClient(ctx context.Context, opts ...Option) (*http.Client, error) {
	options := &httpClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.tlsConfig == nil {
		tlsConfig, err := p.tlsConfigProvider.GetTLSClientConfig(ctx, options.tlsConfigOptions...)
		if err != nil {
			return nil, err
		}
		options.tlsConfig = tlsConfig
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = options.tlsConfig

	var rt http.RoundTripper = transport
	if transportCompressionFeatureFlag.GetValue(ctx) {
		rt = gzhttp.Transport(rt)
	}
	rt = otelhttp.NewTransport(
		rt,
		otelhttp.WithTracerProvider(tracing.FromContext(ctx)),
	)
	if options.cache {
		rt = &httpcache.Transport{
			Transport:           rt,
			Cache:               httpcache.NewMemoryCache(),
			MarkCachedResponses: true,
		}
	}
	rt = &roundTripperWithUserAgent{
		RoundTripper: rt,
		UserAgent:    fmt.Sprintf("TheThingsStack/%s (%s/%s)", version.TTN, runtime.GOOS, runtime.GOARCH),
	}

	return &http.Client{
		Timeout:   defaultHTTPClientTimeout,
		Transport: rt,
	}, nil
}

// NewProvider constructs a Provider on top of the provided TLS configuration provider.
func NewProvider(tlsConfigProvider TLSClientConfigurationProvider) Provider {
	return &provider{tlsConfigProvider: tlsConfigProvider}
}

type roundTripperWithUserAgent struct {
	http.RoundTripper
	UserAgent string
}

func (rt *roundTripperWithUserAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", rt.UserAgent)
	}
	return rt.RoundTripper.RoundTrip(r)
}
