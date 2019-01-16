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

package basicstation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstation/messages"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/web"
)

var (
	errEmptyGatewayEUI = errors.Define("empty_gateway_eui", "empty gateway eui")
)

type srv struct {
	ctx      context.Context
	server   io.Server
	upgrader *websocket.Upgrader
}

func (*srv) Protocol() string   { return "basicstation" }
func (*srv) HasScheduler() bool { return true }

// New returns a new Basic Station frontend that can be registered in the web server.
func New(ctx context.Context, server io.Server) web.Registerer {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/basicstation")
	s := &srv{ctx, server, &websocket.Upgrader{}}
	return s
}

func (s *srv) RegisterRoutes(server *web.Server) {
	group := server.Group(ttnpb.HTTPAPIPrefix + "/gs/io/basicstation")
	group.GET("/discover", s.handleDiscover)
	group.GET("/traffic/:uid", s.handleTraffic)
}

func (s *srv) handleDiscover(c echo.Context) error {
	logger := log.FromContext(s.ctx).WithFields(log.Fields(
		"endpoint", "discover",
		"remote_addr", c.Request().RemoteAddr,
	))
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		return err
	}
	defer ws.Close()

	_, data, err := ws.ReadMessage()
	if err != nil {
		logger.WithError(err).Debug("Failed to read message")
		return err
	}
	var req messages.DiscoverQuery
	if err := json.Unmarshal(data, &req); err != nil {
		logger.WithError(err).Debug("Failed to parse discover query message")
		if err := ws.WriteMessage(websocket.TextMessage, getErrorResponseFromString("Invalid request")); err != nil {
			logger.WithError(err).Warn("Failed to write error response message")
			return err
		}
		return err
	}

	if req.EUI.IsZero() {
		if err := ws.WriteMessage(websocket.TextMessage, getErrorResponseFromString("Invalid request")); err != nil {
			logger.WithError(err).Warn("Failed to write error response message")
			return err
		}
		return errEmptyGatewayEUI
	}

	ids := ttnpb.GatewayIdentifiers{
		EUI: &req.EUI.EUI64,
	}
	ctx, ids, err := s.server.FillGatewayContext(s.ctx, ids)
	if err != nil {
		logger.WithError(err).Debug("Failed to fill gateway context")
		if err := ws.WriteMessage(websocket.TextMessage, getErrorResponseFromString("Router not registered")); err != nil {
			logger.WithError(err).Warn("Failed to write error response message")
			return err
		}
		return err
	}
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	scheme := "ws"
	if c.IsTLS() {
		scheme = "wss"
	}
	res := messages.DiscoverResponse{
		EUI: req.EUI,
		Muxs: messages.EUI{
			Prefix: "muxs",
		},
		URI: fmt.Sprintf("%s://%s%s", scheme, c.Request().Host, c.Echo().URI(s.handleTraffic, uid)),
	}
	data, err = json.Marshal(res)
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal response message")
		if err := ws.WriteMessage(websocket.TextMessage, getErrorResponseFromString("Internal error")); err != nil {
			logger.WithError(err).Warn("Failed to write error response message")
			return err
		}
		return err
	}
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		logger.WithError(err).Warn("Failed to write discover response message")
		return err
	}
	logger.Debug("Sent discover response message")
	return nil
}

func (s *srv) handleTraffic(c echo.Context) error {
	uid := c.Param("uid")
	ids, err := unique.ToGatewayID(uid)
	if err != nil {
		return err
	}
	ctx, err := unique.WithContext(s.ctx, uid)
	if err != nil {
		return err
	}
	ctx = log.NewContextWithField(s.ctx, "gateway_uid", uid)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"endpoint", "traffic",
		"remote_addr", c.Request().RemoteAddr,
	))

	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		return err
	}
	defer ws.Close()

	ctx = rights.NewContext(ctx, rights.Rights{
		GatewayRights: map[string]*ttnpb.Rights{
			uid: {
				Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_LINK},
			},
		},
	})
	conn, err := s.server.Connect(ctx, s, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return err
	}
	if err := s.server.ClaimDownlink(ctx, ids); err != nil {
		logger.WithError(err).Error("Failed to claim downlink")
		return err
	}
	defer func() {
		if err := s.server.UnclaimDownlink(ctx, ids); err != nil {
			logger.WithError(err).Error("Failed to unclaim downlink")
		}
	}()

	// TODO: Start downlink processing in goroutine, see gRPC frontend.
	_ = conn

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			logger.WithError(err).Debug("Failed to read message")
			return err
		}
		typ, err := messages.Type(data)
		if err != nil {
			logger.WithError(err).Debug("Failed to parse message type")
			return err
		}
		switch typ {
		case messages.TypeVersion:
			var version messages.Version
			if err := json.Unmarshal(data, &version); err != nil {
				logger.WithError(err).Debug("Failed to unmarshal version message")
				return err
			}
			logger = logger.WithFields(log.Fields(
				"station", version.Station,
				"firmware", version.Firmware,
				"model", version.Model,
			))
			fp, err := s.server.GetFrequencyPlan(ctx, ids)
			if err != nil {
				logger.WithError(err).Warn("Failed to get frequency plan")
				return err
			}
			// TODO: Send frequency plan, see messages.RouterConfig.
			_ = fp

		// TODO: Add case for uplink messages.
		case messages.TypeJoinRequest:
			// TODO

		default:
			logger.WithField("message_type", typ).Debug("Unknown message type")
		}
	}
}

// getErrorResponseFromString add  the string into the "Error" field of the response message and returns the marshaled byte slice.
func getErrorResponseFromString(msg string) []byte {
	data, _ := json.Marshal(messages.DiscoverResponse{
		Error: msg,
	})
	return data
}
