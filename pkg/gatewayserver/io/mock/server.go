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

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

type server struct {
	*component.Component
	store          *frequencyplans.Store
	gateways       map[string]*ttnpb.Gateway
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
func NewServer(c *component.Component) Server {
	return &server{
		Component:     c,
		store:         frequencyplans.NewStore(test.FrequencyPlansFetcher),
		gateways:      make(map[string]*ttnpb.Gateway),
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
	if ids.GatewayID != "" {
		return ctx, ids, nil
	}
	ids.GatewayID = fmt.Sprintf("eui-%v", strings.ToLower(ids.EUI.String()))
	return ctx, ids, nil
}

// Connect implements io.Server.
func (s *server) Connect(ctx context.Context, frontend io.Frontend, ids ttnpb.GatewayIdentifiers) (*io.Connection, error) {
	if err := rights.RequireGateway(ctx, ids, ttnpb.RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}
	gtw, ok := s.gateways[unique.ID(ctx, ids)]
	if !ok {
		gtw = &ttnpb.Gateway{
			GatewayIdentifiers: ids,
			FrequencyPlanID:    test.EUFrequencyPlanID,
		}
	}
	fp, err := s.store.GetByID(gtw.FrequencyPlanID)
	if err != nil {
		return nil, err
	}
	fps, err := s.GetFrequencyPlans(ctx, ids)
	if err != nil {
		return nil, err
	}
	conn, err := io.NewConnection(ctx, frontend, gtw, fp.BandID, fps, true, nil)
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
	if gtw, ok := s.gateways[unique.ID(ctx, ids)]; ok {
		fpID = gtw.FrequencyPlanID
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

// UnclaimDownlink implements io.Server.
func (s *server) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	s.downlinkClaims.Delete(unique.ID(ctx, ids))
	return nil
}

func (s *server) HasDownlinkClaim(ctx context.Context, ids ttnpb.GatewayIdentifiers) bool {
	_, ok := s.downlinkClaims.Load(unique.ID(ctx, ids))
	return ok
}

func (s *server) RegisterGateway(ctx context.Context, ids ttnpb.GatewayIdentifiers, gateway *ttnpb.Gateway) {
	uid := unique.ID(ctx, ids)
	s.gateways[uid] = gateway
}

func (s *server) GetConnection(ctx context.Context, ids ttnpb.GatewayIdentifiers) *io.Connection {
	return s.connections[unique.ID(ctx, ids)]
}

func (s *server) Connections() <-chan *io.Connection {
	return s.connectionsCh
}
