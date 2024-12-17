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

// Package redis implements a Redis-backed PubSub registry.
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

func TestPubSubRegistry(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)
	cl, flush := test.NewRedis(ctx, "redis_test")
	t.Cleanup(func() {
		flush()
		cl.Close()
	})

	ids := &ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "myapp-00",
		},
		PubSubId: "pubsub-00",
	}
	ids1 := &ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "myapp-01",
		},
		PubSubId: "pubsub-01",
	}
	ids2 := &ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "myapp-02",
		},
		PubSubId: "pubsub-02",
	}
	provider := &ttnpb.ApplicationPubSub_Nats{
		Nats: &ttnpb.ApplicationPubSub_NATSProvider{
			ServerUrl: "nats://localhost",
		},
	}

	registry := &PubSubRegistry{
		Redis:   cl,
		LockTTL: test.Delay << 10,
	}
	if err := registry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	paths := []string{"ids", "provider", "format", "base_topic"}
	format := "json"
	format2 := "xml"
	baseTopic := "app1.ps1"
	baseTopic2 := "app1.ps2"

	createPubSubFunc := func(ps *ttnpb.ApplicationPubSub,
	) (*ttnpb.ApplicationPubSub, []string, error) { // nolint: unparam
		a.So(ps, should.BeNil)
		return &ttnpb.ApplicationPubSub{
			Ids:      ids1,
			Provider: provider,
			Format:   format,
		}, paths, nil
	}
	updatePubSubFunc := func(ps *ttnpb.ApplicationPubSub,
	) (*ttnpb.ApplicationPubSub, []string, error) { // nolint: unparam
		a.So(ps, should.NotBeNil)
		return &ttnpb.ApplicationPubSub{
			Ids:       ids1,
			Provider:  provider,
			Format:    format,
			BaseTopic: baseTopic,
		}, paths, nil
	}
	updateFieldMaskPubSubFunc := func(ps *ttnpb.ApplicationPubSub,
	) (*ttnpb.ApplicationPubSub, []string, error) { // nolint: unparam
		a.So(ps, should.NotBeNil)
		return &ttnpb.ApplicationPubSub{
			Ids:       ids1,
			Provider:  provider,
			Format:    format2,
			BaseTopic: baseTopic2,
		}, []string{"ids", "base_topic"}, nil
	}
	deletePubSubFunc := func(ps *ttnpb.ApplicationPubSub,
	) (*ttnpb.ApplicationPubSub, []string, error) { // nolint: unparam
		a.So(ps, should.NotBeNil)
		return nil, nil, nil
	}
	listPubSubFunc := func(ps *ttnpb.ApplicationPubSub,
	) (*ttnpb.ApplicationPubSub, []string, error) { // nolint: unparam
		a.So(ps, should.BeNil)
		return &ttnpb.ApplicationPubSub{
			Ids:      ids2,
			Provider: provider,
			Format:   format,
		}, paths, nil
	}

	t.Run("GetNonExisting", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		pubsub, err := registry.Get(ctx, ids, []string{"ids"})
		a.So(pubsub, should.BeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})

	t.Run("CreateReadUpdateDelete", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		pubsub, err := registry.Set(ctx, ids1, paths, createPubSubFunc)
		a.So(err, should.BeNil)
		a.So(pubsub, should.NotBeNil)
		a.So(pubsub.Ids, should.Resemble, ids1)
		a.So(pubsub.Format, should.NotBeNil)
		a.So(pubsub.Format, should.Equal, format)

		retrieved, err := registry.Get(ctx, ids1, paths)
		a.So(err, should.BeNil)
		a.So(retrieved, should.NotBeNil)
		a.So(retrieved.Ids, should.Resemble, ids1)
		a.So(retrieved.Format, should.NotBeNil)
		a.So(retrieved.Format, should.Equal, format)

		updated, err := registry.Set(ctx, ids1, paths, updatePubSubFunc)
		a.So(err, should.BeNil)
		a.So(updated, should.NotBeNil)
		a.So(updated.Ids, should.Resemble, ids1)
		a.So(updated.Format, should.NotBeNil)
		a.So(updated.Format, should.Equal, format)
		a.So(updated.BaseTopic, should.NotBeNil)
		a.So(updated.BaseTopic, should.Equal, baseTopic)

		updated2, err := registry.Set(ctx, ids1, paths, updateFieldMaskPubSubFunc)
		a.So(err, should.BeNil)
		a.So(updated2, should.NotBeNil)
		a.So(updated2.Ids, should.Resemble, ids1)
		a.So(updated2.Format, should.NotBeNil)
		a.So(updated2.Format, should.Equal, format)
		a.So(updated2.BaseTopic, should.NotBeNil)
		a.So(updated2.BaseTopic, should.Equal, baseTopic2)

		deleted, err := registry.Set(ctx, ids1, []string{"ids"}, deletePubSubFunc)
		a.So(err, should.BeNil)
		a.So(deleted, should.BeNil)
	})

	t.Run("List", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)
		pubsub, err := registry.Set(ctx, ids2, paths, listPubSubFunc)
		a.So(err, should.BeNil)
		a.So(pubsub, should.NotBeNil)
		a.So(pubsub.Ids, should.Resemble, ids2)
		a.So(pubsub.Format, should.NotBeNil)
		a.So(pubsub.Format, should.Equal, format)

		pubsubs, err := registry.List(ctx, ids2.ApplicationIds, paths)
		a.So(err, should.BeNil)
		a.So(pubsubs, should.HaveLength, 1)
		a.So(pubsubs[0], should.NotBeNil)
		a.So(pubsubs[0].Ids, should.Resemble, ids2)
		a.So(pubsubs[0].Format, should.NotBeNil)
		a.So(pubsubs[0].Format, should.Equal, format)
	})

	t.Run("Pagination", func(t *testing.T) {
		t.Parallel()
		a, ctx := test.New(t)

		ttnredis.SetPaginationDefaults(ttnredis.PaginationDefaults{DefaultLimit: 10})

		for i := 1; i < 21; i++ {
			ids3 := &ttnpb.ApplicationPubSubIdentifiers{
				ApplicationIds: &ttnpb.ApplicationIdentifiers{
					ApplicationId: "myapp-pagination",
				},
				PubSubId: fmt.Sprintf("pubsub-%02d", i),
			}

			pubsub, err := registry.Set(
				ctx,
				ids3,
				paths,
				func(ps *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
					a.So(ps, should.BeNil)
					return &ttnpb.ApplicationPubSub{
						Ids:      ids3,
						Provider: provider,
						Format:   format,
					}, paths, nil
				},
			)
			a.So(err, should.BeNil)
			a.So(pubsub, should.NotBeNil)
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
				idLow:  "pubsub-01",
				idHigh: "pubsub-10",
				length: 10,
			},
			{
				limit:  10,
				page:   1,
				idLow:  "pubsub-01",
				idHigh: "pubsub-10",
				length: 10,
			},
			{
				limit:  10,
				page:   2,
				idLow:  "pubsub-11",
				idHigh: "pubsub-20",
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
				idLow:  "pubsub-01",
				idHigh: "pubsub-10",
				length: 10,
			},
		} {
			t.Run(fmt.Sprintf("limit:%v_page:%v", tc.limit, tc.page),
				func(t *testing.T) {
					t.Parallel()
					var total int64
					paginationCtx := registry.WithPagination(ctx, tc.limit, tc.page, &total)

					pubsubs, err := registry.List(paginationCtx, &ttnpb.ApplicationIdentifiers{
						ApplicationId: "myapp-pagination",
					},
						paths,
					)
					a.So(err, should.BeNil)
					a.So(pubsubs, should.HaveLength, tc.length)
					a.So(total, should.Equal, 20)
					for _, pubsub := range pubsubs {
						a.So(pubsub.Ids.PubSubId, should.BeBetweenOrEqual, tc.idLow, tc.idHigh)
					}
				})
		}
	})
}
