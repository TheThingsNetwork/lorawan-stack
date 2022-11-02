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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api/objects"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestUplinks(t *testing.T) {
	withClient(test.Context(), t, nil,
		func(ctx context.Context, t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
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
						resp, err := u.Send(ctx, objects.DeviceUplinks{
							eui: &objects.LoRaUplink{
								FCnt:      uint32Ptr(42),
								Port:      uint8Ptr(199),
								Payload:   hexPtr(objects.Hex{0x03, 0x04}),
								DR:        uint8Ptr(4),
								Freq:      uint32Ptr(865000000),
								Timestamp: float64Ptr(100.0),
							},
						})
						req := <-reqChan
						a.So(resp, should.Resemble, objects.DeviceUplinkResponses{
							eui: objects.DeviceUplinkResponse{
								Result: objects.ExtendedUplinkResponse{
									UplinkResponse: objects.UplinkResponse{
										Downlink: &objects.LoRaDnlink{
											Port:    22,
											Payload: objects.Hex{0x04, 0x05},
										},
									},
									Raw: &json.RawMessage{0x7b, 0xa, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x22, 0x64, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x22, 0x3a, 0x20, 0x7b, 0xa, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x22, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x3a, 0x20, 0x32, 0x32, 0x2c, 0xa, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x22, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x3a, 0x20, 0x22, 0x30, 0x34, 0x30, 0x35, 0x22, 0xa, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x7d, 0xa, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x9, 0x7d},
								},
							},
						})
						a.So(err, should.BeNil)
						a.So(req.Method, should.Equal, "POST")
						a.So(req.URL, should.Resemble, &url.URL{
							Scheme: "https",
							Host:   "mgs.loracloud.com",
							Path:   "/api/v1/uplink/send",
						})
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					a := assertions.New(t)

					respChan <- &http.Response{
						Body:       io.NopCloser(bytes.NewBufferString(tc.body)),
						StatusCode: http.StatusOK,
					}
					errChan <- tc.err

					tc.assert(a, cl.Uplinks)
				})
			}
		})
}

func uint8Ptr(x uint8) *uint8 {
	return &x
}

func uint32Ptr(x uint32) *uint32 {
	return &x
}

func float64Ptr(x float64) *float64 {
	return &x
}

func hexPtr(x objects.Hex) *objects.Hex {
	return &x
}
