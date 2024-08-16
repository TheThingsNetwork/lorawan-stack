// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package ttigw implements The Things Industries protocol for gateways.
package ttigw

import (
	"context"
	"crypto/x509"
	"fmt"
	stdio "io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	lorav1 "go.thethings.industries/pkg/api/gen/tti/gateway/data/lora/v1"
	ttica "go.thethings.industries/pkg/ca"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/mtls"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"nhooyr.io/websocket"
)

const (
	pingIntervalJitter = 0.1
	pingTimeout        = 10 * time.Second
	subprotocol        = "v1.lora.data.gateway.thethings.industries"
)

// Frontend implements the The Things Industries V1 gateway frontend.
type Frontend struct {
	http.Handler
	server io.Server
	cfg    Config
}

var _ io.Frontend = (*Frontend)(nil)

// New returns a new The Things Industries V1 gateway frontend.
func New(ctx context.Context, server io.Server, cfg Config) (*Frontend, error) {
	var proxyConfiguration webmiddleware.ProxyConfiguration
	if err := proxyConfiguration.ParseAndAddTrusted(server.GetBaseConfig(ctx).HTTP.TrustedProxies...); err != nil {
		return nil, err
	}
	router := mux.NewRouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Recover()),
		otelmux.Middleware("ttn-lw-stack", otelmux.WithTracerProvider(tracing.FromContext(ctx))),
		mux.MiddlewareFunc(webmiddleware.Peer()),
		mux.MiddlewareFunc(webmiddleware.RequestURL()),
		mux.MiddlewareFunc(webmiddleware.RequestID()),
		mux.MiddlewareFunc(webmiddleware.ProxyHeaders(proxyConfiguration)),
		mux.MiddlewareFunc(webmiddleware.MaxBody(1024*4)),
		mux.MiddlewareFunc(webmiddleware.Log(log.FromContext(ctx), nil)),
		ratelimit.HTTPMiddleware(server.RateLimiter(), "gs:accept:ttigw"),
	)

	f := &Frontend{
		Handler: router,
		server:  server,
		cfg:     cfg,
	}

	router.HandleFunc("/api/protocols/tti/v1", f.handleGet).Methods(http.MethodGet)
	return f, nil
}

// DutyCycleStyle implements io.Frontend.
func (*Frontend) DutyCycleStyle() scheduling.DutyCycleStyle {
	return scheduling.DefaultDutyCycleStyle
}

// Protocol implements io.Frontend.
func (*Frontend) Protocol() string {
	return "ttigw"
}

// SupportsDownlinkClaim implements io.Frontend.
func (*Frontend) SupportsDownlinkClaim() bool {
	return false
}

func writeError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	if ttnErr, ok := errors.From(err); ok {
		statusCode = errors.ToHTTPStatusCode(ttnErr)
	}
	if statusCode >= 500 {
		http.Error(w, "internal server error", statusCode)
	} else {
		http.Error(w, err.Error(), statusCode)
	}
}

func (f *Frontend) handleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	cert := mtls.ClientCertificateFromContext(ctx)
	if cert == nil {
		logger.Debug("No client certificate presented")
		http.Error(w, "client certificate required", http.StatusUnauthorized)
		return
	}
	ctx, ids, err := f.authenticate(ctx, cert)
	if err != nil {
		logger.WithError(err).Debug("Client certificate verification failed")
		writeError(w, err)
		return
	}
	logger = log.FromContext(ctx)

	srvConn, err := f.server.Connect(ctx, f, ids, &ttnpb.GatewayRemoteAddress{
		Ip: remoteIP(r),
	})
	if err != nil {
		logger.WithError(err).Info("Failed to connect")
		writeError(w, err)
		return
	}

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{subprotocol},
	})
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade request to websocket connection")
		writeError(w, err)
		return
	}
	defer wsConn.CloseNow() //nolint:errcheck
	if wsConn.Subprotocol() != subprotocol {
		logger.Debug("Subprotocol negotiation failed")
		wsConn.Close(websocket.StatusPolicyViolation, "subprotocol negotiation failed")
		return
	}

	if err := f.handleConnection(wsConn, srvConn); err != nil {
		logger.WithError(err).Debug("Failed to handle connection")
	}
}

var errInvalidGatewayEUI = errors.Define("invalid_gateway_eui", "invalid gateway EUI", "common_name")

func (f *Frontend) authenticate(
	ctx context.Context, cert *x509.Certificate,
) (context.Context, *ttnpb.GatewayIdentifiers, error) {
	var eui types.EUI64
	if err := eui.UnmarshalText([]byte(cert.Subject.CommonName)); err != nil {
		return nil, nil, errInvalidGatewayEUI.WithCause(err).WithAttributes("common_name", cert.Subject.CommonName)
	}

	// Append the The Things Industries gateways CAs to the root CAs in the context:
	// Gateways using this frontend are provisioned by The Things Industries.
	ctx = mtls.AppendRootCAsToContext(ctx, ttica.GatewaysCAs)

	ids := &ttnpb.GatewayIdentifiers{Eui: eui.Bytes()}
	ctx, ids, err := f.server.FillGatewayContext(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)

	return ctx, ids, nil
}

