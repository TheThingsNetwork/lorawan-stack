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

package webmiddleware

import "github.com/gorilla/handlers"

// CORSConfig is the configuration for the CORS middleware.
type CORSConfig struct {
	AllowedHeaders   []string
	AllowedMethods   []string
	AllowedOrigins   []string
	ExposedHeaders   []string
	MaxAge           int
	AllowCredentials bool
}

func (c CORSConfig) options() []handlers.CORSOption {
	var options []handlers.CORSOption
	if len(c.AllowedHeaders) > 0 {
		options = append(options, handlers.AllowedHeaders(c.AllowedHeaders))
	}
	if len(c.AllowedMethods) > 0 {
		options = append(options, handlers.AllowedMethods(c.AllowedMethods))
	}
	if len(c.AllowedOrigins) > 0 {
		options = append(options, handlers.AllowedOrigins(c.AllowedOrigins))
	}
	if len(c.ExposedHeaders) > 0 {
		options = append(options, handlers.ExposedHeaders(c.ExposedHeaders))
	}
	if c.MaxAge > 0 {
		options = append(options, handlers.MaxAge(c.MaxAge))
	}
	if c.AllowCredentials {
		options = append(options, handlers.AllowCredentials())
	}
	return options
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
func CORS(config CORSConfig) MiddlewareFunc {
	return MiddlewareFunc(handlers.CORS(config.options()...))
}
