// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package cryptoutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
)

type cacheKeyCacheEntry struct {
	value any
	err   error
}

// CacheKeyVaultClock provides a time source.
type CacheKeyVaultClock interface {
	Now() time.Time
}

// CacheKeyVaultClockFunc implements CacheKeyVaultClock.
type CacheKeyVaultClockFunc func() time.Time

// Now implements CacheKeyVaultClock.
func (f CacheKeyVaultClockFunc) Now() time.Time {
	return f()
}

type cacheKeyVault struct {
	inner crypto.KeyVault
	cache gcache.Cache
	clock CacheKeyVaultClock
}

type cacheKeyVaultOptions struct {
	ttl   time.Duration
	size  int
	clock CacheKeyVaultClock
}

// CacheKeyVaultOption configures CacheKeyVault.
type CacheKeyVaultOption interface {
	apply(*cacheKeyVaultOptions)
}

type cacheKeyVaultOptionFunc func(*cacheKeyVaultOptions)

func (f cacheKeyVaultOptionFunc) apply(opts *cacheKeyVaultOptions) {
	f(opts)
}

// WithCacheKeyVaultSize configures the size of the cache.
func WithCacheKeyVaultSize(size int) CacheKeyVaultOption {
	return cacheKeyVaultOptionFunc(func(opts *cacheKeyVaultOptions) {
		opts.size = size
	})
}

// WithCacheKeyVaultTTL configures the time-to-live of the cache. If 0, no expiry is used.
func WithCacheKeyVaultTTL(ttl time.Duration) CacheKeyVaultOption {
	return cacheKeyVaultOptionFunc(func(opts *cacheKeyVaultOptions) {
		opts.ttl = ttl
	})
}

// WithCacheKeyVaultClock configures a time source.
// This is useful for testing.
func WithCacheKeyVaultClock(clock CacheKeyVaultClock) CacheKeyVaultOption {
	return cacheKeyVaultOptionFunc(func(opts *cacheKeyVaultOptions) {
		opts.clock = clock
	})
}

// NewCacheKeyVault returns a new crypto.KeyVault that caches the keys in memory.
// Certificates are cached for the duration of their validity minus one hour, maximed by the given time-to-live.
func NewCacheKeyVault(inner crypto.KeyVault, opts ...CacheKeyVaultOption) crypto.KeyVault {
	options := &cacheKeyVaultOptions{
		size:  1000,
		clock: CacheKeyVaultClockFunc(time.Now),
	}
	for _, opt := range opts {
		opt.apply(options)
	}
	builder := gcache.New(options.size).ARC()
	if options.ttl != 0 {
		builder = builder.Expiration(options.ttl)
	}
	if options.clock != nil {
		builder = builder.Clock(options.clock)
	}
	return &cacheKeyVault{
		inner: inner,
		cache: builder.Build(),
		clock: options.clock,
	}
}

func (c *cacheKeyVault) getOrLoad(
	ctx context.Context, cache crypto.CacheKey, label string, loaderFn func() (any, time.Duration, error),
) (any, error) {
	cacheKey := fmt.Sprintf("%s:%s", cache, label)
	val, err := c.cache.Get(cacheKey)
	if err == nil {
		crypto.RegisterCacheHit(ctx, cache)
		entry := val.(*cacheKeyCacheEntry)
		return entry.value, entry.err
	}
	crypto.RegisterCacheMiss(ctx, cache)
	val, ttl, err := loaderFn()
	if ttl != 0 {
		c.cache.SetWithExpire(cacheKey, &cacheKeyCacheEntry{val, err}, ttl) //nolint:errcheck
	} else {
		c.cache.Set(cacheKey, &cacheKeyCacheEntry{val, err}) //nolint:errcheck
	}
	return val, err
}

// Key implements crypto.KeyVault.
func (c *cacheKeyVault) Key(ctx context.Context, label string) ([]byte, error) {
	val, err := c.getOrLoad(ctx, crypto.CacheEncryptionKey, label, func() (any, time.Duration, error) {
		val, err := c.inner.Key(ctx, label)
		return val, 0, err
	})
	return val.([]byte), err
}

func (c *cacheKeyVault) cacheCertificateTTL(crt tls.Certificate) (time.Duration, bool) {
	if len(crt.Certificate) > 0 {
		cert, err := x509.ParseCertificate(crt.Certificate[0])
		if err == nil {
			return cert.NotAfter.Sub(c.clock.Now()) - time.Hour, true
		}
	}
	return 0, false
}

// Certificate implements crypto.KeyVault.
func (c *cacheKeyVault) ServerCertificate(ctx context.Context, label string) (tls.Certificate, error) {
	val, err := c.getOrLoad(ctx, crypto.CacheServerCertificate, label, func() (any, time.Duration, error) {
		val, err := c.inner.ServerCertificate(ctx, label)
		ttl := time.Duration(0)
		if err == nil {
			if certTTL, ok := c.cacheCertificateTTL(val); ok {
				ttl = certTTL
			}
		}
		return val, ttl, err
	})
	return val.(tls.Certificate), err
}

// ClientCertificate implements crypto.KeyVault.
func (c *cacheKeyVault) ClientCertificate(ctx context.Context) (tls.Certificate, error) {
	val, err := c.getOrLoad(ctx, crypto.CacheClientCertificate, "", func() (any, time.Duration, error) {
		val, err := c.inner.ClientCertificate(ctx)
		ttl := time.Duration(0)
		if err == nil {
			if certTTL, ok := c.cacheCertificateTTL(val); ok {
				ttl = certTTL
			}
		}
		return val, ttl, err
	})
	return val.(tls.Certificate), err
}
