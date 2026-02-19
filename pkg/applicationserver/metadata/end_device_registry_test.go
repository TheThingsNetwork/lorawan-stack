// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
	"maps"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
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

	originalAttributes = map[string]string{
		"attr1": "val1",
		"attr2": "val2",
	}

	attributesPatch = map[string]string{
		"attr3": "val3",
		"attr4": "val4",
	}
)

func TestClusterEndDeviceRegistry(t *testing.T) { // nolint:gocyclo
	registeredEndDeviceIDs := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "foo",
		},
		DeviceId: "bar",
	}

	t.Parallel()

	a, ctx := test.New(t)
	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()

	registeredEndDevice := ttnpb.EndDevice{
		Ids:        registeredEndDeviceIDs,
		Locations:  originalLocations,
		Attributes: originalAttributes,
	}
	is.EndDeviceRegistry().Add(ctx, &registeredEndDevice)

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

	registry := metadata.NewClusterEndDeviceRegistry(c, 10*time.Second)

	_, err := registry.Get(ctx, registeredEndDeviceIDs, []string{
		"network_server_address", "application_server_address", "join_server_address",
	})
	a.So(errors.IsInvalidArgument(err), should.BeTrue)

	dev, err := registry.Get(ctx, registeredEndDeviceIDs, []string{"attributes", "locations"})
	if a.So(err, should.BeNil) {
		a.So(dev, should.NotBeNil)

		a.So(dev.Locations, should.NotBeNil)
		a.So(len(dev.Locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range dev.Locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(dev.Locations[k], should.Resemble, v)
		}

		a.So(dev.Attributes, should.NotBeNil)
		a.So(len(dev.Attributes), should.Equal, len(registeredEndDevice.Attributes))
		for k, v := range dev.Attributes {
			a.So(registeredEndDevice.Attributes[k], should.Equal, v)
		}
		for k, v := range originalAttributes {
			a.So(dev.Attributes[k], should.Equal, v)
		}
	}

	_, err = registry.Set(ctx, registeredEndDeviceIDs, []string{
		"network_server_address", "application_server_address", "join_server_address",
	}, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return nil, nil, nil // nolint: nilnil
	})
	a.So(errors.IsInvalidArgument(err), should.BeTrue)

	// Update location and attributes.
	dev, err = registry.Set(ctx, registeredEndDeviceIDs, []string{"locations", "attributes"},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				return nil, nil, errors.New("not found")
			}

			if len(stored.Locations) == 0 {
				stored.Locations = make(map[string]*ttnpb.Location, len(locationsPatch))
			}

			maps.Copy(stored.Locations, locationsPatch)

			if len(stored.Attributes) == 0 {
				stored.Attributes = make(map[string]string, len(attributesPatch))
			}

			maps.Copy(stored.Attributes, attributesPatch)

			return stored, []string{"locations", "attributes"}, nil
		},
	)
	if a.So(err, should.BeNil) {
		a.So(dev, should.NotBeNil)

		a.So(dev.Locations, should.NotBeNil)
		a.So(len(dev.Locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range dev.Locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(dev.Locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(dev.Locations[k], should.Resemble, v)
		}

		a.So(dev.Attributes, should.NotBeNil)
		a.So(len(dev.Attributes), should.Equal, len(registeredEndDevice.Attributes))
		for k, v := range dev.Attributes {
			a.So(registeredEndDevice.Attributes[k], should.Equal, v)
		}
		for k, v := range originalAttributes {
			a.So(dev.Attributes[k], should.Equal, v)
		}
		for k, v := range attributesPatch {
			a.So(dev.Attributes[k], should.Equal, v)
		}
	}

	dev, err = registry.Get(ctx, registeredEndDeviceIDs, []string{"attributes", "locations"})
	if a.So(err, should.BeNil) {
		a.So(dev, should.NotBeNil)

		a.So(dev.Locations, should.NotBeNil)
		a.So(len(dev.Locations), should.Equal, len(registeredEndDevice.Locations))
		for k, v := range dev.Locations {
			a.So(registeredEndDevice.Locations[k], should.Resemble, v)
		}
		for k, v := range originalLocations {
			a.So(dev.Locations[k], should.Resemble, v)
		}
		for k, v := range locationsPatch {
			a.So(dev.Locations[k], should.Resemble, v)
		}

		a.So(dev.Attributes, should.NotBeNil)
		a.So(len(dev.Attributes), should.Equal, len(registeredEndDevice.Attributes))
		for k, v := range dev.Attributes {
			a.So(registeredEndDevice.Attributes[k], should.Equal, v)
		}
		for k, v := range originalAttributes {
			a.So(dev.Attributes[k], should.Equal, v)
		}
		for k, v := range attributesPatch {
			a.So(dev.Attributes[k], should.Equal, v)
		}
	}
}
