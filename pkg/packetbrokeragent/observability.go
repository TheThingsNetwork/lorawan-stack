// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

const subsystem = "pba"

var pbaMetrics = &messageMetrics{
	uplinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_received_total",
			Help:      "Total number of uplinks received from Packet Broker",
		},
		[]string{
			"forwarder_net_id",
			"forwarder_tenant_id",
			"forwarder_cluster_id",
			"home_network_net_id",
			"home_network_tenant_id",
			"home_network_cluster_id",
		},
	),
	downlinkReceived: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_received_total",
			Help:      "Total number of downlinks received from Packet Broker",
		},
		[]string{
			"home_network_net_id",
			"home_network_tenant_id",
			"home_network_cluster_id",
			"forwarder_net_id",
			"forwarder_tenant_id",
			"forwarder_cluster_id",
		},
	),
	uplinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "uplink_forwarded_total",
			Help:      "Total number of uplinks forwarded to Packet Broker",
		},
		[]string{
			"forwarder_net_id",
			"forwarder_tenant_id",
			"forwarder_cluster_id",
		},
	),
	downlinkForwarded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "downlink_forwarded_total",
			Help:      "Total number of downlinks forwarded to Packet Broker",
		},
		[]string{
			"home_network_net_id",
			"home_network_tenant_id",
			"home_network_cluster_id",
			"forwarder_net_id",
			"forwarder_tenant_id",
			"forwarder_cluster_id",
		},
	),
}

func init() {
	metrics.MustRegister(pbaMetrics)
}

type messageMetrics struct {
	uplinkReceived    *metrics.ContextualCounterVec
	downlinkReceived  *metrics.ContextualCounterVec
	uplinkForwarded   *metrics.ContextualCounterVec
	downlinkForwarded *metrics.ContextualCounterVec
}

func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.uplinkReceived.Describe(ch)
	m.downlinkReceived.Describe(ch)
	m.uplinkForwarded.Describe(ch)
	m.downlinkForwarded.Describe(ch)
}

func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.uplinkReceived.Collect(ch)
	m.downlinkReceived.Collect(ch)
	m.uplinkForwarded.Collect(ch)
	m.downlinkForwarded.Collect(ch)
}
