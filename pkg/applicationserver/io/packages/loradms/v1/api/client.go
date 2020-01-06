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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"go.thethings.network/lorawan-stack/pkg/version"
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
	token string
	cl    *http.Client

	Tokens  *Tokens
	Uplinks *Uplinks
}

const (
	baseURL     = "https://dms.loracloud.com/api/v1"
	contentType = "application/json"
)

var (
	userAgent     = "ttn-lw-application-server/" + version.TTN
	parsedBaseURL *url.URL
)

type queryParam struct {
	key, value string
}

func (c *Client) newRequest(method, category, entity, operation string, body io.Reader, queryParams ...queryParam) (*http.Request, error) {
	u := cloneURL(parsedBaseURL)
	u.Path = path.Join(u.Path, category, entity, operation)
	q := u.Query()
	for _, p := range queryParams {
		q.Add(p.key, p.value)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
	return req, nil
}

// Do executes a new HTTP request with the given parameters and body and returns the response.
func (c *Client) Do(method, category, entity, operation string, body io.Reader, queryParams ...queryParam) (*http.Response, error) {
	req, err := c.newRequest(method, category, entity, operation, body, queryParams...)
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

// New creates a new Client with the given options.
func New(cl *http.Client, opts ...Option) (*Client, error) {
	client := &Client{
		cl: cl,
	}
	client.Tokens = &Tokens{client}
	client.Uplinks = &Uplinks{client}
	for _, opt := range opts {
		opt.apply(client)
	}
	return client, nil
}

// cloneURL deep-clones a url.URL.
// Based on $GOROOT/src/net/http/clone.go.
func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}

func init() {
	var err error
	parsedBaseURL, err = url.Parse(baseURL)
	if err != nil {
		panic(fmt.Sprintf("loradms: failed to parse base URL: %v", err))
	}
}
