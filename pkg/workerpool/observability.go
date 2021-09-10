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

package workerpool

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

const (
	subsystem = "workerpool"
	poolLabel = "pool"
)

type workPoolMetrics struct {
	workersStarted *metrics.ContextualCounterVec
	workersStopped *metrics.ContextualCounterVec
	workEnqueued   *metrics.ContextualCounterVec
	workDequeued   *metrics.ContextualCounterVec
	workDropped    *metrics.ContextualCounterVec
}

func (m workPoolMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.workersStarted.Describe(ch)
	m.workersStopped.Describe(ch)
	m.workEnqueued.Describe(ch)
	m.workDequeued.Describe(ch)
	m.workDropped.Describe(ch)
}

func (m workPoolMetrics) Collect(ch chan<- prometheus.Metric) {
	m.workersStarted.Collect(ch)
	m.workersStopped.Collect(ch)
	m.workEnqueued.Collect(ch)
	m.workDequeued.Collect(ch)
	m.workDropped.Collect(ch)
}

var poolMetrics = &workPoolMetrics{
	workersStarted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "workers_started",
			Help:      "Number of workers started",
		},
		[]string{poolLabel},
	),
	workersStopped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "workers_stopped",
			Help:      "Number of workers stopped",
		},
		[]string{poolLabel},
	),
	workEnqueued: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "work_enqueued",
			Help:      "Amount of work enqueued",
		},
		[]string{poolLabel},
	),
	workDequeued: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "work_dequeued",
			Help:      "Amount of work dequeued",
		},
		[]string{poolLabel},
	),
	workDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "work_dropped",
			Help:      "Amount of work dropped",
		},
		[]string{poolLabel},
	),
}

func init() {
	metrics.MustRegister(poolMetrics)
}

func registerWorkerStarted(ctx context.Context, name string) {
	poolMetrics.workersStarted.WithLabelValues(ctx, name).Inc()
	poolMetrics.workersStopped.WithLabelValues(ctx, name)
}

func registerWorkerStopped(ctx context.Context, name string) {
	poolMetrics.workersStopped.WithLabelValues(ctx, name).Inc()
}

func registerWorkEnqueued(ctx context.Context, name string) {
	poolMetrics.workEnqueued.WithLabelValues(ctx, name).Inc()
	poolMetrics.workDequeued.WithLabelValues(ctx, name)
}

func registerWorkDequeued(ctx context.Context, name string) {
	poolMetrics.workDequeued.WithLabelValues(ctx, name).Inc()
}

func registerWorkDropped(ctx context.Context, name string) {
	poolMetrics.workDropped.WithLabelValues(ctx, name).Inc()
}
