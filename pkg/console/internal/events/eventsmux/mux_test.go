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

package eventsmux_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/eventsmux"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/protocol"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/subscriptions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type subscribeRequest struct {
	ID          uint64
	Identifiers []*ttnpb.EntityIdentifiers
	After       *time.Time
	Tail        uint32
	Names       []string

	Response chan<- error
}

type unsubscribeRequest struct {
	ID uint64

	Response chan<- error
}

type mockSubscriptions struct {
	ctx       context.Context
	subReqs   chan subscribeRequest
	unsubReqs chan unsubscribeRequest
	evsCh     chan *subscriptions.SubscriptionEvent
}

// Subscribe implements subscriptions.Interface.
func (m *mockSubscriptions) Subscribe(
	id uint64, identifiers []*ttnpb.EntityIdentifiers, after *time.Time, tail uint32, names []string,
) error {
	ch := make(chan error, 1)
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	case m.subReqs <- subscribeRequest{
		ID:          id,
		Identifiers: identifiers,
		After:       after,
		Tail:        tail,
		Names:       names,

		Response: ch,
	}:
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		case err := <-ch:
			return err
		}
	}
}

// Unsubscribe implements subscriptions.Interface.
func (m *mockSubscriptions) Unsubscribe(id uint64) error {
	ch := make(chan error, 1)
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	case m.unsubReqs <- unsubscribeRequest{
		ID: id,

		Response: ch,
	}:
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		case err := <-ch:
			return err
		}
	}
}

// SubscriptionEvents implements subscriptions.Interface.
func (m *mockSubscriptions) SubscriptionEvents() <-chan *subscriptions.SubscriptionEvent {
	return m.evsCh
}

// Close implements subscriptions.Interface.
func (*mockSubscriptions) Close() error { return nil }

var _ subscriptions.Interface = (*mockSubscriptions)(nil)

func TestMux(t *testing.T) { // nolint:gocyclo
	t.Parallel()

	a, ctx := test.New(t)

	appIDs := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "foo",
	}
	ctx = rights.NewContext(ctx, &rights.Rights{
		ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, appIDs): ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL),
		}),
	})
	ctx = rights.NewContextWithAuthInfo(ctx, &ttnpb.AuthInfoResponse{
		UniversalRights: ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL),
		IsAdmin:         true,
	})

	subs := &mockSubscriptions{
		ctx:       ctx,
		subReqs:   make(chan subscribeRequest, 1),
		unsubReqs: make(chan unsubscribeRequest, 1),
		evsCh:     make(chan *subscriptions.SubscriptionEvent, 1),
	}
	m := eventsmux.New(func(context.Context, func(error)) subscriptions.Interface { return subs })

	go m.Run(ctx) // nolint:errcheck

	now := time.Now()
	select {
	case <-ctx.Done():
		return
	case m.Requests() <- &protocol.SubscribeRequest{
		ID: 42,
		Identifiers: []*ttnpb.EntityIdentifiers{
			appIDs.GetEntityIdentifiers(),
		},
		After: &now,
		Tail:  1,
		Names: []string{"foo"},
	}:
	}
	select {
	case <-ctx.Done():
		return
	case req := <-subs.subReqs:
		a.So(req, should.Resemble, subscribeRequest{
			ID: 42,
			Identifiers: []*ttnpb.EntityIdentifiers{
				appIDs.GetEntityIdentifiers(),
			},
			After: &now,
			Tail:  1,
			Names: []string{"foo"},

			Response: req.Response,
		})
		select {
		case <-ctx.Done():
			return
		case req.Response <- nil:
		}
	}
	select {
	case <-ctx.Done():
		return
	case resp := <-m.Responses():
		a.So(resp, should.Resemble, &protocol.SubscribeResponse{
			ID: 42,
		})
	}

	errAlreadySubscribed := errors.New("already subscribed")
	select {
	case <-ctx.Done():
		return
	case m.Requests() <- &protocol.SubscribeRequest{
		ID: 42,
		Identifiers: []*ttnpb.EntityIdentifiers{
			appIDs.GetEntityIdentifiers(),
		},
		After: &now,
		Tail:  1,
		Names: []string{"foo"},
	}:
	}
	select {
	case <-ctx.Done():
		return
	case req := <-subs.subReqs:
		a.So(req, should.Resemble, subscribeRequest{
			ID: 42,
			Identifiers: []*ttnpb.EntityIdentifiers{
				appIDs.GetEntityIdentifiers(),
			},
			After: &now,
			Tail:  1,
			Names: []string{"foo"},

			Response: req.Response,
		})
		select {
		case <-ctx.Done():
			return
		case req.Response <- errAlreadySubscribed:
		}
	}
	select {
	case <-ctx.Done():
		return
	case resp := <-m.Responses():
		a.So(resp, should.Resemble, &protocol.ErrorResponse{
			ID:    42,
			Error: status.New(codes.Unknown, "already subscribed"),
		})
	}

	ev := events.New(
		ctx,
		"test.evt",
		"test event",
		events.WithIdentifiers(appIDs),
	)
	select {
	case <-ctx.Done():
		return
	case subs.evsCh <- &subscriptions.SubscriptionEvent{
		ID:    42,
		Event: ev,
	}:
	}
	select {
	case <-ctx.Done():
		return
	case resp := <-m.Responses():
		a.So(resp, should.Resemble, &protocol.PublishResponse{
			ID:    42,
			Event: test.Must(events.Proto(ev)),
		})
	}

	select {
	case <-ctx.Done():
		return
	case m.Requests() <- &protocol.UnsubscribeRequest{
		ID: 42,
	}:
	}
	select {
	case <-ctx.Done():
		return
	case req := <-subs.unsubReqs:
		a.So(req, should.Resemble, unsubscribeRequest{
			ID: 42,

			Response: req.Response,
		})
		select {
		case <-ctx.Done():
			return
		case req.Response <- nil:
		}
	}
	select {
	case <-ctx.Done():
		return
	case resp := <-m.Responses():
		a.So(resp, should.Resemble, &protocol.UnsubscribeResponse{
			ID: 42,
		})
	}

	errNotSubscribed := errors.New("not subscribed")
	select {
	case <-ctx.Done():
		return
	case m.Requests() <- &protocol.UnsubscribeRequest{
		ID: 42,
	}:
	}
	select {
	case <-ctx.Done():
		return
	case req := <-subs.unsubReqs:
		a.So(req, should.Resemble, unsubscribeRequest{
			ID: 42,

			Response: req.Response,
		})
		select {
		case <-ctx.Done():
			return
		case req.Response <- errNotSubscribed:
		}
	}
	select {
	case <-ctx.Done():
		return
	case resp := <-m.Responses():
		a.So(resp, should.Resemble, &protocol.ErrorResponse{
			ID:    42,
			Error: status.New(codes.Unknown, "not subscribed"),
		})
	}
}
