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

package applicationserver

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtApplicationSubscribe   = events.Define("as.application.subscribe", "application subscribe")
	evtApplicationUnsubscribe = events.Define("as.application.unsubscribe", "application unsubscribe")

	evtReceiveDataUp = events.Define("as.up.data.receive", "receive uplink data message")
	evtDropDataUp    = events.Define("as.up.data.drop", "drop uplink data message")
	evtForwardDataUp = events.Define("as.up.data.forward", "forward uplink data message")

	evtReceiveJoinAccept = events.Define("as.up.join.receive", "receive join-accept message")
	evtDropJoinAccept    = events.Define("as.up.join.drop", "drop join-accept message")
	evtForwardJoinAccept = events.Define("as.up.join.forward", "forward join-accept message")

	evtReceiveDataDown = events.Define("as.down.data.receive", "receive downlink data message")
	evtDropDataDown    = events.Define("as.down.data.drop", "drop downlink data message")
	evtForwardDataDown = events.Define("as.down.data.forward", "forward downlink data message")
)

const (
	subsystem     = "as"
	unknown       = "unknown"
	networkServer = "network_server"
	protocol      = "protocol"
	applicationID = "application_id"
)

var asMetrics = &messageMetrics{
	subscriptions: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "subscriptions",
			Help:      "Number of subscriptions",
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
		// TODO: Remove label (https://github.com/TheThingsIndustries/lorawan-stack/issues/1039)
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
	downlinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_forwarded_total",
			Help:      "Total number of forwarded downlinks",
		},
		[]string{networkServer},
	),
}

func init() {
	metrics.MustRegister(asMetrics)
}

type messageMetrics struct {
	subscriptions     *metrics.ContextualGaugeVec
	uplinkReceived    *metrics.ContextualCounterVec
	uplinkForwarded   *metrics.ContextualCounterVec
	uplinkDropped     *metrics.ContextualCounterVec
	downlinkForwarded *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.subscriptions.Describe(ch)
	m.uplinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.downlinkForwarded.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.subscriptions.Collect(ch)
	m.uplinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.downlinkForwarded.Collect(ch)
}

func registerSubscribe(ctx context.Context, conn *io.Connection) {
	events.Publish(evtApplicationSubscribe(ctx, conn.ApplicationIdentifiers, nil))
	asMetrics.subscriptions.WithLabelValues(ctx, conn.Protocol()).Inc()
}

func registerUnsubscribe(ctx context.Context, conn *io.Connection) {
	events.Publish(evtApplicationUnsubscribe(ctx, conn.ApplicationIdentifiers, nil))
	asMetrics.subscriptions.WithLabelValues(ctx, conn.Protocol()).Dec()
}

func registerReceiveUplink(ctx context.Context, msg *ttnpb.ApplicationUp, ns string) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtReceiveJoinAccept(ctx, msg.EndDeviceIdentifiers, nil))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtReceiveDataUp(ctx, msg.EndDeviceIdentifiers, nil))
	}
	asMetrics.uplinkReceived.WithLabelValues(ctx, ns).Inc()
}

func registerForwardUplink(ctx context.Context, msg *ttnpb.ApplicationUp) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtForwardJoinAccept(ctx, msg.EndDeviceIdentifiers, nil))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtForwardDataUp(ctx, msg.EndDeviceIdentifiers, nil))
	}
	asMetrics.uplinkForwarded.WithLabelValues(ctx, msg.ApplicationID).Inc()
}

func registerDropUplink(ctx context.Context, msg *ttnpb.ApplicationUp, err error) {
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

func registerForwardDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, ns string) {
	events.Publish(evtForwardDataDown(ctx, ids, nil))
	asMetrics.downlinkForwarded.WithLabelValues(ctx, ns).Inc()
}
