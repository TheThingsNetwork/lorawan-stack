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
	"context"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

type mockTemplateRegisterer struct {
	context.Context
	ttnpb.ApplicationWebhookTemplateRegistryServer
}

func (m *mockTemplateRegisterer) Roles() []ttnpb.PeerInfo_Role {
	return nil
}

func (m *mockTemplateRegisterer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterApplicationWebhookTemplateRegistryServer(s, m.ApplicationWebhookTemplateRegistryServer)
}

func (m *mockTemplateRegisterer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterApplicationWebhookTemplateRegistryHandler(m.Context, s, conn)
}

func TestTemplateStore(t *testing.T) {
	ctx := test.Context()

	for _, tc := range []struct {
		name       string
		fetcher    fetch.Interface
		assertGet  func(*assertions.Assertion, *ttnpb.ApplicationWebhookTemplate, error)
		assertList func(*assertions.Assertion, *ttnpb.ApplicationWebhookTemplates, error)
	}{
		{
			name: "InvalidStore",
			fetcher: fetch.NewMemFetcher(map[string][]byte{
				"templates.json": []byte(`invalid-json`),
			}),
			assertGet: func(a *assertions.Assertion, res *ttnpb.ApplicationWebhookTemplate, err error) {
				a.So(err, should.NotBeNil)
				a.So(res, should.BeNil)
			},
			assertList: func(a *assertions.Assertion, res *ttnpb.ApplicationWebhookTemplates, err error) {
				a.So(err, should.NotBeNil)
				a.So(res, should.BeNil)
			},
		},
		{
			name: "EmptyStore",
			fetcher: fetch.NewMemFetcher(map[string][]byte{
				"templates.json": []byte(`[]`),
			}),
			assertGet: func(a *assertions.Assertion, res *ttnpb.ApplicationWebhookTemplate, err error) {
				a.So(err, should.NotBeNil)
				a.So(res, should.BeNil)
			},
			assertList: func(a *assertions.Assertion, res *ttnpb.ApplicationWebhookTemplates, err error) {
				a.So(err, should.BeNil)
				a.So(res, should.NotBeNil)
				a.So(res.Templates, should.BeEmpty)
			},
		},
		{
			name: "NormalStore",
			fetcher: fetch.NewMemFetcher(map[string][]byte{
				"templates.json": []byte(`["foo"]`),
				"foo.json":       []byte(`{"ids":{"template_id":"foo"},"name":"Foo","description":"Bar "}`),
			}),
			assertGet: func(a *assertions.Assertion, res *ttnpb.ApplicationWebhookTemplate, err error) {
				a.So(err, should.BeNil)
				a.So(res, should.NotBeNil)
				a.So(res.Name, should.NotBeEmpty)
				a.So(res.Description, should.BeEmpty)
			},
			assertList: func(a *assertions.Assertion, res *ttnpb.ApplicationWebhookTemplates, err error) {
				a.So(err, should.BeNil)
				a.So(res, should.NotBeNil)
				a.So(res.Templates, should.HaveLength, 1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			a := assertions.New(t)

			store, err := web.NewTemplateStore(tc.fetcher)
			a.So(err, should.BeNil)

			c := component.MustNew(test.GetLogger(t), &component.Config{})
			c.RegisterGRPC(&mockTemplateRegisterer{ctx, store})
			test.Must(nil, c.Start())
			defer c.Close()

			client := ttnpb.NewApplicationWebhookTemplateRegistryClient(c.LoopbackConn())

			getRes, err := client.Get(ctx, &ttnpb.GetApplicationWebhookTemplateRequest{
				ApplicationWebhookTemplateIdentifiers: ttnpb.ApplicationWebhookTemplateIdentifiers{
					TemplateID: "foo",
				},
			})
			tc.assertGet(a, getRes, err)

			listRes, err := client.List(ctx, &ttnpb.ListApplicationWebhookTemplatesRequest{
				FieldMask: types.FieldMask{
					Paths: []string{
						"name",
						"description",
					},
				},
			})
			tc.assertList(a, listRes, err)
		})
	}
}
