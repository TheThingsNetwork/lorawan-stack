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

package packages_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	registeredApplicationID   = ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}
	registeredApplicationUID  = unique.ID(test.Context(), registeredApplicationID)
	registeredApplicationKey  = "test-key"
	unregisteredApplicationID = ttnpb.ApplicationIdentifiers{ApplicationID: "invalid-app"}
	registeredDeviceID        = ttnpb.EndDeviceIdentifiers{ApplicationIdentifiers: registeredApplicationID, DeviceID: "test-dev"}
	unregisteredDeviceID      = ttnpb.EndDeviceIdentifiers{ApplicationIdentifiers: unregisteredApplicationID, DeviceID: "invalid-dev"}
	registeredAssociationID   = ttnpb.ApplicationPackageAssociationIdentifiers{EndDeviceIdentifiers: registeredDeviceID, FPort: 123}
	unregisteredAssociationID = ttnpb.ApplicationPackageAssociationIdentifiers{EndDeviceIdentifiers: unregisteredDeviceID, FPort: 123}
	registeredApplicationUp1  = ttnpb.ApplicationUp{
		EndDeviceIdentifiers: registeredDeviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				FPort: 123,
			},
		},
	}
	registeredApplicationUp2 = ttnpb.ApplicationUp{
		EndDeviceIdentifiers: registeredDeviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				FPort: 124,
			},
		},
	}
	unregisteredApplicationUp = ttnpb.ApplicationUp{
		EndDeviceIdentifiers: unregisteredDeviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				FPort: 123,
			},
		},
	}

	timeout = (1 << 6) * test.Delay
)

func TestAuthentication(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	a := assertions.New(t)

	is, isAddr := startMockIS(ctx)
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	apRegistry := &redis.ApplicationPackagesRegistry{Redis: redisClient}
	srv, err := packages.New(ctx, as, apRegistry)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	c.RegisterGRPC(srv)
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewApplicationPackageRegistryClient(c.LoopbackConn())

	for _, tc := range []struct {
		ID  ttnpb.EndDeviceIdentifiers
		Key string
		OK  bool
	}{
		{
			ID:  registeredDeviceID,
			Key: registeredApplicationKey,
			OK:  true,
		},
		{
			ID:  registeredDeviceID,
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  unregisteredDeviceID,
			Key: "invalid-key",
			OK:  false,
		},
	} {
		t.Run(fmt.Sprintf("%v:%v", tc.ID.ApplicationID, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			creds := grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     tc.Key,
				AllowInsecure: true,
			})

			_, err := client.List(ctx, &tc.ID, creds)
			if tc.OK && err != nil && !a.So(errors.IsCanceled(err), should.BeTrue) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tc.OK && !a.So(errors.IsCanceled(err), should.BeFalse) {
				t.FailNow()
			}
		})
	}
}

