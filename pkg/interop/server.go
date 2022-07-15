// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package interop

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/fillcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

// Registerer allows components to register their interop services to the web server.
type Registerer interface {
	RegisterInterop(s *Server)
}

// IdentityServer represents an Identity Server.
type IdentityServer interface {
	HomeNSRequest(context.Context, *HomeNSReq) (*TTIHomeNSAns, error)
}

// JoinServer represents a Join Server as specified in LoRaWAN Backend Interfaces.
type JoinServer interface {
	JoinRequest(context.Context, *JoinReq) (*JoinAns, error)
	AppSKeyRequest(context.Context, *AppSKeyReq) (*AppSKeyAns, error)
	HomeNSRequest(context.Context, *HomeNSReq) (*TTIHomeNSAns, error)
}

type noopServer struct{}

func (noopServer) JoinRequest(context.Context, *JoinReq) (*JoinAns, error) {
	return nil, ErrMalformedMessage.New()
}

func (noopServer) AppSKeyRequest(context.Context, *AppSKeyReq) (*AppSKeyAns, error) {
	return nil, ErrMalformedMessage.New()
}

func (noopServer) HomeNSRequest(context.Context, *HomeNSReq) (*TTIHomeNSAns, error) {
	return nil, ErrMalformedMessage.New()
}

// Server is the server.
type Server struct {
	config config.InteropServer

	router *mux.Router

	senderClientCAs    map[string][]*x509.Certificate
	senderClientCAPool *x509.CertPool

	tokenVerifiers map[string]tokenVerifier

	is IdentityServer
	js JoinServer
}

// Component represents the Component to the Interop Server.
type Component interface {
	httpclient.Provider
	Context() context.Context
	RateLimiter() ratelimit.Interface
}

// NewServer builds a new server.
func NewServer(c Component, contextFillers []fillcontext.Filler, conf config.InteropServer) (*Server, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "interop")
	logger := log.FromContext(ctx)

	senderClientCAs, err := fetchSenderClientCAs(ctx, conf, c)
	if err != nil {
		return nil, err
	}
	senderClientCAPool := x509.NewCertPool()
	for _, certs := range senderClientCAs {
		for _, cert := range certs {
			senderClientCAPool.AddCert(cert)
		}
	}

	tokenVerifiers := make(map[string]tokenVerifier)
	if conf.PacketBroker.Enabled {
		iss := conf.PacketBroker.TokenIssuer
		// The token audience must match the configured public TLS address. Therefore, a non-empty value must be set.
		aud := conf.PublicTLSAddress
		if aud == "" {
			return nil, errNoPublicTLSAddress.New()
		}
		tokenVerifier, err := newPacketBrokerTokenVerifier(ctx, iss, aud, c)
		if err != nil {
			return nil, err
		}
		tokenVerifiers[iss] = tokenVerifier
	}

	s := &Server{
		config:             conf,
		senderClientCAs:    senderClientCAs,
		senderClientCAPool: senderClientCAPool,
		tokenVerifiers:     tokenVerifiers,
		js:                 &noopServer{},
	}

	s.router = mux.NewRouter()
	s.router.NotFoundHandler = http.HandlerFunc(webhandlers.NotFound)
	s.router.Use(
		mux.MiddlewareFunc(webmiddleware.Recover()),
		mux.MiddlewareFunc(webmiddleware.FillContext(contextFillers...)),
		mux.MiddlewareFunc(webmiddleware.RequestURL()),
		mux.MiddlewareFunc(webmiddleware.RequestID()),
		mux.MiddlewareFunc(webmiddleware.MaxBody(1<<15)), // 32 kB.
		mux.MiddlewareFunc(webmiddleware.Log(logger, nil)),
		mux.MiddlewareFunc(ratelimit.HTTPMiddleware(c.RateLimiter(), "http:interop")),
	)
	s.router.
		NewRoute().
		Handler(s.handle()).
		Headers("Content-Type", "application/json").
		Methods(http.MethodPost)

	return s, nil
}

// ServeHTTP serves the HTTP request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// RegisterIS registers the Identity Server for answering to HomeNSRequest.
func (s *Server) RegisterIS(is IdentityServer) {
	s.is = is
}

