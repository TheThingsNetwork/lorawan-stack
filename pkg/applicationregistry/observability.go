// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package applicationregistry

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
)

var (
	evtCreateApplication = events.Define("application.create", "Create Application")
	evtUpdateApplication = events.Define("application.update", "Update Application")
	evtDeleteApplication = events.Define("application.delete", "Delete Application")
)

const subsystem = "application_registry"

var latencyBuckets = []float64{}

var latency = metrics.NewHistogramVec(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "call_latency_seconds",
		Help:      "Histogram of latency (seconds) of application registry calls",
		Buckets:   latencyBuckets,
	},
	[]string{"action"},
)

var rangeLatency = metrics.NewHistogramVec(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "range_latency_seconds",
		Help:      "Histogram of latency (seconds) of application registry ranges",
		Buckets:   latencyBuckets,
	},
	[]string{"fields"},
)

func init() {
	metrics.MustRegister(latency, rangeLatency)
}
