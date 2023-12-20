// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package events contains the internal events APi for the Console.
package events

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mileusna/useragent"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/eventsmux"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/middleware"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/subscriptions"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"nhooyr.io/websocket"
)

const (
	authorizationProtocolPrefix = "ttn.lorawan.v3.header.authorization.bearer."
	protocolV1                  = "ttn.lorawan.v3.console.internal.events.v1"

	pingPeriod = time.Minute
	pingJitter = 0.1
)

// Component is the interface of the component to the events API handler.
type Component interface {
	task.Starter
	Context() context.Context
	RateLimiter() ratelimit.Interface
	GetBaseConfig(context.Context) config.ServiceBase
}

type eventsHandler struct {
	component    Component
	subscriber   events.Subscriber
	definedNames map[string]struct{}
}

var _ web.Registerer = (*eventsHandler)(nil)

func (h *eventsHandler) RegisterRoutes(server *web.Server) {
	router := server.APIRouter().PathPrefix(ttnpb.HTTPAPIPrefix + "/console/internal/events/").Subrouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Namespace("console/internal/events")),
		mux.MiddlewareFunc(middleware.ProtocolAuthentication(authorizationProtocolPrefix)),
		mux.MiddlewareFunc(webmiddleware.Metadata("Authorization")),
		ratelimit.HTTPMiddleware(h.component.RateLimiter(), "http:console:internal:events"),
	)
	router.Path("/").HandlerFunc(h.handleEvents).Methods(http.MethodGet)
}

func (h *eventsHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx)

	if err := rights.RequireAuthenticated(ctx); err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	rateLimit, err := makeRateLimiter(ctx, h.component.RateLimiter())
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}

	// Safari versions above 15 cannot handle compression correctly when the
	// `NSURLSession Websocket` experimental feature is enabled (it is enabled by default).
	// Versions above 17 still show the same issues, but the experimental feature is baseline.
	// As such, we disable compression for Safari for all versions in order to ensure the best
	// user experience.
	// https://github.com/TheThingsNetwork/lorawan-stack/issues/6782
	compressionMode := websocket.CompressionContextTakeover
	if ua := useragent.Parse(r.UserAgent()); ua.Name == useragent.Safari {
		compressionMode = websocket.CompressionDisabled
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{protocolV1},
		InsecureSkipVerify: true, // CORS is not enabled for APIs.
		CompressionMode:    compressionMode,
	})
	if err != nil {
		logger.WithError(err).Debug("Failed to accept WebSocket")
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "main task closed")

	ctx, cancel := errorcontext.New(ctx)
	defer cancel(nil)

	var wg sync.WaitGroup
	defer wg.Wait()

	m := eventsmux.New(func(ctx context.Context, cancel func(error)) subscriptions.Interface {
		return subscriptions.New(ctx, cancel, h.subscriber, h.definedNames, h.component)
	})
	for name, f := range map[string]func(context.Context) error{
		"console_events_mux":   makeMuxTask(m, cancel),
		"console_events_read":  makeReadTask(conn, m, rateLimit, cancel),
		"console_events_write": makeWriteTask(conn, m, cancel),
		"console_events_ping":  makePingTask(conn, cancel, random.Jitter(pingPeriod, pingJitter)),
	} {
		wg.Add(1)
		h.component.StartTask(&task.Config{
			Context: ctx,
			ID:      name,
			Func:    f,
			Done:    wg.Done,
			Restart: task.RestartNever,
			Backoff: task.DefaultBackoffConfig,
		})
	}
}

// Option configures the events API handler.
type Option func(*eventsHandler)

// WithSubscriber configures the Subscriber to use for events.
func WithSubscriber(subscriber events.Subscriber) Option {
	return func(h *eventsHandler) {
		h.subscriber = subscriber
	}
}

// New returns an events API handler for the Console.
func New(c Component, opts ...Option) web.Registerer {
	definedNames := make(map[string]struct{})
	for _, def := range events.All().Definitions() {
		definedNames[def.Name()] = struct{}{}
	}
	h := &eventsHandler{
		component:    c,
		subscriber:   events.DefaultPubSub(),
		definedNames: definedNames,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
