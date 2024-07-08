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

// Package mux implements an events.PubSub implementation that uses multiplexing for subscriptions.
package mux

import (
	"context"
	"sort"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"golang.org/x/sync/errgroup"
)

type stream struct {
	ps      events.Subscriber
	matcher Matcher
}

type multiplexer struct {
	c       workerpool.Component
	ps      events.PubSub
	streams []stream
}

// Option is an option for the multiplexer.
type Option interface {
	apply(*multiplexer)
}

type optionFunc func(*multiplexer)

func (f optionFunc) apply(m *multiplexer) { f(m) }

// WithStream creates a new stream with the given matcher and subscriber.
func WithStream(ps events.Subscriber, matcher Matcher) Option {
	return optionFunc(func(m *multiplexer) {
		m.streams = append(m.streams, stream{
			ps:      ps,
			matcher: matcher,
		})
	})
}

// New creates a multiplexer.
// Events are always published to the given PubSub; streams are additional, read-only subscribers.
// If the given PubSub implements events.Store, a store implementation is returned. If the given streams are also
// stores, the store implementation will be used for fetching history. However, if a stream is not a store,
// SubscribeWithHistory will call Subscribe for that stream.
func New(c workerpool.Component, ps events.PubSub, opts ...Option) events.PubSub {
	m := multiplexer{
		c:  c,
		ps: ps,
	}
	for _, opt := range opts {
		opt.apply(&m)
	}
	if _, hasStore := ps.(events.Store); hasStore {
		return &multiplexerStore{
			multiplexer: m,
		}
	}
	return &m
}

func (m *multiplexer) matchingStreams(names ...string) []stream {
	streams := make([]stream, 0, len(m.streams))
nextStream:
	for _, s := range m.streams {
		for _, name := range names {
			if s.matcher.Matches(name) {
				streams = append(streams, s)
				continue nextStream
			}
		}
	}
	return streams
}

// Publish implements events.PubSub.
func (m *multiplexer) Publish(evts ...events.Event) {
	m.ps.Publish(evts...)
}

// Subscribe implements events.PubSub.
func (m *multiplexer) Subscribe(
	ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl events.Handler,
) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	if err := m.ps.Subscribe(ctx, names, identifiers, hdl); err != nil {
		return err
	}
	for _, stream := range m.matchingStreams(names...) {
		if err := stream.ps.Subscribe(ctx, names, identifiers, hdl); err != nil {
			return err
		}
	}
	return nil
}

// multiplexerStore is an [events.Store] implementation that uses multiplexing for subscriptions.
type multiplexerStore struct {
	multiplexer
}

func (*multiplexerStore) fromSubscribers(
	ctx context.Context,
	subs []events.SubscriberWithHistory,
	f func(context.Context, events.SubscriberWithHistory) ([]events.Event, error),
) ([]events.Event, error) {
	group, ctx := errgroup.WithContext(ctx)
	res := make(chan []events.Event, len(subs))
	for _, ps := range subs {
		ps := ps
		group.Go(func() error {
			evts, err := f(ctx, ps)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case res <- evts:
				return nil
			}
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	close(res)
	var evts []events.Event
	for r := range res {
		evts = append(evts, r...)
	}
	sort.Slice(evts, func(i, j int) bool {
		return evts[i].Time().Before(evts[j].Time())
	})
	return evts, nil
}

// FetchHistory implements events.Store.
func (m *multiplexerStore) FetchHistory(
	ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int,
) ([]events.Event, error) {
	var streams []stream
	if len(names) == 0 { // names are optional
		streams = m.streams
	} else {
		streams = m.matchingStreams(names...)
	}
	subs := make([]events.SubscriberWithHistory, 0, len(streams)+1)
	subs = append(subs, m.ps.(events.SubscriberWithHistory))
	for _, s := range streams {
		if subscriberWithHistory, ok := s.ps.(events.SubscriberWithHistory); ok {
			subs = append(subs, subscriberWithHistory)
		}
	}
	return m.fromSubscribers(ctx, subs,
		func(ctx context.Context, s events.SubscriberWithHistory) ([]events.Event, error) {
			return s.FetchHistory(ctx, names, ids, after, tail)
		},
	)
}

// FindRelated implements events.Store.
func (m *multiplexerStore) FindRelated(ctx context.Context, correlationID string) ([]events.Event, error) {
	subs := make([]events.SubscriberWithHistory, 0, len(m.streams)+1)
	subs = append(subs, m.ps.(events.SubscriberWithHistory))
	for _, s := range m.streams {
		if subWithHistory, ok := s.ps.(events.SubscriberWithHistory); ok {
			subs = append(subs, subWithHistory)
		}
	}
	return m.fromSubscribers(ctx, subs,
		func(ctx context.Context, s events.SubscriberWithHistory) ([]events.Event, error) {
			return s.FindRelated(ctx, correlationID)
		},
	)
}

// SubscribeWithHistory implements events.Store.
func (m *multiplexerStore) SubscribeWithHistory(
	ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int, hdl events.Handler,
) error {
	streams := m.matchingStreams(names...)
	subs := make([]events.Subscriber, 0, len(streams)+1)
	subs = append(subs, m.ps)
	for _, s := range streams {
		subs = append(subs, s.ps)
	}
	wg, ctx := errgroup.WithContext(ctx)
	for _, sub := range subs {
		sub := sub
		if subWithHistory, hasHistory := sub.(events.SubscriberWithHistory); hasHistory {
			wg.Go(func() error {
				return subWithHistory.SubscribeWithHistory(ctx, names, ids, after, tail, hdl)
			})
		} else {
			wg.Go(func() error {
				return sub.Subscribe(ctx, names, ids, hdl)
			})
		}
	}
	return wg.Wait()
}
