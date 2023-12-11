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
	"strings"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	logNamespace    = "networkserver"
	tracerNamespace = "go.thethings.network/lorawan-stack/pkg/networkserver"
)

var (
	evtReceiveDataUplink = events.Define(
		"ns.up.data.receive", "receive data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.UplinkMessage{}),
	)
	evtDropDataUplink = events.Define(
		"ns.up.data.drop", "drop data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtProcessDataUplink = events.Define(
		"ns.up.data.process", "successfully processed data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.UplinkMessage{}),
	)
	evtForwardDataUplink = events.Define(
		"ns.up.data.forward", "forward data message to Application Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUp{
			Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{}},
		}),
	)
	evtDropRelayUplink = events.Define(
		"ns.up.relay.drop", "drop relay message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtProcessRelayUplink = events.Define(
		"ns.up.relay.process", "successfully processed relay message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.UplinkMessage{}),
	)
	evtScheduleDataDownlinkAttempt = events.Define(
		"ns.down.data.schedule.attempt", "schedule data downlink for transmission on Gateway Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.DownlinkMessage{}),
	)
	evtScheduleDataDownlinkSuccess = events.Define(
		"ns.down.data.schedule.success", "successfully scheduled data downlink for transmission on Gateway Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ScheduleDownlinkResponse{}),
	)
	evtScheduleDataDownlinkFail = events.Define(
		"ns.down.data.schedule.fail", "schedule data downlink for transmission on Gateway Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
	evtReceiveJoinRequest = events.Define(
		"ns.up.join.receive", "receive join-request",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.UplinkMessage{}),
	)
	evtDropJoinRequest = events.Define(
		"ns.up.join.drop", "drop join-request",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtProcessJoinRequest = events.Define(
		"ns.up.join.process", "successfully processed join-request",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.UplinkMessage{}),
		events.WithPropagateToParent(),
	)
	evtClusterJoinAttempt = events.Define(
		"ns.up.join.cluster.attempt", "send join-request to cluster-local Join Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.JoinRequest{}),
	)
	evtClusterJoinSuccess = events.Define(
		"ns.up.join.cluster.success", "join-request to cluster-local Join Server succeeded",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.JoinResponse{}),
	)
	evtClusterJoinFail = events.Define(
		"ns.up.join.cluster.fail", "join-request to cluster-local Join Server failed",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
	evtInteropJoinAttempt = events.Define(
		"ns.up.join.interop.attempt", "forward join-request to interoperability Join Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.JoinRequest{}),
	)
	evtInteropJoinSuccess = events.Define(
		"ns.up.join.interop.success", "join-request to interoperability Join Server succeeded",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.JoinResponse{}),
	)
	evtInteropJoinFail = events.Define(
		"ns.up.join.interop.fail", "join-request to interoperability Join Server failed",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
	evtForwardJoinAccept = events.Define(
		"ns.up.join.accept.forward", "forward join-accept to Application Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUp{
			Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{}},
		}),
	)
	evtScheduleJoinAcceptAttempt = events.Define(
		"ns.down.join.schedule.attempt", "schedule join-accept for transmission on Gateway Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.DownlinkMessage{}),
	)
	evtScheduleJoinAcceptSuccess = events.Define(
		"ns.down.join.schedule.success", "successfully scheduled join-accept for transmission on Gateway Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ScheduleDownlinkResponse{}),
	)
	evtScheduleJoinAcceptFail = events.Define(
		"ns.down.join.schedule.fail", "schedule join-accept for transmission on Gateway Server",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
	evtTransmissionSuccess = events.Define(
		"ns.down.transmission.success", "downlink successfully transmitted",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.DownlinkMessage{}),
	)
	evtTransmissionFail = events.Define(
		"ns.down.transmission.fail", "transmit downlink",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
	evtRXParametersFail = events.Define(
		"ns.down.rx.parameters.fail", "compute RX parameters",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
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
	uplinkDuplicates: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_duplicates_total",
			Help:      "Total number of duplicate uplinks",
		},
		[]string{messageType},
	),
	uplinkProcessed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_processed_total",
			Help:      "Total number of processed uplinks",
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
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of forwarded uplinks",
		},
		[]string{messageType},
	),
	uplinkGateways: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "uplink_gateways",
			Help:      "Number of gateways that forwarded the uplink (within the deduplication window)",
			Buckets:   []float64{1, 2, 3, 4, 5, 10},
		},
		nil,
	),
	gsNsUplinkLatency: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "gs_ns",
			Name:      "uplink_latency_seconds",
			Help:      "Histogram of uplink latency (seconds) between the Gateway Server (or Packet Broker Agent) and Network Server",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		nil,
	),

	matchCandidatesPerUplink: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "match_candidates_uplink_total",
			Help:      "Histogram of device match candidates per data uplink",
			Buckets:   []float64{0.0, 1.0, 2.0, 4.0, 8.0, 16.0, 32.0},
		},
		nil,
	),
	micComputationsPerUplink: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "mic_computations_uplink_total",
			Help:      "Histogram of MIC computations per data uplink",
			Buckets:   []float64{0.0, 1.0, 2.0, 4.0, 8.0, 16.0, 32.0},
		},
		nil,
	),

	downlinkAttempted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_attempted_total",
			Help:      "Total number of attempted downlinks",
		},
		[]string{messageType},
	),
	downlinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_forwarded_total",
			Help:      "Total number of forwarded downlinks",
		},
		[]string{messageType},
	),
}

