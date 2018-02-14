// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
	store map[ttnpb.GatewayIdentifier]*gatewayStoreEntry

	mu sync.Mutex
}

func (s *gatewayStore) Store(gatewayID ttnpb.GatewayIdentifier, entry *gatewayStoreEntry) {
	s.mu.Lock()

	oldEntry := s.fetch(gatewayID)
	if oldEntry != nil {
		close(oldEntry.channel)
	}
	s.store[gatewayID] = entry

	s.mu.Unlock()
}

func (s *gatewayStore) fetch(gatewayID ttnpb.GatewayIdentifier) *gatewayStoreEntry {
	return s.store[gatewayID]
}

func (s *gatewayStore) Fetch(gatewayID ttnpb.GatewayIdentifier) *gatewayStoreEntry {
	s.mu.Lock()
	entry := s.fetch(gatewayID)
	s.mu.Unlock()

	return entry
}

func (s *gatewayStore) Remove(gatewayID ttnpb.GatewayIdentifier) {
	s.mu.Lock()

	entry := s.fetch(gatewayID)
	if entry != nil {
		close(entry.channel)
	}
	delete(s.store, gatewayID)

	s.mu.Unlock()
}

// PoolSubscription is implemented by ttnpb.GtwGs_LinkServer.
//
// Using this interface and not ttnpb.GtwGs_LinkServer allows for better testing.
type PoolSubscription interface {
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
	Subscribe(gatewayInfo ttnpb.GatewayIdentifier, link PoolSubscription, fp ttnpb.FrequencyPlan) (chan *ttnpb.GatewayUp, error)
	Send(gatewayInfo ttnpb.GatewayIdentifier, downstream *ttnpb.GatewayDown) error

	GetGatewayObservations(gatewayInfo *ttnpb.GatewayIdentifier) (*ttnpb.GatewayObservations, error)
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
			store: map[ttnpb.GatewayIdentifier]*gatewayStoreEntry{},
			mu:    sync.Mutex{},
		},
		sendTimeout: sendTimeout,

		logger: logger,
	}
}

func (p *pool) GetGatewayObservations(gatewayInfo *ttnpb.GatewayIdentifier) (*ttnpb.GatewayObservations, error) {
	gateway := p.store.Fetch(*gatewayInfo)
	if gateway == nil {
		return nil, ErrGatewayNotConnected.New(errors.Attributes{"gateway_id": gatewayInfo.GatewayID})
	}

	gateway.observationsLock.RLock()
	obs := gateway.observations
	gateway.observationsLock.RUnlock()
	return &obs, nil
}
