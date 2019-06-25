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

package networkserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

func defineReceiveMACAcceptEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.answer.accept", name), fmt.Sprintf("%s accept received", desc))
}

func defineReceiveMACAnswerEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.answer", name), fmt.Sprintf("%s answer received", desc))
}

func defineReceiveMACIndicationEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.indication", name), fmt.Sprintf("%s indication received", desc))
}

func defineReceiveMACRejectEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.answer.reject", name), fmt.Sprintf("%s rejection received", desc))
}

func defineReceiveMACRequestEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.request", name), fmt.Sprintf("%s request received", desc))
}

func defineEnqueueMACAnswerEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.answer", name), fmt.Sprintf("%s answer enqueued", desc))
}

func defineEnqueueMACConfirmationEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.confirmation", name), fmt.Sprintf("%s confirmation enqueued", desc))
}

func defineEnqueueMACRequestEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(fmt.Sprintf("ns.mac.%s.request", name), fmt.Sprintf("%s request enqueued", desc))
}

var (
	evtBeginApplicationLink = events.Define("ns.application.begin_link", "begin application link")
	evtEndApplicationLink   = events.Define("ns.application.end_link", "end application link")

	evtReceiveUplink          = events.Define("ns.up.receive", "receive uplink message")
	evtReceiveUplinkDuplicate = events.Define("ns.up.receive_duplicate", "receive duplicate uplink message")

	evtMergeMetadata = events.Define("ns.up.merge_metadata", "merge uplink message metadata")

	evtDropDataUplink    = events.Define("ns.up.data.drop", "drop data message")
	evtForwardDataUplink = events.Define("ns.up.data.forward", "forward data message")

	evtDropJoinRequest    = events.Define("ns.up.join.drop", "drop join-request")
	evtForwardJoinRequest = events.Define("ns.up.join.forward", "forward join-request")

	evtDropRejoinRequest    = events.Define("ns.up.rejoin.drop", "drop rejoin-request")
	evtForwardRejoinRequest = events.Define("ns.up.rejoin.forward", "forward rejoin-request")

	evtEnqueueProprietaryMACAnswer  = defineEnqueueMACAnswerEvent("proprietary", "proprietary MAC command")
	evtEnqueueProprietaryMACRequest = defineEnqueueMACRequestEvent("proprietary", "proprietary MAC command")
	evtReceiveProprietaryMAC        = events.Define("ns.mac.proprietary.receive", "proprietary MAC command received")
)

const (
	subsystem   = "ns"
	unknown     = "unknown"
	messageType = "message_type"
)

var nsMetrics = &messageMetrics{
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of received uplinks (and duplicates)",
		},
		[]string{messageType},
	),
	uplinkUniqueReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_unique_received_total",
			Help:      "Total number of received unique uplinks (without duplicates)",
		},
		[]string{messageType},
	),
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of forwarded uplinks",
		},
		[]string{messageType},
	),
	uplinkDropped: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_dropped_total",
			Help:      "Total number of dropped uplinks",
		},
		[]string{messageType, "error"},
	),
	uplinkGateways: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "uplink_gateways",
			Help:      "Number of gateways that forwarded the uplink (within the deduplication window)",
			Buckets:   []float64{1, 2, 3, 4, 5, 10, 20, 30, 40, 50},
		},
		[]string{},
	),
}

func init() {
	metrics.MustRegister(nsMetrics)
}

type messageMetrics struct {
	uplinkReceived       *metrics.ContextualCounterVec
	uplinkUniqueReceived *metrics.ContextualCounterVec
	uplinkForwarded      *metrics.ContextualCounterVec
	uplinkDropped        *metrics.ContextualCounterVec
	uplinkGateways       *metrics.ContextualHistogramVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.uplinkReceived.Describe(ch)
	m.uplinkUniqueReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.uplinkGateways.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.uplinkReceived.Collect(ch)
	m.uplinkUniqueReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.uplinkGateways.Collect(ch)
}

func uplinkMTypeLabel(msg *ttnpb.UplinkMessage) string {
	return strings.ToLower(msg.Payload.MType.String())
}

func registerReceiveUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkReceived.WithLabelValues(ctx, uplinkMTypeLabel(msg)).Inc()
	nsMetrics.uplinkUniqueReceived.WithLabelValues(ctx, uplinkMTypeLabel(msg)).Inc()
}

func registerReceiveUplinkDuplicate(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkReceived.WithLabelValues(ctx, uplinkMTypeLabel(msg)).Inc()
}

func registerMergeMetadata(ctx context.Context, msg *ttnpb.UplinkMessage) {
	gtws := make(map[string]struct{}, len(msg.RxMetadata))
	for _, md := range msg.RxMetadata {
		gtws[unique.ID(ctx, md.GatewayIdentifiers)] = struct{}{}
	}
	nsMetrics.uplinkGateways.WithLabelValues(ctx).Observe(float64(len(gtws)))
}

func registerForwardDataUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, uplinkMTypeLabel(msg)).Inc()
}

func registerForwardJoinRequest(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, uplinkMTypeLabel(msg)).Inc()
}

func registerForwardRejoinRequest(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, uplinkMTypeLabel(msg)).Inc()
}

func registerDropDataUplink(ctx context.Context, msg *ttnpb.UplinkMessage, err error) {
	if ttnErr, ok := errors.From(err); ok {
		nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg), ttnErr.FullName()).Inc()
	} else {
		nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg), unknown).Inc()
	}
}

func registerDropJoinRequest(ctx context.Context, msg *ttnpb.UplinkMessage, err error) {
	if ttnErr, ok := errors.From(err); ok {
		nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg), ttnErr.FullName()).Inc()
	} else {
		nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg), unknown).Inc()
	}
}

func registerDropRejoinRequest(ctx context.Context, msg *ttnpb.UplinkMessage, err error) {
	if ttnErr, ok := errors.From(err); ok {
		nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg), ttnErr.FullName()).Inc()
	} else {
		nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg), unknown).Inc()
	}
}
