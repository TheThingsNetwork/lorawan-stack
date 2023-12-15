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

package subscriptions_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/subscriptions"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type subscribeRequest struct {
	Context     context.Context
	Names       []string
	Identifiers []*ttnpb.EntityIdentifiers
	After       *time.Time
	Tail        int
	Handler     events.Handler

	Response chan<- error
}

type mockSubscriber struct {
	subReqs chan subscribeRequest
}

func (m *mockSubscriber) subscribeRequests() <-chan subscribeRequest { return m.subReqs }

// Subscribe implements events.Subscriber.
func (m *mockSubscriber) Subscribe(
	ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl events.Handler,
) error {
	ch := make(chan error, 1)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.subReqs <- subscribeRequest{
		Context:     ctx,
		Names:       names,
		Identifiers: identifiers,
		Handler:     hdl,

		Response: ch,
	}:
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-ch:
			return err
		}
	}
}

var _ events.Subscriber = (*mockSubscriber)(nil)

type mockPubSubStore struct {
	subReqs chan subscribeRequest
}

func (m *mockPubSubStore) subscribeRequests() <-chan subscribeRequest { return m.subReqs }

func (*mockPubSubStore) historical() {}

// Publish implements events.Store.
func (*mockPubSubStore) Publish(...events.Event) { panic("not implemented") }

// Subscribe implements events.Store.
func (*mockPubSubStore) Subscribe(context.Context, []string, []*ttnpb.EntityIdentifiers, events.Handler) error {
	panic("not implemented")
}

// FindRelated implements events.Store.
func (*mockPubSubStore) FindRelated(context.Context, string) ([]events.Event, error) {
	panic("not implemented")
}

// FetchHistory implements events.Store.
func (*mockPubSubStore) FetchHistory(
	context.Context, []string, []*ttnpb.EntityIdentifiers, *time.Time, int,
) ([]events.Event, error) {
	panic("not implemented")
}

// SubscribeWithHistory implements events.Store.
func (m *mockPubSubStore) SubscribeWithHistory(
	ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int, hdl events.Handler,
) error {
	ch := make(chan error, 1)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.subReqs <- subscribeRequest{
		Context:     ctx,
		Names:       names,
		Identifiers: ids,
		After:       after,
		Tail:        tail,
		Handler:     hdl,

		Response: ch,
	}:
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-ch:
			return err
		}
	}
}

var _ events.Store = (*mockPubSubStore)(nil)

func runTestSubscriptions(
	t *testing.T,
	subscriber interface {
		events.Subscriber
		subscribeRequests() <-chan subscribeRequest
	},
) {
	t.Helper()

	_, historical := subscriber.(interface{ historical() })

	a, ctx := test.New(t)
	ctx, cancel := errorcontext.New(ctx)
	defer cancel(nil)

	timeout := test.Delay << 3
	app1IDs, app2IDs := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "foo",
	}, &ttnpb.ApplicationIdentifiers{
		ApplicationId: "bar",
	}
	ctx = rights.NewContext(ctx, &rights.Rights{
		ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, app1IDs): ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_ALL),
		}),
	})

	sub := subscriptions.New(
		ctx,
		cancel,
		subscriber,
		map[string]struct{}{
			"test": {},
		},
		task.StartTaskFunc(task.DefaultStartTask),
	)
	defer sub.Close()

	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout):
	case req := <-subscriber.subscribeRequests():
		t.Fatal("Unexpected subscribe request", req)
	}

	now := time.Now()

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sub.Subscribe(
			1,
			[]*ttnpb.EntityIdentifiers{
				app1IDs.GetEntityIdentifiers(),
			},
			&now,
			10,
			[]string{"test"},
		)
		a.So(err, should.BeNil)
	}()
	var handler events.Handler
	select {
	case <-ctx.Done():
		return
	case req := <-subscriber.subscribeRequests():
		a.So(req.Context, should.HaveParentContextOrEqual, ctx)
		a.So(req.Names, should.Resemble, []string{"test"})
		a.So(req.Identifiers, should.Resemble, []*ttnpb.EntityIdentifiers{
			app1IDs.GetEntityIdentifiers(),
		})
		if historical {
			a.So(req.After, should.Resemble, &now)
			a.So(req.Tail, should.Equal, 10)
		}
		a.So(req.Handler, should.NotBeNil)
		if !historical {
			select {
			case <-ctx.Done():
				return
			case req.Response <- nil:
			}
		}
		handler = req.Handler
	}
	wg.Wait()

	err := sub.Subscribe(
		1,
		[]*ttnpb.EntityIdentifiers{
			app1IDs.GetEntityIdentifiers(),
		},
		&now,
		10,
		[]string{"test"},
	)
	a.So(err, should.NotBeNil)

	evt := events.New(
		ctx,
		"test",
		"test",
		events.WithIdentifiers(app2IDs),
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
	)
	handler.Notify(evt)

	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout):
	case subEvt := <-sub.SubscriptionEvents():
		t.Fatal("Unexpected subscription event", subEvt)
	}

	evt = events.New(
		ctx,
		"test",
		"test",
		events.WithIdentifiers(app1IDs),
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
	)
	handler.Notify(evt)

	select {
	case <-ctx.Done():
		return
	case subEvt := <-sub.SubscriptionEvents():
		a.So(subEvt.ID, should.Equal, 1)
		a.So(subEvt.Event, should.ResembleEvent, evt)
	}

	err = sub.Unsubscribe(1)
	a.So(err, should.BeNil)

	err = sub.Unsubscribe(1)
	a.So(err, should.NotBeNil)

	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout):
	case subEvt := <-sub.SubscriptionEvents():
		t.Fatal("Unexpected subscription event", subEvt)
	}
}

func TestSubscriptions(t *testing.T) {
	t.Parallel()
	runTestSubscriptions(
		t,
		&mockSubscriber{
			subReqs: make(chan subscribeRequest, 1),
		},
	)
}

func TestStoreSubscriptions(t *testing.T) {
	t.Parallel()
	runTestSubscriptions(
		t,
		&mockPubSubStore{
			subReqs: make(chan subscribeRequest, 1),
		},
	)
}
