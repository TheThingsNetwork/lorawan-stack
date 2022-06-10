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

package packetbrokeragent_test

import (
	"context"
	"net"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
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

func mustServePBDataPlane(ctx context.Context, tb testing.TB) (*mock.PBDataPlane, net.Addr) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	dp := mock.NewPBDataPlane(tb)
	go dp.Serve(lis)
	go func() {
		<-ctx.Done()
		dp.GracefulStop()
	}()
	return dp, lis.Addr()
}

func mustServePBMapper(ctx context.Context, tb testing.TB) (*mock.PBMapper, net.Addr) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	mp := mock.NewPBMapper(tb)
	go mp.Serve(lis)
	go func() {
		<-ctx.Done()
		mp.GracefulStop()
	}()
	return mp, lis.Addr()
}

func eui64Ptr(v types.EUI64) *types.EUI64 {
	return &v
}
