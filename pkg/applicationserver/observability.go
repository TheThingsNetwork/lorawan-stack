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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/peer"
)

var (
	evtReceiveDataUp = events.Define(
		"as.up.data.receive", "receive uplink data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
	)
	evtDropDataUp = events.Define(
		"as.up.data.drop", "drop uplink data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtForwardDataUp = events.Define(
		"as.up.data.forward", "forward uplink data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUp{}),
	)
	evtDecodeFailDataUp = events.Define(
		"as.up.data.decode.fail", "decode uplink data message failure",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtDecodeWarningDataUp = events.Define(
		"as.up.data.decode.warning", "decode uplink data message warning",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUplink{}),
	)
	evtReceiveJoinAccept = events.Define(
		"as.up.join.receive", "receive join-accept message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
	)
	evtDropJoinAccept = events.Define(
		"as.up.join.drop", "drop join-accept message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtForwardJoinAccept = events.Define(
		"as.up.join.forward", "forward join-accept message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUp{}),
	)
	evtForwardLocationSolved = events.Define(
		"as.up.location.forward", "forward location solved message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUp{}),
	)
	evtForwardServiceData = events.Define(
		"as.up.service.forward", "forward service data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationUp{}),
	)
	evtReceiveDataDown = events.Define(
		"as.down.data.receive", "receive downlink data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationDownlink{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDropDataDown = events.Define(
		"as.down.data.drop", "drop downlink data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtForwardDataDown = events.Define(
		"as.down.data.forward", "forward downlink data message",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationDownlink{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtEncodeFailDataDown = events.Define(
		"as.down.data.encode.fail", "encode downlink data message failure",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtEncodeWarningDataDown = events.Define(
		"as.down.data.encode.warning", "encode downlink data message warning",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationDownlink{}),
	)
	evtDecodeFailDataDown = events.Define(
		"as.down.data.decode.fail", "decode downlink data message failure",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
	evtDecodeWarningDataDown = events.Define(
		"as.down.data.decode.warning", "decode downlink data message warning",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithDataType(&ttnpb.ApplicationDownlink{}),
	)
)

const (
	subsystem     = "as"
	unknown       = "unknown"
	networkServer = "network_server"
	protocol      = "protocol"
	applicationID = "application_id"
)

var asMetrics = &messageMetrics{
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
	nsAsUplinkLatency: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "ns_as",
			Name:      "uplink_latency_seconds",
			Help:      "Histogram of uplink latency (seconds) between the Network Server and Application Server, including deduplication",
			Buckets:   []float64{0.2, 0.25, 0.3, 0.4, 0.5, 0.6, 0.8, 1.0, 2.0},
		},
		nil,
	),
	gtwAsUplinkLatency: metrics.NewContextualHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "gtw_as",
			Name:      "uplink_latency_seconds",
			Help:      "Histogram of uplink latency (seconds) between the Gateway and Application Server",
			Buckets:   []float64{0.2, 0.4, 0.6, 0.8, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 5.0},
		},
		nil,
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
	uplinkReceived     *metrics.ContextualCounterVec
	uplinkForwarded    *metrics.ContextualCounterVec
	uplinkDropped      *metrics.ContextualCounterVec
	nsAsUplinkLatency  *metrics.ContextualHistogramVec
	gtwAsUplinkLatency *metrics.ContextualHistogramVec
	downlinkReceived   *metrics.ContextualCounterVec
	downlinkForwarded  *metrics.ContextualCounterVec
	downlinkDropped    *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.uplinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.uplinkDropped.Describe(ch)
	m.nsAsUplinkLatency.Describe(ch)
	m.gtwAsUplinkLatency.Describe(ch)
	m.downlinkReceived.Describe(ch)
	m.downlinkForwarded.Describe(ch)
	m.downlinkDropped.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.uplinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.uplinkDropped.Collect(ch)
	m.nsAsUplinkLatency.Collect(ch)
	m.gtwAsUplinkLatency.Collect(ch)
	m.downlinkReceived.Collect(ch)
	m.downlinkForwarded.Collect(ch)
	m.downlinkDropped.Collect(ch)
}

func registerReceiveUp(ctx context.Context, msg *ttnpb.ApplicationUp) {
	ns := "application"
	if peer, ok := peer.FromContext(ctx); ok {
		ns = peer.Addr.String()
	}
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtReceiveJoinAccept.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, nil))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtReceiveDataUp.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, nil))
	default:
		return
	}
	asMetrics.uplinkReceived.WithLabelValues(ctx, ns).Inc()
}

