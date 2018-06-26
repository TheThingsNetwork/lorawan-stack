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

package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// NewHistogram returns a new Histogram and sets its namespace.
func NewHistogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	opts.Namespace = Namespace
	return prometheus.NewHistogram(opts)
}

// MustRegisterHistogram is a convenience function for NewHistogram and MustRegister.
func MustRegisterHistogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	metric := NewHistogram(opts)
	MustRegister(metric)
	return metric
}

// NewHistogramVec returns a new HistogramVec and sets its namespace.
func NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	opts.Namespace = Namespace
	return prometheus.NewHistogramVec(opts, labelNames)
}

// MustRegisterHistogramVec is a convenience function for NewHistogramVec and MustRegister.
func MustRegisterHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	metric := NewHistogramVec(opts, labelNames)
	MustRegister(metric)
	return metric
}

// ContextualHistogramVec wraps a HistogramVec in order to get labels from the context.
type ContextualHistogramVec struct {
	*prometheus.HistogramVec
}

// With is the equivalent of HistogramVec.With, but with a context.
func (c ContextualHistogramVec) With(ctx context.Context, labels prometheus.Labels) prometheus.Observer {
	if LabelsFromContext == nil {
		return c.HistogramVec.With(labels)
	}
	return c.HistogramVec.MustCurryWith(LabelsFromContext(ctx)).With(labels)
}

// WithLabelValues is the equivalent of HistogramVec.WithLabelValues, but with a context.
func (c ContextualHistogramVec) WithLabelValues(ctx context.Context, lvs ...string) prometheus.Observer {
	if len(ContextLabelNames) == 0 {
		return c.HistogramVec.WithLabelValues(lvs...)
	}
	return c.HistogramVec.MustCurryWith(LabelsFromContext(ctx)).WithLabelValues(lvs...)
}

// NewContextualHistogramVec returns a new ContextualHistogramVec and sets its namespace.
func NewContextualHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *ContextualHistogramVec {
	opts.Namespace = Namespace
	if len(ContextLabelNames) > 0 {
		labelNames = append(ContextLabelNames, labelNames...)
	}
	return &ContextualHistogramVec{prometheus.NewHistogramVec(opts, labelNames)}
}

// MustRegisterContextualHistogramVec is a convenience function for NewContextualHistogramVec and MustRegister.
func MustRegisterContextualHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *ContextualHistogramVec {
	metric := NewContextualHistogramVec(opts, labelNames)
	MustRegister(metric)
	return metric
}
