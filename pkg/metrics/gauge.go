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
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// NewGauge returns a new Gauge and sets its namespace.
func NewGauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	opts.Namespace = Namespace
	return prometheus.NewGauge(opts)
}

// MustRegisterGauge is a convenience function for NewGauge and MustRegister.
func MustRegisterGauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	metric := NewGauge(opts)
	MustRegister(metric)
	return metric
}

// NewGaugeFunc returns a new GaugeFunc and sets its namespace.
func NewGaugeFunc(opts prometheus.GaugeOpts, function func() float64) prometheus.GaugeFunc {
	opts.Namespace = Namespace
	return prometheus.NewGaugeFunc(opts, function)
}

// MustRegisterGaugeFunc is a convenience function for NewGaugeFunc and MustRegister.
func MustRegisterGaugeFunc(opts prometheus.GaugeOpts, function func() float64) prometheus.GaugeFunc {
	metric := NewGaugeFunc(opts, function)
	MustRegister(metric)
	return metric
}

// NewGaugeVec returns a new GaugeVec and sets its namespace.
func NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	opts.Namespace = Namespace
	return prometheus.NewGaugeVec(opts, labelNames)
}

// MustRegisterGaugeVec is a convenience function for NewGaugeVec and MustRegister.
func MustRegisterGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	metric := NewGaugeVec(opts, labelNames)
	MustRegister(metric)
	return metric
}

// ContextualGaugeVec wraps a GaugeVec in order to get labels from the context.
type ContextualGaugeVec struct {
	*prometheus.GaugeVec
}

// With is the equivalent of GaugeVec.With, but with a context.
func (c ContextualGaugeVec) With(ctx context.Context, labels prometheus.Labels) prometheus.Gauge {
	if LabelsFromContext == nil {
		return c.GaugeVec.With(labels)
	}
	return c.GaugeVec.MustCurryWith(LabelsFromContext(ctx)).With(labels)
}

// WithLabelValues is the equivalent of GaugeVec.WithLabelValues, but with a context.
func (c ContextualGaugeVec) WithLabelValues(ctx context.Context, lvs ...string) prometheus.Gauge {
	if len(ContextLabelNames) == 0 {
		return c.GaugeVec.WithLabelValues(lvs...)
	}
	return c.GaugeVec.MustCurryWith(LabelsFromContext(ctx)).WithLabelValues(lvs...)
}

// NewContextualGaugeVec returns a new ContextualGaugeVec and sets its namespace.
func NewContextualGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *ContextualGaugeVec {
	opts.Namespace = Namespace
	if len(ContextLabelNames) > 0 {
		labelNames = append(ContextLabelNames, labelNames...)
	}
	return &ContextualGaugeVec{prometheus.NewGaugeVec(opts, labelNames)}
}

// MustRegisterContextualGaugeVec is a convenience function for NewContextualGaugeVec and MustRegister.
func MustRegisterContextualGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *ContextualGaugeVec {
	metric := NewContextualGaugeVec(opts, labelNames)
	MustRegister(metric)
	return metric
}