func registerForwardUp(ctx context.Context, msg *ttnpb.ApplicationUp) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtForwardJoinAccept.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, msg))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtForwardDataUp.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, msg))
	case *ttnpb.ApplicationUp_LocationSolved:
		events.Publish(evtForwardLocationSolved.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, msg))
	case *ttnpb.ApplicationUp_ServiceData:
		events.Publish(evtForwardServiceData.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, msg))
	default:
		return
	}
	asMetrics.uplinkForwarded.WithLabelValues(ctx, msg.EndDeviceIds.ApplicationIds.ApplicationId).Inc()
}

func registerDropUp(ctx context.Context, msg *ttnpb.ApplicationUp, err error) {
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		events.Publish(evtDropJoinAccept.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, err))
	case *ttnpb.ApplicationUp_UplinkMessage:
		events.Publish(evtDropDataUp.NewWithIdentifiersAndData(ctx, msg.EndDeviceIds, err))
	default:
		return
	}
	if ttnErr, ok := errors.From(err); ok {
		asMetrics.uplinkDropped.WithLabelValues(ctx, ttnErr.FullName()).Inc()
	} else {
		asMetrics.uplinkDropped.WithLabelValues(ctx, unknown).Inc()
	}
}

func registerUplinkLatency(ctx context.Context, msg *ttnpb.ApplicationUplink) {
	asMetrics.nsAsUplinkLatency.WithLabelValues(ctx).Observe(time.Since(*ttnpb.StdTime(msg.ReceivedAt)).Seconds())
	for _, meta := range msg.RxMetadata {
		if stdTime := ttnpb.StdTime(meta.Time); meta.Time != nil && !stdTime.IsZero() {
			asMetrics.gtwAsUplinkLatency.WithLabelValues(ctx).Observe(time.Since(*stdTime).Seconds())
		}
	}
}

func registerReceiveDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink) {
	events.Publish(evtReceiveDataDown.NewWithIdentifiersAndData(ctx, ids, msg))
	asMetrics.downlinkReceived.WithLabelValues(ctx, ids.ApplicationIds.ApplicationId).Inc()
}

func registerReceiveDownlinks(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) {
	for _, item := range items {
		registerReceiveDownlink(ctx, ids, item)
	}
}

func registerForwardDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, ns string) {
	events.Publish(evtForwardDataDown.NewWithIdentifiersAndData(ctx, ids, msg))
	asMetrics.downlinkForwarded.WithLabelValues(ctx, ns).Inc()
}

func registerDropDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, err error) {
	events.Publish(evtDropDataDown.NewWithIdentifiersAndData(ctx, ids, err))
	if ttnErr, ok := errors.From(err); ok {
		asMetrics.downlinkDropped.WithLabelValues(ctx, ttnErr.FullName()).Inc()
	} else {
		asMetrics.downlinkDropped.WithLabelValues(ctx, unknown).Inc()
	}
}

func (as *ApplicationServer) registerDropDownlinks(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink, err error) {
	var errorDetails ttnpb.ErrorDetails
	if ttnErr, ok := err.(errors.ErrorDetails); ok {
		errorDetails = *ttnpb.ErrorDetailsToProto(ttnErr)
	}
	for _, item := range items {
		if err := as.publishUp(ctx, &ttnpb.ApplicationUp{
			EndDeviceIds:   ids,
			CorrelationIds: item.CorrelationIds,
			Up: &ttnpb.ApplicationUp_DownlinkFailed{
				DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
					Downlink: item,
					Error:    &errorDetails,
				},
			},
		}); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to send upstream message")
		}
		registerDropDownlink(ctx, ids, item, err)
	}
}

func (as *ApplicationServer) registerForwardDownlinks(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink, peerName string) {
	for _, item := range items {
		if err := as.publishUp(ctx, &ttnpb.ApplicationUp{
			EndDeviceIds:   ids,
			CorrelationIds: item.CorrelationIds,
			Up: &ttnpb.ApplicationUp_DownlinkQueued{
				DownlinkQueued: item,
			},
		}); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to send upstream message")
		}
		registerForwardDownlink(ctx, ids, item, peerName)
	}
}
