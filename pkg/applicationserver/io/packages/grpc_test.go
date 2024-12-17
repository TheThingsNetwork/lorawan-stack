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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	registeredApplicationID   = &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}
	registeredApplicationUID  = unique.ID(test.Context(), registeredApplicationID)
	registeredApplicationKey  = "test-key"
	unregisteredApplicationID = &ttnpb.ApplicationIdentifiers{ApplicationId: "invalid-app"}
	registeredDeviceID        = &ttnpb.EndDeviceIdentifiers{ApplicationIds: registeredApplicationID, DeviceId: "test-dev"}
	unregisteredDeviceID      = &ttnpb.EndDeviceIdentifiers{ApplicationIds: unregisteredApplicationID, DeviceId: "invalid-dev"}
	registeredAssociationID   = &ttnpb.ApplicationPackageAssociationIdentifiers{EndDeviceIds: registeredDeviceID, FPort: 123}
	unregisteredAssociationID = &ttnpb.ApplicationPackageAssociationIdentifiers{EndDeviceIds: unregisteredDeviceID, FPort: 123}
	registeredApplicationUp1  = &ttnpb.ApplicationUp{
		EndDeviceIds: registeredDeviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				FPort: 123,
			},
		},
	}
	registeredApplicationUp2 = ttnpb.ApplicationUp{
		EndDeviceIds: registeredDeviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				FPort: 124,
			},
		},
	}
	unregisteredApplicationUp = ttnpb.ApplicationUp{
		EndDeviceIds: unregisteredDeviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				FPort: 123,
			},
		},
	}

	timeout = (1 << 6) * test.Delay
)

