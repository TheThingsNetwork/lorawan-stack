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

package joinserver

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const (
	logNamespace    = "joinserver"
	tracerNamespace = "go.thethings.network/lorawan-stack/pkg/joinserver"
)

var (
	evtRejectJoin = events.Define(
		"js.join.reject", "reject join-request",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
		events.WithPropagateToParent(),
	)
	evtAcceptJoin = events.Define(
		"js.join.accept", "accept join-request",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithPropagateToParent(),
	)
)

const (
	subsystem = "js"
	unknown   = "unknown"
)

var jsMetrics = &messageMetrics{
	joinAccepted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "join_accepted_total",
			Help:      "Total number of accepted joins",
		},
		[]string{"net_id"},
	),
	joinRejected: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "join_rejected_total",
			Help:      "Total number of rejected joins",
		},
		[]string{"error"},
	),
	devNonce: &devNonceMetrics{
		tooSmall: metrics.NewContextualCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "dev_nonce_too_small",
				Help:      "Total number of DevNonces too small errors",
			},
			[]string{"mac_version"},
		),
		reuse: metrics.NewContextualCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "dev_nonce_reuse",
				Help:      "Total number of DevNonces reuse errors",
			},
			[]string{"mac_version"},
		),
	},
}

func init() {
	metrics.MustRegister(jsMetrics)
}

type messageMetrics struct {
	joinAccepted *metrics.ContextualCounterVec
	joinRejected *metrics.ContextualCounterVec
	devNonce     *devNonceMetrics
}

type devNonceMetrics struct {
	tooSmall *metrics.ContextualCounterVec
	reuse    *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.joinAccepted.Describe(ch)
	m.joinRejected.Describe(ch)
	m.devNonce.reuse.Describe(ch)
	m.devNonce.tooSmall.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.joinAccepted.Collect(ch)
	m.joinRejected.Collect(ch)
	m.devNonce.reuse.Collect(ch)
	m.devNonce.tooSmall.Collect(ch)
}

func registerAcceptJoin(ctx context.Context, dev *ttnpb.EndDevice, msg *ttnpb.JoinRequest) {
	events.Publish(evtAcceptJoin.NewWithIdentifiersAndData(ctx, dev.Ids, nil))
	jsMetrics.joinAccepted.WithLabelValues(ctx, types.MustNetID(msg.NetId).OrZero().String()).Inc()
}

func registerRejectJoin(ctx context.Context, req *ttnpb.JoinRequest, err error) {
	events.Publish(evtRejectJoin.NewWithIdentifiersAndData(ctx, nil, err))
	if ttnErr, ok := errors.From(err); ok {
		jsMetrics.joinRejected.WithLabelValues(ctx, ttnErr.FullName()).Inc()
	} else {
		jsMetrics.joinRejected.WithLabelValues(ctx, unknown).Inc()
	}
}

func registerDevNonceReuse(ctx context.Context, msg *ttnpb.JoinRequest) {
	jsMetrics.devNonce.reuse.WithLabelValues(ctx, ttnpb.MACVersion_name[int32(msg.SelectedMacVersion)]).Inc()
}

func registerDevNonceTooSmall(ctx context.Context, msg *ttnpb.JoinRequest) {
	jsMetrics.devNonce.tooSmall.WithLabelValues(ctx, ttnpb.MACVersion_name[int32(msg.SelectedMacVersion)]).Inc()
}
