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
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredDeviceIDs = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-app",
		},
		DeviceID: "foo-device",
		DevAddr:  devAddrPtr(types.DevAddr{0x42, 0xff, 0xff, 0xff}),
	}
	unregisteredDeviceIDs = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "bar-app",
		},
		DeviceID: "bar-device",
		DevAddr:  devAddrPtr(types.DevAddr{0x42, 0x42, 0x42, 0x42}),
	}

	timeout = 10 * test.Delay
)

func devAddrPtr(devAddr types.DevAddr) *types.DevAddr {
	return &devAddr
}

func TestWebhooks(t *testing.T) {
	ctx := test.Context()
	redisClient, flush := test.NewRedis(t, "web_test")
	defer flush()
	defer redisClient.Close()
	registry := &redis.WebhookRegistry{
		Redis: redisClient,
	}
	ids := ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-app",
		},
		WebhookID: "bar-hook",
	}
	registry.Set(ctx, ids, nil, func(_ *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
		hook := &ttnpb.ApplicationWebhook{
			BaseURL: "https://myapp.com/api/ttn/v3",
			Headers: map[string]string{
				"Authorization": "key secret",
			},
			Formatter: "json",
			UplinkMessage: &ttnpb.ApplicationWebhook_Message{
				Path: "up",
			},
			JoinAccept: &ttnpb.ApplicationWebhook_Message{
				Path: "join",
			},
			DownlinkAck: &ttnpb.ApplicationWebhook_Message{
				Path: "down/ack",
			},
			DownlinkNack: &ttnpb.ApplicationWebhook_Message{
				Path: "down/nack",
			},
			DownlinkSent: &ttnpb.ApplicationWebhook_Message{
				Path: "down/sent",
			},
			DownlinkQueued: &ttnpb.ApplicationWebhook_Message{
				Path: "down/queued",
			},
			DownlinkFailed: &ttnpb.ApplicationWebhook_Message{
				Path: "down/failed",
			},
			LocationSolved: &ttnpb.ApplicationWebhook_Message{
				Path: "location",
			},
		}
		paths := []string{
			"base_url",
			"headers",
			"formatter",
			"uplink_message",
			"join_accept",
			"downlink_ack",
			"downlink_nack",
			"downlink_sent",
			"downlink_failed",
			"downlink_queued",
			"location_solved",
		}
		return hook, paths, nil
	})

	sink := &mockSink{
		ch: make(chan *http.Request, 1),
	}
	w := &web.Webhooks{
		Registry: registry,
		Target:   sink,
	}
	sub := w.NewSubscription(ctx)

	for _, tc := range []struct {
		Name    string
		Message *ttnpb.ApplicationUp
		OK      bool
		URL     string
	}{
		{
			Name: "UplinkMessage/RegisteredDevice",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{
						SessionKeyID: "session1",
						FPort:        42,
						FCnt:         42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/up",
		},
		{
			Name: "UplinkMessage/UnregisteredDevice",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: unregisteredDeviceIDs,
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{
						SessionKeyID: "session2",
						FPort:        42,
						FCnt:         42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
					},
				},
			},
			OK: false,
		},
		{
			Name: "JoinAccept",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_JoinAccept{
					JoinAccept: &ttnpb.ApplicationJoinAccept{
						SessionKeyID: "session2",
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/join",
		},
		{
			Name: "DownlinkMessage/Ack",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_DownlinkAck{
					DownlinkAck: &ttnpb.ApplicationDownlink{
						SessionKeyID: "session2",
						FCnt:         42,
						FPort:        42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/down/ack",
		},
		{
			Name: "DownlinkMessage/Nack",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_DownlinkNack{
					DownlinkNack: &ttnpb.ApplicationDownlink{
						SessionKeyID: "session2",
						FCnt:         42,
						FPort:        42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/down/nack",
		},
		{
			Name: "DownlinkMessage/Sent",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_DownlinkSent{
					DownlinkSent: &ttnpb.ApplicationDownlink{
						SessionKeyID: "session2",
						FCnt:         42,
						FPort:        42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/down/sent",
		},
		{
			Name: "DownlinkMessage/Queued",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_DownlinkQueued{
					DownlinkQueued: &ttnpb.ApplicationDownlink{
						SessionKeyID: "session2",
						FCnt:         42,
						FPort:        42,
						FRMPayload:   []byte{0x1, 0x2, 0x3},
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/down/queued",
		},
		{
			Name: "DownlinkMessage/Failed",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_DownlinkFailed{
					DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
						ApplicationDownlink: ttnpb.ApplicationDownlink{
							SessionKeyID: "session2",
							FCnt:         42,
							FPort:        42,
							FRMPayload:   []byte{0x1, 0x2, 0x3},
						},
						Error: ttnpb.ErrorDetails{
							Name: "test",
						},
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/down/failed",
		},
		{
			Name: "LocationSolved",
			Message: &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: registeredDeviceIDs,
				Up: &ttnpb.ApplicationUp_LocationSolved{
					LocationSolved: &ttnpb.ApplicationLocation{
						Location: ttnpb.Location{
							Latitude:  10,
							Longitude: 20,
							Altitude:  30,
						},
						Service: "test",
					},
				},
			},
			OK:  true,
			URL: "https://myapp.com/api/ttn/v3/location",
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			err := sub.SendUp(tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			var req *http.Request
			select {
			case req = <-sink.ch:
				if !tc.OK {
					t.Fatalf("Did not expect message but received: %v", req)
				}
			case <-time.After(timeout):
				if tc.OK {
					t.Fatal("Expected message but nothing received")
				} else {
					return
				}
			}
			a.So(req.URL.String(), should.Equal, tc.URL)
			a.So(req.Header.Get("Authorization"), should.Equal, "key secret")
			a.So(req.Header.Get("Content-Type"), should.Equal, "application/json")
			actualBody, err := ioutil.ReadAll(req.Body)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			expectedBody, err := web.Formatters["json"].Encode(ctx, tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(actualBody, should.Resemble, expectedBody)
		})
	}
}

type mockSink struct {
	ch chan *http.Request
}

func (s *mockSink) Process(req *http.Request) error {
	s.ch <- req
	return nil
}
