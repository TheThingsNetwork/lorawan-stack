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

package lbslns

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type mockServer struct {
	ids *ttnpb.GatewayIdentifiers
}

func (srv mockServer) GetBaseConfig(ctx context.Context) config.ServiceBase {
	return config.ServiceBase{}
}

func (srv mockServer) FillGatewayContext(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (context.Context, *ttnpb.GatewayIdentifiers, error) {
	return ctx, srv.ids, nil
}

func (mockServer) AssertGatewayRights(context.Context, *ttnpb.GatewayIdentifiers, ...ttnpb.Right) error {
	return nil
}

func (mockServer) Connect(
	context.Context, io.Frontend, *ttnpb.GatewayIdentifiers, *ttnpb.GatewayRemoteAddress, ...io.ConnectionOption,
) (*io.Connection, error) {
	return nil, nil
}

func (srv mockServer) GetFrequencyPlans(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error) {
	return nil, nil
}

func (srv mockServer) ClaimDownlink(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	return nil
}

func (srv mockServer) UnclaimDownlink(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	return nil
}

func (srv mockServer) FromRequestContext(ctx context.Context) context.Context {
	return ctx
}

func (srv mockServer) RateLimiter() ratelimit.Interface {
	return nil
}

func (srv mockServer) ValidateGatewayID(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	return ids.ValidateContext(ctx)
}

func (srv mockServer) StartTask(cfg *task.Config) {
	task.DefaultStartTask(cfg)
}
