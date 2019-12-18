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

package api_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/lora-cloud-device-management-v1/api"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/lora-cloud-device-management-v1/api/objects"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestTokens(t *testing.T) {
	withClient(t, nil,
		func(t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			for _, tc := range []struct {
				name   string
				body   string
				err    error
				assert func(*assertions.Assertion, *api.Tokens)
			}{
				{
					name: "List",
					body: `{
					"result": {
						"tokens": [{
							"name": "foo",
							"token": "foobar",
							"capabilities": ["bar"]
						}]
					}
				}`,
					assert: func(a *assertions.Assertion, t *api.Tokens) {
						tok, err := t.List()
						req := <-reqChan
						a.So(tok, should.Resemble, []objects.TokenInfo{
							{
								Name:         "foo",
								Token:        "foobar",
								Capabilities: []string{"bar"},
							},
						})
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "GET")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme: "https",
							Host:   "dms.loracloud.com",
							Path:   "/api/v1/token/list",
						})
					},
				},
				{
					name: "Update with renew",
					body: `{
					"result": {
						"name": "foo",
						"token": "foobar",
						"capabilities": ["bar"]
					}
				}`,
					assert: func(a *assertions.Assertion, t *api.Tokens) {
						tok, err := t.Update("foo", "bar", true)
						req := <-reqChan
						a.So(tok, should.Resemble, &objects.TokenInfo{
							Name:         "foo",
							Token:        "foobar",
							Capabilities: []string{"bar"},
						})
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "PUT")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme:   "https",
							Host:     "dms.loracloud.com",
							Path:     "/api/v1/token/foo/update",
							RawQuery: "name=bar&renew=",
						})
					},
				},
				{
					name: "Update without renew",
					body: `{
					"result": {
						"name": "foo",
						"token": "foobar",
						"capabilities": ["bar"]
					}
				}`,
					assert: func(a *assertions.Assertion, t *api.Tokens) {
						tok, err := t.Update("foo", "bar", false)
						req := <-reqChan
						a.So(tok, should.Resemble, &objects.TokenInfo{
							Name:         "foo",
							Token:        "foobar",
							Capabilities: []string{"bar"},
						})
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "PUT")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme:   "https",
							Host:     "dms.loracloud.com",
							Path:     "/api/v1/token/foo/update",
							RawQuery: "name=bar",
						})
					},
				},
				{
					name: "Add",
					body: `{
					"result": {
						"name": "foo",
						"token": "foobar",
						"capabilities": ["bar"]
					}
				}`,
					assert: func(a *assertions.Assertion, t *api.Tokens) {
						tok, err := t.Add("foo", "bar")
						req := <-reqChan
						a.So(tok, should.Resemble, &objects.TokenInfo{
							Name:         "foo",
							Token:        "foobar",
							Capabilities: []string{"bar"},
						})
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "POST")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme: "https",
							Host:   "dms.loracloud.com",
							Path:   "/api/v1/token/add",
						})
					},
				},
				{
					name: "Remove",
					body: `{}`,
					assert: func(a *assertions.Assertion, t *api.Tokens) {
						err := t.Remove("foo")
						req := <-reqChan
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "DELETE")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme: "https",
							Host:   "dms.loracloud.com",
							Path:   "/api/v1/token/foo",
						})
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					a := assertions.New(t)

					respChan <- &http.Response{
						Body: ioutil.NopCloser(bytes.NewBufferString(tc.body)),
					}
					errChan <- tc.err

					tc.assert(a, cl.Tokens)
				})
			}
		})
}
