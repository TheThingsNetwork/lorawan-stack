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

// Package eventsmux implements the events mux.
package eventsmux

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/protocol"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/subscriptions"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// Interface is the interface for the events mux.
type Interface interface {
	// Requests returns the channel for requests.
	Requests() chan<- protocol.Request
	// Responses returns the channel for responses.
	Responses() <-chan protocol.Response

	// Run runs the events mux.
	Run(context.Context) error
}

type mux struct {
	createSubs func(context.Context, func(error)) subscriptions.Interface

	requestCh  chan protocol.Request
	responseCh chan protocol.Response
}

// Requests implements Interface.
func (m *mux) Requests() chan<- protocol.Request {
	return m.requestCh
}

// Responses implements Interface.
func (m *mux) Responses() <-chan protocol.Response {
	return m.responseCh
}

// Run implements Interface.
func (m *mux) Run(ctx context.Context) (err error) {
	ctx, cancel := errorcontext.New(ctx)
	defer func() { cancel(err) }()
	subs := m.createSubs(ctx, cancel)
	defer subs.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req := <-m.requestCh:
			if err := rights.RequireAuthenticated(ctx); err != nil {
				return err
			}
			var resp protocol.Response
			switch req := req.(type) {
			case *protocol.SubscribeRequest:
				resp = req.Response(subs.Subscribe(req.ID, req.Identifiers, req.After, req.Tail, req.Names))
			case *protocol.UnsubscribeRequest:
				resp = req.Response(subs.Unsubscribe(req.ID))
			default:
				panic("unreachable")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m.responseCh <- resp:
			}
		case subEvt := <-subs.SubscriptionEvents():
			evtPB, err := events.Proto(subEvt.Event)
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to convert event to proto")
				continue
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m.responseCh <- &protocol.PublishResponse{
				ID:    subEvt.ID,
				Event: evtPB,
			}:
			}
		}
	}
}

// New returns a new Interface.
func New(createSubs func(context.Context, func(error)) subscriptions.Interface) Interface {
	return &mux{
		createSubs: createSubs,

		requestCh:  make(chan protocol.Request, 1),
		responseCh: make(chan protocol.Response, 1),
	}
}
