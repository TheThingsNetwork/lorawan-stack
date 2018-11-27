// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestWebhookRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	ctx := newContextWithRightsFetcher(test.Context())

	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	webhookReg := &redis.WebhookRegistry{Redis: redisClient}
	srv := web.NewWebhookRegistryRPC(webhookReg)
	authorizedCtx := contextWithKey(ctx, registeredApplicationKey)

	// Formats.
	{
		res, err := srv.GetFormats(authorizedCtx, ttnpb.Empty)
		a.So(err, should.BeNil)
		a.So(res.Formats, should.HaveSameElementsDeep, map[string]string{
			"json": "JSON",
			"pb":   "Protobuf",
		})
	}

	// Check empty.
	{
		res, err := srv.List(authorizedCtx, &ttnpb.ListApplicationWebhooksRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Webhooks, should.BeEmpty)
	}

	// Add.
	{
		_, err := srv.Set(authorizedCtx, &ttnpb.SetApplicationWebhookRequest{
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
		})
		a.So(err, should.BeNil)
	}

	// List; assert one.
	{
		res, err := srv.List(authorizedCtx, &ttnpb.ListApplicationWebhooksRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Webhooks, should.HaveLength, 1)
		a.So(res.Webhooks[0].BaseURL, should.Equal, "http://localhost/test")
	}

	// Get.
	{
		res, err := srv.Get(authorizedCtx, &ttnpb.GetApplicationWebhookRequest{
			ApplicationWebhookIdentifiers: ttnpb.ApplicationWebhookIdentifiers{
				ApplicationIdentifiers: registeredApplicationID,
				WebhookID:              registeredWebhookID,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.BaseURL, should.Equal, "http://localhost/test")
	}

	// Delete.
	{
		_, err := srv.Delete(authorizedCtx, &ttnpb.ApplicationWebhookIdentifiers{
			ApplicationIdentifiers: registeredApplicationID,
			WebhookID:              registeredWebhookID,
		})
		a.So(err, should.BeNil)
	}

	// Check empty.
	{
		res, err := srv.List(authorizedCtx, &ttnpb.ListApplicationWebhooksRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"base_url"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Webhooks, should.BeEmpty)
	}
}
