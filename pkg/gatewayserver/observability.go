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

package gatewayserver

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtGatewayConnect    = events.Define("gs.gateway.connect", "gateway connect")
	evtGatewayDisconnect = events.Define("gs.gateway.disconnect", "gateway disconnect")

	evtReceiveStatus = events.Define("gs.status.receive", "receive gateway status")

	evtReceiveUp = events.Define("gs.up.receive", "receive uplink message")
	evtDropUp    = events.Define("gs.up.drop", "drop uplink message")
	evtForwardUp = events.Define("gs.up.forward", "forward uplink message")

	evtSendDown = events.Define("gs.down.send", "send downlink message")
)

const (
	subsystem = "gs"
	unknown   = "unknown"
	gatewayID = "gateway_id"
	peer      = "peer"
)

var gsMetrics = &messageMetrics{
	statusReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "status_received_total",
			Help:      "Total number of received gateway statuses",
		},
		// TODO: Remove label (https://github.com/TheThingsIndustries/lorawan-stack/issues/1039)
		[]string{gatewayID},
	),
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of received uplinks",
		},
		// TODO: Remove label (https://github.com/TheThingsIndustries/lorawan-stack/issues/1039)
		[]string{gatewayID},
	),
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of forwarded uplinks",
		},
		[]string{peer},
	),
	uplinkDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_dropped_total",
			Help:      "Total number of dropped uplinks",
		},
		[]string{"error"},
	),
	downlinkSent: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_sent_total",
			Help:      "Total number of sent downlinks",
		},
		// TODO: Remove label (https://github.com/TheThingsIndustries/lorawan-stack/issues/1039)
		[]string{gatewayID},
	),
}

func init() {
	metrics.MustRegister(gsMetrics)
}

type messageMetrics struct {
	statusReceived  *metrics.ContextualCounterVec
	uplinkReceived  *metrics.ContextualCounterVec
	uplinkForwarded *metrics.ContextualCounterVec
	uplinkDropped   *metrics.ContextualCounterVec
	downlinkSent    *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.statusReceived.Describe(ch)
	m.uplinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.downlinkSent.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.statusReceived.Collect(ch)
	m.uplinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.downlinkSent.Collect(ch)
}

func registerReceiveStatus(ctx context.Context, gtw *ttnpb.Gateway, status *ttnpb.GatewayStatus) {
	events.Publish(evtReceiveStatus(ctx, gtw.GatewayIdentifiers, nil))
	gsMetrics.statusReceived.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerReceiveUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage) {
	events.Publish(evtReceiveUp(ctx, gtw.GatewayIdentifiers, nil))
	gsMetrics.uplinkReceived.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerForwardUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, peer cluster.Peer) {
	events.Publish(evtForwardUp(ctx, gtw.GatewayIdentifiers, nil))
	gsMetrics.uplinkForwarded.WithLabelValues(ctx, peer.Name()).Inc()
}

func registerDropUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, err error) {
	events.Publish(evtDropUp(ctx, gtw.GatewayIdentifiers, err))
	if ttnErr, ok := errors.From(err); ok {
		gsMetrics.uplinkDropped.WithLabelValues(ctx, ttnErr.String()).Inc()
	} else {
		gsMetrics.uplinkDropped.WithLabelValues(ctx, unknown).Inc()
	}
}

func registerSendDownlink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.DownlinkMessage) {
	events.Publish(evtSendDown(ctx, msg.EndDeviceIdentifiers, nil))
	gsMetrics.downlinkSent.WithLabelValues(ctx, gtw.GatewayID).Inc()
}
