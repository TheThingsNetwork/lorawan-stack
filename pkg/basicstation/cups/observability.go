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
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

type messageMetrics struct {
	requestReceived  *metrics.ContextualCounterVec
	requestSucceeded *metrics.ContextualCounterVec
	requestFailed    *metrics.ContextualCounterVec
}

// TODO: Fetch this from the device repository (https://github.com/TheThingsIndustries/lorawan-stack/issues/2018).
var allowedModels = []string{
	// Note: Please keep this list sorted
	"arm",
	"browan_mt7620a",
	"corecell",
	"laird",
	"linux",
	"linuxpico",
	"lorix",
	"minihub",
	"mips-openwrt",
	"mlinux",
	"rpi",
	"stm32mp1",
}

// TODO: Fetch this from the device repository (https://github.com/TheThingsIndustries/lorawan-stack/issues/2018).
var allowedTypes = []string{
	// Note: Please keep this list sorted
	"std",
	"debug",
}

var (
	subsystem = "cups"
	request   = "request"
	// Ex: `2.0.4(minihub/debug) 2020-05-07 16:03:52` or  `2.0.4-9-g3d5c686(linux/std) 2021-04-16 15:58:53`
	stationRegex = regexp.MustCompile(`([0-9]\.[0-9]\.[0-9](-[0-9]-[a-z0-9]{8})?)\(([a-z_\-0-9]+)\/([a-z_\-0-9]+)\) [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}$`)
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
		[]string{request, "firmware", "model", "type"},
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

func registerUpdateInfoRequestSucceeded(ctx context.Context, request string, station string) {
	log.FromContext(ctx).WithField("station", station).Debug("Register metrics")
	s := stationRegex.FindStringSubmatch(station)
	var (
		firmware = "unknown"
		model    = "unknown"
		typ      = "unknown"
	)
	if len(s) == 4 {
		firmware = s[1]
		for _, mdl := range allowedModels {
			if s[3] == mdl {
				model = mdl
			}
		}
		for _, t := range allowedTypes {
			if s[4] == t {
				typ = t
			}
		}
	}
	cupsMetrics.requestSucceeded.WithLabelValues(ctx, request, firmware, model, typ).Inc()
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
