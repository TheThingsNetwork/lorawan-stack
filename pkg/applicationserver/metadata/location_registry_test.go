// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package metadata_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockISEndDeviceRegistry struct {
	ttnpb.EndDeviceRegistryServer

	endDevicesMu sync.RWMutex
	endDevices   map[string]*ttnpb.EndDevice
}

func (m *mockISEndDeviceRegistry) add(ctx context.Context, dev *ttnpb.EndDevice) {
	m.endDevicesMu.Lock()
	defer m.endDevicesMu.Unlock()
	m.endDevices[unique.ID(ctx, dev.Ids)] = dev
}

func (m *mockISEndDeviceRegistry) get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevice, bool) {
	m.endDevicesMu.RLock()
	defer m.endDevicesMu.RUnlock()
	dev, ok := m.endDevices[unique.ID(ctx, ids)]
	return dev, ok
}

var errNotFound = errors.DefineNotFound("not_found", "not found")

func (m *mockISEndDeviceRegistry) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	m.endDevicesMu.RLock()
	defer m.endDevicesMu.RUnlock()
	if dev, ok := m.endDevices[unique.ID(ctx, in.EndDeviceIds)]; ok {
		return dev, nil
	}
	return nil, errNotFound.New()
}

func (m *mockISEndDeviceRegistry) Update(ctx context.Context, in *ttnpb.UpdateEndDeviceRequest) (*ttnpb.EndDevice, error) {
	m.endDevicesMu.Lock()
	defer m.endDevicesMu.Unlock()
	dev, ok := m.endDevices[unique.ID(ctx, in.EndDevice.Ids)]
	if !ok {
		return nil, errNotFound.New()
	}
	if err := dev.SetFields(in.EndDevice, in.GetFieldMask().GetPaths()...); err != nil {
		return nil, err
	}
	m.endDevices[unique.ID(ctx, in.EndDevice.Ids)] = dev
	return dev, nil
}

type mockIS struct {
	ttnpb.ApplicationRegistryServer
	ttnpb.ApplicationAccessServer

	endDeviceRegistry *mockISEndDeviceRegistry
}

func startMockIS(ctx context.Context) (*mockIS, string, func()) {
	is := &mockIS{
		endDeviceRegistry: &mockISEndDeviceRegistry{
			endDevices: make(map[string]*ttnpb.EndDevice),
		},
	}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterEndDeviceRegistryServer(srv.Server, is.endDeviceRegistry)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return is, lis.Addr().String(), func() {
		lis.Close()
		srv.GracefulStop()
	}
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

var (
	registeredEndDeviceIDs = &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "foo",
		},
		DeviceId: "bar",
	}
	originalLocations = map[string]*ttnpb.Location{
		"baz": {
			Altitude: 12,
			Latitude: 23,
		},
	}
	locationsPatch = map[string]*ttnpb.Location{
		"bzz": {
			Altitude: 23,
			Latitude: 34,
		},
	}
	Timeout = (1 << 7) * test.Delay
)

func TestClusterEndDeviceLocationRegistry(t *testing.T) {
	a, ctx := test.New(t)
	is, isAddr, closeIS := startMockIS(ctx)
	defer closeIS()

	registeredEndDevice := ttnpb.EndDevice{
		Ids:       registeredEndDeviceIDs,
		Locations: originalLocations,
	}
	is.endDeviceRegistry.add(ctx, &registeredEndDevice)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	registry := metadata.NewClusterEndDeviceLocationRegistry(c, 10*time.Second)

	locations, err := registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
	}

	locations, err = registry.Merge(ctx, registeredEndDeviceIDs, locationsPatch)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}

	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}
}

func TestCachedEndDeviceLocationRegistry(t *testing.T) {
	a, ctx := test.New(t)
	is, isAddr, closeIS := startMockIS(ctx)
	defer closeIS()

	registeredEndDevice := ttnpb.EndDevice{
		Ids:       registeredEndDeviceIDs,
		Locations: originalLocations,
	}
	is.endDeviceRegistry.add(ctx, &registeredEndDevice)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	registry := metadata.NewClusterEndDeviceLocationRegistry(c, 4*Timeout)
	cl, flush := test.NewRedis(ctx, "metadata_redis_test")
	defer flush()
	cache := &redis.EndDeviceLocationCache{
		Redis: cl,
	}
	registry = metadata.NewCachedEndDeviceLocationRegistry(
		ctx, c, registry, cache, 4*Timeout, 8*Timeout, 16*Timeout,
	)

	locations, err := registry.Get(ctx, registeredEndDeviceIDs)
	a.So(err, should.BeNil)
	a.So(locations, should.HaveLength, 0)

	// Wait for the cache to be populated asynchronously.
	time.Sleep(Timeout)

	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(originalLocations))
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
	}

	locations, err = registry.Merge(ctx, registeredEndDeviceIDs, locationsPatch)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}

	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}

	// Wait for the entry to be evicted.
	time.Sleep(20 * Timeout)

	// There is no cached location anymore, and we have triggered an asynchronous refresh.
	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	a.So(err, should.BeNil)
	a.So(locations, should.HaveLength, 0)

	time.Sleep(Timeout)

	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}

	// Simulate a network partition.
	closeIS()
	time.Sleep(Timeout)

	// Do a read that will trigger an asynchronous cache refresh.
	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}

	// Wait for the partition to be detected asynchronously.
	time.Sleep(Timeout)

	// We now serve stale data.
	locations, err = registry.Get(ctx, registeredEndDeviceIDs)
	if a.So(err, should.BeNil) {
		a.So(locations, should.NotBeNil)
		a.So(len(locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(locations[k], should.Resemble, v)
		}
	}
}
