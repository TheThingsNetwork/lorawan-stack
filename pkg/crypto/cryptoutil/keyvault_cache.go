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

package cryptoutil

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
)

type unwrapEntry struct {
	value []byte
	err   error
}

type cachedVault struct {
	crypto.KeyVault
	unwrapCache gcache.Cache
}

func NewCacheKeyVault(main crypto.KeyVault, ttl time.Duration, size int) crypto.KeyVault {
	builder := gcache.New(size).ARC()
	if ttl != 0 {
		builder = builder.Expiration(ttl)
	}
	return &cachedVault{
		KeyVault:    main,
		unwrapCache: builder.Build(),
	}
}

func unwrapCacheKey(ciphertext []byte, kekLabel string) string {
	return fmt.Sprintf("%v:%v", kekLabel, ciphertext)
}

func (c *cachedVault) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	id := unwrapCacheKey(ciphertext, kekLabel)
	if val, err := c.unwrapCache.Get(id); err == nil {
		crypto.RegisterCacheHit(ctx, "unwrap")
		v := val.(*unwrapEntry)
		return v.value, v.err
	}
	v := &unwrapEntry{}
	c.unwrapCache.Set(id, v)
	crypto.RegisterCacheMiss(ctx, "unwrap")
	v.value, v.err = c.KeyVault.Unwrap(ctx, ciphertext, kekLabel)
	return v.value, v.err
}
