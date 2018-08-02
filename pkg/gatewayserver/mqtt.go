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
	"io"

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	mqttlog "github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/mqtt"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

type mqttConnectionHandler struct {
	gs         *GatewayServer
	connection *mqttConnection
	id         ttnpb.GatewayIdentifiers
}

func (h mqttConnectionHandler) Context() context.Context {
	return h.connection.sess.Context()
}

// Topic formats for connect, disconnect, uplink, downlink and status messages
const (
	UplinkTopicSuffix   = "up"
	DownlinkTopicSuffix = "down"
	StatusTopicSuffix   = "status"

	V3TopicPrefix = "v3"
)

type deliveryFunc func(pkt *packet.PublishPacket, gtw connection) error

func (h *mqttConnectionHandler) deliverMQTTMessage(pkt *packet.PublishPacket) {
	delivery, err := h.deliveryFunc(pkt)
	if err != nil {
		log.FromContext(h.Context()).WithField("topic", pkt.TopicName).WithError(err).Warn("Dropping packet")
		return
	}

	if err := delivery(pkt, h.connection); err != nil {
		log.FromContext(h.Context()).WithField("topic", pkt.TopicName).WithError(err).Warn("Could not handle MQTT message")
	}
	return
}

func (h *mqttConnectionHandler) deliveryFunc(pkt *packet.PublishPacket) (deliveryFunc, error) {
	if len(pkt.TopicParts) < 3 {
		// TODO: Support v2 MQTT format https://github.com/TheThingsIndustries/ttn/issues/828
		return nil, errors.New("v2 MQTT format not supported yet")
	}
	if len(pkt.TopicParts) != 3 {
		return nil, errUnsupportedTopicFormat.WithAttributes("topic", pkt.TopicName)
	}
	if version := pkt.TopicParts[0]; version != V3TopicPrefix {
		return nil, errInvalidAPIVersion.WithAttributes("version", version)
	}
	switch pkt.TopicParts[2] {
	case UplinkTopicSuffix:
		return h.deliverMQTTUplink, nil
	case StatusTopicSuffix:
		return h.deliverMQTTStatus, nil
	default:
		return nil, errUnsupportedTopicFormat.WithAttributes("topic", pkt.TopicName)
	}
}

func (h *mqttConnectionHandler) deliverMQTTUplink(pkt *packet.PublishPacket, conn connection) error {
	up := &ttnpb.UplinkMessage{}
	err := up.Unmarshal(pkt.Message)
	if err != nil {
		return errUnmarshalFromProtobuf.WithCause(err)
	}

	return h.gs.handleUplink(h.Context(), up, conn)
}

func (h *mqttConnectionHandler) deliverMQTTStatus(pkt *packet.PublishPacket, conn connection) error {
	status := &ttnpb.GatewayStatus{}
	err := status.Unmarshal(pkt.Message)
	if err != nil {
		return errUnmarshalFromProtobuf.WithCause(err)
	}

	return h.gs.handleStatus(h.Context(), status, conn)
}

// CustomMQTTContextFiller allows for filling the context for the MQTT connection.
var CustomMQTTContextFiller func(ctx context.Context, id ttnpb.GatewayIdentifiers) (context.Context, error)

func (h *mqttConnectionHandler) Connect(ctx context.Context, info *auth.Info) (context.Context, error) {
	var err error
	h.id, err = unique.ToGatewayID(info.Username)
	if err != nil {
		return nil, err
	}

	ctx = h.gs.Component.FillContext(ctx)
	if filler := CustomMQTTContextFiller; filler != nil {
		ctx, err = filler(ctx, h.id)
		if err != nil {
			return nil, err
		}
	}

	is, err := h.gs.getIdentityServer(ctx)
	if err != nil {
		return nil, err
	}

	h.connection.gtw, err = is.GetGateway(ctx, &h.id)
	if err != nil {
		return nil, err
	}

	md := rpcmetadata.MD{
		ID:            h.id.GatewayID,
		AuthType:      "Bearer",
		AuthValue:     string(info.Password),
		AllowInsecure: h.gs.Component.AllowInsecureForCredentials(),
	}
	resp, err := is.ListGatewayRights(ctx, &h.id, grpc.PerRPCCredentials(md))
	if err != nil {
		return nil, err
	}
	info.Metadata = resp.GetRights()

	info.Interface = h

	return ctx, nil
}

