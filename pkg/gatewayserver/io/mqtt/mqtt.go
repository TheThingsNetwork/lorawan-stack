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
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/mqtt"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/metadata"
)

const qosDownlink byte = 0

type srv struct {
	ctx    context.Context
	server io.Server
	lis    mqttnet.Listener
}

// Start starts the MQTT frontend.
func Start(ctx context.Context, server io.Server, listener net.Listener, protocol string) {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/mqtt")
	ctx = mqttlog.NewContext(ctx, mqtt.Logger(log.FromContext(ctx)))
	s := &srv{ctx, server, mqttnet.NewListener(listener, protocol)}
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
			log.FromContext(s.ctx).WithError(err).Warn("Accept failed")
			return
		}

		go func() {
			ctx := log.NewContextWithFields(s.ctx, log.Fields("remote_addr", mqttConn.RemoteAddr().String()))
			conn := &connection{server: s.server, mqtt: mqttConn}
			if conn.setup(ctx); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to setup connection")
				mqttConn.Close()
				return
			}
		}()
	}
}

type connection struct {
	server  io.Server
	mqtt    mqttnet.Conn
	session session.Session
	io      *io.Connection
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
			var err error
			select {
			case <-c.io.Context().Done():
				logger.WithError(c.io.Context().Err()).Debug("Done sending downlink")
				return
			case down := <-c.io.Down():
				msg := &ttnpb.GatewayDown{
					DownlinkMessage: down,
				}
				var buf []byte
				buf, err = msg.Marshal()
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}
				logger.Info("Publishing downlink message")
				topicParts := topics.Downlink(unique.ID(c.io.Context(), c.io.Gateway().GatewayIdentifiers))
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
}

func (c *connection) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	id, err := unique.ToGatewayID(info.Username)
	if err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		"id":            id.GatewayID,
		"authorization": fmt.Sprintf("Key %s", info.Password),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)

	ctx, id, err = c.server.FillGatewayContext(ctx, id)
	if err != nil {
		return nil, err
	}
	uid := unique.ID(ctx, id)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	c.io, err = c.server.Connect(ctx, "mqtt", id)
	if err != nil {
		return nil, err
	}
	if err = c.server.ClaimDownlink(ctx, id); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to claim downlink")
		return nil, err
	}

	info.Metadata = gatewayTopics{
		downlinkTopic: topics.Downlink(uid),
		uplinkTopic:   topics.Uplink(uid),
		statusTopic:   topics.Status(uid),
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
	return topic.MatchPath(topicParts, gt.uplinkTopic) || topic.MatchPath(topicParts, gt.statusTopic)
}

func (c *connection) deliver(pkt *packet.PublishPacket) {
	logger := log.FromContext(c.io.Context()).WithField("topic", pkt.TopicName)
	switch {
	case topics.IsUplink(pkt.TopicParts):
		up := &ttnpb.UplinkMessage{}
		if err := up.Unmarshal(pkt.Message); err != nil {
			logger.WithError(err).Warn("Failed to unmarshal uplink message")
			return
		}
		if err := c.io.HandleUp(up); err != nil {
			logger.WithError(err).Warn("Failed to handle uplink message")
		}
	case topics.IsStatus(pkt.TopicParts):
		status := &ttnpb.GatewayStatus{}
		if err := status.Unmarshal(pkt.Message); err != nil {
			logger.WithError(err).Warn("Failed to unmarshal status message")
			return
		}
		if err := c.io.HandleStatus(status); err != nil {
			logger.WithError(err).Warn("Failed to handle status message")
		}
	}
}
