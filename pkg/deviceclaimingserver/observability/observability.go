// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package observability provides events and metrics for device claiming.
package observability

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtClaimEndDeviceSuccess = events.Define(
		"dcs.end_device.claim.success", "claim end device successful",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
	)
	evtClaimEndDeviceAbort = events.Define(
		"dcs.end_device.claim.abort", "abort end device claim",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
	)
	evtClaimEndDeviceFail = events.Define(
		"dcs.end_device.claim.fail", "claim end device failure",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
		),
	)
	evtClaimGatewaySuccess = events.Define(
		"dcs.gateway.claim.success", "claim gateway successful",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
	)
	evtClaimGatewayAbort = events.Define(
		"dcs.gateway.claim.abort", "abort gateway claim",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
	)
	evtClaimGatewayFail = events.Define(
		"dcs.gateway.claim.fail", "claim gateway failure",
		events.WithVisibility(
			ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
		),
	)

	evtUnclaimEndDeviceSuccess = events.Define(
		"dcs.end_device.unclaim.success", "unclaim end device successful",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
	)
	evtUnclaimEndDeviceFail = events.Define(
		"dcs.end_device.unclaim.fail", "unclaim end device failure",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
		),
	)
	evtUnclaimGatewaySuccess = events.Define(
		"dcs.gateway.unclaim.success", "unclaim gateway successful",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
	)
	evtUnclaimGatewayFail = events.Define(
		"dcs.gateway.unclaim.fail", "unclaim gateway failure",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ),
	)
)

const (
	subsystem  = "dcs"
	unknown    = "unknown"
	entityType = "entity_type"
	id         = "id"
)

var dcsMetrics = &claimMetrics{
	claimSucceeded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "claim_success_total",
			Help:      "Total number of successfully claimed entities",
		},
		[]string{entityType, id},
	),
	claimAborted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "claim_aborted_total",
			Help:      "Total number of claim entity abortions",
		},
		[]string{entityType, id, "error"},
	),
	claimFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "claim_failed_total",
			Help:      "Total number of claim entity failures",
		},
		[]string{entityType, id, "error"},
	),

	unclaimSucceeded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "unclaim_success_total",
			Help:      "Total number of successfully unclaimed entities",
		},
		[]string{entityType, id},
	),
	unclaimFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "unclaim_failed_total",
			Help:      "Total number of unclaim entity failures",
		},
		[]string{entityType, id, "error"},
	),
}

func init() {
	metrics.MustRegister(dcsMetrics)
}

type claimMetrics struct {
	claimSucceeded *metrics.ContextualCounterVec
	claimAborted   *metrics.ContextualCounterVec
	claimFailed    *metrics.ContextualCounterVec

	unclaimSucceeded *metrics.ContextualCounterVec
	unclaimFailed    *metrics.ContextualCounterVec
}

func (m claimMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.claimSucceeded.Describe(ch)
	m.claimAborted.Describe(ch)
	m.claimFailed.Describe(ch)

	m.unclaimSucceeded.Describe(ch)
	m.unclaimFailed.Describe(ch)
}

func (m claimMetrics) Collect(ch chan<- prometheus.Metric) {
	m.claimSucceeded.Collect(ch)
	m.claimAborted.Collect(ch)
	m.claimFailed.Collect(ch)

	m.unclaimSucceeded.Collect(ch)
	m.unclaimFailed.Collect(ch)
}

// RegisterSuccessClaim registers a successful claim.
func RegisterSuccessClaim(ctx context.Context, entityIDs *ttnpb.EntityIdentifiers) {
	var id, entityType string
	switch ids := entityIDs.Ids.(type) {
	case *ttnpb.EntityIdentifiers_DeviceIds:
		id = ids.DeviceIds.GetApplicationIds().GetApplicationId()
		entityType = store.EntityEndDevice
		events.Publish(evtClaimEndDeviceSuccess.NewWithIdentifiersAndData(ctx, entityIDs, nil))
	case *ttnpb.EntityIdentifiers_GatewayIds:
		id = ids.GatewayIds.GatewayId
		entityType = store.EntityGateway
		events.Publish(evtClaimGatewaySuccess.NewWithIdentifiersAndData(ctx, entityIDs, nil))
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", entityIDs.Ids))
	}
	dcsMetrics.claimSucceeded.WithLabelValues(ctx, entityType, id).Inc()
}

