// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package javascript

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/metrics"
)

const subsystem = "javascript"

var runs = metrics.NewCounterVec(
	prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "runs_total",
		Help:      "JavaScript runs",
	},
	[]string{"result"},
)

var runLatency = metrics.NewHistogram(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "run_latency_seconds",
		Help:      "Histogram of latency (seconds) of JavaScript runs",
	},
)

func init() {
	metrics.MustRegister(runs, runLatency)
}
