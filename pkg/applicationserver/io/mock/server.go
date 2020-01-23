// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package mock

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type server struct {
	*component.Component
	subscriptionsMu sync.RWMutex
	subscriptions   map[string][]*io.Subscription
	subscriptionsCh chan *io.Subscription
	downlinkQueueMu sync.RWMutex
	downlinkQueue   map[string][]*ttnpb.ApplicationDownlink
	subscribeError  error
}

// Server represents a mock io.Server.
type Server interface {
	io.Server

	SetSubscribeError(error)
	Subscriptions() <-chan *io.Subscription
}

// NewServer instantiates a new Server.
func NewServer(c *component.Component) Server {
	return &server{
		Component:       c,
		subscriptions:   make(map[string][]*io.Subscription),
		subscriptionsCh: make(chan *io.Subscription, 10),
		downlinkQueue:   make(map[string][]*ttnpb.ApplicationDownlink),
	}
}

// FillContext implements io.Server.
func (s *server) FillContext(ctx context.Context) context.Context {
	return s.Component.FillContext(ctx)
}

func (s *server) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	s.subscriptionsMu.RLock()
	defer s.subscriptionsMu.RUnlock()
	for _, sub := range s.subscriptions[unique.ID(ctx, up.ApplicationIdentifiers)] {
		if err := sub.SendUp(ctx, up); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe implements io.Server.
func (s *server) Subscribe(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*io.Subscription, error) {
	s.subscriptionsMu.RLock()
	err := s.subscribeError
	s.subscriptionsMu.RUnlock()
	if err != nil {
		return nil, err
	}
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	sub := io.NewSubscription(ctx, protocol, &ids)
	s.subscriptionsMu.Lock()
	s.subscriptions[unique.ID(ctx, ids)] = append(s.subscriptions[unique.ID(ctx, ids)], sub)
	s.subscriptionsMu.Unlock()
	select {
	case s.subscriptionsCh <- sub:
	default:
	}
	return sub, nil
}

// DownlinkQueuePush implements io.Server.
func (s *server) DownlinkQueuePush(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	s.downlinkQueueMu.Lock()
	uid := unique.ID(ctx, ids)
	s.downlinkQueue[uid] = append(s.downlinkQueue[uid], io.CleanDownlinks(items)...)
	s.downlinkQueueMu.Unlock()
	return nil
}

// DownlinkQueueReplace implements io.Server.
func (s *server) DownlinkQueueReplace(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	s.downlinkQueueMu.Lock()
	s.downlinkQueue[unique.ID(ctx, ids)] = io.CleanDownlinks(items)
	s.downlinkQueueMu.Unlock()
	return nil
}

// DownlinkQueueList implements io.Server.
func (s *server) DownlinkQueueList(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error) {
	s.downlinkQueueMu.RLock()
	queue := s.downlinkQueue[unique.ID(ctx, ids)]
	s.downlinkQueueMu.RUnlock()
	return queue, nil
}

func (s *server) SetSubscribeError(err error) {
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()
	s.subscribeError = err
}

func (s *server) Subscriptions() <-chan *io.Subscription {
	return s.subscriptionsCh
}
