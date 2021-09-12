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
	workersStarted *prometheus.CounterVec
	workersIdle    *prometheus.GaugeVec
	workersStopped *prometheus.CounterVec
	workQueueSize  *prometheus.GaugeVec
	workProcessed  *metrics.ContextualCounterVec
	workDropped    *metrics.ContextualCounterVec
}

func (m workPoolMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.workersStarted.Describe(ch)
	m.workersIdle.Describe(ch)
	m.workersStopped.Describe(ch)
	m.workQueueSize.Describe(ch)
	m.workProcessed.Describe(ch)
	m.workDropped.Describe(ch)
}

func (m workPoolMetrics) Collect(ch chan<- prometheus.Metric) {
	m.workersStarted.Collect(ch)
	m.workersIdle.Collect(ch)
	m.workersStopped.Collect(ch)
	m.workQueueSize.Collect(ch)
	m.workProcessed.Collect(ch)
	m.workDropped.Collect(ch)
}

var poolMetrics = &workPoolMetrics{
	workersStarted: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "workers_started",
			Help:      "Number of workers started",
		},
		[]string{poolLabel},
	),
	workersIdle: metrics.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "workers_idle",
			Help:      "Number of idle workers",
		},
		[]string{poolLabel},
	),
	workersStopped: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "workers_stopped",
			Help:      "Number of workers stopped",
		},
		[]string{poolLabel},
	),
	workQueueSize: metrics.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "work_queue_size",
			Help:      "Amount of work enqueued",
		},
		[]string{poolLabel},
	),
	workProcessed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "work_processed",
			Help:      "Amount of work processed",
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

func registerWorkerStarted(name string) {
	poolMetrics.workersStarted.WithLabelValues(name).Inc()
	poolMetrics.workersIdle.WithLabelValues(name)
	poolMetrics.workersStopped.WithLabelValues(name)
}

func registerWorkerIdle(name string) {
	poolMetrics.workersIdle.WithLabelValues(name).Inc()
}

func registerWorkerBusy(name string) {
	poolMetrics.workersIdle.WithLabelValues(name).Dec()
}

func registerWorkerStopped(name string) {
	poolMetrics.workersStopped.WithLabelValues(name).Inc()
}

func registerWorkEnqueued(name string) {
	poolMetrics.workQueueSize.WithLabelValues(name).Inc()
}

func registerWorkDequeued(name string) {
	poolMetrics.workQueueSize.WithLabelValues(name).Dec()
}

func registerWorkProcessed(ctx context.Context, name string) {
	poolMetrics.workProcessed.WithLabelValues(ctx, name).Inc()
}

func registerWorkDropped(ctx context.Context, name string) {
	poolMetrics.workDropped.WithLabelValues(ctx, name).Inc()
}
