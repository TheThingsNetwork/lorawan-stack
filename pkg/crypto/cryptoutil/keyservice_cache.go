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
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
)

type cacheKeyServiceEntry struct {
	value any
	err   error
}

type cacheKeyService struct {
	crypto.KeyService
	cache gcache.Cache
}

// NewCacheKeyService returns a new crypto.KeyService that caches the results of Unwrap.
func NewCacheKeyService(inner crypto.KeyService, ttl time.Duration, size int) crypto.KeyService {
	builder := gcache.New(size).ARC()
	if ttl != 0 {
		builder = builder.Expiration(ttl)
	}
	return &cacheKeyService{
		KeyService: inner,
		cache:      builder.Build(),
	}
}

func (c *cacheKeyService) getOrLoad(
	ctx context.Context, cache crypto.CacheKey, key string, loaderFunc func() (any, error),
) (any, error) {
	cacheKey := fmt.Sprintf("%s:%s", cache, key)
	if val, err := c.cache.Get(cacheKey); err == nil {
		crypto.RegisterCacheHit(ctx, cache)
		entry := val.(*cacheKeyServiceEntry)
		return entry.value, entry.err
	}
	crypto.RegisterCacheMiss(ctx, cache)
	val, err := loaderFunc()
	c.cache.Set(cacheKey, &cacheKeyServiceEntry{val, err}) //nolint:errcheck
	return val, err
}

func (c *cacheKeyService) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	res, err := c.getOrLoad(ctx, crypto.CacheUnwrap, fmt.Sprintf("%s:%X", kekLabel, ciphertext),
		func() (any, error) {
			return c.KeyService.Unwrap(ctx, ciphertext, kekLabel)
		},
	)
	return res.([]byte), err
}
