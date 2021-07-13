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

package gatewayserver_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluele/gcache"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockEntityRegistry struct {
	gtw                            *ttnpb.Gateway
	ids                            *ttnpb.GatewayIdentifiers
	GetIdentifiersForEUICalledWith chan *ttnpb.GetGatewayIdentifiersForEUIRequest
	GetCalledWith                  chan *ttnpb.GetGatewayRequest
}

// AssertGatewayRights checks whether the gateway authentication (provied in the context) contains the required rights.
func (is *mockEntityRegistry) AssertGatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error {
	return nil
}

// Get the identifiers of the gateway that has the given EUI registered.
func (is *mockEntityRegistry) GetIdentifiersForEUI(ctx context.Context, in *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	is.GetIdentifiersForEUICalledWith <- in
	return is.ids, nil
}

// Get the gateway with the given identifiers, selecting the fields specified.
func (is *mockEntityRegistry) Get(ctx context.Context, in *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	is.GetCalledWith <- in
	return is.gtw, nil
}

// UpdateAntennas updates the gateway antennas.
func (is *mockEntityRegistry) UpdateAntennas(ctx context.Context, ids ttnpb.GatewayIdentifiers, antennas []ttnpb.GatewayAntenna) error {
	return nil
}

// ValidateGatewayID validates the ID of the gateway.
func (is *mockEntityRegistry) ValidateGatewayID(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return nil
}

func TestCachedEntityRegistry(t *testing.T) {
	er := &mockEntityRegistry{
		gtw:                            &ttnpb.Gateway{Name: "mock-gateway"},
		ids:                            &ttnpb.GatewayIdentifiers{GatewayId: "gateway-id"},
		GetIdentifiersForEUICalledWith: make(chan *ttnpb.GetGatewayIdentifiersForEUIRequest, 1),
		GetCalledWith:                  make(chan *ttnpb.GetGatewayRequest, 1),
	}

	clock := gcache.NewFakeClock()
	config := gatewayserver.EntityRegistryCacheConfig{
		Size:    1000,
		Timeout: time.Minute,
		Clock:   clock,
	}

	is := gatewayserver.EntityRegistryWithCache(er, config)
	var _ gatewayserver.EntityRegistry = is

	t.Run("Get", func(t *testing.T) {
		for _, tc := range []string{"gtw-1", "gtw-2"} {
			t.Run(tc, func(t *testing.T) {
				request := &ttnpb.GetGatewayRequest{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{
						GatewayId: tc,
					},
					FieldMask: &pbtypes.FieldMask{
						Paths: []string{"gateway_id", "description"},
					},
				}

				t.Run("ColdMiss", func(t *testing.T) {
					a := assertions.New(t)
					gtw, err := is.Get(test.Context(), request)
					a.So(gtw, should.Resemble, er.gtw)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected request to the entity registry, but none received")
					case r := <-er.GetCalledWith:
						a.So(r, should.Resemble, request)
					}
				})

				t.Run("Hit", func(t *testing.T) {
					a := assertions.New(t)
					gtw, err := is.Get(test.Context(), request)
					a.So(gtw, should.Resemble, er.gtw)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
					case <-er.GetCalledWith:
						t.Fatal("Received unexpected request to the entity registry")
					}
				})

				t.Run("MissFlags", func(t *testing.T) {
					request2 := deepcopy.Copy(request).(*ttnpb.GetGatewayRequest)
					request2.FieldMask.Paths = []string{"gateway_id", "name", "description"}
					a := assertions.New(t)
					gtw, err := is.Get(test.Context(), request2)
					a.So(gtw, should.Resemble, er.gtw)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected request to the entity registry, but none received")
					case r := <-er.GetCalledWith:
						a.So(r, should.Resemble, request2)
					}
				})

				t.Run("Expire", func(t *testing.T) {
					clock.Advance(config.Timeout + time.Second)
					a := assertions.New(t)
					gtw, err := is.Get(test.Context(), request)
					a.So(gtw, should.Resemble, er.gtw)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected request to the entity registry, but none received")
					case r := <-er.GetCalledWith:
						a.So(r, should.Resemble, request)
					}
				})
			})
		}
	})

	t.Run("GetIdentifiers", func(t *testing.T) {
		for _, tc := range []*ttnpb.GetGatewayIdentifiersForEUIRequest{
			{Eui: types.EUI64{1, 1, 1, 1, 1, 1, 1, 1}},
			{Eui: types.EUI64{1, 1, 1, 1, 1, 1, 1, 2}},
		} {
			t.Run(tc.Eui.String(), func(t *testing.T) {
				request := tc
				t.Run("ColdMiss", func(t *testing.T) {
					a := assertions.New(t)
					ids, err := is.GetIdentifiersForEUI(test.Context(), request)
					a.So(ids, should.Resemble, er.ids)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected request to the entity registry, but none received")
					case r := <-er.GetIdentifiersForEUICalledWith:
						a.So(r, should.Resemble, request)
					}
				})

				t.Run("Hit", func(t *testing.T) {
					a := assertions.New(t)
					ids, err := is.GetIdentifiersForEUI(test.Context(), request)
					a.So(ids, should.Resemble, er.ids)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
					case <-er.GetIdentifiersForEUICalledWith:
						t.Fatal("Received unexpected request to the entity registry")
					}
				})

				t.Run("Expire", func(t *testing.T) {
					clock.Advance(config.Timeout + time.Second)
					a := assertions.New(t)
					ids, err := is.GetIdentifiersForEUI(test.Context(), request)
					a.So(ids, should.Resemble, er.ids)
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected request to the entity registry, but none received")
					case r := <-er.GetIdentifiersForEUICalledWith:
						a.So(r, should.Resemble, request)
					}
				})
			})
		}
	})
}
