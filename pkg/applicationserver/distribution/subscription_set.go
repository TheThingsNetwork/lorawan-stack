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

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// NewSubscriptionSet creates a Distributor that has only one
// underlying set of subscribers. The lifetime of the Distributor
// is bound to the one of the provided context.
func NewSubscriptionSet(ctx context.Context) Distributor {
	set := &subscriptionSet{
		ctx:           ctx,
		subscribeCh:   make(chan *io.Subscription),
		unsubscribeCh: make(chan *io.Subscription),
		upCh:          make(chan *io.ContextualApplicationUp),
	}
	go set.run()
	return set
}

type subscriptionSet struct {
	ctx context.Context

	subscribeCh   chan *io.Subscription
	unsubscribeCh chan *io.Subscription

	upCh chan *io.ContextualApplicationUp
}

// Subscribe implements Distributor.
func (s *subscriptionSet) Subscribe(ctx context.Context, sub *io.Subscription) error {
	select {
	case <-s.ctx.Done():
		// Propagate the subscription set shutdown to the subscription.
		sub.Disconnect(s.ctx.Err())
		return s.ctx.Err()
	case <-ctx.Done():
		// Propagate the main context shutdown to the subscription.
		sub.Disconnect(ctx.Err())
		return ctx.Err()
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
	return nil
}

// SendUp implements Distributor.
func (s *subscriptionSet) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	ctxUp := &io.ContextualApplicationUp{
		Context:       ctx,
		ApplicationUp: up,
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.upCh <- ctxUp:
		return nil
	}
}

func (s *subscriptionSet) run() {
	subscribers := make(map[*io.Subscription]string)

	defer func() {
		for sub, correlationID := range subscribers {
			s.observeUnsubscribe(correlationID, sub)
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case sub := <-s.subscribeCh:
			correlationID := fmt.Sprintf("as:subscriber:%s", events.NewCorrelationID())
			subscribers[sub] = correlationID
			s.observeSubscribe(correlationID, sub)
		case sub := <-s.unsubscribeCh:
			if correlationID, ok := subscribers[sub]; ok {
				delete(subscribers, sub)
				s.observeUnsubscribe(correlationID, sub)
			}
		case up := <-s.upCh:
			for sub := range subscribers {
				if err := sub.SendUp(up.Context, up.ApplicationUp); err != nil {
					log.FromContext(sub.Context()).WithError(err).Warn("Send message failed")
				}
			}
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
