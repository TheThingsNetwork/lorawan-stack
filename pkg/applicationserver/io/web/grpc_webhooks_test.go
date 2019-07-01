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

package web_test

import (
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestWebhookRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	is, isAddr := startMockIS(ctx)
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	c := component.MustNew(test.GetLogger(t), &component.Config{
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
	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	webhookReg := &redis.WebhookRegistry{Redis: redisClient}
	srv := web.NewWebhookRegistryRPC(webhookReg)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	test.Must(nil, c.Start())
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	client := ttnpb.NewApplicationWebhookRegistryClient(c.LoopbackConn())
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})

	// Formats.
	{
		res, err := client.GetFormats(ctx, ttnpb.Empty, creds)
		a.So(err, should.BeNil)
		a.So(res.Formats, should.HaveSameElementsDeep, map[string]string{
			"json":     "JSON",
			"protobuf": "Protocol Buffers",
		})
	}

	// Check empty.
	{
		res, err := client.List(ctx, &ttnpb.ListApplicationWebhooksRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		}, creds)
		a.So(err, should.BeNil)
		a.So(res.Webhooks, should.BeEmpty)
	}

	// Add.
	{
		_, err := client.Set(ctx, &ttnpb.SetApplicationWebhookRequest{
			ApplicationWebhook: ttnpb.ApplicationWebhook{
				ApplicationWebhookIdentifiers: ttnpb.ApplicationWebhookIdentifiers{
					ApplicationIdentifiers: registeredApplicationID,
					WebhookID:              registeredWebhookID,
				},
				BaseURL: "http://localhost/test",
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		}, creds)
		a.So(err, should.BeNil)
	}

	// List; assert one.
	{
		res, err := client.List(ctx, &ttnpb.ListApplicationWebhooksRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		}, creds)
		a.So(err, should.BeNil)
		a.So(res.Webhooks, should.HaveLength, 1)
		a.So(res.Webhooks[0].BaseURL, should.Equal, "http://localhost/test")
	}

	// Get.
	{
		res, err := client.Get(ctx, &ttnpb.GetApplicationWebhookRequest{
			ApplicationWebhookIdentifiers: ttnpb.ApplicationWebhookIdentifiers{
				ApplicationIdentifiers: registeredApplicationID,
				WebhookID:              registeredWebhookID,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		}, creds)
		a.So(err, should.BeNil)
		a.So(res.BaseURL, should.Equal, "http://localhost/test")
	}

	// Delete.
	{
		_, err := client.Delete(ctx, &ttnpb.ApplicationWebhookIdentifiers{
			ApplicationIdentifiers: registeredApplicationID,
			WebhookID:              registeredWebhookID,
		}, creds)
		a.So(err, should.BeNil)
	}

	// Check empty.
	{
		res, err := client.List(ctx, &ttnpb.ListApplicationWebhooksRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		}, creds)
		a.So(err, should.BeNil)
		a.So(res.Webhooks, should.BeEmpty)
	}
}
