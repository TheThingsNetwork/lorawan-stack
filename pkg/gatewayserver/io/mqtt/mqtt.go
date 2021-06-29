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

package mqtt

import (
	"context"
	"fmt"
	stdio "io"
	"net"
	"os"
	"runtime/debug"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	mqttlog "github.com/TheThingsIndustries/mystique/pkg/log"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/mqtt"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc/metadata"
)

const qosDownlink byte = 0

type srv struct {
	ctx    context.Context
	server io.Server
	format Format
	lis    mqttnet.Listener
}

var errMQTTFrontendRecovered = errors.DefineInternal("mqtt_frontend_recovered", "internal server error")

// Serve serves the MQTT frontend.
func Serve(ctx context.Context, server io.Server, listener net.Listener, format Format, protocol string) error {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/mqtt")
	ctx = mqttlog.NewContext(ctx, mqtt.Logger(log.FromContext(ctx)))
	s := &srv{ctx, server, format, mqttnet.NewListener(listener, protocol)}
	go func() {
		<-ctx.Done()
		s.lis.Close()
	}()
	return s.accept()
}

func (s *srv) accept() error {
	for {
		mqttConn, err := s.lis.Accept()
		if err != nil {
			if s.ctx.Err() == nil {
				log.FromContext(s.ctx).WithError(err).Warn("Accept failed")
			}
			return err
		}

		remoteAddr := mqttConn.RemoteAddr().String()
		ctx := log.NewContextWithFields(s.ctx, log.Fields("remote_addr", remoteAddr))

		resource := ratelimit.GatewayAcceptMQTTConnectionResource(remoteAddr)
		if err := ratelimit.Require(s.server.RateLimiter(), resource); err != nil {
			if err := mqttConn.Close(); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Close connection failed")
			}
			log.FromContext(ctx).WithError(err).Debug("Drop connection")
			continue
		}

		go func() {
			conn := &connection{server: s.server, mqtt: mqttConn, format: s.format}
			if err := conn.setup(ctx); err != nil {
				switch err {
				case stdio.EOF, stdio.ErrUnexpectedEOF:
				default:
					log.FromContext(ctx).WithError(err).Warn("Failed to setup connection")
				}
				mqttConn.Close()
				return
			}
		}()
	}
}

type connection struct {
	format  Format
	server  io.Server
	mqtt    mqttnet.Conn
	session session.Session
	io      *io.Connection
	tokens  io.DownlinkTokens

	resource ratelimit.Resource
}

func (*connection) Protocol() string            { return "mqtt" }
func (*connection) SupportsDownlinkClaim() bool { return false }

func (c *connection) setup(ctx context.Context) (err error) {
	ctx = auth.NewContextWithInterface(ctx, c)
	ctx, cancel := errorcontext.New(ctx)
	defer func() {
		retrievedErr := recoverMQTTFrontend(ctx)
		if retrievedErr != nil {
			err = retrievedErr
		}
	}()
	c.session = session.New(ctx, c.mqtt, c.deliver)
	if err := c.session.ReadConnect(); err != nil {
		cancel(err)
		return err
	}
	ctx = c.io.Context()

	logger := log.FromContext(ctx)
	controlCh := make(chan packet.ControlPacket)

	// Read control packets
	go func() {
		defer recoverMQTTFrontend(ctx)
		for {
			pkt, err := c.session.ReadPacket()
			if err != nil {
				if err != stdio.EOF {
					logger.WithError(err).Warn("Error when reading packet")
				}
				cancel(err)
				return
			}
			if pkt != nil {
				logger.Debugf("Schedule %s packet", packet.Name[pkt.PacketType()])
				select {
				case <-ctx.Done():
					return
				case controlCh <- pkt:
				}
			}
		}
	}()

	// Publish downlinks
	go func() {
		defer recoverMQTTFrontend(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case down := <-c.io.Down():
				token := c.tokens.Next(down, time.Now())
				down.CorrelationIDs = append(down.CorrelationIDs, c.tokens.FormatCorrelationID(token))

				buf, err := c.format.FromDownlink(down, c.io.Gateway().GatewayIdentifiers)
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}
				logger.Info("Publish downlink message")
				topicParts := c.format.DownlinkTopic(unique.ID(c.io.Context(), c.io.Gateway().GatewayIdentifiers))
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
		defer recoverMQTTFrontend(ctx)
		for {
			var err error
			select {
			case <-ctx.Done():
				return
			case pkt := <-controlCh:
				err = c.mqtt.Send(pkt)
			case pkt, ok := <-c.session.PublishChan():
				if !ok {
					return
				}
				logger.Debug("Write publish packet")
				err = c.mqtt.Send(pkt)
			}
			if err != nil {
				cancel(err)
				return
			}
		}
	}()

	// Close connection on context closure
	go func() {
		select {
		case <-ctx.Done():
			logger.WithError(ctx.Err()).Info("Disconnected")
			c.session.Close()
			c.mqtt.Close()
		}
	}()

	logger.Info("Connected")
	return nil
}

