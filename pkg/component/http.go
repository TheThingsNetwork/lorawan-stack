// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package component

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gregjones/httpcache"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

// defaultHTTPClientTimeout is the default timeout for the HTTP client.
const defaultHTTPClientTimeout = 10 * time.Second

type httpClientOptions struct {
	transportOptions []HTTPTransportOption
}

// HTTPClientOption is an option for HTTP clients.
type HTTPClientOption func(*httpClientOptions)

// WithTransportOptions constructs a transport with the provided options.
func WithTransportOptions(opts ...HTTPTransportOption) HTTPClientOption {
	return HTTPClientOption(func(o *httpClientOptions) {
		o.transportOptions = opts
	})
}

// HTTPClient returns a new *http.Client with a default timeout and a configured transport.
func (c *Component) HTTPClient(ctx context.Context, opts ...HTTPClientOption) (*http.Client, error) {
	options := &httpClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	tr, err := c.HTTPTransport(ctx, options.transportOptions...)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   defaultHTTPClientTimeout,
		Transport: tr,
	}, nil
}

type httpTransportOptions struct {
	cache bool
}

// HTTPTransportOption is an option for HTTP transports.
type HTTPTransportOption func(*httpTransportOptions)

// WithCache enables caching at transport level.
func WithCache(b bool) HTTPTransportOption {
	return HTTPTransportOption(func(o *httpTransportOptions) {
		o.cache = b
	})
}

// HTTPTransport returns a new http.RoundTripper with TLS client configuration.
func (c *Component) HTTPTransport(ctx context.Context, opts ...HTTPTransportOption) (http.RoundTripper, error) {
	tlsConfig, err := c.GetTLSClientConfig(ctx)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig

	options := &httpTransportOptions{}
	for _, opt := range opts {
		opt(options)
	}

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
