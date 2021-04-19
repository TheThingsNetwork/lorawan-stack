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

package udp

import (
	"context"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type gsWithIdentifiersCache struct {
	io.Server
	cache gcache.Cache
}

type fillGatewayContextResult struct {
	ctx context.Context
	ids ttnpb.GatewayIdentifiers
	err error
}

// FillGatewayContext implements the io.Server interface, and caches the identifiers fetched by the Entity Registry.
func (s *gsWithIdentifiersCache) FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error) {
	if cached, err := s.cache.Get(ids); err == nil {
		if result, ok := cached.(*fillGatewayContextResult); ok {
			return result.ctx, result.ids, result.err
		}
	}
	result := &fillGatewayContextResult{}
	result.ctx, result.ids, result.err = s.Server.FillGatewayContext(ctx, ids)
	if err := s.cache.Set(ids, result); err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to cache gateway identifiers")
	}
	return result.ctx, result.ids, result.err
}

func newServerWithIdentifiersCache(server io.Server, cache gcache.Cache) io.Server {
	if cache == nil {
		return server
	}
	return &gsWithIdentifiersCache{
		Server: server,
		cache:  cache,
	}
}
