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
	"net"
	"sync"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/mqtt"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc/metadata"
)

const qosDownlink byte = 0

// Serve serves the MQTT frontend.
func Serve(ctx context.Context, server io.Server, listener net.Listener, format Format, protocol string) error {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/mqtt")
	lis := mqttnet.NewListener(listener, protocol)
	go func() {
		<-ctx.Done()
		lis.Close()
	}()
	return mqtt.RunListener(
		ctx, lis, server,
		ratelimit.GatewayAcceptMQTTConnectionResource, server.RateLimiter(),
		func(ctx context.Context, mqttConn mqttnet.Conn) error {
			return setupConnection(ctx, mqttConn, format, server)
		},
	)
}

type connection struct {
	format   Format
	server   io.Server
	io       *io.Connection
	tokens   io.DownlinkTokens
	resource ratelimit.Resource
}

func (*connection) Protocol() string            { return "mqtt" }
func (*connection) SupportsDownlinkClaim() bool { return false }

func setupConnection(ctx context.Context, mqttConn mqttnet.Conn, format Format, server io.Server) error {
	c := &connection{
		format: format,
		server: server,
	}

	ctx = auth.NewContextWithInterface(ctx, c)
	session := session.New(ctx, mqttConn, c.deliver)
	if err := session.ReadConnect(); err != nil {
		if c.io != nil {
			c.io.Disconnect(err)
		}
		return err
	}
	ctx = c.io.Context()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	f := func(ctx context.Context) error {
		logger := log.FromContext(ctx)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case down := <-c.io.Down():
				token := c.tokens.Next(down, time.Now())
				down.CorrelationIds = append(down.CorrelationIds, c.tokens.FormatCorrelationID(token))

				buf, err := format.FromDownlink(down, c.io.Gateway().GetIds())
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}
				logger.Info("Publish downlink message")
				topicParts := format.DownlinkTopic(unique.ID(c.io.Context(), c.io.Gateway().GetIds()))
				session.Publish(&packet.PublishPacket{
					TopicName:  topic.Join(topicParts),
					TopicParts: topicParts,
					QoS:        qosDownlink,
					Message:    buf,
				})
			}
		}
	}
	server.StartTask(&task.Config{
		Context: ctx,
		ID:      "mqtt_publish_downlinks",
		Func:    f,
		Done:    wg.Done,
		Restart: task.RestartNever,
		Backoff: task.DefaultBackoffConfig,
	})

	mqtt.RunSession(ctx, c.io.Disconnect, server, session, mqttConn, wg)

	return nil
}

type topicAccess struct {
	gtwUID string
	reads  [][]string
	writes [][]string
}

func (c *connection) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	ids := &ttnpb.GatewayIdentifiers{
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
		up, err := c.format.ToUplink(pkt.Message, c.io.Gateway().GetIds())
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal uplink message")
			return
		}
		up.ReceivedAt = ttnpb.ProtoTimePtr(pkt.Received)
		if err := c.io.HandleUp(up, nil); err != nil {
			logger.WithError(err).Warn("Failed to handle uplink message")
		}
	case c.format.IsStatusTopic(pkt.TopicParts):
		status, err := c.format.ToStatus(pkt.Message, c.io.Gateway().GetIds())
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal status message")
			return
		}
		if err := c.io.HandleStatus(status); err != nil {
			logger.WithError(err).Warn("Failed to handle status message")
		}
	case c.format.IsTxAckTopic(pkt.TopicParts):
		ack, err := c.format.ToTxAck(pkt.Message, c.io.Gateway().GetIds())
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal Tx acknowledgment message")
			return
		}
		if token, ok := c.tokens.ParseTokenFromCorrelationIDs(ack.GetCorrelationIds()); ok {
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