func remoteIP(r *http.Request) string {
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		addr = ""
	}
	if xRealIP := r.Header.Get("x-real-ip"); xRealIP != "" {
		addr = xRealIP
	}
	return addr
}

func (f *Frontend) ping(ctx context.Context, wsConn *websocket.Conn, srvConn *io.Connection) error {
	if !random.CanJitter(f.cfg.WSPingInterval, pingIntervalJitter) {
		return nil
	}
	var (
		ticker      = time.NewTicker(random.Jitter(f.cfg.WSPingInterval, pingIntervalJitter))
		missedPongs = 0
	)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			pingCtx, cancelPing := context.WithTimeout(ctx, pingTimeout)
			start := time.Now()
			err := wsConn.Ping(pingCtx)
			duration := time.Since(start)
			cancelPing()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					missedPongs++
					if missedPongs >= f.cfg.MissedPongThreshold {
						return err
					}
					continue
				}
				return err
			}
			missedPongs = 0
			srvConn.RecordRTT(duration, start)
		}
	}
}

func (f *Frontend) handleConnection(wsConn *websocket.Conn, srvConn *io.Connection) error {
	ctx := srvConn.Context()
	logger := log.FromContext(ctx)

	gwConfig, err := buildLoRaGatewayConfig(srvConn.PrimaryFrequencyPlan())
	if err != nil {
		logger.WithError(err).Warn("Failed to build LoRa gateway configuration")
		wsConn.Close(websocket.StatusInternalError, "failed to build LoRa gateway configuration")
		return err
	}

	go func() {
		if err := f.ping(ctx, wsConn, srvConn); err != nil && !errors.Is(err, context.Canceled) {
			logger.WithError(err).Info("Ping failed")
			wsConn.Close(websocket.StatusPolicyViolation, "ping failed")
		}
	}()

	msgCh := make(chan *lorav1.NetworkServerMessage, 2)
	msgCh <- &lorav1.NetworkServerMessage{
		Message: &lorav1.NetworkServerMessage_ServerHelloNotification{
			ServerHelloNotification: &lorav1.ServerHelloNotification{},
		},
	}
	msgCh <- &lorav1.NetworkServerMessage{
		Message: &lorav1.NetworkServerMessage_ConfigureLoraGatewayRequest{
			ConfigureLoraGatewayRequest: &lorav1.ConfigureLoraGatewayRequest{
				Config: gwConfig,
			},
		},
	}

	dlTokens := &io.DownlinkTokens{}
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return sendMessages(ctx, wsConn, msgCh)
	})
	wg.Go(func() error {
		return enqueueNetworkServerMessages(ctx, srvConn, msgCh, dlTokens)
	})
	wg.Go(func() error {
		return readGatewayMessages(ctx, wsConn, srvConn, dlTokens)
	})
	return wg.Wait()
}

func sendMessages(
	ctx context.Context, wsConn *websocket.Conn, msgCh <-chan *lorav1.NetworkServerMessage,
) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgCh:
			buf, err := proto.Marshal(msg)
			if err != nil {
				logger.WithError(err).Warn("Failed to marshal message")
				continue
			}
			if err := wsConn.Write(ctx, websocket.MessageBinary, buf); err != nil {
				logger.WithError(err).Warn("Failed to write message")
				return err
			}
		}
	}
}

func enqueueNetworkServerMessages(
	ctx context.Context,
	srvConn *io.Connection,
	msgCh chan<- *lorav1.NetworkServerMessage,
	dlTokens *io.DownlinkTokens,
) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case down := <-srvConn.Down():
			logger.Debug("Send downlink message")
			dlToken := dlTokens.Next(down, time.Now())
			msg, err := fromDownlinkMessage(srvConn.PrimaryFrequencyPlan(), down)
			if err != nil {
				logger.WithError(err).Warn("Failed to convert downlink message")
				continue
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case msgCh <- &lorav1.NetworkServerMessage{
				TransactionId: uint32(dlToken),
				Message: &lorav1.NetworkServerMessage_TransmitDownlinkRequest{
					TransmitDownlinkRequest: &lorav1.TransmitDownlinkRequest{
						Message: msg,
					},
				},
			}:
			}
		}
	}
}

var (
	errUnsupportedMessageType = errors.DefineCorruption("unsupported_message_type", "unsupported message type")
	errNoClientHello          = errors.DefineCorruption("no_client_hello", "no client hello")
)

