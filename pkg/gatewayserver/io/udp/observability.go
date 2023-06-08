// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
)

const subsystem = "gs_io_udp"

var udpMetrics = &messageMetrics{
	messageReceived: prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "message_received_total",
			Help:      "Total number of received UDP messages",
		},
	),
	messageForwarded: prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "message_forwarded_total",
			Help:      "Total number of forwarded UDP messages",
		},
		[]string{"type"},
	),
	messageDropped: prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "message_dropped_total",
			Help:      "Total number of dropped UDP messages",
		},
		[]string{"error"},
	),

	unmarshalTypeErrors: prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "unmarshal_type_errors_total",
			Help:      "Total number of unmarshal type errors",
		},
		[]string{"type", "struct", "field"},
	),
}

func init() {
	metrics.MustRegister(udpMetrics)
}

type messageMetrics struct {
	messageReceived  prometheus.Counter
	messageForwarded *prometheus.CounterVec
	messageDropped   *prometheus.CounterVec

	unmarshalTypeErrors *prometheus.CounterVec
}

// Describe implements prometheus.Collector.
func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.messageReceived.Describe(ch)
	m.messageForwarded.Describe(ch)
	m.messageDropped.Describe(ch)

	m.unmarshalTypeErrors.Describe(ch)
}

// Collect implements prometheus.Collector.
func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.messageReceived.Collect(ch)
	m.messageForwarded.Collect(ch)
	m.messageDropped.Collect(ch)

	m.unmarshalTypeErrors.Collect(ch)
}

func registerMessageReceived(_ context.Context) {
	udpMetrics.messageReceived.Inc()
}

func registerMessageForwarded(_ context.Context, tp encoding.PacketType) {
	udpMetrics.messageForwarded.WithLabelValues(tp.String()).Inc()
}

func registerMessageDropped(_ context.Context, err error) {
	errorLabel := "unknown"
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	} else if jsonErr := (&json.SyntaxError{}); errors.As(err, &jsonErr) {
		errorLabel = "encoding/json:syntax"
	} else if jsonErr := (&json.UnmarshalTypeError{}); errors.As(err, &jsonErr) {
		errorLabel = "encoding/json:unmarshal_type"
		udpMetrics.unmarshalTypeErrors.WithLabelValues(jsonErr.Type.Name(), jsonErr.Struct, jsonErr.Field).Inc()
	}
	udpMetrics.messageDropped.WithLabelValues(errorLabel).Inc()
}
