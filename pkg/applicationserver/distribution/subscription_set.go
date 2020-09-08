// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package distribution

import (
	"context"
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const subscriptionSetBufferSize = 16

// newSubscriptionSet creates a new subscription set. The timeout represents
// the period after which the set will shut down if empty. If the timeout is
// zero, the set never timeouts.
func newSubscriptionSet(ctx context.Context, timeout time.Duration) *subscriptionSet {
	ctx, cancel := errorcontext.New(ctx)
	s := &subscriptionSet{
		ctx:           ctx,
		cancel:        cancel,
		timeout:       timeout,
		subscribeCh:   make(chan *io.Subscription, 1),
		unsubscribeCh: make(chan *io.Subscription, 1),
		upCh:          make(chan *io.ContextualApplicationUp, subscriptionSetBufferSize),
	}
	go s.run()
	return s
}

type subscriptionSet struct {
	ctx    context.Context
	cancel errorcontext.CancelFunc

	timeout time.Duration

	subscribeCh   chan *io.Subscription
	unsubscribeCh chan *io.Subscription
	upCh          chan *io.ContextualApplicationUp
}

// Context returns the context of the set.
func (s *subscriptionSet) Context() context.Context {
	return s.ctx
}

// Cancel cancels the set and the associated subscriptions.
func (s *subscriptionSet) Cancel(err error) {
	s.cancel(err)
}

// Subscribe creates a subscription for the provided application with the given protocol.
func (s *subscriptionSet) Subscribe(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers) (*io.Subscription, error) {
	sub := io.NewSubscription(ctx, protocol, ids)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.ctx.Done():
		// Propagate the subscription set shutdown to the subscription.
		sub.Disconnect(s.ctx.Err())
		return nil, s.ctx.Err()
	case s.subscribeCh <- sub:
	}
	go func() {
		select {
		case <-s.ctx.Done():
			// Propagate the subscription set shutdown to the subscription.
			sub.Disconnect(s.ctx.Err())
		case <-sub.Context().Done():
			select {
			case <-s.ctx.Done():
			case s.unsubscribeCh <- sub:
			}
		}
	}()
	return sub, nil
}

// Publish publishes the upstream traffic to the subscribers.
func (s *subscriptionSet) Publish(ctx context.Context, up *ttnpb.ApplicationUp) error {
	ctxUp := &io.ContextualApplicationUp{
		Context:       ctx,
		ApplicationUp: up,
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return s.ctx.Err()
	case s.upCh <- ctxUp:
		return nil
	}
}

var errEmptySet = errors.DefineAborted("empty_set", "empty set")

func (s *subscriptionSet) run() {
	registerSubscriptionSetStart(s.ctx)
	defer registerSubscriptionSetStop(s.ctx)

	subscribers := make(map[*io.Subscription]string)

	defer func() {
		for sub, correlationID := range subscribers {
			s.observeUnsubscribe(correlationID, sub)
		}
	}()

	tickCh, tickStop := newTicker(s.timeout)
	defer tickStop()
	lastAction := time.Now()
	for {
		select {
		case <-s.ctx.Done():
			return
		case sub := <-s.subscribeCh:
			correlationID := fmt.Sprintf("as:subscriber:%s", events.NewCorrelationID())
			subscribers[sub] = correlationID
			s.observeSubscribe(correlationID, sub)
			lastAction = time.Now()
		case sub := <-s.unsubscribeCh:
			if correlationID, ok := subscribers[sub]; ok {
				delete(subscribers, sub)
				s.observeUnsubscribe(correlationID, sub)
			}
			lastAction = time.Now()
		case up := <-s.upCh:
			for sub := range subscribers {
				if err := sub.Publish(up.Context, up.ApplicationUp); err != nil {
					log.FromContext(sub.Context()).WithError(err).Warn("Send message failed")
				}
			}
		case <-tickCh:
			if len(subscribers) > 0 {
				// There are still subscribers in the set.
				continue
			}
			if time.Now().Sub(lastAction) < s.timeout {
				// We had activity in the last period.
				continue
			}
			s.cancel(errEmptySet.New())
			return
		}
	}
}

func (s *subscriptionSet) observeSubscribe(correlationID string, sub *io.Subscription) {
	registerSubscribe(events.ContextWithCorrelationID(s.ctx, correlationID), sub)
	log.FromContext(sub.Context()).Debug("Subscribed")
}

func (s *subscriptionSet) observeUnsubscribe(correlationID string, sub *io.Subscription) {
	registerUnsubscribe(events.ContextWithCorrelationID(s.ctx, correlationID), sub)
	log.FromContext(sub.Context()).Debug("Unsubscribed")
}

// newTicker creates a ticking channel similar to time.Ticker.
// If the period is 0, the ticker will never fire.
func newTicker(d time.Duration) (<-chan time.Time, func()) {
	if d == 0 {
		return make(chan time.Time), func() {}
	}
	t := time.NewTicker(d)
	return t.C, t.Stop
}