func init() {
	metrics.MustRegister(nsMetrics)
}

type messageMetrics struct {
	uplinkReceived    *metrics.ContextualCounterVec
	uplinkDuplicates  *metrics.ContextualCounterVec
	uplinkProcessed   *metrics.ContextualCounterVec
	uplinkForwarded   *metrics.ContextualCounterVec
	uplinkDropped     *metrics.ContextualCounterVec
	uplinkGateways    *metrics.ContextualHistogramVec
	gsNsUplinkLatency *metrics.ContextualHistogramVec

	matchCandidatesPerUplink *metrics.ContextualHistogramVec
	micComputationsPerUplink *metrics.ContextualHistogramVec

	downlinkAttempted *metrics.ContextualCounterVec
	downlinkForwarded *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.uplinkReceived.Describe(ch)
	m.uplinkDuplicates.Describe(ch)
	m.uplinkProcessed.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.uplinkGateways.Describe(ch)
	m.gsNsUplinkLatency.Describe(ch)

	m.matchCandidatesPerUplink.Describe(ch)
	m.micComputationsPerUplink.Describe(ch)

	m.downlinkAttempted.Describe(ch)
	m.downlinkForwarded.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.uplinkReceived.Collect(ch)
	m.uplinkDuplicates.Collect(ch)
	m.uplinkProcessed.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.uplinkGateways.Collect(ch)
	m.gsNsUplinkLatency.Collect(ch)

	m.matchCandidatesPerUplink.Collect(ch)
	m.micComputationsPerUplink.Collect(ch)

	m.downlinkAttempted.Collect(ch)
	m.downlinkForwarded.Collect(ch)
}

func mTypeLabel(mType ttnpb.MType) string {
	return strings.ToLower(mType.String())
}

func registerReceiveUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkReceived.WithLabelValues(ctx, mTypeLabel(msg.Payload.MHdr.MType)).Inc()
}

func registerReceiveDuplicateUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkDuplicates.WithLabelValues(ctx, mTypeLabel(msg.Payload.MHdr.MType)).Inc()
}

func registerProcessUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkProcessed.WithLabelValues(ctx, mTypeLabel(msg.Payload.MHdr.MType)).Inc()
}

func registerForwardDataUplink(ctx context.Context, msg *ttnpb.ApplicationUplink) {
	mType := ttnpb.MType_UNCONFIRMED_UP
	if msg.Confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, mTypeLabel(mType)).Inc()
}

