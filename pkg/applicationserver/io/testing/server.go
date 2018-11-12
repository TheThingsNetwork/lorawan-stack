// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package testing

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type server struct {
	connectionsCh   chan *io.Connection
	downlinkQueueMu sync.RWMutex
	downlinkQueue   map[string][]*ttnpb.ApplicationDownlink
}

// Server represents a testing io.Server.
type Server interface {
	io.Server

	Connections() <-chan *io.Connection
}

// NewServer instantiates a new Server.
func NewServer() Server {
	return &server{
		connectionsCh: make(chan *io.Connection, 10),
		downlinkQueue: make(map[string][]*ttnpb.ApplicationDownlink),
	}
}

// Connect implements io.Server.
func (s *server) Connect(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*io.Connection, error) {
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	conn := io.NewConnection(ctx, protocol, &ids)
	select {
	case s.connectionsCh <- conn:
	default:
	}
	return conn, nil
}

// DownlinkQueuePush implements io.Server.
func (s *server) DownlinkQueuePush(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	s.downlinkQueueMu.Lock()
	uid := unique.ID(ctx, ids)
	s.downlinkQueue[uid] = append(s.downlinkQueue[uid], items...)
	s.downlinkQueueMu.Unlock()
	return nil
}

// DownlinkQueueReplace implements io.Server.
func (s *server) DownlinkQueueReplace(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	s.downlinkQueueMu.Lock()
	s.downlinkQueue[unique.ID(ctx, ids)] = items
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

func (s *server) Connections() <-chan *io.Connection {
	return s.connectionsCh
}
