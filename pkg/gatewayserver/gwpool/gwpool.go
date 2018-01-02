// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package gwpool abstracts the polling and sending procedures between gateways and the gateway server.
package gwpool

import (
	"context"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type gatewayStore struct {
	store map[ttnpb.GatewayIdentifier]chan *ttnpb.GatewayDown

	mu sync.Mutex
}

func (s *gatewayStore) Store(gatewayID ttnpb.GatewayIdentifier, outgoingChannel chan *ttnpb.GatewayDown) {
	s.mu.Lock()

	res, err := s.fetch(gatewayID)
	if err == nil {
		close(res)
	}
	s.store[gatewayID] = outgoingChannel

	s.mu.Unlock()
}

func (s *gatewayStore) fetch(gatewayID ttnpb.GatewayIdentifier) (chan *ttnpb.GatewayDown, error) {
	outgoingChannel, ok := s.store[gatewayID]
	if !ok {
		return nil, errors.New("Gateway not found")
	}

	return outgoingChannel, nil
}

func (s *gatewayStore) Fetch(gatewayID ttnpb.GatewayIdentifier) (chan *ttnpb.GatewayDown, error) {
	s.mu.Lock()
	res, err := s.fetch(gatewayID)
	s.mu.Unlock()

	return res, err
}

func (s *gatewayStore) Remove(gatewayID ttnpb.GatewayIdentifier) {
	s.mu.Lock()

	res, err := s.fetch(gatewayID)
	if err == nil {
		close(res)
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
	Subscribe(gatewayInfo ttnpb.GatewayIdentifier, link PoolSubscription) chan *ttnpb.GatewayUp
	Send(gatewayInfo ttnpb.GatewayIdentifier, downstream *ttnpb.GatewayDown) error
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
			store: map[ttnpb.GatewayIdentifier]chan *ttnpb.GatewayDown{},
			mu:    sync.Mutex{},
		},
		sendTimeout: sendTimeout,

		logger: logger,
	}
}