func registerForwardJoinRequest(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, mTypeLabel(msg.Payload.MHdr.MType)).Inc()
}

func registerDropUplink(ctx context.Context, msg *ttnpb.UplinkMessage, err error) {
	cause := unknown
	if ttnErr, ok := errors.From(err); ok {
		cause = ttnErr.FullName()
	}
	nsMetrics.uplinkDropped.WithLabelValues(ctx, mTypeLabel(msg.Payload.MHdr.MType), cause).Inc()
}

func registerUplinkLatency(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.gsNsUplinkLatency.WithLabelValues(ctx).Observe(time.Since(*ttnpb.StdTime(msg.ReceivedAt)).Seconds())
}

func registerMergeMetadata(ctx context.Context, msg *ttnpb.UplinkMessage) {
	gtwCount, _ := RXMetadataStats(ctx, msg.RxMetadata)
	nsMetrics.uplinkGateways.WithLabelValues(ctx).Observe(float64(gtwCount))
}

type matchStatsKeyType struct{}

var matchStatsKey matchStatsKeyType

type matchStats struct {
	matchCandidates int64
	micComputations int64
}

func newContextWithMatchStats(ctx context.Context) (context.Context, func()) {
	stats := &matchStats{}
	return context.WithValue(ctx, matchStatsKey, stats), func() {
		nsMetrics.matchCandidatesPerUplink.WithLabelValues(ctx).Observe(float64(stats.matchCandidates))
		nsMetrics.micComputationsPerUplink.WithLabelValues(ctx).Observe(float64(stats.micComputations))
	}
}

func registerMatchStats(ctx context.Context, f func(*matchStats)) {
	if stats, ok := ctx.Value(matchStatsKey).(*matchStats); ok {
		f(stats)
		return
	}
	panic("match stats not found in context")
}

func registerMatchCandidate(ctx context.Context) {
	registerMatchStats(ctx, func(stats *matchStats) {
		atomic.AddInt64(&stats.matchCandidates, 1)
	})
}

func registerMICComputation(ctx context.Context) {
	registerMatchStats(ctx, func(stats *matchStats) {
		atomic.AddInt64(&stats.micComputations, 1)
	})
}

var (
	unconfirmedDownlinkMTypeLabel = mTypeLabel(ttnpb.MType_UNCONFIRMED_DOWN)
	confirmedDownlinkMTypeLabel   = mTypeLabel(ttnpb.MType_CONFIRMED_DOWN)
	joinAcceptDownlinkMTypeLabel  = mTypeLabel(ttnpb.MType_JOIN_ACCEPT)
)

func registerAttemptUnconfirmedDataDownlink(ctx context.Context) {
	nsMetrics.downlinkAttempted.WithLabelValues(ctx, unconfirmedDownlinkMTypeLabel).Inc()
}

func registerAttemptConfirmedDataDownlink(ctx context.Context) {
	nsMetrics.downlinkAttempted.WithLabelValues(ctx, confirmedDownlinkMTypeLabel).Inc()
}

func registerAttemptJoinAcceptDownlink(ctx context.Context) {
	nsMetrics.downlinkAttempted.WithLabelValues(ctx, joinAcceptDownlinkMTypeLabel).Inc()
}

func registerForwardUnconfirmedDataDownlink(ctx context.Context) {
	nsMetrics.downlinkForwarded.WithLabelValues(ctx, unconfirmedDownlinkMTypeLabel).Inc()
}

func registerForwardConfirmedDataDownlink(ctx context.Context) {
	nsMetrics.downlinkForwarded.WithLabelValues(ctx, confirmedDownlinkMTypeLabel).Inc()
}

func registerForwardJoinAcceptDownlink(ctx context.Context) {
	nsMetrics.downlinkForwarded.WithLabelValues(ctx, joinAcceptDownlinkMTypeLabel).Inc()
}
