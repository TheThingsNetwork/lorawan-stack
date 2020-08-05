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

package ws

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/basicstation"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/web/middleware"
	"google.golang.org/grpc/metadata"
)

var (
	errEmptyGatewayEUI = errors.Define("empty_gateway_eui", "empty gateway EUI")
	errListener        = errors.DefineFailedPrecondition(
		"listener",
		"failed to serve Basic Station frontend listener",
	)
	errGatewayID      = errors.DefineInvalidArgument("invalid_gateway_id", "invalid gateway id `{id}`")
	errNoAuthProvided = errors.DefineUnauthenticated("no_auth_provided", "no auth provided for gateway id `{id}`")
)

type srv struct {
	ctx                  context.Context
	server               io.Server
	webServer            *echo.Echo
	upgrader             *websocket.Upgrader
	tokens               io.DownlinkTokens
	useTrafficTLSAddress bool
	wsPingInterval       time.Duration
	cfg                  Config
	formatter             Formatter
}

func (*srv) Protocol() string            { return "basicstation" }
func (*srv) SupportsDownlinkClaim() bool { return false }

// New creates the LoRa Basics Station front end.
func New(ctx context.Context, server io.Server, formatter Formatter, cfg Config) *echo.Echo {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/basicstation")

	webServer := echo.New()
	webServer.Logger = web.NewNoopLogger()
	webServer.HTTPErrorHandler = errorHandler
	webServer.Use(
		middleware.ID(""),
		echomiddleware.BodyLimit("16M"),
		middleware.Log(log.FromContext(ctx)),
		middleware.Recover(),
	)

	s := &srv{
		ctx:       ctx,
		server:    server,
		upgrader:  &websocket.Upgrader{},
		webServer: webServer,
		formatter:    formatter,
		cfg:       cfg,
	}

	webServer.GET("/router-info", s.handleDiscover)
	webServer.GET("/traffic/:id", s.handleTraffic)

	go func() {
		<-ctx.Done()
		webServer.Close()
	}()

	return webServer
}

func (s *srv) handleDiscover(c echo.Context) error {
	ctx := c.Request().Context()
	logger := log.FromContext(ctx).WithFields(log.Fields(
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
	var req DiscoverQuery
	if err := json.Unmarshal(data, &req); err != nil {
		logger.WithError(err).Debug("Failed to parse discover query message")
		return err
	}

	if req.EUI.IsZero() {
		writeDiscoverError(ctx, ws, "Empty router EUI provided")
		return errEmptyGatewayEUI.New()
	}

	ids := ttnpb.GatewayIdentifiers{
		EUI: &req.EUI.EUI64,
	}
	filledCtx, ids, err := s.server.FillGatewayContext(ctx, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to fetch gateway")
		writeDiscoverError(ctx, ws, fmt.Sprintf("Failed to fetch gateway: %s", err.Error()))
		return err
	}

	ctx = filledCtx

	scheme := "ws"
	if c.IsTLS() || s.cfg.UseTrafficTLSAddress {
		scheme = "wss"
	}

	euiWithPrefix := fmt.Sprintf("eui-%s", ids.EUI.String())
	res := DiscoverResponse{
		EUI: req.EUI,
		Muxs: basicstation.EUI{
			Prefix: "muxs",
		},
		URI: fmt.Sprintf("%s://%s%s", scheme, c.Request().Host, c.Echo().URI(s.handleTraffic, euiWithPrefix)),
	}
	data, err = json.Marshal(res)
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal response message")
		writeDiscoverError(ctx, ws, "Router not provisioned")
		return err
	}
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		logger.WithError(err).Warn("Failed to write discover response message")
		return err
	}
	logger.Debug("Sent discover response message")
	return nil
}

var euiHexPattern = regexp.MustCompile("^eui-([a-f0-9A-F]{16})$")

