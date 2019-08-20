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
	"strconv"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestLink(t *testing.T) {
	a := assertions.New(t)

	registeredApplicationKey := "secret"

	ctx := test.Context()
	ns, nsAddr := startMockNS(ctx, func(md rpcmetadata.MD) bool {
		return md.AuthType == "Bearer" && md.AuthValue == registeredApplicationKey
	})

	paths := []string{
		"network_server_address",
		"api_key",
		"default_formatters",
	}

	// app1 is added to the link registry, app2 will be linked at runtime.
	app1 := ttnpb.ApplicationIdentifiers{ApplicationID: "app1"}
	app2 := ttnpb.ApplicationIdentifiers{ApplicationID: "app2"}
	app3 := ttnpb.ApplicationIdentifiers{ApplicationID: "app3"}
	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	linkRegistry := &redis.LinkRegistry{Redis: redisClient}
	linkRegistry.Set(ctx, app1, paths, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &ttnpb.ApplicationLink{
			APIKey: registeredApplicationKey,
		}, paths, nil
	})

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: config.Cluster{
				NetworkServer: nsAddr,
			},
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
	as, err := applicationserver.New(c, &applicationserver.Config{
		LinkMode: "all",
		Links:    linkRegistry,
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	test.Must(nil, c.Start())
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)

	// Expect app1 to be linked through the registry.
	{
		select {
		case ids := <-ns.linkCh:
			a.So(ids, should.Resemble, app1)
		case <-time.After(Timeout):
			t.Fatal("Expect link timeout")
		}
	}

	// app2: expect no link, set link, expect link, delete link and expect link to be gone.
	t.Run("Success", func(t *testing.T) {
		for i, link := range []ttnpb.ApplicationLink{
			{
				// Cluster-local Network Server.
				APIKey: registeredApplicationKey,
			},
			{
				// External Network Server.
				NetworkServerAddress: nsAddr,
				APIKey:               registeredApplicationKey,
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)
				ctx := rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, app2): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_LINK},
						},
					},
				})

				// Expect no link.
				_, err := as.GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
					ApplicationIdentifiers: app2,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				_, err = as.GetLinkStats(ctx, &app2)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Set link, expect link to establish.
				_, err = as.SetLink(ctx, &ttnpb.SetApplicationLinkRequest{
					ApplicationIdentifiers: app2,
					ApplicationLink:        link,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				select {
				case ids := <-ns.linkCh:
					a.So(ids, should.Resemble, app2)
				case <-time.After(Timeout):
					t.Fatal("Expect link timeout")
				}
				actual, err := as.GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
					ApplicationIdentifiers: app2,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				a.So(*actual, should.Resemble, link)
				stats, err := as.GetLinkStats(ctx, &app2)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				a.So(stats.NetworkServerAddress, should.Equal, link.NetworkServerAddress)

				// Wait for link to subscribe internally.
				time.Sleep(Timeout)

				// Delete link.
				_, err = as.DeleteLink(ctx, &app2)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				select {
				case ids := <-ns.unlinkCh:
					a.So(ids, should.Resemble, app2)
				case <-time.After(Timeout):
					t.Fatal("Expect unlink timeout")
				}
				_, err = as.GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
					ApplicationIdentifiers: app2,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				_, err = as.GetLinkStats(ctx, &app2)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			})
		}
	})

	// app3: expect no link, set link with invalid auth, expect error, delete link and expect link to be gone.
	t.Run("InvalidAuth", func(t *testing.T) {
		for i, link := range []ttnpb.ApplicationLink{
			{
				// Cluster-local Network Server.
				APIKey: "invalid",
			},
			{
				// External Network Server.
				NetworkServerAddress: nsAddr,
				APIKey:               "invalid",
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)
				ctx := rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, app3): {
							Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_LINK},
						},
					},
				})

				// Expect no link.
				_, err := as.GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
					ApplicationIdentifiers: app3,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				_, err = as.GetLinkStats(ctx, &app3)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Set link, expect link to establish.
				_, err = as.SetLink(ctx, &ttnpb.SetApplicationLinkRequest{
					ApplicationIdentifiers: app3,
					ApplicationLink:        link,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				select {
				case <-ns.linkCh:
					t.Fatal("Expect no link with invalid authentication")
				case <-time.After(Timeout):
				}

				actual, err := as.GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
					ApplicationIdentifiers: app3,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				a.So(*actual, should.Resemble, link)
				_, err = as.GetLinkStats(ctx, &app3)
				if !a.So(err, should.NotBeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Delete link.
				_, err = as.DeleteLink(ctx, &app3)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Unexpected error: %v", err)
				}
				_, err = as.GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
					ApplicationIdentifiers: app3,
					FieldMask: pbtypes.FieldMask{
						Paths: paths,
					},
				})
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				_, err = as.GetLinkStats(ctx, &app3)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			})
		}
	})
}
