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

package mqtt

import (
	"context"
	"fmt"
	stdio "io"
	"net"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	mqttlog "github.com/TheThingsIndustries/mystique/pkg/log"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/mqtt"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/metadata"
)

const qosDownlink byte = 0

type Marshaler interface {
	Version() topics.Version

	MarshalDownlink(down *ttnpb.DownlinkMessage) ([]byte, error)
	UnmarshalUplink(message []byte) (*ttnpb.UplinkMessage, error)
	UnmarshalStatus(message []byte) (*ttnpb.GatewayStatus, error)
}

type srv struct {
	ctx       context.Context
	server    io.Server
	marshaler Marshaler
	lis       mqttnet.Listener
}

// Start starts the MQTT frontend.
func Start(ctx context.Context, server io.Server, listener net.Listener, marshaler Marshaler, protocol string) {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/mqtt")
	ctx = mqttlog.NewContext(ctx, mqtt.Logger(log.FromContext(ctx)))
	s := &srv{ctx, server, marshaler, mqttnet.NewListener(listener, protocol)}
	go s.accept()
	go func() {
		<-ctx.Done()
		s.lis.Close()
	}()
}

func (s *srv) accept() {
	for {
		mqttConn, err := s.lis.Accept()
		if err != nil {
			if s.ctx.Err() == nil {
				log.FromContext(s.ctx).WithError(err).Warn("Accept failed")
			}
			return
		}

		go func() {
			ctx := log.NewContextWithFields(s.ctx, log.Fields("remote_addr", mqttConn.RemoteAddr().String()))
			conn := &connection{server: s.server, mqtt: mqttConn, marshaler: s.marshaler}
			if conn.setup(ctx); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to setup connection")
				mqttConn.Close()
				return
			}
		}()
	}
}

type connection struct {
	marshaler Marshaler
	server    io.Server
	mqtt      mqttnet.Conn
	session   session.Session
	io        *io.Connection
}

func (c *connection) setup(ctx context.Context) error {
	ctx = auth.NewContextWithInterface(ctx, c)
	c.session = session.New(ctx, c.mqtt, c.deliver)
	if err := c.session.ReadConnect(); err != nil {
		return err
	}
	ctx = c.io.Context()

	logger := log.FromContext(ctx)
	errCh := make(chan error)
	controlCh := make(chan packet.ControlPacket)

	// Read control packets
	go func() {
		for {
			pkt, err := c.session.ReadPacket()
			if err != nil {
				if err != stdio.EOF {
					logger.WithError(err).Warn("Error when reading packet")
				}
				errCh <- err
				return
			}
			if pkt != nil {
				logger.Debugf("Scheduling %s packet", packet.Name[pkt.PacketType()])
				controlCh <- pkt
			}
		}
	}()

	// Publish downlinks
	go func() {
		for {
			select {
			case <-c.io.Context().Done():
				logger.WithError(c.io.Context().Err()).Debug("Done sending downlink")
				return
			case down := <-c.io.Down():
				buf, err := c.marshaler.MarshalDownlink(down)
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}
				logger.Info("Publishing downlink message")
				topicParts := topics.Downlink(unique.ID(c.io.Context(), c.io.Gateway().GatewayIdentifiers), c.marshaler.Version())
				c.session.Publish(&packet.PublishPacket{
					TopicName:  topic.Join(topicParts),
					TopicParts: topicParts,
					QoS:        qosDownlink,
					Message:    buf,
				})
			}
		}
	}()

	// Write packets
	go func() {
		for {
			var err error
			select {
			case err = <-errCh:
			case pkt, ok := <-controlCh:
				if !ok {
					controlCh = nil
					continue
				}
				err = c.mqtt.Send(pkt)
			case pkt, ok := <-c.session.PublishChan():
				if !ok {
					return
				}
				logger.Debug("Writing publish packet")
				err = c.mqtt.Send(pkt)
			}
			if err != nil {
				if err != stdio.EOF {
					logger.WithError(err).Error("Send failed, closing session")
				} else {
					logger.Info("Disconnected")
				}
				c.session.Close()
				c.io.Disconnect(err)
				return
			}
		}
	}()

	logger.Info("Connected")
	return nil
}

type gatewayTopics struct {
	downlinkTopic []string
	uplinkTopic   []string
	statusTopic   []string
	ackTopic      []string
}

func (c *connection) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	ids, err := unique.ToGatewayID(info.Username)
	if err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		"id":            ids.GatewayID,
		"authorization": fmt.Sprintf("Key %s", info.Password),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)

	ctx, ids, err = c.server.FillGatewayContext(ctx, ids)
	if err != nil {
		return nil, err
	}
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	c.io, err = c.server.Connect(ctx, "mqtt", ids)
	if err != nil {
		return nil, err
	}
	if err = c.server.ClaimDownlink(ctx, ids); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to claim downlink")
		return nil, err
	}

	info.Metadata = gatewayTopics{
		downlinkTopic: topics.Downlink(uid, c.marshaler.Version()),
		uplinkTopic:   topics.Uplink(uid, c.marshaler.Version()),
		statusTopic:   topics.Status(uid, c.marshaler.Version()),
		ackTopic:      topics.TxAck(uid, c.marshaler.Version()),
	}
	info.Interface = c
	return c.io.Context(), nil
}

var errNotAuthorized = errors.DefinePermissionDenied("not_authorized", "not authorized")

func (c *connection) Subscribe(info *auth.Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	gt := info.Metadata.(gatewayTopics)
	if !topic.MatchPath(gt.downlinkTopic, topic.Split(requestedTopic)) {
		return "", 0, errNotAuthorized
	}

	acceptedTopic = topic.Join(gt.downlinkTopic)
	acceptedQoS = requestedQoS
	return
}

func (c *connection) CanRead(info *auth.Info, topicParts ...string) bool {
	gt := info.Metadata.(gatewayTopics)
	return topic.MatchPath(topicParts, gt.downlinkTopic)
}

func (c *connection) CanWrite(info *auth.Info, topicParts ...string) bool {
	gt := info.Metadata.(gatewayTopics)
	return topic.MatchPath(topicParts, gt.uplinkTopic) || topic.MatchPath(topicParts, gt.statusTopic) || topic.MatchPath(topicParts, gt.ackTopic)
}

func (c *connection) deliver(pkt *packet.PublishPacket) {
	logger := log.FromContext(c.io.Context()).WithField("topic", pkt.TopicName)
	switch {
	case topics.IsUplink(pkt.TopicParts):
		up, err := c.marshaler.UnmarshalUplink(pkt.Message)
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal uplink message")
			return
		}
		if err := c.io.HandleUp(up); err != nil {
			logger.WithError(err).Warn("Failed to handle uplink message")
		}
	case topics.IsStatus(pkt.TopicParts):
		status, err := c.marshaler.UnmarshalStatus(pkt.Message)
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal status message")
			return
		}
		if err := c.io.HandleStatus(status); err != nil {
			logger.WithError(err).Warn("Failed to handle status message")
		}
	case topics.IsTxAck(pkt.TopicParts):
		ack := &ttnpb.TxAcknowledgment{}
		if err := ack.Unmarshal(pkt.Message); err != nil {
			logger.WithError(err).Warn("Failed to unmarshal Tx acknowledgment message")
			return
		}
		if err := c.io.HandleTxAck(ack); err != nil {
			logger.WithError(err).Warn("Failed to handle Tx acknowledgment message")
		}
	}
}
