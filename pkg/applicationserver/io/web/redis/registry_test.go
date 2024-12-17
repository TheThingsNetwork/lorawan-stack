// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package redis implements a Redis-backed Webhook registry.
package redis

import (
	"fmt"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var Timeout = 10 * test.Delay

func TestWebhookRegistry(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)
	cl, flush := test.NewRedis(ctx, "redis_test")
	t.Cleanup(func() {
		flush()
		cl.Close()
	})

	ids := &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "myapp-00",
		},
		WebhookId: "webhook-00",
	}
	ids1 := &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "myapp-01",
		},
		WebhookId: "webhook-01",
	}
	ids2 := &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "myapp-02",
		},
		WebhookId: "webhook-02",
	}

	registry := &WebhookRegistry{
		Redis:   cl,
		LockTTL: test.Delay << 10,
	}
	if err := registry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	paths := []string{"ids", "base_url", "format"}
	format := "json"
	format2 := "xml"
	baseURL := "http://localhost/test1"
	baseURL2 := "http://localhost/test2"

	createWebhookFunc := func(ps *ttnpb.ApplicationWebhook,
	) (*ttnpb.ApplicationWebhook, []string, error) { // nolint: unparam
		a.So(ps, should.BeNil)
		return &ttnpb.ApplicationWebhook{
			Ids:     ids1,
			Format:  format,
			BaseUrl: baseURL,
		}, paths, nil
	}
	updateWebhookFunc := func(ps *ttnpb.ApplicationWebhook,
	) (*ttnpb.ApplicationWebhook, []string, error) { // nolint: unparam
		a.So(ps, should.NotBeNil)
		return &ttnpb.ApplicationWebhook{
			Ids:     ids1,
			Format:  format,
			BaseUrl: baseURL,
		}, paths, nil
	}
	updateFieldMaskWebhookFunc := func(ps *ttnpb.ApplicationWebhook,
	) (*ttnpb.ApplicationWebhook, []string, error) { // nolint: unparam
		a.So(ps, should.NotBeNil)
		return &ttnpb.ApplicationWebhook{
			Ids:     ids1,
			Format:  format2,
			BaseUrl: baseURL2,
		}, []string{"ids", "base_url"}, nil
	}
	deleteWebhookFunc := func(ps *ttnpb.ApplicationWebhook,
	) (*ttnpb.ApplicationWebhook, []string, error) { // nolint: unparam
		a.So(ps, should.NotBeNil)
		return nil, nil, nil
	}
	listWebhookFunc := func(ps *ttnpb.ApplicationWebhook,
	) (*ttnpb.ApplicationWebhook, []string, error) { // nolint: unparam
		a.So(ps, should.BeNil)
		return &ttnpb.ApplicationWebhook{
			Ids:     ids2,
			Format:  format,
			BaseUrl: baseURL,
		}, paths, nil
	}

	t.Run("GetNonExisting", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		webhook, err := registry.Get(ctx, ids, []string{"ids"})
		a.So(webhook, should.BeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})

	t.Run("CreateReadUpdateDelete", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		webhook, err := registry.Set(ctx, ids1, paths, createWebhookFunc)
		a.So(err, should.BeNil)
		a.So(webhook, should.NotBeNil)
		a.So(webhook.Ids, should.Resemble, ids1)
		a.So(webhook.Format, should.NotBeNil)
		a.So(webhook.Format, should.Equal, format)

		retrieved, err := registry.Get(ctx, ids1, paths)
		a.So(err, should.BeNil)
		a.So(retrieved, should.NotBeNil)
		a.So(retrieved.Ids, should.Resemble, ids1)
		a.So(retrieved.Format, should.NotBeNil)
		a.So(retrieved.Format, should.Equal, format)

		updated, err := registry.Set(ctx, ids1, paths, updateWebhookFunc)
		a.So(err, should.BeNil)
		a.So(updated, should.NotBeNil)
		a.So(updated.Ids, should.Resemble, ids1)
		a.So(updated.Format, should.NotBeNil)
		a.So(updated.Format, should.Equal, format)
		a.So(updated.BaseUrl, should.NotBeNil)
		a.So(updated.BaseUrl, should.Equal, baseURL)

		updated2, err := registry.Set(ctx, ids1, paths, updateFieldMaskWebhookFunc)
		a.So(err, should.BeNil)
		a.So(updated2, should.NotBeNil)
		a.So(updated2.Ids, should.Resemble, ids1)
		a.So(updated2.Format, should.NotBeNil)
		a.So(updated2.Format, should.Equal, format)
		a.So(updated2.BaseUrl, should.NotBeNil)
		a.So(updated2.BaseUrl, should.Equal, baseURL2)

		deleted, err := registry.Set(ctx, ids1, []string{"ids"}, deleteWebhookFunc)
		a.So(err, should.BeNil)
		a.So(deleted, should.BeNil)
	})

	t.Run("List", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		webhook, err := registry.Set(ctx, ids2, paths, listWebhookFunc)
		a.So(err, should.BeNil)
		a.So(webhook, should.NotBeNil)
		a.So(webhook.Ids, should.Resemble, ids2)
		a.So(webhook.Format, should.NotBeNil)
		a.So(webhook.Format, should.Equal, format)

		webhooks, err := registry.List(ctx, ids2.ApplicationIds, paths)
		a.So(err, should.BeNil)
		a.So(webhooks, should.HaveLength, 1)
		a.So(webhooks[0], should.NotBeNil)
		a.So(webhooks[0].Ids, should.Resemble, ids2)
		a.So(webhooks[0].Format, should.NotBeNil)
		a.So(webhooks[0].Format, should.Equal, format)
	})

	t.Run("Pagination", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)

		ttnredis.SetPaginationDefaults(ttnredis.PaginationDefaults{DefaultLimit: 10})

		for i := 1; i < 21; i++ {
			ids3 := &ttnpb.ApplicationWebhookIdentifiers{
				ApplicationIds: &ttnpb.ApplicationIdentifiers{
					ApplicationId: "myapp-pagination",
				},
				WebhookId: fmt.Sprintf("webhook-%02d", i),
			}

			webhook, err := registry.Set(
				ctx,
				ids3,
				paths,
				func(ps *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
					a.So(ps, should.BeNil)
					return &ttnpb.ApplicationWebhook{
						Ids:     ids3,
						Format:  format,
						BaseUrl: baseURL,
					}, paths, nil
				},
			)
			a.So(err, should.BeNil)
			a.So(webhook, should.NotBeNil)
		}

		for _, tc := range []struct {
			limit  uint32
			page   uint32
			idLow  string
			idHigh string
			length int
		}{
			{
				limit:  10,
				page:   0,
				idLow:  "webhook-01",
				idHigh: "webhook-10",
				length: 10,
			},
			{
				limit:  10,
				page:   1,
				idLow:  "webhook-01",
				idHigh: "webhook-10",
				length: 10,
			},
			{
				limit:  10,
				page:   2,
				idLow:  "webhook-11",
				idHigh: "webhook-20",
				length: 10,
			},
			{
				limit:  10,
				page:   3,
				length: 0,
			},
			{
				limit:  0,
				page:   0,
				idLow:  "webhook-01",
				idHigh: "webhook-10",
				length: 10,
			},
		} {
			t.Run(fmt.Sprintf("limit:%v_page:%v", tc.limit, tc.page),
				func(t *testing.T) {
					t.Parallel()
					var total int64
					paginationCtx := registry.WithPagination(ctx, tc.limit, tc.page, &total)

					webhooks, err := registry.List(paginationCtx, &ttnpb.ApplicationIdentifiers{
						ApplicationId: "myapp-pagination",
					},
						paths,
					)
					a.So(err, should.BeNil)
					a.So(webhooks, should.HaveLength, tc.length)
					a.So(total, should.Equal, 20)
					for _, webhook := range webhooks {
						a.So(webhook.Ids.WebhookId, should.BeBetweenOrEqual, tc.idLow, tc.idHigh)
					}
				})
		}
	})
}
