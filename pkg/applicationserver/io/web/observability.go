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
	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

const (
	subsystem = "as_webhook"
	unknown   = "unknown"
)

var webhookMetrics = &messageMetrics{
	webhookQueue: metrics.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "queue_size",
			Help:      "Webhook queue size",
		},
	),
	webhooksSent: metrics.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "sent_total",
			Help:      "Total number of sent webhooks",
		},
	),
	webhooksFailed: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "failed_total",
			Help:      "Total number of failed webhooks",
		},
		[]string{"error"},
	),
}

func init() {
	webhookMetrics.webhookQueue.Set(0)
	webhookMetrics.webhooksSent.Add(0)
	metrics.MustRegister(webhookMetrics)
}

type messageMetrics struct {
	webhookQueue   prometheus.Gauge
	webhooksSent   prometheus.Counter
	webhooksFailed *prometheus.CounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.webhookQueue.Describe(ch)
	m.webhooksSent.Describe(ch)
	m.webhooksFailed.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.webhookQueue.Collect(ch)
	m.webhooksSent.Collect(ch)
	m.webhooksFailed.Collect(ch)
}

func registerWebhookQueued() {
	webhookMetrics.webhookQueue.Inc()
}

func registerWebhookDequeued() {
	webhookMetrics.webhookQueue.Dec()
}

func registerWebhookSent() {
	webhookMetrics.webhooksSent.Inc()
}

func registerWebhookFailed(err error) {
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	webhookMetrics.webhooksFailed.WithLabelValues(errorLabel).Inc()
}
