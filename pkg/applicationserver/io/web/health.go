// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package web

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type healthStatusRegistry struct {
	registry WebhookRegistry
}

// Get implements HealthStatusRegistry.
func (reg *healthStatusRegistry) Get(ctx context.Context) (*ttnpb.ApplicationWebhookHealth, error) {
	ids := internal.WebhookIDFromContext(ctx)
	web, err := reg.registry.Get(ctx, ids, []string{"health_status"})
	if err != nil {
		return nil, err
	}
	return web.HealthStatus, nil
}

// Set implements HealthStatusRegistry.
func (reg *healthStatusRegistry) Set(
	ctx context.Context, f func(*ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error),
) error {
	ids := internal.WebhookIDFromContext(ctx)
	_, err := reg.registry.Set(
		ctx,
		ids,
		[]string{"health_status"},
		func(wh *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
			if wh == nil {
				// The webhook has been deleted during execution.
				return nil, nil, nil
			}
			updated, err := f(wh.HealthStatus)
			if err != nil {
				return nil, nil, err
			}
			wh.HealthStatus = updated
			return wh, []string{"health_status"}, nil
		},
	)
	return err
}

// NewHealthStatusRegistry constructs a HealthStatusRegistry on top of the provided WebhookRegistry.
func NewHealthStatusRegistry(registry WebhookRegistry) sink.HealthStatusRegistry {
	return &healthStatusRegistry{registry}
}

type cachedHealthStatusRegistry struct {
	registry sink.HealthStatusRegistry
}

// Get implements HealthStatusRegistry.
func (reg *cachedHealthStatusRegistry) Get(ctx context.Context) (*ttnpb.ApplicationWebhookHealth, error) {
	if h, ok := internal.WebhookHealthFromContext(ctx); ok {
		return h, nil
	}
	return reg.registry.Get(ctx)
}

// Set implements HealthStatusRegistry.
func (reg *cachedHealthStatusRegistry) Set(
	ctx context.Context, f func(*ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error),
) error {
	return reg.registry.Set(ctx, f)
}

// NewCachedHealthStatusRegistry constructs a HealthStatusRegistry which allows the Get response to be cached.
func NewCachedHealthStatusRegistry(registry sink.HealthStatusRegistry) sink.HealthStatusRegistry {
	return &cachedHealthStatusRegistry{registry}
}
