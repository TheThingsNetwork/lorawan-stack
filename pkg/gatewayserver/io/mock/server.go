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
	"fmt"
	"strings"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

type server struct {
	*component.Component
	store          *frequencyplans.Store
	identityStore  *mockis.MockDefinition
	connections    map[string]*io.Connection
	connectionsCh  chan *io.Connection
	downlinkClaims sync.Map
}

// Server represents a testing io.Server.
type Server interface {
	io.Server

	HasDownlinkClaim(context.Context, ttnpb.GatewayIdentifiers) bool
	RegisterGateway(ctx context.Context, ids ttnpb.GatewayIdentifiers, gateway *ttnpb.Gateway)
	GetConnection(ctx context.Context, ids ttnpb.GatewayIdentifiers) *io.Connection
	Connections() <-chan *io.Connection
}

// NewServer instantiates a new Server.
func NewServer(c *component.Component, is *mockis.MockDefinition) Server {
	return &server{
		Component:     c,
		store:         frequencyplans.NewStore(test.FrequencyPlansFetcher),
		identityStore: is,
		connections:   make(map[string]*io.Connection),
		connectionsCh: make(chan *io.Connection, 10),
	}
}

// FillContext implements io.Server.
func (s *server) FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error) {
	ctx = s.FillContext(ctx)
	if ids.IsZero() {
		return nil, ttnpb.GatewayIdentifiers{}, errors.New("the identifiers are zero")
	}
	if ids.GatewayId != "" {
		return ctx, ids, nil
	}
	ids.GatewayId = fmt.Sprintf("eui-%v", strings.ToLower(ids.Eui.String()))
	return ctx, ids, nil
}

// Connect implements io.Server.
func (s *server) Connect(ctx context.Context, frontend io.Frontend, ids ttnpb.GatewayIdentifiers) (*io.Connection, error) {
	if err := rights.RequireGateway(ctx, ids, ttnpb.Right_RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}
	gtw, err := s.identityStore.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: &ids})
	if err != nil {
		gtw = &ttnpb.Gateway{
			Ids:             &ids,
			FrequencyPlanId: test.EUFrequencyPlanID,
		}
	}
	fps, err := s.FrequencyPlansStore(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := io.NewConnection(ctx, frontend, gtw, fps, true, nil)
	if err != nil {
		return nil, err
	}

	s.connections[unique.ID(ctx, ids)] = conn
	select {
	case s.connectionsCh <- conn:
	default:
	}
	return conn, nil
}

// GetFrequencyPlans implements io.Server.
func (s *server) GetFrequencyPlans(ctx context.Context, ids ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error) {
	var fpID string
	if gtw, err := s.identityStore.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{GatewayIds: &ids}); err == nil {
		fpID = gtw.FrequencyPlanId
	} else {
		fpID = test.EUFrequencyPlanID
	}
	fp, err := s.store.GetByID(fpID)
	if err != nil {
		return nil, err
	}
	fps := make(map[string]*frequencyplans.FrequencyPlan)
	fps[fpID] = fp
	return fps, nil
}

// ClaimDownlink implements io.Server.
func (s *server) ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	s.downlinkClaims.Store(unique.ID(ctx, ids), true)
	return nil
}

func (s *server) ValidateGatewayID(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return ids.ValidateContext(ctx)
}

// UnclaimDownlink implements io.Server.
func (s *server) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	s.downlinkClaims.Delete(unique.ID(ctx, ids))
	return nil
}

// StartTask implements io.Server.
func (s *server) StartTask(cfg *task.Config) {
	task.DefaultStartTask(cfg)
}

func (s *server) HasDownlinkClaim(ctx context.Context, ids ttnpb.GatewayIdentifiers) bool {
	_, ok := s.downlinkClaims.Load(unique.ID(ctx, ids))
	return ok
}

func (s *server) RegisterGateway(ctx context.Context, ids ttnpb.GatewayIdentifiers, gateway *ttnpb.Gateway) {
	if len(gateway.FrequencyPlanIds) > 0 {
		gateway.FrequencyPlanId = gateway.FrequencyPlanIds[0]
	}

	gtwRights := []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO, ttnpb.Right_RIGHT_GATEWAY_LINK}
	s.identityStore.GatewayRegistry().Add(ctx, ids, "default-key", gateway, gtwRights...)
}

func (s *server) GetConnection(ctx context.Context, ids ttnpb.GatewayIdentifiers) *io.Connection {
	return s.connections[unique.ID(ctx, ids)]
}

func (s *server) Connections() <-chan *io.Connection {
	return s.connectionsCh
}
