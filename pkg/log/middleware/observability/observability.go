// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

// Package observability implements a pkg/log.Handler that exports metrics for the logged messages.
package observability

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

type logMessageMetrics struct {
	logMessages *metrics.ContextualCounterVec
}

const (
	subsystem = "log"
	level     = "level"
	namespace = "namespace"
)

func (m logMessageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.logMessages.Describe(ch)
}

func (m logMessageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.logMessages.Collect(ch)
}

var logMetrics = &logMessageMetrics{
	logMessages: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "log_messages_total",
			Help:      "Total number of logged messages",
		},
		[]string{level, namespace},
	),
}

func init() {
	metrics.MustRegister(logMetrics)
}

// observability is a log.Handler that tracks metrics for logged messages.
type observability struct{}

// New creates a new observability log middleware.
func New() log.Middleware {
	return &observability{}
}

// Wrap an existing log handler with observability.
func (o *observability) Wrap(next log.Handler) log.Handler {
	return log.HandlerFunc(func(entry log.Entry) error {
		namespace := "unknown"
		if ns, ok := entry.Fields().Fields()["namespace"]; ok {
			if ns, ok := ns.(string); ok {
				namespace = ns
			}
		}
		logMetrics.logMessages.WithLabelValues(context.Background(), entry.Level().String(), namespace).Inc()
		return next.HandleLog(entry)
	})
}
