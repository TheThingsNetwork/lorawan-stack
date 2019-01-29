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

package applicationserver

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtApplicationSubscribe   = events.Define("as.application.subscribe", "application subscribe")
	evtApplicationUnsubscribe = events.Define("as.application.unsubscribe", "application unsubscribe")

	evtReceiveDataUp    = events.Define("as.up.data.receive", "receive uplink data message")
	evtDropDataUp       = events.Define("as.up.data.drop", "drop uplink data message")
	evtForwardDataUp    = events.Define("as.up.data.forward", "forward uplink data message")
	evtDecodeFailDataUp = events.Define("as.up.data.decode.fail", "decode uplink data message fail")

	evtReceiveJoinAccept = events.Define("as.up.join.receive", "receive join-accept message")
	evtDropJoinAccept    = events.Define("as.up.join.drop", "drop join-accept message")
	evtForwardJoinAccept = events.Define("as.up.join.forward", "forward join-accept message")

	evtReceiveDataDown      = events.Define("as.down.data.receive", "receive downlink data message")
	evtDropDataDown         = events.Define("as.down.data.drop", "drop downlink data message")
	evtForwardDataDown      = events.Define("as.down.data.forward", "forward downlink data message")
	evtLostQueueDataDown    = events.Define("as.down.data.queue.lost", "lost downlink data queue")
	evtInvalidQueueDataDown = events.Define("as.down.data.queue.invalid", "invalid downlink data queue")
)

const (
	subsystem     = "as"
	unknown       = "unknown"
	networkServer = "network_server"
	protocol      = "protocol"
	applicationID = "application_id"
)

var asMetrics = &messageMetrics{
	subscriptionsStarted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscriptions_started",
			Help:      "Number of subscriptions started",
		},
		[]string{protocol},
	),
	subscriptionsEnded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "subscriptions_ended",
			Help:      "Number of subscriptions ended",
		},
		[]string{protocol},
	),
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of received uplinks (join-accepts and data)",
		},
		[]string{networkServer},
	),
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of forwarded uplinks (join-accepts and data)",
		},
		[]string{applicationID},
	),
	uplinkDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_dropped_total",
			Help:      "Total number of dropped uplinks (join-accepts and data)",
		},
		[]string{"error"},
	),
	downlinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_received_total",
			Help:      "Total number of received downlinks",
		},
		[]string{applicationID},
	),
	downlinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_forwarded_total",
			Help:      "Total number of forwarded downlinks",
		},
		[]string{networkServer},
	),
	downlinkDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_dropped_total",
			Help:      "Total number of dropped downlinks (join-accepts and data)",
		},
		[]string{"error"},
	),
}

func init() {
	metrics.MustRegister(asMetrics)
}

type messageMetrics struct {
	subscriptionsStarted *metrics.ContextualCounterVec
	subscriptionsEnded   *metrics.ContextualCounterVec
	uplinkReceived       *metrics.ContextualCounterVec
	uplinkForwarded      *metrics.ContextualCounterVec
	uplinkDropped        *metrics.ContextualCounterVec
	downlinkReceived     *metrics.ContextualCounterVec
	downlinkForwarded    *metrics.ContextualCounterVec
	downlinkDropped      *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.subscriptionsStarted.Describe(ch)
	m.subscriptionsEnded.Describe(ch)
	m.uplinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.downlinkReceived.Describe(ch)
	m.downlinkForwarded.Describe(ch)
	m.downlinkDropped.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.subscriptionsStarted.Collect(ch)
	m.subscriptionsEnded.Collect(ch)
	m.uplinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.downlinkReceived.Collect(ch)
	m.downlinkForwarded.Collect(ch)
	m.downlinkDropped.Collect(ch)
}

func registerSubscribe(ctx context.Context, sub *io.Subscription) {
	var ids ttnpb.Identifiers
	if appIDs := sub.ApplicationIDs(); appIDs != nil {
		ids = appIDs
	}
	events.Publish(evtApplicationSubscribe(ctx, ids, nil))
	asMetrics.subscriptionsStarted.WithLabelValues(ctx, sub.Protocol()).Inc()
}

func registerUnsubscribe(ctx context.Context, sub *io.Subscription) {
	var ids ttnpb.Identifiers
	if appIDs := sub.ApplicationIDs(); appIDs != nil {
		ids = appIDs
	}
	events.Publish(evtApplicationUnsubscribe(ctx, ids, nil))
	asMetrics.subscriptionsEnded.WithLabelValues(ctx, sub.Protocol()).Inc()
}

func registerReceiveUp(ctx context.Context, msg *ttnpb.ApplicationUp, ns string) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtReceiveJoinAccept(ctx, msg.EndDeviceIdentifiers, nil))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtReceiveDataUp(ctx, msg.EndDeviceIdentifiers, nil))
	}
	asMetrics.uplinkReceived.WithLabelValues(ctx, ns).Inc()
}

func registerForwardUp(ctx context.Context, msg *ttnpb.ApplicationUp) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtForwardJoinAccept(ctx, msg.EndDeviceIdentifiers, nil))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtForwardDataUp(ctx, msg.EndDeviceIdentifiers, nil))
	}
	asMetrics.uplinkForwarded.WithLabelValues(ctx, msg.ApplicationID).Inc()
}

func registerDropUp(ctx context.Context, msg *ttnpb.ApplicationUp, err error) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtDropJoinAccept(ctx, msg.EndDeviceIdentifiers, nil))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtDropDataUp(ctx, msg.EndDeviceIdentifiers, nil))
	}
	if ttnErr, ok := errors.From(err); ok {
		asMetrics.uplinkDropped.WithLabelValues(ctx, ttnErr.String()).Inc()
	} else {
		asMetrics.uplinkDropped.WithLabelValues(ctx, unknown).Inc()
	}
}

func registerReceiveDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink) {
	events.Publish(evtReceiveDataDown(ctx, ids, nil))
	asMetrics.downlinkReceived.WithLabelValues(ctx, ids.ApplicationID).Inc()
}

func registerForwardDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, ns string) {
	events.Publish(evtForwardDataDown(ctx, ids, nil))
	asMetrics.downlinkForwarded.WithLabelValues(ctx, ns).Inc()
}

func registerDropDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, err error) {
	events.Publish(evtDropDataDown(ctx, ids, nil))
	if ttnErr, ok := errors.From(err); ok {
		asMetrics.downlinkDropped.WithLabelValues(ctx, ttnErr.String()).Inc()
	} else {
		asMetrics.downlinkDropped.WithLabelValues(ctx, unknown).Inc()
	}
}
