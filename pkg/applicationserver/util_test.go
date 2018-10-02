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

package applicationserver

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type memStore struct {
	mu    sync.RWMutex
	items map[string][]byte
	New   func() proto.Unmarshaler
}

var errNotFound = errors.DefineNotFound("not_found", "not found")

func (s *memStore) Get(uid string) (proto.Unmarshaler, error) {
	s.mu.RLock()
	buf, ok := s.items[uid]
	s.mu.RUnlock()
	if !ok {
		return nil, errNotFound
	}
	v := s.New()
	if err := v.Unmarshal(buf); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *memStore) Set(uid string, f func(proto.Unmarshaler) (proto.Marshaler, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var v proto.Unmarshaler
	buf, ok := s.items[uid]
	if ok {
		v = s.New()
		if err := v.Unmarshal(buf); err != nil {
			return err
		}
	}
	n, err := f(v)
	if err != nil {
		return err
	}
	if n == nil {
		delete(s.items, uid)
	} else if buf, err := n.Marshal(); err != nil {
		return err
	} else {
		s.items[uid] = buf
	}
	return nil
}

func (s *memStore) Range(f func(string, proto.Unmarshaler) bool) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for uid, buf := range s.items {
		v := s.New()
		if err := v.Unmarshal(buf); err != nil {
			return err
		}
		if !f(uid, v) {
			break
		}
	}
	return nil
}

type memDeviceRegistry struct {
	store memStore
}

func newMemDeviceRegistry() DeviceRegistry {
	return &memDeviceRegistry{
		store: memStore{
			items: make(map[string][]byte),
			New:   func() proto.Unmarshaler { return new(ttnpb.EndDevice) },
		},
	}
}

func (r *memDeviceRegistry) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevice, error) {
	v, err := r.store.Get(unique.ID(ctx, ids))
	if err != nil {
		return nil, err
	}
	return v.(*ttnpb.EndDevice), nil
}

func (r *memDeviceRegistry) Set(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, error)) error {
	return r.store.Set(unique.ID(ctx, ids), func(v proto.Unmarshaler) (proto.Marshaler, error) {
		var ed *ttnpb.EndDevice
		if v != nil {
			ed = v.(*ttnpb.EndDevice)
		}
		return f(ed)
	})
}

type memLinkRegistry struct {
	store memStore
}

func newMemLinkRegistry() LinkRegistry {
	return &memLinkRegistry{
		store: memStore{
			items: make(map[string][]byte),
			New:   func() proto.Unmarshaler { return new(ttnpb.ApplicationLink) },
		},
	}
}

func (r *memLinkRegistry) Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*ttnpb.ApplicationLink, error) {
	v, err := r.store.Get(unique.ID(ctx, ids))
	if err != nil {
		return nil, err
	}
	return v.(*ttnpb.ApplicationLink), nil
}

func (r *memLinkRegistry) Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, error)) error {
	return r.store.Set(unique.ID(ctx, ids), func(v proto.Unmarshaler) (proto.Marshaler, error) {
		var l *ttnpb.ApplicationLink
		if v != nil {
			l = v.(*ttnpb.ApplicationLink)
		}
		return f(l)
	})
}

func (r *memLinkRegistry) Range(ctx context.Context, f func(ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error {
	var ferr error
	err := r.store.Range(func(uid string, v proto.Unmarshaler) bool {
		ids, ferr := unique.ToApplicationID(uid)
		if ferr != nil {
			return false
		}
		return f(ids, v.(*ttnpb.ApplicationLink))
	})
	if err != nil {
		return err
	}
	if ferr != nil {
		return ferr
	}
	return nil
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.PeerInfo_Role) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if peer := c.GetPeer(ctx, role, nil); peer != nil {
			return
		}
	}
	panic("could not connect to peer")
}

type mockNS struct {
	ttnpb.AsNsServer
	linkCh chan ttnpb.ApplicationIdentifiers
	upCh   chan *ttnpb.ApplicationUp
}

func startMockNS(ctx context.Context) (*mockNS, string) {
	ns := &mockNS{
		linkCh: make(chan ttnpb.ApplicationIdentifiers, 1),
		upCh:   make(chan *ttnpb.ApplicationUp, 1),
	}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterAsNsServer(srv.Server, ns)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return ns, lis.Addr().String()
}

func (ns *mockNS) LinkApplication(ids *ttnpb.ApplicationIdentifiers, stream ttnpb.AsNs_LinkApplicationServer) error {
	ns.linkCh <- *ids
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case up := <-ns.upCh:
			if err := stream.Send(up); err != nil {
				return err
			}
		}
	}
}
