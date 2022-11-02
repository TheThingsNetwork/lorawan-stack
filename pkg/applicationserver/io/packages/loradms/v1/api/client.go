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

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	urlutil "go.thethings.network/lorawan-stack/v3/pkg/util/url"
)

// Option is an option for the API client.
type Option interface {
	apply(*Client)
}

// OptionFunc is an Option implemented as a function.
type OptionFunc func(*Client)

func (f OptionFunc) apply(c *Client) { f(c) }

// Client is an API client for the LoRa Cloud Device Management v1 service.
type Client struct {
	token   string
	baseURL *url.URL
	cl      *http.Client

	Uplinks *Uplinks
}

const (
	contentType      = "application/json"
	defaultServerURL = "https://mgs.loracloud.com"
	basePath         = "/api/v1"
)

// DefaultServerURL is the default server URL for LoRa Cloud Device Management v1.
var DefaultServerURL = func() *url.URL {
	parsed, err := url.Parse(defaultServerURL)
	if err != nil {
		panic(fmt.Sprintf("loradms: failed to parse base URL: %v", err))
	}
	return parsed
}()

func (c *Client) newRequest(ctx context.Context, method, category, entity, operation string, body io.Reader) (*http.Request, error) {
	u := urlutil.CloneURL(c.baseURL)
	u.Path = path.Join(basePath, category, entity, operation)
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
	return req, nil
}

// Do executes a new HTTP request with the given parameters and body and returns the response.
func (c *Client) Do(ctx context.Context, method, category, entity, operation string, body interface{}) (*http.Response, error) {
	buffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"method", method,
		"category", category,
		"entity", entity,
		"operation", operation,
		"body", buffer.String(),
	)).Debug("Run DAS request")
	req, err := c.newRequest(ctx, method, category, entity, operation, buffer)
	if err != nil {
		return nil, err
	}
	return c.cl.Do(req)
}

// WithToken uses the given authentication token in the client.
func WithToken(token string) Option {
	return OptionFunc(func(c *Client) {
		c.token = token
	})
}

// WithBaseURL uses the given base URL for the requests of the client.
func WithBaseURL(baseURL *url.URL) Option {
	return OptionFunc(func(c *Client) {
		c.baseURL = baseURL
	})
}

// New creates a new Client with the given options.
func New(cl *http.Client, opts ...Option) (*Client, error) {
	client := &Client{
		cl:      cl,
		baseURL: urlutil.CloneURL(DefaultServerURL),
	}
	client.Uplinks = &Uplinks{client}
	for _, opt := range opts {
		opt.apply(client)
	}
	return client, nil
}