func TestAuthentication(t *testing.T) {
	a, ctx := test.New(t)

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, ttnpb.Right_RIGHT_APPLICATION_SETTINGS_PACKAGES)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	redisClient, flush := test.NewRedis(ctx, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	apRegistry, err := redis.NewApplicationPackagesRegistry(ctx, redisClient, 10*time.Second)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	srv, err := packages.New(ctx, as, apRegistry, map[string]packages.ApplicationPackageHandler{}, 1, 10*time.Second)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	c.RegisterGRPC(&grpcServiceRegistererWrapper{srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewApplicationPackageRegistryClient(c.LoopbackConn())

	for _, tc := range []struct {
		ID  *ttnpb.EndDeviceIdentifiers
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
		t.Run(fmt.Sprintf("%v:%v", tc.ID.ApplicationIds.ApplicationId, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			creds := grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     tc.Key,
				AllowInsecure: true,
			})

			_, err := client.List(ctx, tc.ID, creds)
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
	a, ctx := test.New(t)

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, ttnpb.Right_RIGHT_APPLICATION_SETTINGS_PACKAGES)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	redisClient, flush := test.NewRedis(ctx, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	apRegistry, err := redis.NewApplicationPackagesRegistry(ctx, redisClient, 10*time.Second)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	handleUpCh := make(chan *handleUpRequest, 4)
	mockHandler := createMockPackageHandler(handleUpCh)
	handlers := map[string]packages.ApplicationPackageHandler{
		mockHandler.Package().Name: mockHandler,
	}
	srv, err := packages.New(ctx, as, apRegistry, handlers, 1, 10*time.Second)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	c.RegisterGRPC(&grpcServiceRegistererWrapper{srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

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
		res, err := client.List(ctx, registeredDeviceID, creds)
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
			Ids: registeredAssociationID,
		}, creds)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		res, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
			Ids: registeredDeviceID,
		}, creds)
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res.Associations, should.HaveLength, 0)
	})

	association := &ttnpb.ApplicationPackageAssociation{
		Ids:         registeredAssociationID,
		PackageName: "test-package",
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"state": {
					Kind: &structpb.Value_NumberValue{
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
			Association: association,
			FieldMask:   ttnpb.FieldMask("package_name", "data"),
		}, creds)
		a.So(err, should.BeNil)
		association.CreatedAt = res.CreatedAt
		association.UpdatedAt = res.UpdatedAt
		a.So(res, should.Resemble, association)
	})

	// Check that the association is available.
	t.Run("AssociationsFound", func(t *testing.T) {
		a := assertions.New(t)
		res1, err := client.GetAssociation(ctx, &ttnpb.GetApplicationPackageAssociationRequest{
			Ids:       registeredAssociationID,
			FieldMask: ttnpb.FieldMask("package_name", "data"),
		}, creds)
		a.So(err, should.BeNil)
		a.So(res1, should.Resemble, association)

		res2, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
			Ids:       registeredDeviceID,
			FieldMask: ttnpb.FieldMask("package_name", "data"),
		}, creds)
		a.So(err, should.BeNil)
		a.So(res2, should.NotBeNil)
		a.So(res2.Associations, should.HaveLength, 1)
		a.So(res2.Associations[0], should.Resemble, association)
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
				up:    registeredApplicationUp1,
				valid: true,
			},
			{
				name:  "Different FPort",
				up:    &registeredApplicationUp2,
				valid: true,
			},
			{
				name:  "Wrong application",
				up:    &unregisteredApplicationUp,
				valid: false,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				a := assertions.New(t)

				err := as.Publish(ctx, tc.up)
				a.So(err, should.BeNil)

				select {
				case up := <-handleUpCh:
					{
						if !tc.valid {
							t.Fatal("unexpected uplink")
						} else {
							a.So(up.ctx, should.NotBeNil)
							a.So(up.assoc, should.Resemble, association)
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

		_, err := client.DeleteAssociation(ctx, registeredAssociationID, creds)
		a.So(err, should.BeNil)

		_, err = client.GetAssociation(ctx, &ttnpb.GetApplicationPackageAssociationRequest{
			Ids: registeredAssociationID,
		}, creds)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		res, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
			Ids: registeredDeviceID,
		}, creds)
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res.Associations, should.BeEmpty)

		err = as.Publish(ctx, registeredApplicationUp1)
		a.So(err, should.BeNil)
		select {
		case <-handleUpCh:
			t.Fatal("unexpected uplink arrived")
		case <-time.After(2 * timeout):
			break
		}
	})

	// Check if list pagination is correct.
	t.Run("Pagination", func(t *testing.T) {
		a := assertions.New(t)

		ttnredis.SetPaginationDefaults(ttnredis.PaginationDefaults{DefaultLimit: 10})

		for i := 1; i < 21; i++ {
			association := &ttnpb.ApplicationPackageAssociation{
				Ids: &ttnpb.ApplicationPackageAssociationIdentifiers{
					EndDeviceIds: registeredDeviceID,
					FPort:        uint32(i),
				},
				PackageName: fmt.Sprintf("test-package-%v", i),
			}
			res, err := client.SetAssociation(ctx, &ttnpb.SetApplicationPackageAssociationRequest{
				Association: association,
				FieldMask:   ttnpb.FieldMask("package_name"),
			}, creds)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
		}

		for _, tc := range []struct {
			limit    uint32
			page     uint32
			portLow  uint32
			portHigh uint32
			length   int
		}{
			{
				limit:    10,
				page:     0,
				portLow:  1,
				portHigh: 10,
				length:   10,
			},
			{
				limit:    10,
				page:     1,
				portLow:  1,
				portHigh: 10,
				length:   10,
			},
			{
				limit:    10,
				page:     2,
				portLow:  11,
				portHigh: 20,
				length:   10,
			},
			{
				limit:  10,
				page:   3,
				length: 0,
			},
			{
				limit:    0,
				page:     0,
				portLow:  1,
				portHigh: 10,
				length:   10,
			},
		} {
			t.Run(fmt.Sprintf("limit:%v_page:%v", tc.limit, tc.page),
				func(t *testing.T) {
					a := assertions.New(t)

					res, err := client.ListAssociations(ctx, &ttnpb.ListApplicationPackageAssociationRequest{
						Ids:       registeredDeviceID,
						Limit:     tc.limit,
						Page:      tc.page,
						FieldMask: ttnpb.FieldMask("package_name"),
					}, creds)
					a.So(err, should.BeNil)
					a.So(res, should.NotBeNil)
					a.So(res.Associations, should.HaveLength, tc.length)
					for _, association := range res.Associations {
						a.So(association.GetIds().FPort, should.BeBetweenOrEqual, tc.portLow, tc.portHigh)
					}
				})
		}
	})
}
