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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtGatewayConnect = events.Define(
		"gs.gateway.connect", "connect gateway",
		events.WithVisibility(
			ttnpb.Right_RIGHT_GATEWAY_LINK,
			ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtGatewayDisconnect = events.Define(
		"gs.gateway.disconnect", "disconnect gateway",
		events.WithVisibility(
			ttnpb.Right_RIGHT_GATEWAY_LINK,
			ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
		),
		events.WithErrorDataType(),
	)
	evtReceiveStatus = events.Define(
		"gs.status.receive", "receive gateway status",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
		events.WithDataType(&ttnpb.GatewayStatus{}),
	)
	evtDropStatus = events.Define(
		"gs.status.drop", "drop gateway status",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
		events.WithErrorDataType(),
	)
	evtReceiveUp = events.Define(
		"gs.up.receive", "receive uplink message",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
		events.WithDataType(&ttnpb.UplinkMessage{}),
	)
	evtDropUp = events.Define(
		"gs.up.drop", "drop uplink message",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtForwardUp = events.Define(
		"gs.up.forward", "forward uplink message",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
	)
	evtSendDown = events.Define(
		"gs.down.send", "send downlink message",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
		events.WithDataType(&ttnpb.DownlinkMessage{}),
	)
	evtTxSuccessDown = events.Define(
		"gs.down.tx.success", "transmit downlink message successful",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
	)
	evtTxFailureDown = events.Define(
		"gs.down.tx.fail", "transmit downlink message failure",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
		events.WithDataType(ttnpb.TxAcknowledgment_COLLISION_PACKET),
	)
	evtReceiveTxAck = events.Define(
		"gs.txack.receive", "receive transmission acknowledgement",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
		events.WithDataType(&ttnpb.TxAcknowledgment{}),
	)
	evtDropTxAck = events.Define(
		"gs.txack.drop", "drop transmission acknowledgement",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
		events.WithErrorDataType(),
	)
	evtForwardTxAck = events.Define(
		"gs.txack.forward", "forward transmission acknowledgement",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ),
	)
)

const (
	subsystem = "gs"
	unknown   = "unknown"
	protocol  = "protocol"
	gatewayID = "gateway_id"
	host      = "host"
)

var gsMetrics = &messageMetrics{
	gatewaysConnected: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "connected_gateways",
			Help:      "Number of currently connected gateways",
		},
		[]string{protocol},
	),
	statusReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "status_received_total",
			Help:      "Total number of received gateway statuses",
		},
		[]string{protocol},
	),
	statusForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "status_forwarded_total",
			Help:      "Total number of forwarded gateway statuses",
		},
		[]string{host},
	),
	statusDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "status_dropped_total",
			Help:      "Total number of dropped gateway statuses",
		},
		[]string{host, "error"},
	),
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of received uplinks",
		},
		[]string{protocol},
	),
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of forwarded uplinks",
		},
		[]string{host},
	),
	uplinkDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_dropped_total",
			Help:      "Total number of dropped uplinks",
		},
		[]string{host, "error"},
	),
	downlinkSent: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_sent_total",
			Help:      "Total number of sent downlinks",
		},
		[]string{protocol},
	),
	downlinkTxSucceeded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_tx_success_total",
			Help:      "Total number of successfully emitted downlinks",
		},
		[]string{protocol},
	),
	downlinkTxFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_tx_failed_total",
			Help:      "Total number of unsuccessfully emitted downlinks",
		},
		[]string{protocol, "result"},
	),
	txAckReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "txack_received_total",
			Help:      "Total number of received gateway transmission acknowledgements",
		},
		[]string{protocol},
	),
	txAckForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "txack_forwarded_total",
			Help:      "Total number of forwarded gateway transmission acknowledgements",
		},
		[]string{host},
	),
	txAckDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "txack_dropped_total",
			Help:      "Total number of dropped gateway transmission acknowledgements",
		},
		[]string{host, "error"},
	),
}

func init() {
	metrics.MustRegister(gsMetrics)
}

type messageMetrics struct {
	gatewaysConnected   *metrics.ContextualGaugeVec
	statusReceived      *metrics.ContextualCounterVec
	statusForwarded     *metrics.ContextualCounterVec
	statusDropped       *metrics.ContextualCounterVec
	uplinkReceived      *metrics.ContextualCounterVec
	uplinkForwarded     *metrics.ContextualCounterVec
	uplinkDropped       *metrics.ContextualCounterVec
	downlinkSent        *metrics.ContextualCounterVec
	downlinkTxSucceeded *metrics.ContextualCounterVec
	downlinkTxFailed    *metrics.ContextualCounterVec
	txAckReceived       *metrics.ContextualCounterVec
	txAckForwarded      *metrics.ContextualCounterVec
	txAckDropped        *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.gatewaysConnected.Describe(ch)
	m.statusReceived.Describe(ch)
	m.statusForwarded.Describe(ch)
	m.statusDropped.Describe(ch)
	m.uplinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.downlinkSent.Describe(ch)
	m.downlinkTxSucceeded.Describe(ch)
	m.downlinkTxFailed.Describe(ch)
	m.txAckReceived.Describe(ch)
	m.txAckForwarded.Describe(ch)
	m.txAckDropped.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.gatewaysConnected.Collect(ch)
	m.statusReceived.Collect(ch)
	m.statusForwarded.Collect(ch)
	m.statusDropped.Collect(ch)
	m.uplinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.downlinkSent.Collect(ch)
	m.downlinkTxSucceeded.Collect(ch)
	m.downlinkTxFailed.Collect(ch)
	m.txAckReceived.Collect(ch)
	m.txAckForwarded.Collect(ch)
	m.txAckDropped.Collect(ch)
}

