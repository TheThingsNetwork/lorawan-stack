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

package packages

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

const (
	subsystem = "as_packages"
	unknown   = "unknown"
)

var packagesMetrics = &messageMetrics{
	messagesProcessed: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "processed_total",
			Help:      "Total number of processed messages",
		},
		[]string{"package"},
	),
	messagesFailed: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "failed_total",
			Help:      "Total number of failed messages",
		},
		[]string{"package", "error"},
	),
}

func init() {
	metrics.MustRegister(packagesMetrics)
}

type messageMetrics struct {
	messagesProcessed *prometheus.CounterVec
	messagesFailed    *prometheus.CounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.messagesProcessed.Describe(ch)
	m.messagesFailed.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.messagesProcessed.Collect(ch)
	m.messagesFailed.Collect(ch)
}

func registerMessageProcessed(name string) {
	packagesMetrics.messagesProcessed.WithLabelValues(name).Inc()
}

func registerMessageFailed(name string, err error) {
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	packagesMetrics.messagesFailed.WithLabelValues(name, errorLabel).Inc()
}
