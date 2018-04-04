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

// Package pool abstracts the polling and sending procedures between gateways and the gateway server.
package pool

import (
	"context"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/scheduling"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type gatewayStoreEntry struct {
	channel chan *ttnpb.GatewayDown

	scheduler scheduling.Scheduler

	observations     ttnpb.GatewayObservations
	observationsLock sync.RWMutex
}

type gatewayStore struct {
	store map[ttnpb.GatewayIdentifiers]*gatewayStoreEntry

	mu sync.Mutex
}

func (s *gatewayStore) Store(gatewayID ttnpb.GatewayIdentifiers, entry *gatewayStoreEntry) {
	s.mu.Lock()

	oldEntry := s.fetch(gatewayID)
	if oldEntry != nil {
		close(oldEntry.channel)
	}
	s.store[gatewayID] = entry

	s.mu.Unlock()
}

func (s *gatewayStore) fetch(gatewayID ttnpb.GatewayIdentifiers) *gatewayStoreEntry {
	return s.store[gatewayID]
}

func (s *gatewayStore) Fetch(gatewayID ttnpb.GatewayIdentifiers) *gatewayStoreEntry {
	s.mu.Lock()
	entry := s.fetch(gatewayID)
	s.mu.Unlock()

	return entry
}

func (s *gatewayStore) Remove(gatewayID ttnpb.GatewayIdentifiers) {
	s.mu.Lock()

	entry := s.fetch(gatewayID)
	if entry != nil {
		close(entry.channel)
	}
	delete(s.store, gatewayID)

	s.mu.Unlock()
}

// Subscription is implemented by ttnpb.GtwGs_LinkServer.
//
// Using this interface and not ttnpb.GtwGs_LinkServer allows for better testing.
type Subscription interface {
	Send(*ttnpb.GatewayDown) error
	Recv() (*ttnpb.GatewayUp, error)
	Context() context.Context
}

// Pool is a connection pool for every component that receives linking connections from gateways. At the time, this only means the gateway server. It abstracts:
//
// - Keeping track of gateway connections
//
// - Scheduling of downlinks
type Pool interface {
	Subscribe(gatewayInfo ttnpb.GatewayIdentifiers, link Subscription, fp ttnpb.FrequencyPlan) (chan *ttnpb.GatewayUp, error)
	Send(gatewayInfo ttnpb.GatewayIdentifiers, downstream *ttnpb.GatewayDown) error

	GetGatewayObservations(gatewayInfo *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error)
}

type pool struct {
	store *gatewayStore

	sendTimeout time.Duration
	logger      log.Interface
}

// NewPool returns a new empty gateway pool.
func NewPool(logger log.Interface, sendTimeout time.Duration) Pool {
	return &pool{
		store: &gatewayStore{
			store: map[ttnpb.GatewayIdentifiers]*gatewayStoreEntry{},
			mu:    sync.Mutex{},
		},
		sendTimeout: sendTimeout,

		logger: logger,
	}
}

func (p *pool) GetGatewayObservations(gatewayInfo *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error) {
	gateway := p.store.Fetch(*gatewayInfo)
	if gateway == nil {
		return nil, ErrGatewayNotConnected.New(errors.Attributes{"gateway_id": gatewayInfo.GatewayID})
	}

	gateway.observationsLock.RLock()
	obs := gateway.observations
	gateway.observationsLock.RUnlock()
	return &obs, nil
}
