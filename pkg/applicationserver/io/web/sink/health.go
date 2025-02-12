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

package sink

import (
	"context"
	"net/http"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HealthStatusRegistry is a registry for webhook health status.
type HealthStatusRegistry interface {
	Get(context.Context) (*ttnpb.ApplicationWebhookHealth, error)
	Set(context.Context, func(*ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error)) error
}

type healthCheckSink struct {
	sink     Sink
	registry HealthStatusRegistry

	unhealthyAttemptsThreshold int
	unhealthyRetryInterval     time.Duration
}

// Process runs the health checks and sends the request to the underlying sink
// if they pass.
func (hcs *healthCheckSink) Process(req *http.Request) error {
	ctx := req.Context()
	lastKnownState, err := hcs.preRunCheck(ctx)
	if err != nil {
		registerWebhookFailed(ctx, err, true)
		return err
	}
	return hcs.executeAndRecord(ctx, req, lastKnownState)
}

type healthState int

const (
	healthStateUnknown healthState = iota
	healthStateHealthy
	healthStateMonitorSkipRecord
	healthStateMonitorRecord
	healthStateUnhealthy
)

var errWebhookDisabled = errors.DefineAborted(
	"webhook_disabled", "webhook {webhook_id} of application {application_id} is disabled",
)

// preRunCheck verifies if the webhook should be executed.
func (hcs *healthCheckSink) preRunCheck(ctx context.Context) (healthState, error) {
	h, err := hcs.registry.Get(ctx)
	if err != nil {
		return healthStateUnknown, err
	}

	switch {
	case h == nil, h.Status == nil:
		return healthStateUnknown, nil

	case h.GetHealthy() != nil:
		return healthStateHealthy, nil

	case h.GetUnhealthy() != nil:
		h := h.GetUnhealthy()
		monitorOnly := hcs.unhealthyAttemptsThreshold <= 0 || hcs.unhealthyRetryInterval <= 0
		nextAttemptAt := ttnpb.StdTime(h.LastFailedAttemptAt).Add(hcs.unhealthyRetryInterval)
		retryIntervalPassed := time.Now().After(nextAttemptAt)
		switch {
		case monitorOnly:
			// The system only monitors the health status, but does not block execution.
			if retryIntervalPassed {
				return healthStateMonitorRecord, nil
			}
			return healthStateMonitorSkipRecord, nil

		case h.FailedAttempts < uint64(hcs.unhealthyAttemptsThreshold):
			// The webhook is unhealthy but it has not failed enough times to be disabled yet.
			// This comparison is racing, as we may allow multiple webhooks at a time to execute
			// under the assumption that we are still under the threshold. However, serializing the
			// execution of unhealthy webhooks is considered costly, so we allow the race to occur.
			return healthStateUnhealthy, nil

		case h.FailedAttempts >= uint64(hcs.unhealthyAttemptsThreshold) && retryIntervalPassed:
			// The webhook is above the threshold, but the cooldown period has elapsed.
			return healthStateUnhealthy, nil

		default:
			// The webhook is above the threshold, and the cooldown period has not passed yet.
			return healthStateUnhealthy, errWebhookDisabled.WithAttributes(
				"application_id", internal.WebhookIDFromContext(ctx).ApplicationIds.ApplicationId,
				"webhook_id", internal.WebhookIDFromContext(ctx).WebhookId,
			)
		}

	default:
		panic("unreachable")
	}
}

// executeAndRecord runs the provided request using the underlying sink and records the health status.
func (hcs *healthCheckSink) executeAndRecord(
	ctx context.Context, req *http.Request, lastKnownState healthState,
) error {
	sinkErr := hcs.sink.Process(req)

	// Fast path 1: the health status is available, the request did not error, and the webhook is healthy.
	if sinkErr == nil && lastKnownState == healthStateHealthy {
		return nil
	}

	// Fast path 2: the health status is available, the request did error, and the webhook is unhealthy.
	if sinkErr != nil && lastKnownState == healthStateMonitorSkipRecord {
		return sinkErr
	}

	// Slow path: the request did error, or the webhook is unhealthy.
	var details *ttnpb.ErrorDetails
	if sinkErr != nil {
		if ttnErr, ok := errors.From(sinkErr); ok {
			details = ttnpb.ErrorDetailsToProto(ttnErr)
		}
	}
	f := func(h *ttnpb.ApplicationWebhookHealth) (*ttnpb.ApplicationWebhookHealth, error) {
		if sinkErr == nil {
			return &ttnpb.ApplicationWebhookHealth{
				Status: &ttnpb.ApplicationWebhookHealth_Healthy{
					Healthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy{},
				},
			}, nil
		}
		return &ttnpb.ApplicationWebhookHealth{
			Status: &ttnpb.ApplicationWebhookHealth_Unhealthy{
				Unhealthy: &ttnpb.ApplicationWebhookHealth_WebhookHealthStatusUnhealthy{
					FailedAttempts:           h.GetUnhealthy().GetFailedAttempts() + 1,
					LastFailedAttemptAt:      timestamppb.Now(),
					LastFailedAttemptDetails: details,
				},
			},
		}, nil
	}
	if err := hcs.registry.Set(ctx, f); err != nil {
		return err
	}
	return sinkErr
}

// NewHealthCheckSink creates a Sink that records the health status of the webhooks and stops them from executing if
// too many fail in a specified interval of time.
func NewHealthCheckSink(
	sink Sink, registry HealthStatusRegistry, unhealthyAttemptsThreshold int, unhealthyRetryInterval time.Duration,
) Sink {
	return &healthCheckSink{
		sink:                       sink,
		registry:                   registry,
		unhealthyAttemptsThreshold: unhealthyAttemptsThreshold,
		unhealthyRetryInterval:     unhealthyRetryInterval,
	}
}
