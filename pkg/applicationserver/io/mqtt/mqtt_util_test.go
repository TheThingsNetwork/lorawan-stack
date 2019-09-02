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

package mqtt_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/metadata"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

type mockIS struct {
	ttnpb.ApplicationRegistryServer
	ttnpb.ApplicationAccessServer
	applications     map[string]*ttnpb.Application
	applicationAuths map[string][]string
}

func startMockIS(ctx context.Context) (*mockIS, string) {
	is := &mockIS{
		applications:     make(map[string]*ttnpb.Application),
		applicationAuths: make(map[string][]string),
	}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterApplicationRegistryServer(srv.Server, is)
	ttnpb.RegisterApplicationAccessServer(srv.Server, is)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return is, lis.Addr().String()
}

func (is *mockIS) add(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) {
	uid := unique.ID(ctx, ids)
	is.applications[uid] = &ttnpb.Application{
		ApplicationIdentifiers: ids,
	}
	if key != "" {
		is.applicationAuths[uid] = []string{fmt.Sprintf("Bearer %v", key)}
	}
}

var errNotFound = errors.DefineNotFound("not_found", "not found")

func (is *mockIS) Get(ctx context.Context, req *ttnpb.GetApplicationRequest) (*ttnpb.Application, error) {
	uid := unique.ID(ctx, req.ApplicationIdentifiers)
	app, ok := is.applications[uid]
	if !ok {
		return nil, errNotFound
	}
	return app, nil
}

func (is *mockIS) ListRights(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (res *ttnpb.Rights, err error) {
	res = &ttnpb.Rights{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	authorization, ok := md["authorization"]
	if !ok || len(authorization) == 0 {
		return
	}
	auths, ok := is.applicationAuths[unique.ID(ctx, *ids)]
	if !ok {
		return
	}
	for _, auth := range auths {
		if auth == authorization[0] {
			res.Rights = append(res.Rights,
				ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
				ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
			)
		}
	}
	return
}
