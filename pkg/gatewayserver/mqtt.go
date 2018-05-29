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

	"github.com/TheThingsIndustries/mystique/pkg/auth"
	mqttlog "github.com/TheThingsIndustries/mystique/pkg/log"
	"github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/mqtt"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

type mqttConnectionHandler struct {
	gs         *GatewayServer
	connection *mqttConnection
}

func (h mqttConnectionHandler) Context() context.Context {
	return h.connection.sess.Context()
}

// Topic formats for connect, disconnect, uplink, downlink and status messages
const (
	UplinkTopicSuffix   = "up"
	DownlinkTopicSuffix = "down"
	StatusTopicStatus   = "status"

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
		return nil, ErrUnsupportedTopicFormat.New(errors.Attributes{
			"topic": pkt.TopicName,
		})
	}
	if version := pkt.TopicParts[0]; version != V3TopicPrefix {
		return nil, ErrInvalidAPIVersion.New(errors.Attributes{
			"version": version,
		})
	}
	switch pkt.TopicParts[2] {
	case UplinkTopicSuffix:
		return h.deliverMQTTUplink, nil
	case StatusTopicStatus:
		return h.deliverMQTTStatus, nil
	default:
		return nil, ErrUnsupportedTopicFormat.New(errors.Attributes{
			"topic": pkt.TopicName,
		})
	}
}

func (h *mqttConnectionHandler) deliverMQTTUplink(pkt *packet.PublishPacket, conn connection) error {
	up := &ttnpb.UplinkMessage{}
	err := up.Unmarshal(pkt.Message)
	if err != nil {
		return common.ErrUnmarshalPayloadFailed.NewWithCause(nil, err)
	}

	return h.gs.handleUplink(h.Context(), up, conn)
}

func (h *mqttConnectionHandler) deliverMQTTStatus(pkt *packet.PublishPacket, _ connection) error {
	status := &ttnpb.GatewayStatus{}
	err := status.Unmarshal(pkt.Message)
	if err != nil {
		return common.ErrUnmarshalPayloadFailed.NewWithCause(nil, err)
	}

	return h.gs.handleStatus(h.Context(), status)
}

func (h *mqttConnectionHandler) Connect(info *auth.Info) error {
	identifiers, err := ttnpb.GatewayIdentifiersFromUniqueID(info.Username)
	if err != nil {
		return err
	}
	md := rpcmetadata.MD{
		ID:            identifiers.GetGatewayID(),
		AuthType:      "Bearer",
		AuthValue:     string(info.Password),
		AllowInsecure: true,
	}

	is, err := h.gs.getIdentityServer()
	if err != nil {
		return err
	}
	h.connection.gtw, err = is.GetGateway(h.Context(), identifiers)
	if err != nil {
		return err
	}
	resp, err := is.ListGatewayRights(h.Context(), identifiers, grpc.PerRPCCredentials(md))
	if err != nil {
		return err
	}
	info.Metadata = resp.GetRights()
	info.Interface = h
	return nil
}

func (h *mqttConnectionHandler) Subscribe(info *auth.Info, requestedTopic string, requestedQoS byte) (acceptedTopic string, acceptedQoS byte, err error) {
	if info.Metadata == nil {
		err = errors.New("No metadata present")
		return
	}
	rights := info.Metadata.([]ttnpb.Right)
	// TODO: Support v2 MQTT format https://github.com/TheThingsIndustries/ttn/issues/828
	downlinkTopic := topic.Join([]string{V3TopicPrefix, info.Username, DownlinkTopicSuffix})
	if requestedTopic != downlinkTopic || !ttnpb.IncludesRights(rights, ttnpb.RIGHT_GATEWAY_LINK) {
		err = common.ErrPermissionDenied.New(nil)
		return
	}
	identifiers, err := ttnpb.GatewayIdentifiersFromUniqueID(info.Username)
	if err != nil {
		return "", 0, err
	}

	if h.connection.scheduler == nil {
		fp, err := h.gs.getGatewayFrequencyPlan(h.gs.Context(), identifiers)
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
	go func() {
		<-ctx.Done()
		lis.Close()
	}()
	logger := log.FromContext(ctx)
	ctx = mqttlog.NewContext(ctx, mqtt.Logger(logger.WithField("namespace", "mqtt")))

	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.WithError(err).Error("Cannot continue accepting MQTT connections")
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

	if err := session.ReadConnect(); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Could not read CONNECT message, dropping connection")
		return
	}
	defer session.Close()

	uid := session.AuthInfo().Username
	logger := log.FromContext(ctx).WithField("gateway_uid", uid)
	gtwIDs, err := ttnpb.GatewayIdentifiersFromUniqueID(uid)
	if err != nil {
		logger.WithError(err).Warn("Could not extract gateway identifiers")
		return
	}

	logger.Debug("MQTT connection opened")
	g.signalStartServingGateway(ctx, gtwIDs)

	go func() {
		<-ctx.Done()
		logger.Debug("MQTT connection closed")
		g.signalStopServingGateway(g.Context(), gtwIDs)
	}()

	readErr := make(chan error)
	control := make(chan packet.ControlPacket)
	go func() {
		for {
			response, err := session.ReadPacket()
			if err != nil {
				if !errors.ErrEOF.Describes(err) {
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
			if !errors.ErrEOF.Describes(err) {
				logger.WithError(err).Error("Cannot continue to serve MQTT connection")
			}
			return
		}
	}
}
