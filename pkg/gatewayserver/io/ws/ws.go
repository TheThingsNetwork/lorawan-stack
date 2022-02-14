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
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"google.golang.org/grpc/metadata"
)

var (
	errGatewayID          = errors.DefineInvalidArgument("invalid_gateway_id", "invalid gateway ID `{id}`")
	errNoAuthProvided     = errors.DefineUnauthenticated("no_auth_provided", "no auth provided `{uid}`")
	errMissedTooManyPongs = errors.Define("missed_too_many_pongs", "gateway missed too many pongs")
)

type srv struct {
	ctx       context.Context
	server    io.Server
	upgrader  *websocket.Upgrader
	cfg       Config
	formatter Formatter
}

func (s *srv) Protocol() string            { return "ws" }
func (s *srv) SupportsDownlinkClaim() bool { return false }

// New creates a new WebSocket frontend.
func New(ctx context.Context, server io.Server, formatter Formatter, cfg Config) (*web.Server, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/ws")

	s := &srv{
		ctx:    ctx,
		server: server,
		upgrader: &websocket.Upgrader{
			HandshakeTimeout: 120 * time.Second,
			WriteBufferPool:  &sync.Pool{},
			Error: func(w http.ResponseWriter, r *http.Request, _ int, err error) {
				webhandlers.Error(w, r, err)
			},
		},
		formatter: formatter,
		cfg:       cfg,
	}

	web, err := web.New(ctx, web.WithDisableWarnings(true))
	if err != nil {
		return nil, err
	}

	router := web.RootRouter()
	router.Use(
		ratelimit.HTTPMiddleware(server.RateLimiter(), "gs:accept:ws"),
	)

	eps := s.formatter.Endpoints()
	router.HandleFunc(eps.ConnectionInfo, s.handleConnectionInfo).Methods(http.MethodGet)
	router.HandleFunc(eps.Traffic, func(w http.ResponseWriter, r *http.Request) {
		if err := s.handleTraffic(w, r); err != nil {
			webhandlers.Error(w, r, err)
		}
	}).Methods(http.MethodGet)

	return web, nil
}

func (s *srv) handleConnectionInfo(w http.ResponseWriter, r *http.Request) {
	eps := s.formatter.Endpoints()
	ctx := log.NewContextWithFields(r.Context(), log.Fields(
		"endpoint", eps.ConnectionInfo,
		"remote_addr", r.RemoteAddr,
	))
	logger := log.FromContext(ctx)
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		return
	}
	defer ws.Close()

	_, data, err := ws.ReadMessage()
	if err != nil {
		logger.WithError(err).Debug("Failed to read message")
		return
	}

	scheme := "ws"
	if r.TLS != nil || s.cfg.UseTrafficTLSAddress {
		scheme = "wss"
	}

	info := ServerInfo{
		Scheme:  scheme,
		Address: r.Host,
	}

	resp := s.formatter.HandleConnectionInfo(ctx, data, s.server, info, time.Now())
	if err := ws.WriteMessage(websocket.TextMessage, resp); err != nil {
		logger.WithError(err).Warn("Failed to write connection info response message")
		return
	}
	logger.Debug("Sent connection info response message")
}

var euiHexPattern = regexp.MustCompile("^eui-([a-f0-9A-F]{16})$")