// RegisterJS registers the Join Server for AS-JS, hNS-JS and vNS-JS messages.
// If an Identity Server is registered with RegisterIS, the Identity Server takes precedence for handling HomeNSRequest.
func (s *Server) RegisterJS(js JoinServer) {
	s.js = js
}

// ClientCAPool returns a certificate pool of all configured client CAs.
func (s *Server) ClientCAPool() *x509.CertPool {
	return s.senderClientCAPool
}

// SenderClientCAs returns the client certificate authorities that are trusted for the given SenderID.
// The SenderID is typically a NetID, but an AS-ID or JoinEUI can also be used to trust Application Servers
// and Join Servers respectively.
func (s *Server) SenderClientCAs(_ context.Context, senderID string) ([]*x509.Certificate, error) {
	// TODO: Lookup partner CA by SenderID with DNS (https://github.com/TheThingsNetwork/lorawan-stack/issues/718).
	return s.senderClientCAs[senderID], nil
}

func (s *Server) handle() http.Handler {
	senderAuthenticators := map[MessageType]senderAuthenticator{
		MessageTypeJoinReq:    senderAuthenticatorFunc(s.authenticateNS),
		MessageTypeRejoinReq:  senderAuthenticatorFunc(s.authenticateNS),
		MessageTypeAppSKeyReq: senderAuthenticatorFunc(s.authenticateAS),
		MessageTypeHomeNSReq:  senderAuthenticatorFunc(s.authenticateNS),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := fmt.Sprintf("interop:%s", r.Header.Get("X-Request-ID"))
		ctx := events.ContextWithCorrelationID(r.Context(), cid)
		logger := log.FromContext(ctx)

		data, err := io.ReadAll(r.Body)
		if err != nil {
			logger.WithError(err).Debug("Failed to read body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var header MessageHeader
		if err := json.Unmarshal(data, &header); err != nil {
			logger.WithError(err).Debug("Failed to unmarshal body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		logger = logger.WithFields(log.Fields(
			"message_type", header.MessageType,
			"protocol_version", header.ProtocolVersion,
			"sender_id", header.SenderID,
			"receiver_id", header.ReceiverID,
		))
		ctx = log.NewContext(ctx, logger)

		if err := header.MessageType.Validate(header.ProtocolVersion); err != nil {
			logger.WithError(err).Debug("Invalid protocol version or message type")
			writeError(w, r, header, err)
			return
		}

		senderAuthenticator, ok := senderAuthenticators[header.MessageType]
		if !ok {
			writeError(w, r, header, ErrMalformedMessage.New())
			return
		}
		ctx, err = senderAuthenticator.Authenticate(ctx, r, data)
		if err != nil {
			logger.WithError(err).Warn("Failed to authenticate")
			writeError(w, r, header, err)
			return
		}

		var msg interface{}
		switch header.MessageType {
		case MessageTypeJoinReq, MessageTypeRejoinReq:
			msg = &JoinReq{}
		case MessageTypeAppSKeyReq:
			msg = &AppSKeyReq{}
		case MessageTypeHomeNSReq:
			msg = &HomeNSReq{}
		default:
			writeError(w, r, header, ErrMalformedMessage.New())
			return
		}

		if err := json.Unmarshal(data, msg); err != nil {
			writeError(w, r, header, ErrMalformedMessage.WithCause(err))
			return
		}

		var ans interface{}
		switch req := msg.(type) {
		case *JoinReq:
			ans, err = s.js.JoinRequest(ctx, req)
		case *HomeNSReq:
			// The registered Identity Server takes precedence over a registered Join Server to handle HomeNSRequest.
			js := s.is
			if js == nil {
				js = s.js
			}
			ans, err = js.HomeNSRequest(ctx, req)
		case *AppSKeyReq:
			ans, err = s.js.AppSKeyRequest(ctx, req)
		default:
			writeError(w, r, header, ErrMalformedMessage.New())
			return
		}
		if err != nil {
			logger.WithError(err).Warn("Failed to handle request")
			writeError(w, r, header, err)
			return
		}

		json.NewEncoder(w).Encode(ans) //nolint:errcheck
	})
}