func registerGatewayConnect(ctx context.Context, ids ttnpb.GatewayIdentifiers, protocol string) {
	events.Publish(evtGatewayConnect.NewWithIdentifiersAndData(ctx, &ids, nil))
	gsMetrics.gatewaysConnected.WithLabelValues(ctx, protocol).Inc()
}

func registerGatewayDisconnect(ctx context.Context, ids ttnpb.GatewayIdentifiers, protocol string, err error) {
	events.Publish(evtGatewayDisconnect.NewWithIdentifiersAndData(ctx, &ids, err))
	gsMetrics.gatewaysConnected.WithLabelValues(ctx, protocol).Dec()
}

func registerReceiveStatus(ctx context.Context, gtw *ttnpb.Gateway, status *ttnpb.GatewayStatus, protocol string) {
	events.Publish(evtReceiveStatus.NewWithIdentifiersAndData(ctx, gtw, status))
	gsMetrics.statusReceived.WithLabelValues(ctx, protocol).Inc()
}

func registerForwardStatus(ctx context.Context, gtw *ttnpb.Gateway, status *ttnpb.GatewayStatus, host string) {
	gsMetrics.statusForwarded.WithLabelValues(ctx, host).Inc()
}

func registerDropStatus(ctx context.Context, gtw *ttnpb.Gateway, status *ttnpb.GatewayStatus, host string, err error) {
	events.Publish(evtDropStatus.NewWithIdentifiersAndData(ctx, gtw, err))
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	gsMetrics.statusDropped.WithLabelValues(ctx, host, errorLabel).Inc()
}

func registerReceiveUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, protocol string) {
	events.Publish(evtReceiveUp.NewWithIdentifiersAndData(ctx, gtw, msg))
	gsMetrics.uplinkReceived.WithLabelValues(ctx, protocol).Inc()
}

func registerForwardUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.UplinkMessage, host string) {
	events.Publish(evtForwardUp.NewWithIdentifiersAndData(ctx, gtw, host))
	gsMetrics.uplinkForwarded.WithLabelValues(ctx, host).Inc()
}

func registerDropUplink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.GatewayUplinkMessage, host string, err error) {
	events.Publish(evtDropUp.NewWithIdentifiersAndData(ctx, gtw, err))
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	gsMetrics.uplinkDropped.WithLabelValues(ctx, host, errorLabel).Inc()
}

func registerSendDownlink(ctx context.Context, gtw *ttnpb.Gateway, msg *ttnpb.DownlinkMessage, protocol string) {
	events.Publish(evtSendDown.NewWithIdentifiersAndData(ctx, gtw, msg))
	gsMetrics.downlinkSent.WithLabelValues(ctx, protocol).Inc()
}

func registerSuccessDownlink(ctx context.Context, gtw *ttnpb.Gateway, protocol string) {
	events.Publish(evtTxSuccessDown.NewWithIdentifiersAndData(ctx, gtw, nil))
	gsMetrics.downlinkTxSucceeded.WithLabelValues(ctx, protocol).Inc()
}

func registerFailDownlink(ctx context.Context, gtw *ttnpb.Gateway, txAck *ttnpb.TxAcknowledgment, protocol string) {
	events.Publish(evtTxFailureDown.NewWithIdentifiersAndData(ctx, gtw, txAck.Result))
	gsMetrics.downlinkTxFailed.WithLabelValues(ctx, protocol, txAck.Result.String()).Inc()
}

func registerReceiveTxAck(ctx context.Context, gtw *ttnpb.Gateway, txAck *ttnpb.TxAcknowledgment, protocol string) {
	events.Publish(evtReceiveTxAck.NewWithIdentifiersAndData(ctx, gtw, txAck))
	gsMetrics.txAckReceived.WithLabelValues(ctx, protocol).Inc()
}

func registerForwardTxAck(ctx context.Context, gtw *ttnpb.Gateway, txAck *ttnpb.TxAcknowledgment, host string) {
	events.Publish(evtForwardTxAck.NewWithIdentifiersAndData(ctx, gtw, host))
	gsMetrics.txAckForwarded.WithLabelValues(ctx, host).Inc()
}

func registerDropTxAck(ctx context.Context, gtw *ttnpb.Gateway, txAck *ttnpb.TxAcknowledgment, host string, err error) {
	events.Publish(evtDropTxAck.NewWithIdentifiersAndData(ctx, gtw, err))
	errorLabel := unknown
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	}
	gsMetrics.txAckDropped.WithLabelValues(ctx, host, errorLabel).Inc()
}
