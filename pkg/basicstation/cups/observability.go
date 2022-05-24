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

package cups

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

type messageMetrics struct {
	requestReceived  *metrics.ContextualCounterVec
	requestSucceeded *metrics.ContextualCounterVec
	requestFailed    *metrics.ContextualCounterVec
}

var (
	subsystem = "cups"
	request   = "request"
)

var cupsMetrics = &messageMetrics{
	requestReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "request_received_total",
			Help:      "Total number of requests received",
		},
		[]string{request},
	),
	requestSucceeded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "request_succeeded_total",
			Help:      "Total number of requests succeeded",
		},
		[]string{request},
	),
	requestFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "request_failed_total",
			Help:      "Total number of requests failed",
		},
		[]string{request, "error"},
	),
}

// Describe implements prometheus.Collector.
func (m *messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.requestReceived.Describe(ch)
	m.requestSucceeded.Describe(ch)
	m.requestFailed.Describe(ch)
}

// Collect implements prometheus.Collector.
func (m *messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.requestReceived.Collect(ch)
	m.requestSucceeded.Collect(ch)
	m.requestFailed.Collect(ch)
}

func registerUpdateInfoRequestReceived(ctx context.Context, request string) {
	cupsMetrics.requestReceived.WithLabelValues(ctx, request).Inc()
}

func registerUpdateInfoRequestSucceeded(ctx context.Context, request string) {
	cupsMetrics.requestSucceeded.WithLabelValues(ctx, request).Inc()
}

func registerUpdateInfoRequestFailed(ctx context.Context, request string, err error) {
	if ttnErr, ok := errors.From(err); ok {
		cupsMetrics.requestFailed.WithLabelValues(ctx, request, ttnErr.FullName()).Inc()
	} else {
		cupsMetrics.requestFailed.WithLabelValues(ctx, request, "unknown").Inc()
	}
}

func init() {
	metrics.MustRegister(cupsMetrics)
}
