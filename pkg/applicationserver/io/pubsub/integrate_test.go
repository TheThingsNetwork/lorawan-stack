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

package pubsub_test

import (
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	mock_server "go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	mock_provider "go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider/mock"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/redis"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestIntegrate(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	is, isAddr := startMockIS(ctx)
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	mockProvider, err := provider.GetProvider(ttnpb.ApplicationPubSub_AWSSNSSQS)
	a.So(mockProvider, should.NotBeNil)
	a.So(err, should.BeNil)
	mockImpl := mockProvider.(*mock_provider.Impl)

	paths := []string{
		"format",
		"attributes",
	}

	// ps1 is added to the pubsub registry, app2 will be integrated at runtime.
	ps1 := ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		PubSubID:               "ps1",
	}
	ps2 := ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		PubSubID:               "ps2",
	}
	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	pubsubRegistry := &redis.PubSubRegistry{Redis: redisClient}
	_, err = pubsubRegistry.Set(ctx, ps1, paths, func(_ *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
		return &ttnpb.ApplicationPubSub{
			ApplicationPubSubIdentifiers: ps1,
			Attributes: map[string]string{
				mock_provider.MockAckDeadline: timeout.String(),
			},
			Format: "json",
		}, append(paths, "ids.application_ids", "ids.pubsub_id"), nil
	})
	a.So(err, should.BeNil)

	io := mock_server.NewServer()
	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9185",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	})
	srv, err := pubsub.Start(c, io, pubsubRegistry)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	c.RegisterGRPC(&mockRegisterer{srv})
	test.Must(nil, c.Start())
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	conn, err := grpc.Dial(":9185", append(rpcclient.DefaultDialOptions(ctx), grpc.WithInsecure(), grpc.WithBlock())...)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer conn.Close()
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})
	ps := ttnpb.NewApplicationPubSubRegistryClient(conn)

	// Expect ps1 to be integrated through the registry.
	t.Run("AlreadyExisting", func(t *testing.T) {
		select {
		case conn := <-mockImpl.OpenConnectionCh:
			a.So(conn.ApplicationPubSubIdentifiers, should.Resemble, ps1)
		case <-time.After(timeout):
			t.Fatal("Expect integration timeout")
		}
		select {
		case sub := <-io.Subscriptions():
			a.So(*sub.ApplicationIDs(), should.Resemble, registeredApplicationID)
		case <-time.After(timeout):
			t.Fatal("Expect integration timeout")
		}
	})

	// ps2: expect no integration, set integration, expect integration, delete integration and expect integration to be gone.
	t.Run("RuntimeCreation", func(t *testing.T) {
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredApplicationID): {
					Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ},
				},
			},
		})

		integration := ttnpb.ApplicationPubSub{
			ApplicationPubSubIdentifiers: ps2,
			Attributes: map[string]string{
				mock_provider.MockAckDeadline: timeout.String(),
			},
			Format: "json",
		}

		// Expect no integration.
		_, err := ps.Get(ctx, &ttnpb.GetApplicationPubSubRequest{
			ApplicationPubSubIdentifiers: ps2,
			FieldMask: pbtypes.FieldMask{
				Paths: paths,
			},
		}, creds)
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Set integration, expect integration to establish.
		_, err = ps.Set(ctx, &ttnpb.SetApplicationPubSubRequest{
			ApplicationPubSub: integration,
			FieldMask: pbtypes.FieldMask{
				Paths: paths,
			},
		}, creds)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Unexpected error: %v", err)
		}
		select {
		case conn := <-mockImpl.OpenConnectionCh:
			a.So(conn.ApplicationPubSubIdentifiers, should.Resemble, ps2)
		case <-time.After(timeout):
			t.Fatal("Expect integration timeout")
		}
		actual, err := ps.Get(ctx, &ttnpb.GetApplicationPubSubRequest{
			ApplicationPubSubIdentifiers: ps2,
			FieldMask: pbtypes.FieldMask{
				Paths: paths,
			},
		}, creds)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Unexpected error: %v", err)
		}
		actual.CreatedAt = time.Time{}
		actual.UpdatedAt = time.Time{}
		a.So(*actual, should.Resemble, integration)

		// Delete integration.
		_, err = ps.Delete(ctx, &ps2, creds)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Unexpected error: %v", err)
		}
		select {
		case conn := <-mockImpl.ShutdownCh:
			a.So(conn.ApplicationPubSubIdentifiers, should.Resemble, ps2)
		case <-time.After(timeout):
			t.Fatal("Expect integration timeout")
		}
		_, err = ps.Get(ctx, &ttnpb.GetApplicationPubSubRequest{
			ApplicationPubSubIdentifiers: ps2,
			FieldMask: pbtypes.FieldMask{
				Paths: paths,
			},
		}, creds)
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}