func readGatewayMessages(
	ctx context.Context,
	wsConn *websocket.Conn,
	srvConn *io.Connection,
	dlTokens *io.DownlinkTokens,
) error {
	var (
		logger              = log.FromContext(ctx)
		clientHelloReceived bool
	)
	for {
		typ, buf, err := wsConn.Read(ctx)
		if err != nil {
			switch websocket.CloseStatus(err) {
			case -1:
				if errors.Is(err, stdio.EOF) || errors.Is(err, syscall.ECONNRESET) {
					logger.Info("Connection closed unexpectedly")
				} else {
					logger.WithError(err).Warn("Failed to read from websocket connection")
				}
			case websocket.StatusNormalClosure, websocket.StatusGoingAway:
				logger.Debug("Websocket connection closed normally")
			default:
				logger.WithError(err).Info("Websocket connection closed with error")
			}
			return err
		}
		receivedAt := time.Now()

		if typ != websocket.MessageBinary {
			logger.Debug("Received message with non-binary message type")
			wsConn.Close(websocket.StatusUnsupportedData, "unsupported message type")
			return errUnsupportedMessageType.New()
		}

		var envelope lorav1.GatewayMessage
		if err := proto.Unmarshal(buf, &envelope); err != nil {
			logger.WithError(err).Debug("Failed to unmarshal message")
			wsConn.Close(websocket.StatusInvalidFramePayloadData, "invalid message")
			return err
		}

		if !clientHelloReceived {
			clientHello := envelope.GetClientHelloNotification()
			if clientHello == nil {
				logger.Debug("Received message without client hello")
				wsConn.Close(websocket.StatusPolicyViolation, "client hello required")
				return errNoClientHello.New()
			}
			if err := srvConn.HandleStatus(gatewayStatusFromClientHello(clientHello)); err != nil {
				logger.WithError(err).Warn("Failed to handle status message")
			}
			clientHelloReceived = true
			continue
		}

		if err := processGatewayMessage(ctx, srvConn, dlTokens, &envelope, receivedAt); err != nil {
			logger.WithError(err).Warn("Failed to handle message")
		}
	}
}

var errUnknownMessageType = errors.DefineInvalidArgument("unknown_message_type", "unknown message type", "type")

func processGatewayMessage(
	ctx context.Context,
	srvConn *io.Connection,
	dlTokens *io.DownlinkTokens,
	envelope *lorav1.GatewayMessage,
	receivedAt time.Time,
) error {
	logger := log.FromContext(ctx)
	switch msg := envelope.Message.(type) {
	case *lorav1.GatewayMessage_ErrorNotification:
		txAckResult, isTxAckResult := toTxAcknowledgmentResult[msg.ErrorNotification.Code]
		down, _, isTxResponse := dlTokens.Get(uint16(envelope.TransactionId), receivedAt)
		if isTxAckResult || isTxResponse {
			if !isTxAckResult {
				txAckResult = ttnpb.TxAcknowledgment_UNKNOWN_ERROR
			}
			txAck := &ttnpb.TxAcknowledgment{
				Result: txAckResult,
			}
			if isTxResponse {
				txAck.DownlinkMessage = down
				txAck.CorrelationIds = down.CorrelationIds
			}
			if err := srvConn.HandleTxAck(txAck); err != nil {
				logger.WithError(err).Warn("Failed to handle Tx acknowledgment")
			}
		} else {
			logger.WithFields(log.Fields(
				"transaction_id", envelope.TransactionId,
				"code", msg.ErrorNotification.Code.String(),
				"details", msg.ErrorNotification.Details,
			)).Warn("Received error notification")
		}

	case *lorav1.GatewayMessage_StatusNotification:
		logger.Debug("Received status notification")

	case *lorav1.GatewayMessage_ConfigureLoraGatewayResponse:
		logger.Debug("Configured LoRa gateway")

	case *lorav1.GatewayMessage_UplinkMessagesNotification:
		logger.WithField("count", len(msg.UplinkMessagesNotification.Messages)).Debug("Received uplink messages")
		uplinkMessages := make([]*ttnpb.UplinkMessage, 0, len(msg.UplinkMessagesNotification.Messages))
		for _, uplinkMsg := range msg.UplinkMessagesNotification.Messages {
			up, err := toUplinkMessage(srvConn.Gateway().Ids, srvConn.PrimaryFrequencyPlan(), uplinkMsg)
			if err != nil {
				logger.WithError(err).Warn("Failed to convert uplink message")
				continue
			}
			up.ReceivedAt = timestamppb.New(receivedAt)
			uplinkMessages = append(uplinkMessages, up)
		}
		for _, up := range io.UniqueUplinkMessagesByRSSI(uplinkMessages) {
			if err := srvConn.HandleUp(up, nil); err != nil {
				logger.WithError(err).Warn("Failed to handle uplink message")
			}
		}

	case *lorav1.GatewayMessage_TransmitDownlinkResponse:
		logger.Debug("Received transmit downlink response")
		txAck := &ttnpb.TxAcknowledgment{
			Result: ttnpb.TxAcknowledgment_SUCCESS,
		}
		if down, _, ok := dlTokens.Get(uint16(envelope.TransactionId), receivedAt); ok {
			txAck.DownlinkMessage = down
			txAck.CorrelationIds = down.CorrelationIds
		}
		if err := srvConn.HandleTxAck(txAck); err != nil {
			logger.WithError(err).Warn("Failed to handle Tx acknowledgment")
		}

	default:
		return errUnknownMessageType.WithAttributes("type", fmt.Sprintf("%T", msg))
	}
	return nil
}
