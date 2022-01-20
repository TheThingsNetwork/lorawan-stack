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

package web

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var evtWebhookFail = events.Define(
	"as.webhook.fail", "fail to send webhook",
	events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
	events.WithErrorDataType(),
)

const (
	subsystem = "as_webhook"
	unknown   = "unknown"
)

var webhookMetrics = &messageMetrics{
	webhooksSent: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "sent_total",
			Help:      "Total number of sent webhooks",
		},
		[]string{},
	),
	webhooksFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "failed_total",
			Help:      "Total number of failed webhooks",
		},
		[]string{"error"},
	),
}

func init() {
	metrics.MustRegister(webhookMetrics)
}

type messageMetrics struct {
	webhooksSent   *metrics.ContextualCounterVec
	webhooksFailed *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.webhooksSent.Describe(ch)
	m.webhooksFailed.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.webhooksSent.Collect(ch)
	m.webhooksFailed.Collect(ch)
}

func registerWebhookSent(ctx context.Context) {
	webhookMetrics.webhooksSent.WithLabelValues(ctx).Inc()
}

func registerWebhookFailed(ctx context.Context, err error) {
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	webhookMetrics.webhooksFailed.WithLabelValues(ctx, errorLabel).Inc()
	ids := deviceIDFromContext(ctx)
	events.Publish(evtWebhookFail.NewWithIdentifiersAndData(ctx, ids, err))
}
