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
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var withIdentifiersOption = events.WithDataType(&ttnpb.ApplicationPubSubIdentifiers{
	ApplicationIds: &ttnpb.ApplicationIdentifiers{
		ApplicationId: "application-id",
	},
	PubSubId: "pubsub-id",
})

var (
	evtSetPubSub = events.Define(
		"as.pubsub.set", "set pub/sub",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC),
		withIdentifiersOption,
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeletePubSub = events.Define(
		"as.pubsub.delete", "delete pub/sub",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC),
		withIdentifiersOption,
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPubSubStart = events.Define(
		"as.pubsub.start", "start pub/sub",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
		),
		withIdentifiersOption,
	)
	evtPubSubStop = events.Define(
		"as.pubsub.stop", "stop pub/sub",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
		),
		withIdentifiersOption,
	)
	evtPubSubFail = events.Define(
		"as.pubsub.fail", "fail pub/sub",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
		),
		events.WithErrorDataType(),
	)
)

const (
	subsystem     = "as_pubsub"
	unknown       = "unknown"
	providerLabel = "provider"
)

var pubsubMetrics = &integrationsMetrics{
	integrationsStarted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "integrations_started_total",
			Help:      "Number of integrations started",
		},
		[]string{providerLabel},
	),
	integrationsStopped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "integrations_stopped_total",
			Help:      "Number of integrations stopped",
		},
		[]string{providerLabel},
	),
	integrationsFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "integrations_failed_total",
			Help:      "Number of integrations failed",
		},
		[]string{providerLabel},
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

var psTypeName = fmt.Sprintf("%T", &ttnpb.ApplicationPubSub{})

func providerLabelValue(i *integration) string {
	return strings.ToLower(strings.TrimPrefix(fmt.Sprintf("%T", i.ApplicationPubSub.GetProvider()), psTypeName+"_"))
}

func registerIntegrationStart(ctx context.Context, i *integration) {
	events.Publish(evtPubSubStart.NewWithIdentifiersAndData(ctx, i.Ids.ApplicationIds, i.Ids))
	labelValue := providerLabelValue(i)
	pubsubMetrics.integrationsStarted.WithLabelValues(ctx, labelValue).Inc()
	pubsubMetrics.integrationsStopped.WithLabelValues(ctx, labelValue) // Initialize the "stopped" counter.
}

func registerIntegrationStop(ctx context.Context, i *integration) {
	events.Publish(evtPubSubStop.NewWithIdentifiersAndData(ctx, i.Ids.ApplicationIds, i.Ids))
	pubsubMetrics.integrationsStopped.WithLabelValues(ctx, providerLabelValue(i)).Inc()
}

var errIntegrationFailed = errors.DefineAborted("integration_failed", "integration `{pub_sub_id}` failed")

func registerIntegrationFail(ctx context.Context, i *integration, err error) {
	err = errIntegrationFailed.
		WithAttributes(
			"application_uid", unique.ID(ctx, i.Ids.ApplicationIds),
			"pub_sub_id", i.Ids.PubSubId,
		).
		WithCause(err)
	events.Publish(evtPubSubFail.NewWithIdentifiersAndData(ctx, i.Ids.ApplicationIds, err))
	pubsubMetrics.integrationsFailed.WithLabelValues(ctx, providerLabelValue(i)).Inc()
}
