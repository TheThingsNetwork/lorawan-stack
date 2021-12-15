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
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

// TLSClientConfigurationProvider provides a *tls.Config to be used by TLS clients.
type TLSClientConfigurationProvider interface {
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
}

// Provider constructs *http.Clients.
type Provider interface {
	HTTPClient(context.Context, ...Option) (*http.Client, error)
}

type provider struct {
	tlsConfigProvider TLSClientConfigurationProvider
}

// NewProvider constructs a Provider on top of the provided TLS configuration provider.
func NewProvider(tlsConfigProvider TLSClientConfigurationProvider) Provider {
	return &provider{tlsConfigProvider: tlsConfigProvider}
}

// defaultHTTPClientTimeout is the default timeout for the HTTP client.
const defaultHTTPClientTimeout = 10 * time.Second

type httpClientOptions struct {
	transportOptions []TransportOption
}

// Option is an option for HTTP clients.
type Option func(*httpClientOptions)

// WithTransportOptions constructs a transport with the provided options.
func WithTransportOptions(opts ...TransportOption) Option {
	return Option(func(o *httpClientOptions) {
		o.transportOptions = opts
	})
}

// HTTPClient returns a new *http.Client with a default timeout and a configured transport.
func (p *provider) HTTPClient(ctx context.Context, opts ...Option) (*http.Client, error) {
	options := &httpClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	tr, err := p.HTTPTransport(ctx, options.transportOptions...)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   defaultHTTPClientTimeout,
		Transport: tr,
	}, nil
}

type httpTransportOptions struct {
	cache            bool
	tlsConfigOptions []tlsconfig.Option
}

// TransportOption is an option for HTTP transports.
type TransportOption func(*httpTransportOptions)

// WithCache enables caching at transport level.
func WithCache(b bool) TransportOption {
	return TransportOption(func(o *httpTransportOptions) {
		o.cache = b
	})
}

// WithTLSConfigurationOptions provides the given tlsconfig.ConfigOption to the underlying TLS configuration provider.
func WithTLSConfigurationOptions(opts ...tlsconfig.Option) TransportOption {
	return TransportOption(func(o *httpTransportOptions) {
		o.tlsConfigOptions = opts
	})
}

// HTTPTransport returns a new http.RoundTripper with TLS client configuration.
func (p *provider) HTTPTransport(ctx context.Context, opts ...TransportOption) (http.RoundTripper, error) {
	options := &httpTransportOptions{}
	for _, opt := range opts {
		opt(options)
	}

	tlsConfig, err := p.tlsConfigProvider.GetTLSClientConfig(ctx, options.tlsConfigOptions...)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig

	rt := http.RoundTripper(transport)
	if options.cache {
		rt = &httpcache.Transport{
			Transport:           rt,
			Cache:               httpcache.NewMemoryCache(),
			MarkCachedResponses: true,
		}
	}

	return &roundTripperWithUserAgent{
		RoundTripper: rt,
		UserAgent:    fmt.Sprintf("TheThingsStack/%s (%s/%s)", version.TTN, runtime.GOOS, runtime.GOARCH),
	}, nil
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
