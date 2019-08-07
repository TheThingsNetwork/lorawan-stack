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

// Package metrics creates the metrics registry and exposes some common metrics.
package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/version"
)

// Namespace for metrics.
const Namespace = "ttn_lw"

// ContextLabelNames are the label names that can be retrieved from a context for XXXVec metrics.
var ContextLabelNames []string

// LabelsFromContext returns the values for ContextLabelNames.
var LabelsFromContext func(ctx context.Context) prometheus.Labels

var ttnInfo = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: Namespace,
	Name:      "info",
	Help:      "Information about The Things Stack for LoRaWAN",
	ConstLabels: prometheus.Labels{
		"version":    version.TTN,
		"build_date": version.BuildDate,
		"git_commit": version.GitCommit,
	},
})

func init() {
	ttnInfo.Set(1)
}
