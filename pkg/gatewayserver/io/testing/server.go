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
	stdio "io"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

type server struct {
	localStore     test.FrequencyPlansStore
	store          *frequencyplans.Store
	connectionsCh  chan *io.Connection
	downlinkClaims sync.Map
}

// Server represents a testing io.Server.
type Server interface {
	stdio.Closer
	io.Server

	Connections() <-chan *io.Connection
}

// MustNewServer instantiates a new Server and panics on failure.
func MustNewServer() Server {
	localStore, err := test.NewFrequencyPlansStore()
	if err != nil {
		panic(err)
	}
	return &server{
		localStore:    localStore,
		store:         frequencyplans.NewStore(fetch.FromFilesystem(localStore.Directory())),
		connectionsCh: make(chan *io.Connection, 10),
	}
}

// Close cleans up temporary files.
func (s *server) Close() error {
	return s.localStore.Destroy()
}

// Connect implements io.Server.
func (s *server) Connect(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*io.Connection, error) {
	if err := rights.RequireGateway(ctx, ids, ttnpb.RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}
	gtw := &ttnpb.Gateway{GatewayIdentifiers: ids}
	fp, err := s.store.GetByID("EU_863_870")
	if err != nil {
		return nil, err
	}
	scheduler, err := scheduling.FrequencyPlanScheduler(ctx, fp)
	if err != nil {
		return nil, err
	}
	conn := io.NewConnection(ctx, gtw, scheduler)
	select {
	case s.connectionsCh <- conn:
	default:
	}
	return conn, nil
}

// GetFrequencyPlan implements io.Server.
func (s *server) GetFrequencyPlan(ctx context.Context, id string) (*ttnpb.FrequencyPlan, error) {
	fp, err := s.store.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &fp, nil
}

// ClaimDownlink implements io.Server.
func (s *server) ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	s.downlinkClaims.Store(unique.ID(ctx, ids), true)
	return nil
}

// UnclaimDownlink implements io.Server.
func (s *server) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	s.downlinkClaims.Delete(unique.ID(ctx, ids))
	return nil
}

// GetFrequencyPlan implements io.Server.
func (s *server) HasDownlinkClaim(ctx context.Context, ids ttnpb.GatewayIdentifiers) (bool, error) {
	_, ok := s.downlinkClaims.Load(unique.ID(ctx, ids))
	return ok, nil
}

func (s *server) Connections() <-chan *io.Connection {
	return s.connectionsCh
}
