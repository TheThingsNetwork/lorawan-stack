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

package events

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/metrics"
)

const subsystem = "events"

var publishes = metrics.NewContextualCounterVec(
	prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "publishes_total",
		Help:      "Number of Publishes",
	},
	[]string{"name"},
)

var subscriptions = metrics.NewGaugeVec(
	prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "subscriptions",
		Help:      "Number of Subscriptions",
	},
	[]string{"name"},
)

var channelDropped = metrics.NewContextualCounterVec(
	prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "channel_dropped_total",
		Help:      "Number of events dropped from event channels",
	},
	[]string{"name"},
)

func initMetrics(name string) {
	ctx := context.Background()
	publishes.WithLabelValues(ctx, name).Add(0)
	subscriptions.WithLabelValues(name).Add(0)
	channelDropped.WithLabelValues(ctx, name).Add(0)
}

func init() {
	metrics.MustRegister(publishes, subscriptions, channelDropped)
}
