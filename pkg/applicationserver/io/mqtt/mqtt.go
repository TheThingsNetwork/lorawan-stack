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

// Package mqtt implements the MQTT frontend.
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
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/mqtt"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/metadata"
)

const qosUpstream byte = 0

type srv struct {
	ctx    context.Context
	server io.Server
	format Format
	lis    mqttnet.Listener
}

// Start starts the MQTT frontend.
func Start(ctx context.Context, server io.Server, listener net.Listener, format Format, protocol string) {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/mqtt")
	ctx = mqttlog.NewContext(ctx, mqtt.Logger(log.FromContext(ctx)))
	s := &srv{ctx, server, format, mqttnet.NewListener(listener, protocol)}
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
			conn := &connection{server: s.server, mqtt: mqttConn, format: s.format}
			if err := conn.setup(ctx); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to setup connection")
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
	io      *io.Subscription
}

func (c *connection) setup(ctx context.Context) error {
	ctx = auth.NewContextWithInterface(ctx, c)
	ctx, cancel := errorcontext.New(ctx)
	c.session = session.New(ctx, c.mqtt, c.deliver)
	if err := c.session.ReadConnect(); err != nil {
		return err
	}
	ctx = c.io.Context()

	logger := log.FromContext(ctx)
	controlCh := make(chan packet.ControlPacket)

	// Read control packets
	go func() {
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
				controlCh <- pkt
			}
		}
	}()

	// Publish upstream
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.io.Context().Done():
				err := c.io.Context().Err()
				cancel(err)
				logger.WithError(err).Debug("Subscription cancelled")
				return
			case up := <-c.io.Up():
				logger := logger.WithField("device_uid", unique.ID(up.Context, up.EndDeviceIdentifiers))
				var topicParts []string
				switch up.Up.(type) {
				case *ttnpb.ApplicationUp_UplinkMessage:
					topicParts = c.format.UplinkTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_JoinAccept:
					topicParts = c.format.JoinAcceptTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_DownlinkAck:
					topicParts = c.format.DownlinkAckTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_DownlinkNack:
					topicParts = c.format.DownlinkNackTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_DownlinkSent:
					topicParts = c.format.DownlinkSentTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_DownlinkFailed:
					topicParts = c.format.DownlinkFailedTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_DownlinkQueued:
					topicParts = c.format.DownlinkQueuedTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				case *ttnpb.ApplicationUp_LocationSolved:
					topicParts = c.format.LocationSolvedTopic(unique.ID(up.Context, c.io.ApplicationIDs()), up.DeviceID)
				}
				if topicParts == nil {
					continue
				}
				buf, err := c.format.FromUp(up.ApplicationUp)
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal upstream message")
					continue
				}
				logger.Debug("Publish upstream message")
				c.session.Publish(&packet.PublishPacket{
					TopicName:  topic.Join(topicParts),
					TopicParts: topicParts,
					QoS:        qosUpstream,
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
			case <-ctx.Done():
				return
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
				logger.Debug("Write publish packet")
				err = c.mqtt.Send(pkt)
			}
			if err != nil {
				if err != stdio.EOF {
					logger.WithError(err).Error("Send failed, closing session")
				} else {
					logger.Info("Disconnected")
				}
				cancel(err)
				return
			}
		}
	}()

	// Close connection on context closure
	go func() {
		select {
		case <-ctx.Done():
			c.session.Close()
			c.mqtt.Close()
		}
	}()

	logger.Info("Connected")
	return nil
}

type topicAccess struct {
	appUID string
	reads  [][]string
	writes [][]string
}

func (c *connection) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	ids := ttnpb.ApplicationIdentifiers{
		ApplicationID: info.Username,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		"id":            ids.ApplicationID,
		"authorization": fmt.Sprintf("Bearer %s", info.Password),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)

	ctx = c.server.FillContext(ctx)
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "application_uid", uid)

	var err error
	c.io, err = c.server.Subscribe(ctx, "mqtt", ids)
	if err != nil {
		return nil, err
	}
	ctx = c.io.Context()
	access := topicAccess{
		appUID: uid,
	}
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err == nil {
		access.reads = append(access.reads,
			c.format.UplinkTopic(uid, topic.PartWildcard),
			c.format.JoinAcceptTopic(uid, topic.PartWildcard),
			c.format.DownlinkAckTopic(uid, topic.PartWildcard),
			c.format.DownlinkNackTopic(uid, topic.PartWildcard),
			c.format.DownlinkSentTopic(uid, topic.PartWildcard),
			c.format.DownlinkFailedTopic(uid, topic.PartWildcard),
			c.format.DownlinkQueuedTopic(uid, topic.PartWildcard),
			c.format.LocationSolvedTopic(uid, topic.PartWildcard),
		)
	}
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE); err == nil {
		access.writes = append(access.writes,
			c.format.DownlinkPushTopic(uid, topic.PartWildcard),
			c.format.DownlinkReplaceTopic(uid, topic.PartWildcard),
		)
	}
	info.Metadata = access
	info.Interface = c
	return ctx, nil
}

var errNotAuthorized = errors.DefinePermissionDenied("not_authorized", "not authorized")

func (c *connection) Subscribe(info *auth.Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	access := info.Metadata.(topicAccess)
	accepted, ok := c.format.AcceptedTopic(access.appUID, topic.Split(requestedTopic))
	if !ok {
		return "", 0, errNotAuthorized
	}
	acceptedTopic = topic.Join(accepted)
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
	var deviceID string
	var op func(io.Server, context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
	switch {
	case c.format.IsDownlinkPushTopic(pkt.TopicParts):
		deviceID = c.format.ParseDownlinkPushTopic(pkt.TopicParts)
		op = io.Server.DownlinkQueuePush
	case c.format.IsDownlinkReplaceTopic(pkt.TopicParts):
		deviceID = c.format.ParseDownlinkReplaceTopic(pkt.TopicParts)
		op = io.Server.DownlinkQueueReplace
	default:
		logger.Error("Invalid topic path")
		return
	}
	items, err := c.format.ToDownlinks(pkt.Message)
	if err != nil {
		logger.WithError(err).Warn("Failed to decode downlink messages")
		return
	}
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: *c.io.ApplicationIDs(),
		DeviceID:               deviceID,
	}
	logger.WithFields(log.Fields(
		"device_uid", unique.ID(c.io.Context(), ids),
		"count", len(items.Downlinks),
	)).Debug("Handle downlink messages")
	if err := op(c.server, c.io.Context(), ids, items.Downlinks); err != nil {
		logger.WithError(err).Warn("Failed to handle downlink messages")
	}
}
