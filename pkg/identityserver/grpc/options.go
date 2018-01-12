// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

// Option is the type that defines a Client option.
type Option func(*Client)

// WithCache sets the given cache to the Client.
func WithCache(cache Cache) Option {
	return func(c *Client) {
		c.cache = cache
	}
}
