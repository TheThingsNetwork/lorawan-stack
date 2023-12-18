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

// Package subscriptions implements the events mux subscriptions.
package subscriptions

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights/rightsutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// SubscriptionEvent wraps an events.Event with a subscription ID.
type SubscriptionEvent struct {
	ID    uint64
	Event events.Event
}

// Interface is the interface for the events mux subscriptions.
type Interface interface {
	// Subscribe subscribes to events.
	Subscribe(
		id uint64, identifiers []*ttnpb.EntityIdentifiers, after *time.Time, tail uint32, names []string,
	) error
	// Unsubscribe unsubscribe to events.
	Unsubscribe(id uint64) error

	// SubscriptionEvents provides the events for the underlying subscriptions.
	SubscriptionEvents() <-chan *SubscriptionEvent

	// Close closes all of the underlying subscriptions and waits for the background tasks to finish.
	Close() error
}

type subscription struct {
	id           uint64
	cancel       func(error)
	wg           sync.WaitGroup
	cancelParent func(error)
	inputCh      <-chan events.Event
	outputCh     chan<- *SubscriptionEvent
}

func (s *subscription) run(ctx context.Context) (err error) {
	defer func() {
		select {
		case <-ctx.Done():
		default:
			s.cancelParent(err)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt := <-s.inputCh:
			isVisible, err := rightsutil.EventIsVisible(ctx, evt)
			if err != nil {
				if err := rights.RequireAny(ctx, evt.Identifiers()...); err != nil {
					return err
				}
				log.FromContext(ctx).WithError(err).Warn("Failed to check event visibility")
				continue
			}
			if !isVisible {
				continue
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case s.outputCh <- &SubscriptionEvent{
				ID:    s.id,
				Event: evt,
			}:
			}
		}
	}
}

type subscriptions struct {
	ctx          context.Context
	cancel       func(error)
	subscriber   events.Subscriber
	definedNames map[string]struct{}
	taskStarter  task.Starter

	wg   sync.WaitGroup
	ch   chan *SubscriptionEvent
	subs map[uint64]*subscription
}

var _ Interface = (*subscriptions)(nil)

// Close implements Interface.
func (s *subscriptions) Close() error {
	for id, sub := range s.subs {
		delete(s.subs, id)
		sub.cancel(nil)
		sub.wg.Wait()
	}
	s.wg.Wait()
	return nil
}

// SubscriptionEvents implements Interface.
func (s *subscriptions) SubscriptionEvents() <-chan *SubscriptionEvent { return s.ch }

var (
	errAlreadySubscribed = errors.DefineAlreadyExists("already_subscribed", "already subscribed with ID `{id}`")
	errNoIdentifiers     = errors.DefineInvalidArgument("no_identifiers", "no identifiers")
)

// Subscribe implements Interface.
func (s *subscriptions) Subscribe(
	id uint64, identifiers []*ttnpb.EntityIdentifiers, after *time.Time, tail uint32, names []string,
) (err error) {
	if err := s.validateSubscribe(id, identifiers); err != nil {
		return err
	}
	names, err = events.NamesFromPatterns(s.definedNames, names)
	if err != nil {
		return err
	}
	ch := make(chan events.Event, channelSize(tail))
	ctx, cancel := errorcontext.New(s.ctx)
	defer func() {
		if err != nil {
			cancel(err)
		}
	}()
	if store, hasStore := s.subscriber.(events.Store); hasStore {
		if after == nil && tail == 0 {
			now := time.Now()
			after = &now
		}
		f := func(ctx context.Context) (err error) {
			defer func() {
				select {
				case <-ctx.Done():
				default:
					s.cancel(err)
				}
			}()
			return store.SubscribeWithHistory(ctx, names, identifiers, after, int(tail), events.Channel(ch))
		}
		s.wg.Add(1)
		s.taskStarter.StartTask(&task.Config{
			Context: ctx,
			ID:      "console_events_subscribe",
			Func:    f,
			Done:    s.wg.Done,
			Restart: task.RestartNever,
			Backoff: task.DefaultBackoffConfig,
		})
	} else {
		if err := s.subscriber.Subscribe(ctx, names, identifiers, events.Channel(ch)); err != nil {
			return err
		}
	}
	sub := &subscription{
		id:           id,
		cancel:       cancel,
		cancelParent: s.cancel,
		inputCh:      ch,
		outputCh:     s.ch,
	}
	sub.wg.Add(1)
	s.taskStarter.StartTask(&task.Config{
		Context: ctx,
		ID:      "console_events_filter",
		Func:    sub.run,
		Done:    sub.wg.Done,
		Restart: task.RestartNever,
		Backoff: task.DefaultBackoffConfig,
	})
	s.subs[id] = sub
	return nil
}

var errNotSubscribed = errors.DefineNotFound("not_subscribed", "not subscribed with ID `{id}`")

// Unsubscribe implements Interface.
func (s *subscriptions) Unsubscribe(id uint64) error {
	sub, ok := s.subs[id]
	if !ok {
		return errNotSubscribed.WithAttributes("id", id)
	}
	delete(s.subs, id)
	sub.cancel(nil)
	sub.wg.Wait()
	return nil
}

// New returns a new Interface.
func New(
	ctx context.Context,
	cancel func(error),
	subscriber events.Subscriber,
	definedNames map[string]struct{},
	taskStarter task.Starter,
) Interface {
	return &subscriptions{
		ctx:          ctx,
		cancel:       cancel,
		subscriber:   subscriber,
		definedNames: definedNames,
		taskStarter:  taskStarter,
		ch:           make(chan *SubscriptionEvent, 1),
		subs:         make(map[uint64]*subscription),
	}
}

func (s *subscriptions) validateSubscribe(id uint64, identifiers []*ttnpb.EntityIdentifiers) error {
	if _, ok := s.subs[id]; ok {
		return errAlreadySubscribed.WithAttributes("id", id)
	}
	if len(identifiers) == 0 {
		return errNoIdentifiers.New()
	}
	for _, ids := range identifiers {
		if err := ids.ValidateFields(); err != nil {
			return err
		}
	}
	return rights.RequireAny(s.ctx, identifiers...)
}

func channelSize(n uint32) uint32 {
	if n < 8 {
		n = 8
	}
	if n > 1024 {
		n = 1024
	}
	return n
}
