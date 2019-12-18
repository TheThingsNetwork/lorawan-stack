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

func TestUplinks(t *testing.T) {
	withClient(t, nil,
		func(t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			for _, tc := range []struct {
				name   string
				body   string
				err    error
				assert func(*assertions.Assertion, *api.Uplinks)
			}{
				{
					name: "Send",
					body: `{
					"result": {
						"01-02-03-04-05-06-07-08": {
							"result": {
								"dnlink": {
									"port": 22,
									"payload": "0405"
								}
							}
						}
					}
				}`,
					assert: func(a *assertions.Assertion, u *api.Uplinks) {
						eui := objects.EUI{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
						resp, err := u.Send(objects.DeviceUplinks{
							eui: objects.LoRaUplink{
								FCnt:      42,
								Port:      200,
								Payload:   objects.Hex{0x03, 0x04},
								DR:        4,
								Freq:      865000000,
								Timestamp: 100.0,
							},
						})
						req := <-reqChan
						a.So(resp, should.Resemble, objects.DeviceUplinkResponses{
							eui: objects.DeviceUplinkResponse{
								Result: objects.UplinkResponse{
									Downlink: &objects.LoRaDnlink{
										Port:    22,
										Payload: objects.Hex{0x04, 0x05},
									},
								},
							},
						})
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "POST")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme: "https",
							Host:   "dms.loracloud.com",
							Path:   "/api/v1/uplink/send",
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

					tc.assert(a, cl.Uplinks)
				})
			}
		})
}
