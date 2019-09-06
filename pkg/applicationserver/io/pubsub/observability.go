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

package pubsub

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtSetPubSub = events.Define(
		"as.pubsub.set", "set pubsub",
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	)
	evtDeletePubSub = events.Define(
		"as.pubsub.delete", "delete pubsub",
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	)
	evtPubSubStart = events.Define(
		"as.pubsub.start", "start pubsub",
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	)
	evtPubSubStop = events.Define(
		"as.pubsub.stop", "stop pubsub",
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	)
	evtPubSubFail = events.Define(
		"as.pubsub.fail", "fail pubsub",
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	)
)

const (
	subsystem     = "as_pubsub"
	unknown       = "unknown"
	applicationID = "application_id"
)

var pubsubMetrics = &integrationsMetrics{
	integrationsStarted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "integrations_started_total",
			Help:      "Number of integrations started",
		},
		[]string{applicationID},
	),
	integrationsStopped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "integrations_stopped_total",
			Help:      "Number of integrations stopped",
		},
		[]string{applicationID},
	),
	integrationsFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "integrations_failed_total",
			Help:      "Number of integrations failed",
		},
		[]string{applicationID},
	),
}

func init() {
	metrics.MustRegister(pubsubMetrics)
}

type integrationsMetrics struct {
	integrationsStarted *metrics.ContextualCounterVec
	integrationsStopped *metrics.ContextualCounterVec
	integrationsFailed  *metrics.ContextualCounterVec
}

func (m integrationsMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.integrationsStarted.Describe(ch)
	m.integrationsStopped.Describe(ch)
	m.integrationsFailed.Describe(ch)
}

func (m integrationsMetrics) Collect(ch chan<- prometheus.Metric) {
	m.integrationsStarted.Collect(ch)
	m.integrationsStopped.Collect(ch)
	m.integrationsFailed.Collect(ch)
}

func registerIntegrationStart(ctx context.Context, i *integration) {
	events.Publish(evtPubSubStart(ctx, i.ApplicationIdentifiers, i.ApplicationPubSubIdentifiers))
	pubsubMetrics.integrationsStarted.WithLabelValues(ctx, i.ApplicationID).Inc()
}

func registerIntegrationStop(ctx context.Context, i *integration) {
	events.Publish(evtPubSubStop(ctx, i.ApplicationIdentifiers, i.ApplicationPubSubIdentifiers))
	pubsubMetrics.integrationsStopped.WithLabelValues(ctx, i.ApplicationID).Inc()
}

var errIntegrationFailed = errors.DefineAborted("integration_failed", "integration {pub_sub_id} failed")

func registerIntegrationFail(ctx context.Context, i *integration, err error) {
	err = errIntegrationFailed.
		WithAttributes(
			"application_uid", unique.ID(ctx, i.ApplicationIdentifiers),
			"pub_sub_id", i.PubSubID).
		WithCause(err)
	events.Publish(evtPubSubFail(ctx, i.ApplicationIdentifiers, err))
	pubsubMetrics.integrationsFailed.WithLabelValues(ctx, i.ApplicationID).Inc()
}