func (s *srv) handleTraffic(w http.ResponseWriter, r *http.Request) (err error) {
	const noPongReceived int64 = -1
	var (
		id           = mux.Vars(r)["id"]
		auth         = r.Header.Get("Authorization")
		ctx          = r.Context()
		eps          = s.formatter.Endpoints()
		missedPongs  = noPongReceived
		pongCh       = make(chan struct{}, 1)
		downstreamCh = make(chan []byte, 1)
	)

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"endpoint", eps.Traffic,
		"remote_addr", r.RemoteAddr,
	))
	ctx = NewContextWithSession(ctx, &Session{})

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

	ctx, ids, err := s.server.FillGatewayContext(ctx, ttnpb.GatewayIdentifiers{Eui: &eui})
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
			"id":            ids.GatewayId,
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
						Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK},
					},
				},
			})
		} else {
			// We error here directly as there is no need make an RPC call to the IS to get a failed rights check due to no Auth.
			return errNoAuthProvided.WithAttributes("uid", uid)
		}
	}

	logger := log.FromContext(ctx)

	conn, err := s.server.Connect(ctx, s, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return err
	}

	defer func() {
		conn.Disconnect(err)
		err = nil // Errors are sent over the websocket connection that is established by this point.
	}()

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Debug("Failed to upgrade request to websocket connection")
		return err
	}
	defer ws.Close()
	pingTicker := time.NewTicker(random.Jitter(s.cfg.WSPingInterval, 0.1))
	defer pingTicker.Stop()

	ws.SetPingHandler(func(data string) error {
		logger.Debug("Received ping from gateway, send pong")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pongCh <- struct{}{}:
		}
		return nil
	})

	// Not all gateways support pongs to the server's pings.
	ws.SetPongHandler(func(data string) error {
		atomic.StoreInt64(&missedPongs, 0)
		logger.Debug("Received pong from gateway")
		return nil
	})

	var timeSyncTickerC <-chan time.Time
	if s.cfg.TimeSyncInterval > 0 {
		ticker := time.NewTicker(random.Jitter(s.cfg.TimeSyncInterval, 0.1))
		timeSyncTickerC = ticker.C
		defer ticker.Stop()
	}

	go func() (err error) {
		defer ws.Close()
		defer func() {
			if err != nil {
				conn.Disconnect(err)
			}
		}()
		for {
			select {
			case <-conn.Context().Done():
				return
			case <-pingTicker.C:
				if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
					logger.WithError(err).Warn("Failed to send ping message")
					return err
				}
				if atomic.LoadInt64(&missedPongs) == noPongReceived {
					continue
				}
				if atomic.AddInt64(&missedPongs, 1) == int64(s.cfg.MissedPongThreshold) {
					err := errMissedTooManyPongs.New()
					logger.WithError(err).Warn("Disconnect gateway")
					return err
				}
			case <-pongCh:
				if err := ws.WriteMessage(websocket.PongMessage, nil); err != nil {
					logger.WithError(err).Warn("Failed to send pong")
					return err
				}
			case <-timeSyncTickerC:
				// TODO: Use GPS timestamp from a overlapping frames.
				// https://github.com/TheThingsNetwork/lorawan-stack/issues/4852
				b, err := s.formatter.TransferTime(ctx, time.Now(), nil, nil)
				if err != nil {
					logger.WithError(err).Warn("Failed to generate time transfer")
					return err
				}
				if b == nil {
					continue
				}
				if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
					logger.WithError(err).Warn("Failed to transfer time")
					return err
				}
			case down := <-conn.Down():
				concentratorTime, ok := conn.TimeFromTimestampTime(down.GetScheduled().Timestamp)
				if !ok {
					logger.Warn("No clock synchronization")
					continue
				}
				dnmsg, err := s.formatter.FromDownlink(ctx, *down, conn.BandID(), concentratorTime, time.Now())
				if err != nil {
					logger.WithError(err).Warn("Failed to marshal downlink message")
					continue
				}
				if err = ws.WriteMessage(websocket.TextMessage, dnmsg); err != nil {
					logger.WithError(err).Warn("Failed to send downlink message")
					return err
				}
			case downstream := <-downstreamCh:
				if err := ws.WriteMessage(websocket.TextMessage, downstream); err != nil {
					logger.WithError(err).Warn("Failed to send message downstream")
					return err
				}
			}
		}
	}()

	resource := ratelimit.GatewayUpResource(ctx, ids)
	for {
		if err := ratelimit.Require(s.server.RateLimiter(), resource); err != nil {
			logger.WithError(err).Warn("Terminate connection")
			return err
		}
		_, data, err := ws.ReadMessage()
		if err != nil {
			logger.WithError(err).Debug("Failed to read message")
			return err
		}
		downstream, err := s.formatter.HandleUp(ctx, data, ids, conn, time.Now())
		if err != nil {
			return err
		}
		if downstream == nil {
			continue
		}
		logger.Info("Send downstream message")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case downstreamCh <- downstream:
		}
	}
}
