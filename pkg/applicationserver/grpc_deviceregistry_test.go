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

package applicationserver_test

import (
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestDeviceRegistry(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	is, isAddr := startMockIS(ctx)

	// Register the application in the Entity Registry.
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: redisClient}

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9184",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	})
	config := &applicationserver.Config{
		LinkMode: "explicit",
		Devices:  deviceRegistry,
	}
	_, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	test.Must(nil, c.Start())
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	conn, err := grpc.Dial(":9184", grpc.WithInsecure(), grpc.WithBlock())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer conn.Close()
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Key",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})
	client := ttnpb.NewAsEndDeviceRegistryClient(conn)

	// Unauthorized: no credentials.
	{
		_, err := client.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
		})
		a.So(errors.IsUnauthenticated(err), should.BeTrue)

		_, err = client.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: ttnpb.EndDevice{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FrequencyPlanID:      "EU_863_870",
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"frequency_plan_id"},
			},
		})
		a.So(errors.IsUnauthenticated(err), should.BeTrue)

		_, err = client.Delete(ctx, &registeredDevice.EndDeviceIdentifiers)
		a.So(errors.IsUnauthenticated(err), should.BeTrue)
	}

	// Unauthorized: wrong application.
	{
		otherID := ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "other-app",
			},
			DeviceID: "other-device",
		}

		_, err := client.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: otherID,
		}, creds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = client.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: ttnpb.EndDevice{
				EndDeviceIdentifiers: otherID,
				FrequencyPlanID:      "EU_863_870",
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"frequency_plan_id"},
			},
		}, creds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = client.Delete(ctx, &otherID, creds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	}

	// Happy flow: create, update and delete.
	{
		// Assert the device doesn't exist yet.
		_, err := client.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids", "version_ids", "formatters"},
			},
		}, creds)
		a.So(errors.IsNotFound(err), should.BeTrue)

		// Create and assert resemblance.
		_, err = client.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: *registeredDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids", "version_ids", "formatters"},
			},
		}, creds)
		a.So(err, should.BeNil)
		dev, err := client.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids", "version_ids", "formatters"},
			},
		}, creds)
		a.So(err, should.BeNil)
		registeredDevice.CreatedAt = dev.CreatedAt
		registeredDevice.UpdatedAt = dev.UpdatedAt
		a.So(dev, should.HaveEmptyDiff, registeredDevice)

		// Update and assert new value.
		registeredDevice.Formatters.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_NONE
		_, err = client.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: *registeredDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"formatters"},
			},
		}, creds)
		a.So(err, should.BeNil)
		dev, err = client.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids", "version_ids", "formatters"},
			},
		}, creds)
		a.So(err, should.BeNil)
		registeredDevice.CreatedAt = dev.CreatedAt
		registeredDevice.UpdatedAt = dev.UpdatedAt
		a.So(dev, should.HaveEmptyDiff, registeredDevice)

		// Delete and assert it's gone.
		_, err = client.Delete(ctx, &registeredDevice.EndDeviceIdentifiers, creds)
		a.So(err, should.BeNil)
		_, err = client.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids", "version_ids", "formatters"},
			},
		}, creds)
		a.So(errors.IsNotFound(err), should.BeTrue)
	}
}
