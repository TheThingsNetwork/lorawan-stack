// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package redis

import (
	"net"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

var (
	bytesRx = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "redis_client",
			Name:      "receive_bytes_total",
			Help:      "Total number of bytes received by the Redis client",
		},
		[]string{"remote_address"},
	)
	bytesTx = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "redis_client",
			Name:      "transmit_bytes_total",
			Help:      "Total number of bytes transmitted by the Redis client",
		},
		[]string{"remote_address"},
	)
	protosUnmarshaled = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: "redis",
			Name:      "protos_unmarshaled_total",
			Help:      "Total number of protos unmarshaled by the Redis client",
		},
	)
	protosMarshaled = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: "redis",
			Name:      "protos_marshaled_total",
			Help:      "Total number of protos marshaled by the Redis client",
		},
	)
)

func init() {
	metrics.MustRegister(
		bytesRx,
		bytesTx,
		protosUnmarshaled,
		protosMarshaled,
	)
}

type observableConn struct {
	addr string
	net.Conn
}

func (c *observableConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	bytesRx.WithLabelValues(c.addr).Add(float64(n))
	return n, err
}

func (c *observableConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	bytesTx.WithLabelValues(c.addr).Add(float64(n))
	return n, err
}
