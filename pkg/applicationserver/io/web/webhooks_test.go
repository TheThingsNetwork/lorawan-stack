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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/formatters"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestWebhooks(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	redisClient, flush := test.NewRedis(t, "web_test")
	defer flush()
	defer redisClient.Close()
	registry := &redis.WebhookRegistry{
		Redis: redisClient,
	}
	ids := ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		WebhookID:              registeredWebhookID,
	}
	_, err := registry.Set(ctx, ids, nil, func(_ *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
		return &ttnpb.ApplicationWebhook{
				ApplicationWebhookIdentifiers: ids,
				BaseURL:                       "https://myapp.com/api/ttn/v3",
				Headers: map[string]string{
					"Authorization": "key secret",
				},
				Format: "json",
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
			},
			[]string{
				"base_url",
				"downlink_ack",
				"downlink_failed",
				"downlink_nack",
				"downlink_queued",
				"downlink_sent",
				"format",
				"headers",
				"ids",
				"join_accept",
				"location_solved",
				"uplink_message",
			}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set webhook in registry: %s", err)
	}

	t.Run("Upstream", func(t *testing.T) {
		testSink := &mockSink{
			ch: make(chan *http.Request, 1),
		}
		for _, sink := range []web.Sink{
			testSink,
			&web.QueuedSink{
				Target:  testSink,
				Queue:   make(chan *http.Request, 4),
				Workers: 1,
			},
			&web.QueuedSink{
				Target: &web.QueuedSink{
					Target:  testSink,
					Queue:   make(chan *http.Request, 4),
					Workers: 1,
				},
				Queue:   make(chan *http.Request, 4),
				Workers: 1,
			},
		} {
			t.Run(fmt.Sprintf("%T", sink), func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				if controllable, ok := sink.(web.ControllableSink); ok {
					go controllable.Run(ctx)
				}
				w := web.NewWebhooks(ctx, nil, registry, sink)
				sub := w.NewSubscription()
				for _, tc := range []struct {
					Name    string
					Message *ttnpb.ApplicationUp
					OK      bool
					URL     string
				}{
					{
						Name: "UplinkMessage/RegisteredDevice",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x11},
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
							EndDeviceIdentifiers: unregisteredDeviceID,
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x22},
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
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: []byte{0x22},
								},
							},
						},
						OK:  true,
						URL: "https://myapp.com/api/ttn/v3/join",
					},
					{
						Name: "DownlinkMessage/Ack",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_DownlinkAck{
								DownlinkAck: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x22},
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
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_DownlinkNack{
								DownlinkNack: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x22},
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
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_DownlinkSent{
								DownlinkSent: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x22},
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
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_DownlinkQueued{
								DownlinkQueued: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x22},
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
							EndDeviceIdentifiers: registeredDeviceID,
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									ApplicationDownlink: ttnpb.ApplicationDownlink{
										SessionKeyID: []byte{0x22},
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
							EndDeviceIdentifiers: registeredDeviceID,
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
						err := sub.SendUp(ctx, tc.Message)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						var req *http.Request
						select {
						case req = <-testSink.ch:
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
						expectedBody, err := formatters.JSON.FromUp(tc.Message)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
						a.So(actualBody, should.Resemble, expectedBody)
					})
				}
			})
		}
	})

	t.Run("Downstream", func(t *testing.T) {
		is, isAddr := startMockIS(ctx)
		is.add(ctx, registeredApplicationID, registeredApplicationKey)
		httpAddress := "0.0.0.0:8098"
		conf := &component.Config{
			ServiceBase: config.ServiceBase{
				GRPC: config.GRPC{
					Listen:                      ":0",
					AllowInsecureForCredentials: true,
				},
				Cluster: config.Cluster{
					IdentityServer: isAddr,
				},
				HTTP: config.HTTP{
					Listen: httpAddress,
				},
			},
		}
		c := component.MustNew(test.GetLogger(t), conf)
		io := mock.NewServer(c)
		testSink := &mockSink{
			Component: c,
			Server:    io,
		}
		w := web.NewWebhooks(ctx, testSink, registry, testSink)
		c.RegisterWeb(w)
		test.Must(nil, c.Start())
		defer c.Close()

		mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

		t.Run("Authorization", func(t *testing.T) {
			for _, tc := range []struct {
				Name       string
				ID         ttnpb.ApplicationIdentifiers
				Key        string
				ExpectCode int
			}{
				{
					Name:       "Valid",
					ID:         registeredApplicationID,
					Key:        registeredApplicationKey,
					ExpectCode: http.StatusOK,
				},
				{
					Name:       "InvalidKey",
					ID:         registeredApplicationID,
					Key:        "invalid key",
					ExpectCode: http.StatusForbidden,
				},
				{
					Name:       "InvalidIDAndKey",
					ID:         ttnpb.ApplicationIdentifiers{ApplicationID: "--invalid-id"},
					Key:        "invalid key",
					ExpectCode: http.StatusBadRequest,
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf("http://%s/api/v3/as/applications/%s/webhooks/%s/devices/%s/down/replace",
						httpAddress, tc.ID.ApplicationID, registeredWebhookID, registeredDeviceID.DeviceID,
					)
					body := bytes.NewReader([]byte(`{"downlinks":[]}`))
					req, err := http.NewRequest(http.MethodPost, url, body)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
					res, err := http.DefaultClient.Do(req)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					a.So(res.StatusCode, should.Equal, tc.ExpectCode)
					downlinks, err := io.DownlinkQueueList(ctx, registeredDeviceID)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					a.So(downlinks, should.Resemble, []*ttnpb.ApplicationDownlink{})
				})
			}
		})
	})
}

type mockSink struct {
	*component.Component
	io.Server
	ch chan *http.Request
}

func (s *mockSink) FillContext(ctx context.Context) context.Context {
	return s.Component.FillContext(ctx)
}

func (s *mockSink) Process(req *http.Request) error {
	s.ch <- req
	return nil
}