func (s *srv) handleTraffic(c echo.Context) (err error) {
	var sessionID int32
	id := c.Param("id")
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	ctx := c.Request().Context()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"endpoint", "traffic",
		"remote_addr", c.Request().RemoteAddr,
	))

	// Convert the ID to EUI.
	str := euiHexPattern.FindStringSubmatch(id)
	if len(str) != 2 {
		return errGatewayID.WithAttributes("id", id)
	}
	hexValue, err := hex.DecodeString(str[1])
	if err != nil {
		return errGatewayID.WithAttributes("id", id)
	}
	var eui types.EUI64
	eui.UnmarshalBinary(hexValue)

	ctx, ids, err := s.server.FillGatewayContext(ctx, ttnpb.GatewayIdentifiers{EUI: &eui})
	if err != nil {
		return err
	}

	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	var md metadata.MD

	if auth != "" {
		if !strings.HasPrefix(auth, "Bearer ") {
			auth = fmt.Sprintf("Bearer %s", auth)
		}
		md = metadata.New(map[string]string{
			"id":            ids.GatewayID,
			"authorization": auth,
		})
	}

	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	// If a fallback frequency is defined in the server context, inject it into local the context.
	if fallback, ok := frequencyplans.FallbackIDFromContext(s.ctx); ok {
		ctx = frequencyplans.WithFallbackID(ctx, fallback)
	}

	if auth == "" {
		// If the server allows unauthenticated connections (for local testing), we provide the link rights ourselves.
		if s.cfg.AllowUnauthenticated {
			ctx = rights.NewContext(ctx, rights.Rights{
				GatewayRights: map[string]*ttnpb.Rights{
					uid: {
						Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_LINK},
					},
				},
			})
		} else {
			// We error here directly as there is no need make an RPC call to the IS to get a failed rights check due to no Auth.
			return errNoAuthProvided
		}
	}

	conn, err := s.server.Connect(ctx, s, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return err
	}

	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		conn.Disconnect(err)
		return err
	}
	defer ws.Close()
	wsWriteMu := &sync.Mutex{}

	defer func() {
		conn.Disconnect(err)
		err = nil // Errors are sent over the websocket connection that is established by this point.
	}()

	fps := conn.FrequencyPlans()
	bandID := conn.BandID()

	pingTicker := time.NewTicker(s.cfg.WSPingInterval)
	defer pingTicker.Stop()

	ws.SetPingHandler(func(data string) error {
		logger.Debug("Received ping from gateway, send pong")
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		if err := ws.WriteMessage(websocket.PongMessage, nil); err != nil {
			logger.WithError(err).Warn("Failed to send pong")
			return err
		}
		return nil
	})

	// Not all gateways support pongs to the server's pings.
	ws.SetPongHandler(func(data string) error {
		logger.Debug("Received pong from gateway")
		return nil
	})

	go func() {
		for {
			select {
			case <-conn.Context().Done():
				ws.Close()
				return
			case <-pingTicker.C:
				wsWriteMu.Lock()
				err := ws.WriteMessage(websocket.PingMessage, nil)
				wsWriteMu.Unlock()
				if err != nil {
					logger.WithError(err).Warn("Failed to send ping message")
					conn.Disconnect(err)
					ws.Close()
					return
				}
			case down := <-conn.Down():
				dlTime := time.Now()
				scheduledMsg := down.GetScheduled()

				// The first 16 bits of XTime gets the session ID from the upstream latestXTime and the other 48 bits are concentrator timestamp accounted for rollover.
				sID := atomic.LoadInt32(&sessionID)
				concentratorTime, ok := conn.TimeFromTimestampTime(scheduledMsg.Timestamp)
				if !ok {
					logger.Warn("No clock synchronization")
					continue
				}
				xTime := int64(sID)<<48 | (int64(concentratorTime) / int64(time.Microsecond) & 0xFFFFFFFFFF)
				dnmsg, err := s.formatter.FromDownlink(ids, down.GetRawPayload(), scheduledMsg, int64(s.tokens.Next(down.CorrelationIDs, dlTime)), dlTime, xTime)
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}

				logger.Info("Send downlink message")
				wsWriteMu.Lock()
				err = ws.WriteMessage(websocket.TextMessage, dnmsg)
				wsWriteMu.Unlock()
				if err != nil {
					logger.WithError(err).Warn("Failed to send downlink message")
					conn.Disconnect(err)
					return
				}
			}
		}
	}()
	recordTime := func(parsedTime ParsedTime, server time.Time) {
		// The session is the 16 MSB.
		atomic.StoreInt32(&sessionID, int32(parsedTime.XTime>>48))
		conn.SyncWithGatewayConcentrator(
			// The concentrator timestamp is the 32 LSB.
			uint32(parsedTime.XTime&0xFFFFFFFF),
			server,
			// The Basic Station epoch is the 48 LSB.
			scheduling.ConcentratorTime(time.Duration(parsedTime.XTime&0xFFFFFFFFFF)*time.Microsecond),
		)
		sec, nsec := math.Modf(parsedTime.RefTime)
		if sec != 0 {
			ref := time.Unix(int64(sec), int64(nsec*1e9))
			conn.RecordRTT(server.Sub(ref), server)
		}
	}

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			logger.WithError(err).Debug("Failed to read message")
			return err
		}

		typ, err := Type(data)
		if err != nil {
			logger.WithError(err).Debug("Failed to parse message type")
			continue
		}
		logger := logger.WithFields(log.Fields(
			"upstream_type", typ,
		))
		receivedAt := time.Now()

		switch typ {
		case TypeUpstreamVersion:
			ctx, msg, stat, err := s.formatter.GetRouterConfig(ctx, data, bandID, fps, time.Now())
			if err != nil {
				logger.WithError(err).Warn("Failed to generate router configuration")
				return err
			}
			logger := log.FromContext(ctx)
			wsWriteMu.Lock()
			err = ws.WriteMessage(websocket.TextMessage, msg)
			wsWriteMu.Unlock()
			if err != nil {
				logger.WithError(err).Warn("Failed to send router configuration")
				return err
			}
			if err := conn.HandleStatus(stat); err != nil {
				logger.WithError(err).Warn("Failed to send version response message")
			}

		case TypeUpstreamJoinRequest, TypeUpstreamUplinkDataFrame:
			up, parsedTime, err := s.formatter.ToUplink(ctx, data, ids, bandID, receivedAt, typ)
			if err != nil {
				logger.WithError(err).Debug("Failed to parse upstream message")
				return err
			}
			// TODO: Remove (https://github.com/lorabasics/basicstation/issues/74)
			if parsedTime.XTime == 0 {
				logger.Warn("Received join-request without xtime, drop message")
				break
			}
			if err := conn.HandleUp(up); err != nil {
				logger.WithError(err).Warn("Failed to handle upstream message")
			}
			recordTime(parsedTime, receivedAt)

		case TypeUpstreamTxConfirmation:
			txAck, parsedTime, err := s.formatter.ToTxAck(ctx, data, s.tokens, receivedAt)
			if err != nil {
				logger.WithError(err).Debug("Failed to parse tx confirmation frame")
				return err
			}
			if err := conn.HandleTxAck(txAck); err != nil {
				logger.WithError(err).Warn("Failed to handle tx ack message")
			}
			recordTime(parsedTime, receivedAt)

		case TypeUpstreamProprietaryDataFrame, TypeUpstreamRemoteShell, TypeUpstreamTimeSync:
			logger.WithField("message_type", typ).Debug("Message type not implemented")

		default:
			logger.WithField("message_type", typ).Debug("Unknown message type")
		}
	}
}

// writeDiscoverError sends the error messages during the discovery on the WS connection to the station.
func writeDiscoverError(ctx context.Context, ws *websocket.Conn, msg string) {
	logger := log.FromContext(ctx)
	errMsg, err := json.Marshal(DiscoverResponse{Error: msg})
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal error message")
		return
	}
	if err := ws.WriteMessage(websocket.TextMessage, errMsg); err != nil {
		logger.WithError(err).Warn("Failed to write error response message")
	}
}

type errorMessage struct {
	Message string `json:"message"`
}

// errorHandler is an echo.HTTPErrorHandler.
func errorHandler(err error, c echo.Context) {
	if httpErr, ok := err.(*echo.HTTPError); ok {
		c.JSON(httpErr.Code, httpErr.Message)
		return
	}

	statusCode, description := http.StatusInternalServerError, ""
	if ttnErr, ok := errors.From(err); ok {
		if !errors.IsInternal(ttnErr) {
			description = ttnErr.Error()
		}
		statusCode = errors.ToHTTPStatusCode(ttnErr)
	}
	if description != "" {
		c.JSON(statusCode, errorMessage{description})
	} else {
		c.NoContent(statusCode)
	}
}
