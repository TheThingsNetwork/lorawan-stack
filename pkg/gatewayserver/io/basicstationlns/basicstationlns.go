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

package basicstationlns

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/gorilla/websocket"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstationlns/messages"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/log"
	pfconfig "go.thethings.network/lorawan-stack/pkg/pfconfig/basicstationlns"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/middleware"
	"google.golang.org/grpc/metadata"
)

var (
	errEmptyGatewayEUI = errors.Define("empty_gateway_eui", "empty gateway EUI")
	errListener        = errors.DefineFailedPrecondition(
		"listener",
		"failed to serve Basic Station frontend listener",
	)
	errGatewayID = errors.DefineInvalidArgument("invalid_gateway_id", "invalid gateway id `{id}`")
)

type srv struct {
	ctx                  context.Context
	server               io.Server
	webServer            *echo.Echo
	upgrader             *websocket.Upgrader
	tokens               io.DownlinkTokens
	useTrafficTLSAddress bool
}

func (*srv) Protocol() string            { return "basicstation" }
func (*srv) SupportsDownlinkClaim() bool { return false }

// New creates the Basic Station front end.
func New(ctx context.Context, server io.Server, useTrafficTLSAddress bool) *echo.Echo {
	webServer := echo.New()
	webServer.Logger = web.NewNoopLogger()
	webServer.HTTPErrorHandler = errorHandler
	webServer.Use(
		middleware.ID(""),
		echomiddleware.BodyLimit("16M"),
		middleware.Recover(),
	)

	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/basicstation")
	s := &srv{
		ctx:                  ctx,
		server:               server,
		upgrader:             &websocket.Upgrader{},
		webServer:            webServer,
		useTrafficTLSAddress: useTrafficTLSAddress,
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
	var req messages.DiscoverQuery
	if err := json.Unmarshal(data, &req); err != nil {
		logger.WithError(err).Debug("Failed to parse discover query message")
		return err
	}

	if req.EUI.IsZero() {
		writeDiscoverError(s.ctx, ws, "Empty router EUI provided")
		return errEmptyGatewayEUI
	}

	ids := ttnpb.GatewayIdentifiers{
		EUI: &req.EUI.EUI64,
	}
	ctx, ids, err = s.server.FillGatewayContext(ctx, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to fetch gateway")
		writeDiscoverError(s.ctx, ws, fmt.Sprintf("Failed to fetch gateway: %s", err.Error()))
		return err
	}

	scheme := "ws"
	if c.IsTLS() || s.useTrafficTLSAddress {
		scheme = "wss"
	}

	euiWithPrefix := fmt.Sprintf("eui-%s", ids.EUI.String())
	res := messages.DiscoverResponse{
		EUI: req.EUI,
		Muxs: basicstation.EUI{
			Prefix: "muxs",
		},
		URI: fmt.Sprintf("%s://%s%s", scheme, c.Request().Host, c.Echo().URI(s.handleTraffic, euiWithPrefix)),
	}
	data, err = json.Marshal(res)
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal response message")
		writeDiscoverError(s.ctx, ws, "Router not provisioned")
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
	var md metadata.MD

	if auth != "" {
		if !strings.Contains(auth, "Bearer") {
			auth = fmt.Sprintf("Bearer %s", auth)
		}
		md = metadata.New(map[string]string{
			"id":            id,
			"authorization": auth,
		})
	}

	if ctxMd, ok := metadata.FromIncomingContext(s.ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	ctx = metadata.NewIncomingContext(s.ctx, md)
	// If a fallback frequency is defined in the server context, inject it into local the context.
	if fallback, ok := frequencyplans.FallbackIDFromContext(s.ctx); ok {
		ctx = frequencyplans.WithFallbackID(ctx, fallback)
	}

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

	// For gateways with valid EUIs and no auth, we provide the link rights ourselves as in the udp frontend.
	if auth == "" {
		ctx = rights.NewContext(ctx, rights.Rights{
			GatewayRights: map[string]*ttnpb.Rights{
				uid: {
					Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_LINK},
				},
			},
		})
	}

	conn, err := s.server.Connect(ctx, s, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return err
	}
	defer func() {
		conn.Disconnect(err)
	}()

	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		return err
	}
	defer ws.Close()

	fp := conn.FrequencyPlan()

	go func() {
		for {
			select {
			case <-conn.Context().Done():
				return
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
				xTime := int64(sID)<<48 | int64(concentratorTime)/int64(time.Microsecond)
				dnmsg := messages.FromDownlinkMessage(ids, down.GetRawPayload(), scheduledMsg, int64(s.tokens.Next(down.CorrelationIDs, dlTime)), dlTime, xTime)
				msg, err := dnmsg.MarshalJSON()
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}

				logger.Info("Send downlink message")
				if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					logger.WithError(err).Warn("Failed to send downlink message")
					conn.Disconnect(err)
					return
				}
			}
		}
	}()

	var syncedConcentratorTime bool
	recordXTime := func(timestamp int64, server time.Time) {
		// The session is the 16 MSB.
		atomic.StoreInt32(&sessionID, int32(timestamp>>48))
		if !syncedConcentratorTime {
			conn.SyncWithGatewayConcentrator(
				// The concentrator timestamp is the 32 LSB.
				uint32(timestamp&0xFFFFFFFF),
				server,
				// The Basic Station epoch is the 48 LSB.
				scheduling.ConcentratorTime(timestamp&0xFFFFFFFFFF),
			)
			syncedConcentratorTime = true
		}
	}

	for {
		select {
		case <-conn.Context().Done():
			return conn.Context().Err()
		default:
			_, data, err := ws.ReadMessage()
			if err != nil {
				logger.WithError(err).Debug("Failed to read message")
				conn.Disconnect(err)
				return nil
			}

			typ, err := messages.Type(data)
			if err != nil {
				logger.WithError(err).Debug("Failed to parse message type")
				continue
			}
			logger := logger.WithFields(log.Fields(
				"upstream_type", typ,
			))
			receivedAt := time.Now()

			switch typ {
			case messages.TypeUpstreamVersion:
				var version messages.Version
				if err := json.Unmarshal(data, &version); err != nil {
					logger.WithError(err).Debug("Failed to unmarshal version message")
					return err
				}
				logger := logger.WithFields(log.Fields(
					"station", version.Station,
					"firmware", version.Firmware,
					"model", version.Model,
				))
				cfg, err := pfconfig.GetRouterConfig(*fp, version.IsProduction(), time.Now())
				if err != nil {
					logger.WithError(err).Warn("Failed to generate router configuration")
					return err
				}
				data, err = cfg.MarshalJSON()
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal response message")
					return err
				}
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					logger.WithError(err).Warn("Failed to send router configuration")
					return err
				}
				stat := &ttnpb.GatewayStatus{
					Time: receivedAt,
					Versions: map[string]string{
						"station":  version.Station,
						"firmware": version.Firmware,
						"package":  version.Package,
					},
					Advanced: &pbtypes.Struct{
						Fields: map[string]*pbtypes.Value{
							"model": {
								Kind: &pbtypes.Value_StringValue{
									StringValue: version.Model,
								},
							},
							"features": {
								Kind: &pbtypes.Value_StringValue{
									StringValue: version.Features,
								},
							},
						},
					},
				}
				if err := conn.HandleStatus(stat); err != nil {
					logger.WithError(err).Warn("Failed to send status message")
				}

			case messages.TypeUpstreamJoinRequest:
				var jreq messages.JoinRequest
				if err := json.Unmarshal(data, &jreq); err != nil {
					logger.WithError(err).Debug("Failed to unmarshal join-request message")
					return nil
				}
				up, err := jreq.ToUplinkMessage(ids, fp.BandID, receivedAt)
				if err != nil {
					logger.WithError(err).Debug("Failed to parse join-request message")
					return nil
				}
				recordXTime(jreq.UpInfo.XTime, up.ReceivedAt)
				if err := conn.HandleUp(up); err != nil {
					logger.WithError(err).Warn("Failed to handle uplink message")
				}
				recordRTT(conn, receivedAt, jreq.RefTime)

			case messages.TypeUpstreamUplinkDataFrame:
				var updf messages.UplinkDataFrame
				if err := json.Unmarshal(data, &updf); err != nil {
					logger.WithError(err).Debug("Failed to unmarshal uplink data frame")
					return nil
				}
				up, err := updf.ToUplinkMessage(ids, fp.BandID, receivedAt)
				if err != nil {
					logger.WithError(err).Debug("Failed to parse uplink data frame")
					return nil
				}
				recordXTime(updf.UpInfo.XTime, up.ReceivedAt)
				if err := conn.HandleUp(up); err != nil {
					logger.WithError(err).Warn("Failed to handle uplink message")
				}
				recordRTT(conn, receivedAt, updf.RefTime)

			case messages.TypeUpstreamTxConfirmation:
				var txConf messages.TxConfirmation
				if err := json.Unmarshal(data, &txConf); err != nil {
					logger.WithError(err).Debug("Failed to unmarshal Tx acknowledgement frame")
					return nil
				}
				if cids, _, ok := s.tokens.Get(uint16(txConf.Diid), receivedAt); ok {
					txAck := messages.ToTxAcknowledgment(cids)
					if err := conn.HandleTxAck(&txAck); err != nil {
						logger.WithField("diid", txConf.Diid).Warn("Failed to handle Tx acknowledgement")
					}
				} else {
					logger.WithField("diid", txConf.Diid).Debug("Tx acknowledgement either does not correspond to a downlink message or arrived too late")
				}
				recordRTT(conn, receivedAt, txConf.RefTime)

			case messages.TypeUpstreamProprietaryDataFrame, messages.TypeUpstreamRemoteShell, messages.TypeUpstreamTimeSync:
				logger.WithField("message_type", typ).Debug("Message type not implemented")

			default:
				logger.WithField("message_type", typ).Debug("Unknown message type")
			}
		}
	}
}

// writeDiscoverError sends the error messages during the discovery on the WS connection to the station.
func writeDiscoverError(ctx context.Context, ws *websocket.Conn, msg string) {
	logger := log.FromContext(ctx)
	errMsg, err := json.Marshal(messages.DiscoverResponse{Error: msg})
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal error message")
		return
	}
	if err := ws.WriteMessage(websocket.TextMessage, errMsg); err != nil {
		logger.WithError(err).Warn("Failed to write error response message")
	}
}

func recordRTT(conn *io.Connection, receivedAt time.Time, refTime float64) {
	sec, nsec := math.Modf(refTime)
	if sec != 0 {
		ref := time.Unix(int64(sec), int64(nsec*1e9))
		conn.RecordRTT(receivedAt.Sub(ref))
	}
}

// errorHandler is an echo.HTTPErrorHandler.
func errorHandler(err error, c echo.Context) {
	if httpErr, ok := err.(*echo.HTTPError); ok {
		c.JSON(httpErr.Code, httpErr.Message)
		return
	}
}