type topicAccess struct {
	gtwUID string
	reads  [][]string
	writes [][]string
}

func (c *connection) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	ids := ttnpb.GatewayIdentifiers{
		GatewayId: info.Username,
	}
	if err := c.server.ValidateGatewayID(ctx, ids); err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		"id":            ids.GatewayId,
		"authorization": fmt.Sprintf("Bearer %s", info.Password),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)

	ctx, ids, err := c.server.FillGatewayContext(ctx, ids)
	if err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)
	c.io, err = c.server.Connect(ctx, c, ids)
	if err != nil {
		return nil, err
	}
	c.resource = ratelimit.GatewayUpResource(ctx, ids)

	access := topicAccess{
		gtwUID: uid,
		reads: [][]string{
			c.format.DownlinkTopic(uid),
		},
		writes: [][]string{
			c.format.BirthTopic(uid),
			c.format.LastWillTopic(uid),
			c.format.UplinkTopic(uid),
			c.format.StatusTopic(uid),
			c.format.TxAckTopic(uid),
		},
	}
	info.Metadata = access
	info.Interface = c
	return c.io.Context(), nil
}

var errNotAuthorized = errors.DefinePermissionDenied("not_authorized", "not authorized")

func (c *connection) Subscribe(info *auth.Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	access := info.Metadata.(topicAccess)
	acceptedTopicParts := c.format.DownlinkTopic(access.gtwUID)
	if !topic.MatchPath(acceptedTopicParts, topic.Split(requestedTopic)) {
		return "", 0, errNotAuthorized.New()
	}
	acceptedTopic = topic.Join(acceptedTopicParts)
	acceptedQoS = requestedQoS
	return
}

func (c *connection) CanRead(info *auth.Info, topicParts ...string) bool {
	access := info.Metadata.(topicAccess)
	for _, reads := range access.reads {
		if topic.MatchPath(topicParts, reads) {
			return true
		}
	}
	return false
}

func (c *connection) CanWrite(info *auth.Info, topicParts ...string) bool {
	access := info.Metadata.(topicAccess)
	for _, writes := range access.writes {
		if topic.MatchPath(topicParts, writes) {
			return true
		}
	}
	return false
}

func (c *connection) deliver(pkt *packet.PublishPacket) {
	logger := log.FromContext(c.io.Context()).WithField("topic", pkt.TopicName)

	if err := ratelimit.Require(c.server.RateLimiter(), c.resource); err != nil {
		logger.WithError(err).Warn("Terminate connection")
		c.io.Disconnect(err)
		return
	}

	switch {
	case c.format.IsBirthTopic(pkt.TopicParts):
	case c.format.IsLastWillTopic(pkt.TopicParts):
	case c.format.IsUplinkTopic(pkt.TopicParts):
		up, err := c.format.ToUplink(pkt.Message, c.io.Gateway().GatewayIdentifiers)
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal uplink message")
			return
		}
		up.ReceivedAt = pkt.Received
		if err := c.io.HandleUp(up); err != nil {
			logger.WithError(err).Warn("Failed to handle uplink message")
		}
	case c.format.IsStatusTopic(pkt.TopicParts):
		status, err := c.format.ToStatus(pkt.Message, c.io.Gateway().GatewayIdentifiers)
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal status message")
			return
		}
		if err := c.io.HandleStatus(status); err != nil {
			logger.WithError(err).Warn("Failed to handle status message")
		}
	case c.format.IsTxAckTopic(pkt.TopicParts):
		ack, err := c.format.ToTxAck(pkt.Message, c.io.Gateway().GatewayIdentifiers)
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal Tx acknowledgment message")
			return
		}
		if token, ok := c.tokens.ParseTokenFromCorrelationIDs(ack.GetCorrelationIDs()); ok {
			if down, _, ok := c.tokens.Get(token, time.Now()); ok {
				ack.DownlinkMessage = down
			}
		}
		if err := c.io.HandleTxAck(ack); err != nil {
			logger.WithError(err).Warn("Failed to handle Tx acknowledgment message")
		}
	default:
		logger.Debug("Publish to invalid topic")
	}
}

func recoverMQTTFrontend(ctx context.Context) error {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, p)
		os.Stderr.Write(debug.Stack())
		var err error
		if pErr, ok := p.(error); ok {
			err = errMQTTFrontendRecovered.WithCause(pErr)
		} else {
			err = errMQTTFrontendRecovered.WithAttributes("panic", p)
		}
		log.FromContext(ctx).WithError(err).Error("MQTT frontend failed")
		return err
	}
	return nil
}
