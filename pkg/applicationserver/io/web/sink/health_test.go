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

package sink_test

import (
	"net/http"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHealthCheckSink(t *testing.T) { // nolint:gocyclo
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

	sinkCh := make(chan *http.Request, 1)
	target := mock.New(sinkCh)

	registry := web.NewHealthStatusRegistry(webhookRegistry)
	healthSink := sink.NewHealthCheckSink(target, registry, 4, 8*timeout)

	ctx = internal.WithWebhookData(ctx, &internal.WebhookData{
		EndDeviceIDs: registeredDeviceID,
		WebhookIDs:   registeredWebhookIDs,
	})
	r, err := http.NewRequestWithContext(
		ctx, http.MethodPost, "http://foo.bar", nil,
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	// No processing error - the webhook should move to healthy.
	err = healthSink.Process(r)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	select {
	case <-sinkCh:
	case <-time.After(timeout):
		t.Fatal("expected request")
	}

	// The stored status is now healthy.
	health, err := registry.Get(ctx)
	a.So(err, should.BeNil)
	a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
		Status: &ttnpb.ApplicationWebhookHealth_Healthy{
			Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
		},
	})

	// The sink starts erroring out.
	requestErr := errors.DefineUnavailable("request", "request").New()
	target.SetError(requestErr)

	var lastFailedAttemptAt time.Time
	for i := 1; i <= 4; i++ {
		now := time.Now()
		err = healthSink.Process(r)
		if !a.So(err, should.HaveSameErrorDefinitionAs, requestErr) {
			t.FailNow()
		}
		select {
		case <-sinkCh:
		case <-time.After(timeout):
			t.Fatal("expected request")
		}

		// The errors should be recorded.
		health, err := registry.Get(ctx)
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
		if !a.So(err, should.NotBeNil) {
			t.FailNow()
		}
		select {
		case <-sinkCh:
			t.Fatal("unexpected request")
		case <-time.After(timeout):
		}

		// The number of attempts should stay the same.
		health, err := registry.Get(ctx)
		a.So(err, should.BeNil)
		if unhealthy := health.GetUnhealthy(); a.So(unhealthy, should.NotBeNil) {
			a.So(unhealthy.FailedAttempts, should.Equal, 4)
			a.So(*ttnpb.StdTime(unhealthy.LastFailedAttemptAt), should.Equal, lastFailedAttemptAt)
			a.So(unhealthy.LastFailedAttemptDetails, should.Resemble, ttnpb.ErrorDetailsToProto(requestErr))
		}
	}

	// We wait for the cooldown period to pass.
	time.Sleep(8 * timeout)

	// The sink should now do one attempt, and fail the rest.
	for i := 1; i <= 4; i++ {
		now := time.Now()
		err = healthSink.Process(r)
		if !a.So(err, should.NotBeNil) {
			t.FailNow()
		}
		if i == 1 {
			select {
			case <-sinkCh:
			case <-time.After(timeout):
				t.Fatal("expected request")
			}
		} else {
			select {
			case <-sinkCh:
				t.Fatal("unexpected request")
			case <-time.After(timeout):
			}
		}

		// The errors should be recorded.
		health, err := registry.Get(ctx)
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
	time.Sleep(8 * timeout)

	// We reset the error and expect the health status to recover.
	target.SetError(nil)

	// No processing error - the webhook should move to healthy.
	for i := 1; i <= 4; i++ {
		err = healthSink.Process(r)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		select {
		case <-sinkCh:
		case <-time.After(timeout):
			t.Fatal("expected request")
		}

		// The stored status is healthy.
		health, err = registry.Get(ctx)
		a.So(err, should.BeNil)
		a.So(health, should.Resemble, &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Healthy{
				Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
			},
		})
	}
}
