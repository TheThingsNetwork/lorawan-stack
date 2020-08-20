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

package oauthclient

import (
	echo "github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// Option is an OAuth2Client configuration option.
type Option func(*OAuthClient)

// OAuth2ConfigProvider provides an OAuth2 client config based on the context.
type OAuth2ConfigProvider func(echo.Context) *oauth2.Config

// WithOAuth2ConfigProvider overrides the default OAuth2 configuration provider.
func WithOAuth2ConfigProvider(provider OAuth2ConfigProvider) Option {
	return func(o *OAuthClient) {
		o.config.customProvider = true
		o.oauth = provider
	}
}

// WithNextKey overrides the default query parameter used for callback return.
func WithNextKey(key string) Option {
	return func(o *OAuthClient) {
		o.nextKey = key
	}
}

// Callback occurs after the OAuth2 token exchange has been performed successfully.
type Callback func(echo.Context, *oauth2.Token, string) error

// WithCallback adds a callback to be executed at the end of the OAuth2
// token exchange.
func WithCallback(cb Callback) Option {
	return func(o *OAuthClient) {
		o.callback = cb
	}
}
