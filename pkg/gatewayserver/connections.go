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
	"sync"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/udp"
	"go.thethings.network/lorawan-stack/pkg/toa"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const pullDataExpiration = 1 * time.Minute

type connection interface {
	addUpstreamObservations(*ttnpb.GatewayUp)
	addDownstreamObservations(*ttnpb.GatewayDown)
	getObservations() ttnpb.GatewayObservations

	gateway() *ttnpb.Gateway
	send(*ttnpb.GatewayDown) error
	Close() error
}

type connectionData struct {
	observations   ttnpb.GatewayObservations
	observationsMu sync.RWMutex

	scheduler scheduling.Scheduler
}

func (c *connectionData) schedule(down *ttnpb.GatewayDown) (err error) {
	span := scheduling.Span{
		Start: scheduling.ConcentratorTime(down.DownlinkMessage.TxMetadata.Timestamp),
	}
	span.Duration, err = toa.Compute(down.DownlinkMessage.RawPayload, down.DownlinkMessage.Settings)
	if err != nil {
		return errors.NewWithCause(err, "Could not compute time-on-air of the downlink")
	}

	err = c.scheduler.ScheduleAt(span, down.DownlinkMessage.Settings.Frequency)
	if err != nil {
		return errors.NewWithCause(err, "Could not schedule downlink")
	}

	return nil
}

func (c *connectionData) getObservations() ttnpb.GatewayObservations {
	c.observationsMu.Lock()
	observations := c.observations
	c.observationsMu.Unlock()
	return observations
}

func (c *connectionData) addUpstreamObservations(up *ttnpb.GatewayUp) {
	now := time.Now().UTC()

	c.observationsMu.Lock()

	if up.GatewayStatus != nil {
		c.observations.LastStatus = up.GatewayStatus
		c.observations.LastStatusReceivedAt = &now
	}

	if nbUplinks := len(up.UplinkMessages); nbUplinks > 0 {
		c.observations.LastUplinkReceivedAt = &now
	}

	c.observationsMu.Unlock()
}

func (c *connectionData) addDownstreamObservations(down *ttnpb.GatewayDown) {
	now := time.Now().UTC()

	c.observationsMu.Lock()
	c.observations.LastDownlinkReceivedAt = &now
	c.observationsMu.Unlock()
}

type gRPCConnection struct {
	connectionData

	link   ttnpb.GtwGs_LinkServer
	cancel context.CancelFunc

	gtw *ttnpb.Gateway
}

func (c *gRPCConnection) send(down *ttnpb.GatewayDown) error {
	if err := c.schedule(down); err != nil {
		return err
	}

	return c.link.Send(down)
}

func (c *gRPCConnection) gateway() *ttnpb.Gateway {
	return c.gtw
}

func (c *gRPCConnection) Close() error {
	c.cancel()
	return nil
}

type udpConnection struct {
	connectionData

	gtw atomic.Value

	lastPullDataStorage atomic.Value
	lastPullDataTime    atomic.Value
}

func (c *udpConnection) lastPullData() *udp.Packet {
	packet, _ := c.lastPullDataStorage.Load().(*udp.Packet)
	return packet
}

func (c *udpConnection) pullDataExpired() bool {
	lastReceived, ok := c.lastPullDataTime.Load().(time.Time)
	if !ok {
		return true
	}
	return time.Since(lastReceived) > pullDataExpiration
}

func (c *udpConnection) send(down *ttnpb.GatewayDown) error {
	if c.pullDataExpired() {
		return errors.NewWithCausef(ErrGatewayNotConnected.New(errors.Attributes{
			"gateway_id": c.gateway().GetGatewayID(),
		}), "No PULL_DATA received in the last %s", pullDataExpiration.String())
	}

	downstream, err := udp.TranslateDownstream(down)
	if err != nil {
		return ErrTranslationFromProtobuf.New(nil)
	}

	if err := c.schedule(down); err != nil {
		return err
	}

	packet := *c.lastPullData()
	packet.PacketType = udp.PullResp
	packet.Data = &downstream

	// TODO: Add a delay before the packet is sent: https://github.com/TheThingsIndustries/ttn/issues/726
	return packet.GatewayConn.Write(&packet)
}

func (c *udpConnection) gateway() *ttnpb.Gateway {
	gtw, _ := c.gtw.Load().(*ttnpb.Gateway)
	return gtw
}

func (c *udpConnection) Close() error { return nil }
