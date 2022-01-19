// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package distribution

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtApplicationSubscribe = events.Define(
		"as.application.subscribe", "subscribe application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_LINK),
	)
	evtApplicationUnsubscribe = events.Define(
		"as.application.unsubscribe", "unsubscribe application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_LINK),
	)
)

const (
	subsystem = "as"
	protocol  = "protocol"
	unknown   = "unknown"
)

var subMetrics = &subscriptionMetrics{
	subscriptionSetsStarted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscription_sets_started_total",
			Help:      "Number of subscription sets started",
		},
		[]string{},
	),
	subscriptionSetsStopped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscription_sets_stopped_total",
			Help:      "Number of subscription sets stopped",
		},
		[]string{},
	),
	subscriptionSetsPublishSuccess: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscription_sets_publish_success_total",
			Help:      "Number of successful publish attempts",
		},
		[]string{protocol},
	),
	subscriptionSetsPublishFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscription_sets_publish_failed_total",
			Help:      "Number of failed publish attempts",
		},
		[]string{protocol, "error"},
	),
	subscriptionsStarted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscriptions_started_total",
			Help:      "Number of subscriptions started",
		},
		[]string{protocol},
	),
	subscriptionsStopped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscriptions_stopped_total",
			Help:      "Number of subscriptions stopped",
		},
		[]string{protocol},
	),
}

func init() {
	metrics.MustRegister(subMetrics)
}

type subscriptionMetrics struct {
	subscriptionSetsStarted        *metrics.ContextualCounterVec
	subscriptionSetsStopped        *metrics.ContextualCounterVec
	subscriptionSetsPublishSuccess *metrics.ContextualCounterVec
	subscriptionSetsPublishFailed  *metrics.ContextualCounterVec
	subscriptionsStarted           *metrics.ContextualCounterVec
	subscriptionsStopped           *metrics.ContextualCounterVec
}

func (m subscriptionMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.subscriptionSetsStarted.Describe(ch)
	m.subscriptionSetsStopped.Describe(ch)
	m.subscriptionSetsPublishSuccess.Describe(ch)
	m.subscriptionSetsPublishFailed.Describe(ch)
	m.subscriptionsStarted.Describe(ch)
	m.subscriptionsStopped.Describe(ch)
}

func (m subscriptionMetrics) Collect(ch chan<- prometheus.Metric) {
	m.subscriptionSetsStarted.Collect(ch)
	m.subscriptionSetsStopped.Collect(ch)
	m.subscriptionSetsPublishSuccess.Collect(ch)
	m.subscriptionSetsPublishFailed.Collect(ch)
	m.subscriptionsStarted.Collect(ch)
	m.subscriptionsStopped.Collect(ch)
}

func registerSubscriptionSetStart(ctx context.Context) {
	subMetrics.subscriptionSetsStarted.WithLabelValues(ctx).Inc()
	subMetrics.subscriptionSetsStopped.WithLabelValues(ctx) // Initialize the "stopped" counter.
}

func registerSubscriptionSetStop(ctx context.Context) {
	subMetrics.subscriptionSetsStopped.WithLabelValues(ctx).Inc()
}

func registerSubscribe(ctx context.Context, sub *io.Subscription) {
	var ids events.EntityIdentifiers
	if appIDs := sub.ApplicationIDs(); appIDs != nil {
		ids = appIDs
	}
	events.Publish(evtApplicationSubscribe.NewWithIdentifiersAndData(ctx, ids, nil))
	subMetrics.subscriptionsStarted.WithLabelValues(ctx, sub.Protocol()).Inc()
	subMetrics.subscriptionsStopped.WithLabelValues(ctx, sub.Protocol()) // Initialize the "stopped" counter.
}

func registerUnsubscribe(ctx context.Context, sub *io.Subscription) {
	var ids events.EntityIdentifiers
	if appIDs := sub.ApplicationIDs(); appIDs != nil {
		ids = appIDs
	}
	events.Publish(evtApplicationUnsubscribe.NewWithIdentifiersAndData(ctx, ids, nil))
	subMetrics.subscriptionsStopped.WithLabelValues(ctx, sub.Protocol()).Inc()
}

func registerPublishSuccess(ctx context.Context, sub *io.Subscription) {
	subMetrics.subscriptionSetsPublishSuccess.WithLabelValues(ctx, sub.Protocol()).Inc()
}

func registerPublishFailed(ctx context.Context, sub *io.Subscription, err error) {
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	subMetrics.subscriptionSetsPublishFailed.WithLabelValues(ctx, sub.Protocol(), errorLabel).Inc()
}
