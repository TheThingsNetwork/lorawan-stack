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

// Package ws provides common interface for Web Socket front end.
package ws

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
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
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
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

const pingIntervalJitter = 0.1

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

func (*srv) Protocol() string            { return "ws" }
func (*srv) SupportsDownlinkClaim() bool { return false }
func (*srv) DutyCycleStyle() scheduling.DutyCycleStyle {
	return scheduling.DutyCycleStyleBlockingWindow
}

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

	w, err := web.New(
		ctx,
		web.WithDisableWarnings(true),
		web.WithTrustedProxies(server.GetBaseConfig(ctx).HTTP.TrustedProxies...),
	)
	if err != nil {
		return nil, err
	}

	router := w.RootRouter()
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

	return w, nil
}

func (s *srv) handleConnectionInfo(w http.ResponseWriter, r *http.Request) {
	eps := s.formatter.Endpoints()
	ctx := log.NewContextWithFields(r.Context(), log.Fields(
		"endpoint", eps.ConnectionInfo,
		"remote_addr", r.RemoteAddr,
	))
	logger := log.FromContext(ctx)

	assertAuth := func(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
		ctx, hasAuth := withForwardedAuth(ctx, ids, r.Header.Get("Authorization"))
		if !hasAuth {
			if !s.cfg.AllowUnauthenticated {
				return errNoAuthProvided.WithAttributes("uid", unique.ID(ctx, ids))
			}
			return nil
		}
		return s.server.AssertGatewayRights(ctx, ids, ttnpb.Right_RIGHT_GATEWAY_LINK)
	}

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

	var (
		scheme = "ws"
		port   = "80"
	)
	if r.TLS != nil || s.cfg.UseTrafficTLSAddress {
		scheme = "wss"
		port = "443"
	}

	// If port is retrievable from the host, use it.
	host, p, err := net.SplitHostPort(r.Host)
	if err == nil {
		// Both `host` and `p` are valid, since we have no error.
		port = p
	} else {
		// `host` and `p` are unknown/empty. Reset `host` to `r.Host` since it does not contain the port.
		host = r.Host
	}

	info := ServerInfo{
		Scheme:  scheme,
		Address: net.JoinHostPort(host, port),
	}

	resp := s.formatter.HandleConnectionInfo(ctx, data, s.server, info, assertAuth)
	if err := ws.WriteMessage(websocket.TextMessage, resp); err != nil {
		logger.WithError(err).Warn("Failed to write connection info response message")
		return
	}
	logger.Debug("Sent connection info response message")
}

var euiHexPattern = regexp.MustCompile("^eui-([a-f0-9A-F]{16})$")

func (s *srv) handleTraffic(w http.ResponseWriter, r *http.Request) (err error) {
	var (
		id           = mux.Vars(r)["id"]
		auth         = r.Header.Get("Authorization")
		ctx          = r.Context()
		eps          = s.formatter.Endpoints()
		missingPongs = int64(0)
		pongCount    = int64(0)
		pongCh       = make(chan []byte, 1)
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

	ctx, ids, err := s.server.FillGatewayContext(ctx, &ttnpb.GatewayIdentifiers{Eui: eui.Bytes()})
	if err != nil {
		return err
	}

	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	// If a fallback frequency is defined in the server context, inject it into local the context.
	if fallback, ok := frequencyplans.FallbackIDFromContext(s.ctx); ok {
		ctx = frequencyplans.WithFallbackID(ctx, fallback)
	}

	ctx, hasAuth := withForwardedAuth(ctx, ids, auth)
	if !hasAuth {
		if !s.cfg.AllowUnauthenticated {
			// We error here directly as there is no auth.
			return errNoAuthProvided.WithAttributes("uid", uid)
		}
		// If the server allows unauthenticated connections (for local testing), we provide the link rights ourselves.
		ctx = rights.NewContext(ctx, &rights.Rights{
			GatewayRights: *rights.NewMap(map[string]*ttnpb.Rights{
				uid: {
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK},
				},
			}),
		})
	}

	logger := log.FromContext(ctx)
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err
	}
	if xRealIP := r.Header[http.CanonicalHeaderKey("X-Real-IP")]; len(xRealIP) == 1 {
		addr = xRealIP[0]
	}

	conn, err := s.server.Connect(ctx, s, ids, &ttnpb.GatewayRemoteAddress{
		Ip: addr,
	})
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

	var pingTickerC <-chan time.Time
	if s.cfg.MissedPongThreshold > 0 && random.CanJitter(s.cfg.WSPingInterval, pingIntervalJitter) {
		pingTicker := time.NewTicker(random.Jitter(s.cfg.WSPingInterval, pingIntervalJitter))
		pingTickerC = pingTicker.C
		defer pingTicker.Stop()
	}

	ws.SetPingHandler(func(data string) error {
		logger.Debug("Received client ping")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pongCh <- []byte(data):
		}
		return nil
	})

	// Not all gateways support pongs to the server's pings.
	ws.SetPongHandler(func(data string) error {
		logger.Debug("Received client pong")
		for n := atomic.LoadInt64(&missingPongs); ; n = atomic.LoadInt64(&missingPongs) {
			if n == 0 {
				logger.Warn("Unsolicited client pong")
				return nil
			}
			if atomic.CompareAndSwapInt64(&missingPongs, n, n-1) {
				break
			}
		}
		atomic.AddInt64(&pongCount, 1)
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
			case <-pingTickerC:
				if atomic.AddInt64(&missingPongs, 1) > int64(s.cfg.MissedPongThreshold) &&
					atomic.LoadInt64(&pongCount) > 0 {
					err := errMissedTooManyPongs.New()
					logger.WithError(err).Warn("Gateway missed too many pings")
					return err
				}
				if err := ws.WriteControl(websocket.PingMessage, nil, time.Time{}); err != nil {
					logger.WithError(err).Warn("Failed to send ping message")
					return err
				}
				logger.Debug("Server ping sent")
			case data := <-pongCh:
				if err := ws.WriteControl(websocket.PongMessage, data, time.Time{}); err != nil {
					logger.WithError(err).Warn("Failed to send pong")
					return err
				}
				logger.Debug("Server pong sent")
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
				dnmsg, err := s.formatter.FromDownlink(ctx, down, conn.BandID(), time.Now())
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

func withForwardedAuth(ctx context.Context, ids *ttnpb.GatewayIdentifiers, auth string) (context.Context, bool) {
	var md metadata.MD
	var hasAuth bool
	if auth != "" {
		if !strings.HasPrefix(auth, "Bearer ") {
			auth = fmt.Sprintf("Bearer %s", auth)
		}
		m := map[string]string{"authorization": auth}
		if ids != nil {
			m["id"] = ids.GatewayId
		}
		md = metadata.New(m)
		if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
			md = metadata.Join(ctxMd, md)
		}
		ctx = metadata.NewIncomingContext(ctx, md)
		hasAuth = true
	}
	return ctx, hasAuth
}
