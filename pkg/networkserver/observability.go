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
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func defineReceiveMACAcceptEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer.accept", name), fmt.Sprintf("%s accept received", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineReceiveMACAnswerEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer", name), fmt.Sprintf("%s answer received", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineReceiveMACIndicationEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.indication", name), fmt.Sprintf("%s indication received", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineReceiveMACRejectEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer.reject", name), fmt.Sprintf("%s rejection received", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineReceiveMACRequestEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.request", name), fmt.Sprintf("%s request received", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineEnqueueMACAnswerEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer", name), fmt.Sprintf("%s answer enqueued", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineEnqueueMACConfirmationEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.confirmation", name), fmt.Sprintf("%s confirmation enqueued", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineEnqueueMACRequestEvent(name, desc string) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.request", name), fmt.Sprintf("%s request enqueued", desc),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

func defineClassSwitchEvent(class rune) func() events.Definition {
	return events.DefineFunc(
		fmt.Sprintf("ns.class.switch.%c", class), fmt.Sprintf("switched to class %c", unicode.ToUpper(class)),
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
}

var (
	evtBeginApplicationLink = events.Define(
		"ns.application.link.begin", "begin application link",
		ttnpb.RIGHT_APPLICATION_LINK,
	)
	evtEndApplicationLink = events.Define(
		"ns.application.link.end", "end application link",
		ttnpb.RIGHT_APPLICATION_LINK,
	)
	evtReceiveDataUplink = events.Define(
		"ns.up.data.receive", "receive data message",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtDropDataUplink = events.Define(
		"ns.up.data.drop", "drop data message",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtProcessDataUplink = events.Define(
		"ns.up.data.process", "successfully processed data message",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtForwardDataUplink = events.Define(
		"ns.up.data.forward", "forward data message to Application Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtScheduleDataDownlinkAttempt = events.Define(
		"ns.down.data.schedule.attempt", "schedule data downlink for transmission on Gateway Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtScheduleDataDownlinkSuccess = events.Define(
		"ns.down.data.schedule.success", "successfully scheduled data downlink for transmission on Gateway Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtScheduleDataDownlinkFail = events.Define(
		"ns.down.data.schedule.fail", "failed to schedule data downlink for transmission on Gateway Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtReceiveJoinRequest = events.Define(
		"ns.up.join.receive", "receive join-request",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtDropJoinRequest = events.Define(
		"ns.up.join.drop", "drop join-request",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtProcessJoinRequest = events.Define(
		"ns.up.join.process", "successfully processed join-request",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtClusterJoinAttempt = events.Define(
		"ns.up.join.cluster.attempt", "send join-request to cluster-local Join Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtClusterJoinSuccess = events.Define(
		"ns.up.join.cluster.success", "join-request to cluster-local Join Server succeeded",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtClusterJoinFail = events.Define(
		"ns.up.join.cluster.fail", "join-request to cluster-local Join Server failed",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtInteropJoinAttempt = events.Define(
		"ns.up.join.interop.attempt", "forward join-request to interoperability Join Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtInteropJoinSuccess = events.Define(
		"ns.up.join.interop.success", "join-request to interoperability Join Server succeeded",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtInteropJoinFail = events.Define(
		"ns.up.join.interop.fail", "join-request to interoperability Join Server failed",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtForwardJoinAccept = events.Define(
		"ns.up.join.accept.forward", "forward join-accept to Application Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtScheduleJoinAcceptAttempt = events.Define(
		"ns.down.join.schedule.attempt", "schedule join-accept for transmission on Gateway Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtScheduleJoinAcceptSuccess = events.Define(
		"ns.down.join.schedule.success", "successfully scheduled join-accept for transmission on Gateway Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtScheduleJoinAcceptFail = events.Define(
		"ns.down.join.schedule.fail", "failed to schedule join-accept for transmission on Gateway Server",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)
	evtEnqueueProprietaryMACAnswer  = defineEnqueueMACAnswerEvent("proprietary", "proprietary MAC command")
	evtEnqueueProprietaryMACRequest = defineEnqueueMACRequestEvent("proprietary", "proprietary MAC command")
	evtReceiveProprietaryMAC        = events.Define(
		"ns.mac.proprietary.receive", "receive proprietary MAC command",
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	)

	evtClassASwitch = defineClassSwitchEvent('a')()
	evtClassBSwitch = defineClassSwitchEvent('b')()
	evtClassCSwitch = defineClassSwitchEvent('c')()
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

func uplinkMTypeLabel(mType ttnpb.MType) string {
	return strings.ToLower(mType.String())
}

func registerReceiveUniqueUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkUniqueReceived.WithLabelValues(ctx, uplinkMTypeLabel(msg.Payload.MType)).Inc()
}

func registerReceiveUplink(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkReceived.WithLabelValues(ctx, uplinkMTypeLabel(msg.Payload.MType)).Inc()
}

func registerMergeMetadata(ctx context.Context, msg *ttnpb.UplinkMessage) {
	gtwCount, _ := rxMetadataStats(ctx, msg.RxMetadata)
	nsMetrics.uplinkGateways.WithLabelValues(ctx).Observe(float64(gtwCount))
}

func registerForwardDataUplink(ctx context.Context, msg *ttnpb.ApplicationUplink) {
	mType := ttnpb.MType_UNCONFIRMED_UP
	if msg.Confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, uplinkMTypeLabel(mType)).Inc()
}

func registerForwardJoinRequest(ctx context.Context, msg *ttnpb.UplinkMessage) {
	nsMetrics.uplinkForwarded.WithLabelValues(ctx, uplinkMTypeLabel(msg.Payload.MType)).Inc()
}

func registerDropUplink(ctx context.Context, msg *ttnpb.UplinkMessage, err error) {
	cause := unknown
	if ttnErr, ok := errors.From(err); ok {
		cause = ttnErr.FullName()
	}
	nsMetrics.uplinkDropped.WithLabelValues(ctx, uplinkMTypeLabel(msg.Payload.MType), cause).Inc()
}
