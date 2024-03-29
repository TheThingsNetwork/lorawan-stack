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

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var registry = prometheus.NewRegistry()

// Exporter for the metrics registry.
var Exporter http.Handler = promhttp.InstrumentMetricHandler(
	registry,
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
)

// Registry for metrics.
var Registry prometheus.Registerer = registry

func init() {
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(ttnInfo)
}

// Register registers the given Collector in the registry.
func Register(c prometheus.Collector) error {
	return registry.Register(c)
}

// MustRegister registers the given Collectors in the registry and panics on errors.
func MustRegister(cs ...prometheus.Collector) {
	registry.MustRegister(cs...)
}

// Unregister the given Collector from the Prometheus registry.
func Unregister(c prometheus.Collector) bool {
	return registry.Unregister(c)
}