func (h *mqttConnectionHandler) Subscribe(info *auth.Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	if info.Metadata == nil {
		err = errNoMetadata
		return
	}
	rights := info.Metadata.([]ttnpb.Right)
	// TODO: Support v2 MQTT format https://github.com/TheThingsIndustries/ttn/issues/828
	downlinkTopic := topic.Join([]string{V3TopicPrefix, info.Username, DownlinkTopicSuffix})
	splitRequestedTopic := topic.Split(requestedTopic)
	switch {
	case len(splitRequestedTopic) != 3:
		err = errUnsupportedTopicFormat.WithAttributes("topic", requestedTopic)
	case splitRequestedTopic[0] != V3TopicPrefix:
		err = errInvalidAPIVersion.WithAttributes("version", splitRequestedTopic[0])
	case splitRequestedTopic[1] != info.Username || splitRequestedTopic[2] != DownlinkTopicSuffix:
		err = errPermissionDeniedForThisTopic.WithAttributes("topic", requestedTopic)
	case !ttnpb.IncludesRights(rights, ttnpb.RIGHT_GATEWAY_LINK):
		err = errAPIKeyNeedsRights.WithAttributes(
			"gateway_uid", info.Username,
			"rights", ttnpb.RIGHT_GATEWAY_LINK.String(),
		)
	}
	if err != nil {
		return
	}

	if h.connection.scheduler == nil {
		fp, err := h.gs.getGatewayFrequencyPlan(h.gs.Context(), &h.id)
		if err != nil {
			return "", 0, err
		}
		scheduler, err := scheduling.FrequencyPlanScheduler(h.gs.Context(), fp)
		if err != nil {
			return "", 0, err
		}
		h.connection.scheduler = scheduler
		h.gs.setupConnection(info.Username, h.connection)
	}
	return downlinkTopic, requestedQoS, nil
}

func (h *mqttConnectionHandler) CanRead(info *auth.Info, topic ...string) bool {
	var username, message string
	switch len(topic) {
	// TODO: Support v2 MQTT format https://github.com/TheThingsIndustries/ttn/issues/828
	case 3:
		if topic[0] != V3TopicPrefix {
			return false
		}
		username = topic[1]
		message = topic[2]
	default:
		return false
	}
	return username == info.Username && message == "down"
}

func (h *mqttConnectionHandler) CanWrite(info *auth.Info, topic ...string) bool {
	var username, message string
	switch len(topic) {
	// TODO: Support v2 MQTT format https://github.com/TheThingsIndustries/ttn/issues/828
	case 3:
		if topic[0] != V3TopicPrefix {
			return false
		}
		username = topic[1]
		message = topic[2]
	default:
		return false
	}
	return username == info.Username && (message == "up" || message == "status")
}

func (g *GatewayServer) runMQTTEndpoint(lis net.Listener) {
	ctx := g.Context()
	logger := log.FromContext(ctx)
	ctx = mqttlog.NewContext(ctx, mqtt.Logger(logger.WithField("namespace", "mqtt")))

	for {
		conn, err := lis.Accept()
		if err != nil {
			if ctx.Err() == nil {
				logger.WithError(err).Error("Cannot continue accepting MQTT connections")
			}
			return
		}
		go g.handleMQTTConnection(ctx, conn)
	}
}

func (g *GatewayServer) handleMQTTConnection(ctx context.Context, conn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer conn.Close()

	handler := &mqttConnectionHandler{gs: g}
	ctx = auth.NewContextWithInterface(ctx, handler)
	session := session.New(ctx, conn, handler.deliverMQTTMessage)
	mqttConn := &mqttConnection{sess: session}
	handler.connection = mqttConn

	logger := log.FromContext(ctx)
	if err := session.ReadConnect(); err != nil {
		logger.WithError(err).Warn("Could not read CONNECT message, dropping connection")
		return
	}
	defer session.Close()
	ctx = session.Context()
	logger = logger.WithField("gateway_uid", unique.ID(ctx, handler.id))
	logger.Debug("MQTT connection opened")

	if err := g.ClaimIDs(ctx, handler.id); err != nil {
		logger.WithError(err).Warn("Could not claim identifiers")
		return
	}

	go func() {
		<-ctx.Done()
		logger.Debug("MQTT connection closed")
		if err := g.UnclaimIDs(ctx, handler.id); err != nil {
			logger.WithError(err).Debug("Could not unclaim identifiers")
		}
	}()

	readErr := make(chan error)
	control := make(chan packet.ControlPacket)
	go func() {
		for {
			response, err := session.ReadPacket()
			if err != nil {
				if err != io.EOF {
					logger.WithError(err).Warn("Error when reading packet")
				}
				readErr <- err
				return
			}
			if response != nil {
				logger.Debugf("Write %s packet", packet.Name[response.PacketType()])
				control <- response
			}
		}
	}()

	for {
		var err error
		select {
		case err = <-readErr:
		case pkt, ok := <-control:
			if !ok {
				control = nil
				continue
			}
			err = conn.Send(pkt)
		case pkt, ok := <-session.PublishChan():
			if !ok {
				return
			}
			logger.Debug("Writing publish packet")
			err = conn.Send(pkt)
		}
		if err != nil {
			if err != io.EOF {
				logger.WithError(err).Error("Cannot continue to serve MQTT connection")
			}
			return
		}
	}
}