func TestAssociations(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	a := assertions.New(t)

	is, isAddr := startMockIS(ctx)
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	apRegistry := &redis.ApplicationPackagesRegistry{Redis: redisClient}

	handleUpCh := make(chan *handleUpRequest, 4)
	applicationPackageFactory = packages.CreateApplicationPackage(
		func(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
			a.So(server, should.Equal, as)
			a.So(registry, should.Equal, apRegistry)
			return createMockPackageHandler(handleUpCh)
		},
	)

	srv, err := packages.New(ctx, as, apRegistry)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	c.RegisterGRPC(srv)
	componenttest.StartComponent(t, c)
	defer c.Close()
	sub := srv.NewSubscription()

	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewApplicationPackageRegistryClient(c.LoopbackConn())

	// Check that the test package is registered.
	t.Run("AvailablePackages", func(t *testing.T) {
		a := assertions.New(t)
		res, err := client.List(ctx, &registeredDeviceID, creds)
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res.Packages, should.Resemble, []*ttnpb.ApplicationPackage{
			{
				Name:         "test-package",
				DefaultFPort: 123,
			},
		})
	})

	// Check that no associations exist initially.
	// TODO: changes this after "Multi-package access protocol" is added.
	// https://github.com/TheThingsNetwork/lorawan-stack/issues/1328
	t.Run("AssociationsNotFound", func(t *testing.T) {
		a := assertions.New(t)
		_, err = client.GetAssociation(ctx, &ttnpb.GetApplicationPackageAssociationRequest{
			ApplicationPackageAssociationIdentifiers: registeredAssociationID,
		}, creds)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		res, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
			EndDeviceIdentifiers: registeredDeviceID,
		}, creds)
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res.Associations, should.HaveLength, 0)
	})

	association := ttnpb.ApplicationPackageAssociation{
		ApplicationPackageAssociationIdentifiers: registeredAssociationID,
		PackageName:                              "test-package",
		Data: &types.Struct{
			Fields: map[string]*types.Value{
				"state": {
					Kind: &types.Value_NumberValue{
						NumberValue: 0,
					},
				},
			},
		},
	}

	// Create the association with the test package.
	t.Run("Create", func(t *testing.T) {
		a := assertions.New(t)
		res, err := client.SetAssociation(ctx, &ttnpb.SetApplicationPackageAssociationRequest{
			ApplicationPackageAssociation: association,
			FieldMask: types.FieldMask{
				Paths: []string{
					"package_name",
					"data",
				},
			},
		}, creds)
		a.So(err, should.BeNil)
		association.CreatedAt = res.CreatedAt
		association.UpdatedAt = res.UpdatedAt
		a.So(res, should.Resemble, &association)
	})

	// Check that the association is available.
	t.Run("AssociationsFound", func(t *testing.T) {
		a := assertions.New(t)
		res1, err := client.GetAssociation(ctx, &ttnpb.GetApplicationPackageAssociationRequest{
			ApplicationPackageAssociationIdentifiers: registeredAssociationID,
			FieldMask: types.FieldMask{
				Paths: []string{
					"package_name",
					"data",
				},
			},
		}, creds)
		a.So(err, should.BeNil)
		a.So(res1, should.Resemble, &association)

		res2, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
			EndDeviceIdentifiers: registeredDeviceID,
			FieldMask: types.FieldMask{
				Paths: []string{
					"package_name",
					"data",
				},
			},
		}, creds)
		a.So(err, should.BeNil)
		a.So(res2, should.NotBeNil)
		a.So(res2.Associations, should.HaveLength, 1)
		a.So(res2.Associations[0], should.Resemble, &association)
	})

	// Send traffic and expect to arrive in the correct handler.
	t.Run("Traffic1", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			up    *ttnpb.ApplicationUp
			valid bool
		}{
			{
				name:  "Valid",
				up:    &registeredApplicationUp1,
				valid: true,
			},
			{
				name:  "Wrong FPort",
				up:    &registeredApplicationUp2,
				valid: false,
			},
			{
				name:  "Wrong application",
				up:    &unregisteredApplicationUp,
				valid: false,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				a := assertions.New(t)

				err := sub.SendUp(ctx, tc.up)
				a.So(err, should.BeNil)

				select {
				case up := <-handleUpCh:
					{
						if !tc.valid {
							t.Fatal("unexpected uplink")
						} else {
							a.So(up.ctx, should.NotBeNil)
							a.So(up.assoc, should.Resemble, &association)
						}
					}
				case <-time.After(2 * timeout):
					{
						if tc.valid {
							t.Fatal("expected uplink timeout")
						}
					}
				}
			})
		}
	})

	// Check that after the deletion no traces are left and traffic is no longer handled.
	t.Run("Deletion", func(t *testing.T) {
		a := assertions.New(t)

		_, err := client.DeleteAssociation(ctx, &registeredAssociationID, creds)
		a.So(err, should.BeNil)

		_, err = client.GetAssociation(ctx, &ttnpb.GetApplicationPackageAssociationRequest{
			ApplicationPackageAssociationIdentifiers: registeredAssociationID,
		}, creds)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		res, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
			EndDeviceIdentifiers: registeredDeviceID,
		}, creds)
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res.Associations, should.BeEmpty)

		err = sub.SendUp(ctx, &registeredApplicationUp1)
		a.So(err, should.BeNil)
		select {
		case <-handleUpCh:
			t.Fatal("unexpected uplink arrived")
		case <-time.After(2 * timeout):
			break
		}
	})
}

var applicationPackageFactory = func(io.Server, packages.Registry) packages.ApplicationPackageHandler {
	return &mockPackageHandler{}
}

func init() {
	p := ttnpb.ApplicationPackage{
		Name:         "test-package",
		DefaultFPort: 123,
	}
	packages.RegisterPackage(p, packages.CreateApplicationPackage(
		func(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
			return applicationPackageFactory(server, registry)
		},
	))
}
