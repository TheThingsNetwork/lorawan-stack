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

package web_test

import (
	"fmt"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHealthStatusRegistry(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)

	redisClient, flush := test.NewRedis(ctx, "web_test")
	defer flush()
	defer redisClient.Close()
	webhookRegistry := &redis.WebhookRegistry{
		Redis:   redisClient,
		LockTTL: test.Delay << 10,
	}
	if err := webhookRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	_, err := webhookRegistry.Set(
		ctx,
		registeredWebhookIDs,
		nil,
		func(*ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
			return &ttnpb.ApplicationWebhook{
				Ids:     registeredWebhookIDs,
				BaseUrl: "http://example.com",
				Format:  "json",
			}, []string{"ids", "base_url", "format"}, nil
		},
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ctx = internal.WithWebhookData(ctx, &internal.WebhookData{WebhookIDs: registeredWebhookIDs})
	registry := web.NewHealthStatusRegistry(webhookRegistry)

	// Initially no status is stored.
	health, err := registry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.BeNil)

	// Callback errors are propagated.
	err = registry.Set(ctx, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		a.So(health, should.BeNil)
		return nil, fmt.Errorf("internal failure")
	})
	a.So(err, should.NotBeNil)

	// Store the health status.
	err = registry.Set(ctx, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		a.So(health, should.BeNil)
		return &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Healthy{
				Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
			},
		}, nil
	})
	a.So(err, should.BeNil)

	// Get is consistent with Set.
	health, err = registry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})

	// Store an unhealthy status.
	errorDetails := errors.Define("failure", "random failure").New()
	err = registry.Set(ctx, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Healthy{
				Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
			},
		})
		return &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Unhealthy{
				Unhealthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusUnhealthy{
					FailedAttempts:           123,
					LastFailedAttemptAt:      timestamppb.New(time.Unix(123, 234)),
					LastFailedAttemptDetails: ttnpb.ErrorDetailsToProto(errorDetails),
				},
			},
		}, nil
	})
	a.So(err, should.BeNil)

	// Get is consistent with Set.
	health, err = registry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Unhealthy{
			Unhealthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusUnhealthy{
				FailedAttempts:           123,
				LastFailedAttemptAt:      timestamppb.New(time.Unix(123, 234)),
				LastFailedAttemptDetails: ttnpb.ErrorDetailsToProto(errorDetails),
			},
		},
	})

	cachedRegistry := web.NewCachedHealthStatusRegistry(registry)
	// Initially the cached registry just defers to the underlying registry.
	health, err = cachedRegistry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Unhealthy{
			Unhealthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusUnhealthy{
				FailedAttempts:           123,
				LastFailedAttemptAt:      timestamppb.New(time.Unix(123, 234)),
				LastFailedAttemptDetails: ttnpb.ErrorDetailsToProto(errorDetails),
			},
		},
	})

	// Cache a status inside the context.
	ctx = internal.WithWebhookData(ctx, &internal.WebhookData{
		WebhookIDs: registeredWebhookIDs,
		Health: &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Healthy{
				Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
			},
		},
	})
	// Expect the status to take precedence to the underlying registry.
	health, err = cachedRegistry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})

	// Update stored value.
	err = registry.Set(ctx, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Unhealthy{
				Unhealthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusUnhealthy{
					FailedAttempts:           123,
					LastFailedAttemptAt:      timestamppb.New(time.Unix(123, 234)),
					LastFailedAttemptDetails: ttnpb.ErrorDetailsToProto(errorDetails),
				},
			},
		})
		return &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Unhealthy{
				Unhealthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusUnhealthy{
					FailedAttempts:           234,
					LastFailedAttemptAt:      timestamppb.New(time.Unix(345, 567)),
					LastFailedAttemptDetails: ttnpb.ErrorDetailsToProto(errorDetails),
				},
			},
		}, nil
	})
	a.So(err, should.BeNil)

	// The cached value takes precedence to the update in the underlying registry.
	health, err = cachedRegistry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})
}
