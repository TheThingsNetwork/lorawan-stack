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
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtStartGatewayLink = events.Define("gs.gateway.start_link", "start gateway link")
	evtEndGatewayLink   = events.Define("gs.gateway.end_link", "end gateway link")

	evtReceiveUp     = events.Define("gs.up.receive", "receive uplink message")
	evtReceiveStatus = events.Define("gs.status.receive", "receive status message")
	evtSendDown      = events.Define("gs.down.send", "send downlink message")
)

const (
	subsystem = "gs"
	gatewayID = "gateway_id"
)

var gsMetrics = &messageMetrics{
	gatewayLinks: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "gateway_links",
			Help:      "Number of gateway links",
		},
		[]string{gatewayID},
	),
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of received uplinks",
		},
		[]string{gatewayID},
	),
	statusReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "status_received_total",
			Help:      "Total number of received statuses",
		},
		[]string{gatewayID},
	),
	downlinkSent: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_sent_total",
			Help:      "Total number of sent downlinks",
		},
		[]string{gatewayID},
	),
}

func init() {
	metrics.MustRegister(gsMetrics)
}

type messageMetrics struct {
	gatewayLinks   *metrics.ContextualGaugeVec
	uplinkReceived *metrics.ContextualCounterVec
	statusReceived *metrics.ContextualCounterVec
	downlinkSent   *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.gatewayLinks.Describe(ch)
	m.uplinkReceived.Describe(ch)
	m.statusReceived.Describe(ch)
	m.downlinkSent.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.gatewayLinks.Collect(ch)
	m.uplinkReceived.Collect(ch)
	m.statusReceived.Collect(ch)
	m.downlinkSent.Collect(ch)
}

func registerStartGatewayLink(ctx context.Context, gtw ttnpb.GatewayIdentifiers) {
	events.Publish(evtStartGatewayLink(ctx, gtw, nil))
	gsMetrics.gatewayLinks.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerEndGatewayLink(ctx context.Context, gtw ttnpb.GatewayIdentifiers) {
	events.Publish(evtEndGatewayLink(ctx, gtw, nil))
	gsMetrics.gatewayLinks.WithLabelValues(ctx, gtw.GatewayID).Dec()
}

func registerReceiveUplink(ctx context.Context, gtw ttnpb.GatewayIdentifiers, msg *ttnpb.UplinkMessage) {
	events.Publish(evtReceiveUp(ctx, ttnpb.CombineIdentifiers(gtw, msg.EndDeviceIdentifiers), nil))
	gsMetrics.uplinkReceived.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerReceiveStatus(ctx context.Context, gtw ttnpb.GatewayIdentifiers, msg *ttnpb.GatewayStatus) {
	events.Publish(evtReceiveStatus(ctx, gtw, nil))
	gsMetrics.statusReceived.WithLabelValues(ctx, gtw.GatewayID).Inc()
}

func registerSendDownlink(ctx context.Context, gtw ttnpb.GatewayIdentifiers, msg *ttnpb.DownlinkMessage) {
	events.Publish(evtSendDown(ctx, ttnpb.CombineIdentifiers(gtw, msg.EndDeviceIdentifiers), nil))
	gsMetrics.downlinkSent.WithLabelValues(ctx, gtw.GatewayID).Inc()
}
