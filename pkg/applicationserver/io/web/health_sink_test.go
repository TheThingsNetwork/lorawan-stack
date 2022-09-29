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

package web_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var registeredWebhookIDs = &ttnpb.ApplicationWebhookIdentifiers{
	ApplicationIds: registeredApplicationID,
	WebhookId:      registeredWebhookID,
}

func TestHealthStatusRegistry(t *testing.T) {
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

	_, err := webhookRegistry.Set(ctx, registeredWebhookIDs, nil, func(wh *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
		return &ttnpb.ApplicationWebhook{
			Ids:     registeredWebhookIDs,
			BaseUrl: "http://example.com",
			Format:  "json",
		}, []string{"ids", "base_url", "format"}, nil
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	r, err := http.NewRequestWithContext(web.WithWebhookID(ctx, registeredWebhookIDs), http.MethodPost, "http://foo.bar", nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	registry := web.NewHealthStatusRegistry(webhookRegistry)

	// Initially no status is stored.
	health, err := registry.Get(r)
	a.So(err, should.BeNil)
	a.So(health, should.BeNil)

	// Callback errors are propagated.
	err = registry.Set(r, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		a.So(health, should.BeNil)
		return nil, fmt.Errorf("internal failure")
	})
	a.So(err, should.NotBeNil)

	// Store the health status.
	err = registry.Set(r, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		a.So(health, should.BeNil)
		return &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Healthy{
				Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
			},
		}, nil
	})
	a.So(err, should.BeNil)

	// Get is consistent with Set.
	health, err = registry.Get(r)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})

	// Store an unhealthy status.
	errorDetails := errors.Define("failure", "random failure").New()
	err = registry.Set(r, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
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
	health, err = registry.Get(r)
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
	health, err = cachedRegistry.Get(r)
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
	r = r.WithContext(web.WithCachedHealthStatus(r.Context(), &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	}))
	// Expect the status to take precedence to the underlying registry.
	health, err = cachedRegistry.Get(r)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})

	// Update stored value.
	err = registry.Set(r, func(health *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
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
	health, err = cachedRegistry.Get(r)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})
}

func TestHealthCheckSink(t *testing.T) {
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

	_, err := webhookRegistry.Set(ctx, registeredWebhookIDs, nil, func(wh *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
		return &ttnpb.ApplicationWebhook{
			Ids:     registeredWebhookIDs,
			BaseUrl: "http://example.com",
			Format:  "json",
		}, []string{"ids", "base_url", "format"}, nil
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	sink := &mockSink{
		ch: make(chan *http.Request, 1),
	}

	registry := web.NewHealthStatusRegistry(webhookRegistry)
	healthSink := web.NewHealthCheckSink(sink, registry, 4, 8*Timeout)

	r, err := http.NewRequestWithContext(web.WithWebhookID(ctx, registeredWebhookIDs), http.MethodPost, "http://foo.bar", nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	// No processing error - the webhook should move to healthy.
	err = healthSink.Process(r)
	select {
	case <-sink.ch:
	case <-time.After(Timeout):
		t.Fatal("expected request")
	}

	// The stored status is now healthy.
	health, err := registry.Get(r)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})

	// The sink starts erroring out.
	requestErr := errors.DefineUnavailable("request", "request").New()
	sink.err = requestErr

	var lastFailedAttemptAt time.Time
	for i := 1; i <= 4; i++ {
		now := time.Now()
		err = healthSink.Process(r)
		select {
		case <-sink.ch:
		case <-time.After(Timeout):
			t.Fatal("expected request")
		}

		// The errors should be recorded.
		health, err := registry.Get(r)
		a.So(err, should.BeNil)
		if unhealthy := health.GetUnhealthy(); a.So(unhealthy, should.NotBeNil) {
			lastFailedAttemptAt = *ttnpb.StdTime(unhealthy.LastFailedAttemptAt)

			a.So(unhealthy.FailedAttempts, should.Equal, i)
			a.So(lastFailedAttemptAt, should.HappenBetween, now.Add(-time.Second), now.Add(time.Second))
			a.So(unhealthy.LastFailedAttemptDetails, should.Resemble, ttnpb.ErrorDetailsToProto(requestErr))
		}
	}

	// The sink should now no longer receive the messages.
	for i := 1; i <= 4; i++ {
		// The request should not reach the underlying sink.
		err = healthSink.Process(r)
		select {
		case <-sink.ch:
			t.Fatal("unexpected request")
		case <-time.After(Timeout):
		}

		// The number of attempts should stay the same.
		health, err := registry.Get(r)
		a.So(err, should.BeNil)
		if unhealthy := health.GetUnhealthy(); a.So(unhealthy, should.NotBeNil) {
			a.So(unhealthy.FailedAttempts, should.Equal, 4)
			a.So(*ttnpb.StdTime(unhealthy.LastFailedAttemptAt), should.Equal, lastFailedAttemptAt)
			a.So(unhealthy.LastFailedAttemptDetails, should.Resemble, ttnpb.ErrorDetailsToProto(requestErr))
		}
	}

	// We wait for the cooldown period to pass.
	time.Sleep(8 * Timeout)

	// The sink should now do one attempt, and fail the rest.
	for i := 1; i <= 4; i++ {
		now := time.Now()
		err = healthSink.Process(r)
		if i == 1 {
			select {
			case <-sink.ch:
			case <-time.After(Timeout):
				t.Fatal("expected request")
			}
		} else {
			select {
			case <-sink.ch:
				t.Fatal("unexpected request")
			case <-time.After(Timeout):
			}
		}

		// The errors should be recorded.
		health, err := registry.Get(r)
		a.So(err, should.BeNil)
		if unhealthy := health.GetUnhealthy(); a.So(unhealthy, should.NotBeNil) {
			if i == 1 {
				lastFailedAttemptAt = *ttnpb.StdTime(unhealthy.LastFailedAttemptAt)
				a.So(lastFailedAttemptAt, should.HappenBetween, now.Add(-time.Second), now.Add(time.Second))
			} else {
				a.So(*ttnpb.StdTime(unhealthy.LastFailedAttemptAt), should.Equal, lastFailedAttemptAt)
			}

			a.So(unhealthy.FailedAttempts, should.Equal, 5)
			a.So(unhealthy.LastFailedAttemptDetails, should.Resemble, ttnpb.ErrorDetailsToProto(requestErr))
		}
	}

	// We wait for the cooldown period to pass.
	time.Sleep(8 * Timeout)

	// We reset the error and expect the health status to recover.
	sink.err = nil

	// No processing error - the webhook should move to healthy.
	for i := 1; i <= 4; i++ {
		err = healthSink.Process(r)
		select {
		case <-sink.ch:
		case <-time.After(Timeout):
			t.Fatal("expected request")
		}

		// The stored status is healthy.
		health, err = registry.Get(r)
		a.So(err, should.BeNil)
		a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Healthy{
				Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
			},
		})
	}
}
