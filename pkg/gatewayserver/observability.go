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

package gatewayserver

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtGatewayConnect = events.Define(
		"gs.gateway.connect", "connect gateway",
		ttnpb.RIGHT_GATEWAY_LINK,
		ttnpb.RIGHT_GATEWAY_STATUS_READ,
	)
	evtGatewayDisconnect = events.Define(
		"gs.gateway.disconnect", "disconnect gateway",
		ttnpb.RIGHT_GATEWAY_LINK,
		ttnpb.RIGHT_GATEWAY_STATUS_READ,
	)
	evtReceiveStatus = events.Define(
		"gs.status.receive", "receive gateway status",
		ttnpb.RIGHT_GATEWAY_STATUS_READ,
	)
	evtReceiveUp = events.Define(
		"gs.up.receive", "receive uplink message",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
	evtDropUp = events.Define(
		"gs.up.drop", "drop uplink message",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
	evtForwardUp = events.Define(
		"gs.up.forward", "forward uplink message",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
	evtFailUp = events.Define(
		"gs.up.fail", "fail to handle uplink message",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
	evtSendDown = events.Define(
		"gs.down.send", "send downlink message",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
	evtTxSuccessDown = events.Define(
		"gs.down.tx.success", "transmit downlink message successful",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
	evtTxFailureDown = events.Define(
		"gs.down.tx.fail", "transmit downlink message failure",
		ttnpb.RIGHT_GATEWAY_TRAFFIC_READ,
	)
)

const (
	subsystem     = "gs"
	unknown       = "unknown"
	gatewayID     = "gateway_id"
	networkServer = "network_server"
)

var gsMetrics = &messageMetrics{
	gatewaysConnected: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "connected_gateways",
			Help:      "Number of currently connected gateways",
		},
		[]string{gatewayID},
	),
	statusReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "status_received_total",
			Help:      "Total number of received gateway statuses",
		},
		[]string{gatewayID},
	),
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of received uplinks",
		},
		[]string{networkServer, gatewayID},
	),
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of forwarded uplinks",
		},
		[]string{networkServer},
	),
	uplinkDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_dropped_total",
			Help:      "Total number of dropped uplinks",
		},
		[]string{networkServer, "error"},
	),
	uplinkFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_failed_total",
			Help:      "Total number of failed uplinks",
		},
		[]string{networkServer},
	),
	downlinkSent: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_sent_total",
			Help:      "Total number of sent downlinks",
		},
		[]string{gatewayID},
	),
	downlinkTxSucceeded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_tx_success_total",
			Help:      "Total number of successfully emitted downlinks",
		},
		[]string{gatewayID},
	),
	downlinkTxFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_tx_failed_total",
			Help:      "Total number of unsuccessfully emitted downlinks",
		},
		[]string{gatewayID},
	),
}

func init() {
	metrics.MustRegister(gsMetrics)
}

type messageMetrics struct {
	gatewaysConnected   *metrics.ContextualGaugeVec
	statusReceived      *metrics.ContextualCounterVec
	uplinkReceived      *metrics.ContextualCounterVec
	uplinkForwarded     *metrics.ContextualCounterVec
	uplinkDropped       *metrics.ContextualCounterVec
	uplinkFailed        *metrics.ContextualCounterVec
	downlinkSent        *metrics.ContextualCounterVec
	downlinkTxSucceeded *metrics.ContextualCounterVec
	downlinkTxFailed    *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.gatewaysConnected.Describe(ch)
	m.statusReceived.Describe(ch)
	m.uplinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.uplinkFailed.Describe(ch)
	m.downlinkSent.Describe(ch)
	m.downlinkTxSucceeded.Describe(ch)
	m.downlinkTxFailed.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.gatewaysConnected.Collect(ch)
	m.statusReceived.Collect(ch)
	m.uplinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.uplinkFailed.Collect(ch)
	m.downlinkSent.Collect(ch)
	m.downlinkTxSucceeded.Collect(ch)
	m.downlinkTxFailed.Collect(ch)
}

func registerGatewayConnect(ctx context.Context, ids ttnpb.GatewayIdentifiers) {
	events.Publish(evtGatewayConnect(ctx, ids, nil))
	gsMetrics.gatewaysConnected.WithLabelValues(ctx, ids.GatewayID).Inc()
}

func registerGatewayDisconnect(ctx context.Context, ids ttnpb.GatewayIdentifiers) {
	events.Publish(evtGatewayDisconnect(ctx, ids, nil))
	gsMetrics.gatewaysConnected.WithLabelValues(ctx, ids.GatewayID).Dec()
}

func registerReceiveStatus(ctx context.Context, gtw *ttnpb.Gateway, status *ttnpb.GatewayStatus) {
	events.Publish(evtReceiveStatus(ctx, gtw, status))
	gsMetrics.statusReceived.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerReceiveUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, ns string) {
	events.Publish(evtReceiveUp(ctx, gtw, msg))
	gsMetrics.uplinkReceived.WithLabelValues(ctx, ns, gtw.GatewayID).Inc()
}

func registerForwardUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, ns string) {
	events.Publish(evtForwardUp(ctx, gtw, nil))
	gsMetrics.uplinkForwarded.WithLabelValues(ctx, ns).Inc()
}

func registerDropUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, ns string, err error) {
	events.Publish(evtDropUp(ctx, gtw, err))
	if ttnErr, ok := errors.From(err); ok {
		gsMetrics.uplinkDropped.WithLabelValues(ctx, ns, ttnErr.FullName()).Inc()
	} else {
		gsMetrics.uplinkDropped.WithLabelValues(ctx, ns, unknown).Inc()
	}
}

func registerFailUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, ns string) {
	events.Publish(evtFailUp(ctx, gtw, nil))
	gsMetrics.uplinkFailed.WithLabelValues(ctx, ns).Inc()
}

func registerSendDownlink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.DownlinkMessage) {
	events.Publish(evtSendDown(ctx, gtw, msg))
	gsMetrics.downlinkSent.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerSuccessDownlink(ctx context.Context, gtw *ttnpb.Gateway) {
	events.Publish(evtTxSuccessDown(ctx, gtw, nil))
	gsMetrics.downlinkSent.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerFailDownlink(ctx context.Context, gtw *ttnpb.Gateway, ack *ttnpb.TxAcknowledgment) {
	events.Publish(evtTxFailureDown(ctx, gtw, ack.Result))
	gsMetrics.downlinkTxFailed.WithLabelValues(ctx, gtw.GatewayID).Inc()
}