// RegisterAbortClaim registers an aborted claim.
func RegisterAbortClaim(ctx context.Context, entityIDs *ttnpb.EntityIdentifiers, err error) {
	var id, entityType string
	switch ids := entityIDs.Ids.(type) {
	case *ttnpb.EntityIdentifiers_DeviceIds:
		id = ids.DeviceIds.GetApplicationIds().GetApplicationId()
		entityType = store.EntityEndDevice
		events.Publish(evtClaimEndDeviceAbort.NewWithIdentifiersAndData(ctx, entityIDs, err))
	case *ttnpb.EntityIdentifiers_GatewayIds:
		id = ids.GatewayIds.GatewayId
		entityType = store.EntityGateway
		events.Publish(evtClaimGatewayAbort.NewWithIdentifiersAndData(ctx, entityIDs, err))
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", entityIDs.Ids))
	}
	if ttnErr, ok := errors.From(err); ok {
		dcsMetrics.claimAborted.WithLabelValues(ctx, entityType, id, ttnErr.FullName()).Inc()
	} else {
		dcsMetrics.claimAborted.WithLabelValues(ctx, entityType, id, unknown).Inc()
	}
}

// RegisterFailClaim registers an failed claim.
func RegisterFailClaim(ctx context.Context, entityIDs *ttnpb.EntityIdentifiers, err error) {
	var id, entityType string
	switch ids := entityIDs.Ids.(type) {
	case *ttnpb.EntityIdentifiers_DeviceIds:
		id = ids.DeviceIds.GetApplicationIds().GetApplicationId()
		entityType = "end_device"
		events.Publish(evtClaimEndDeviceFail.NewWithIdentifiersAndData(ctx, entityIDs, err))
	case *ttnpb.EntityIdentifiers_GatewayIds:
		id = ids.GatewayIds.GatewayId
		entityType = "gateway"
		events.Publish(evtClaimGatewayFail.NewWithIdentifiersAndData(ctx, entityIDs, err))
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", entityIDs.Ids))
	}
	if ttnErr, ok := errors.From(err); ok {
		dcsMetrics.claimFailed.WithLabelValues(ctx, entityType, id, ttnErr.FullName()).Inc()
	} else {
		dcsMetrics.claimFailed.WithLabelValues(ctx, entityType, id, unknown).Inc()
	}
}

// RegisterSuccessUnclaim registers a successful unclaim.
func RegisterSuccessUnclaim(ctx context.Context, entityIDs *ttnpb.EntityIdentifiers) {
	var id, entityType string
	switch ids := entityIDs.Ids.(type) {
	case *ttnpb.EntityIdentifiers_DeviceIds:
		id = ids.DeviceIds.GetApplicationIds().GetApplicationId()
		entityType = "end_device"
		events.Publish(evtUnclaimEndDeviceSuccess.NewWithIdentifiersAndData(ctx, entityIDs, nil))
	case *ttnpb.EntityIdentifiers_GatewayIds:
		id = ids.GatewayIds.GatewayId
		entityType = "gateway"
		events.Publish(evtUnclaimGatewaySuccess.NewWithIdentifiersAndData(ctx, entityIDs, nil))
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", entityIDs.Ids))
	}
	dcsMetrics.unclaimSucceeded.WithLabelValues(ctx, entityType, id).Inc()
}

// RegisterFailUnclaim registers an failed unclaim.
func RegisterFailUnclaim(ctx context.Context, entityIDs *ttnpb.EntityIdentifiers, err error) {
	var id, entityType string
	switch ids := entityIDs.Ids.(type) {
	case *ttnpb.EntityIdentifiers_DeviceIds:
		id = ids.DeviceIds.GetApplicationIds().GetApplicationId()
		entityType = "end_device"
		events.Publish(evtUnclaimEndDeviceFail.NewWithIdentifiersAndData(ctx, entityIDs, err))
	case *ttnpb.EntityIdentifiers_GatewayIds:
		id = ids.GatewayIds.GatewayId
		entityType = "gateway"
		events.Publish(evtUnclaimGatewayFail.NewWithIdentifiersAndData(ctx, entityIDs, err))
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", entityIDs.Ids))
	}
	if ttnErr, ok := errors.From(err); ok {
		dcsMetrics.unclaimFailed.WithLabelValues(ctx, entityType, id, ttnErr.FullName()).Inc()
	} else {
		dcsMetrics.unclaimFailed.WithLabelValues(ctx, entityType, id, unknown).Inc()
	}
}
