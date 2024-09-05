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
	stdio "io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/formatters"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink"
	mocksink "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockComponent struct{}

func (mockComponent) StartTask(conf *task.Config) {
	task.DefaultStartTask(conf)
}

func (mockComponent) FromRequestContext(ctx context.Context) context.Context {
	return ctx
}

func pooledSink(ctx context.Context, target sink.Sink) sink.Sink {
	return sink.NewPooledSink(ctx, mockComponent{}, target, 1, 4)
}

func TestWebhooks(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	redisClient, flush := test.NewRedis(ctx, "web_test")
	defer flush()
	defer redisClient.Close()
	downlinks := web.DownlinksConfig{
		PublicAddress: "https://example.com/api/v3",
	}
	registry := &redis.WebhookRegistry{
		Redis:   redisClient,
		LockTTL: test.Delay << 10,
	}
	if err := registry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	ids := &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIds: registeredApplicationID,
		WebhookId:      registeredWebhookID,
	}

	// Use a dummy JWT for auth check.
	longAuthHeader := "Bearer eyJhbGciOiJIUzI1NiJ9.eyJSb2xlIjoidGVzdHdocm9sZSIsIklzc3VlciI6Iklzc3VlciIsIlVzZXJuYW1lIjoidGVzdHVzZXIiLCJleHAiOjE2NDI0MjU3NDgsImlhdCI6MTY0MjQyNTc0OH0.imuGY_5xnhZYSqjPrc6EUoYV1eapswDBUIBXKVCIYSw" //nolint:lll

	//nolint:paralleltest
	for _, tc := range []struct {
		prefix string
		suffix string
	}{
		{
			prefix: "",
			suffix: "",
		},
		{
			prefix: "",
			suffix: "/",
		},
		{
			prefix: "/",
			suffix: "",
		},
		{
			prefix: "/",
			suffix: "/",
		},
	} {
		t.Run(fmt.Sprintf("Prefix%q/Suffix%q", tc.prefix, tc.suffix), func(t *testing.T) {
			_, ctx := test.New(t)
			_, err := registry.Set(ctx, ids, nil, func(_ *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
				return &ttnpb.ApplicationWebhook{
						Ids:     ids,
						BaseUrl: "https://myapp.com/api/ttn/v3{/appID,devID}" + tc.suffix,
						Headers: map[string]string{
							"Authorization": longAuthHeader,
						},
						DownlinkApiKey: "foo.secret",
						Format:         "json",
						UplinkMessage: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "up{?devEUI}",
						},
						UplinkNormalized: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "up/normalized{?devEUI}",
						},
						JoinAccept: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "join{?joinEUI}",
						},
						DownlinkAck: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "down/ack",
						},
						DownlinkNack: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "down/nack",
						},
						DownlinkSent: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "down/sent",
						},
						DownlinkQueued: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "down/queued",
						},
						DownlinkQueueInvalidated: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "down/invalidated",
						},
						DownlinkFailed: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "down/failed",
						},
						LocationSolved: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "location",
						},
						ServiceData: &ttnpb.ApplicationWebhook_Message{
							Path: tc.prefix + "service/data",
						},
						Paused: false,
						FieldMask: ttnpb.FieldMask(
							"correlation_ids",
							"end_device_ids",
							"received_at",
							"simulated",
							"up.downlink_ack",
							"up.downlink_failed",
							"up.downlink_nack",
							"up.downlink_queue_invalidated",
							"up.downlink_queued",
							"up.downlink_sent",
							"up.join_accept",
							"up.location_solved",
							"up.paused",
							"up.service_data",
							"up.uplink_message",
							"up.uplink_normalized",
						),
					},
					[]string{
						"base_url",
						"downlink_ack",
						"downlink_api_key",
						"downlink_failed",
						"downlink_nack",
						"downlink_queue_invalidated",
						"downlink_queued",
						"downlink_sent",
						"field_mask",
						"format",
						"headers",
						"ids",
						"join_accept",
						"location_solved",
						"service_data",
						"paused",
						"uplink_message",
						"uplink_normalized",
					}, nil
			})
			if err != nil {
				t.Fatalf("Failed to set webhook in registry: %s", err)
			}

			t.Run("Upstream", func(t *testing.T) {
				baseURL := fmt.Sprintf(
					"https://myapp.com/api/ttn/v3/%s/%s", registeredApplicationID.ApplicationId, registeredDeviceID.DeviceId,
				)
				sinkCh := make(chan *http.Request, 1)
				testSink := mocksink.New(sinkCh)
				for _, sink := range []sink.Sink{
					testSink,
					pooledSink(ctx, testSink),
					pooledSink(ctx,
						pooledSink(ctx, testSink),
					),
				} {
					t.Run(fmt.Sprintf("%T", sink), func(t *testing.T) {
						ctx, cancel := context.WithCancel(ctx)
						defer cancel()
						c := componenttest.NewComponent(t, &component.Config{})
						componenttest.StartComponent(t, c)
						defer c.Close()
						as := mock.NewServer(c)
						_, err := web.NewWebhooks(ctx, as, registry, sink, downlinks)
						if err != nil {
							t.Fatalf("Unexpected error %v", err)
						}
						for _, tc := range []struct {
							Name    string
							Message *ttnpb.ApplicationUp
							OK      bool
							URL     string
						}{
							{
								Name: "UplinkMessage/RegisteredDevice",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_UplinkMessage{
										UplinkMessage: &ttnpb.ApplicationUplink{
											SessionKeyId: []byte{0x11},
											FPort:        42,
											FCnt:         42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/up?devEUI=%s", baseURL, types.MustEUI64(registeredDeviceID.DevEui)),
							},
							{
								Name: "UplinkNormalized/RegisteredDevice",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_UplinkNormalized{
										UplinkNormalized: &ttnpb.ApplicationUplinkNormalized{
											SessionKeyId: []byte{0x11},
											FPort:        42,
											FCnt:         42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
											NormalizedPayload: &structpb.Struct{
												Fields: map[string]*structpb.Value{
													"air": {
														Kind: &structpb.Value_StructValue{
															StructValue: &structpb.Struct{
																Fields: map[string]*structpb.Value{
																	"temperature": {
																		Kind: &structpb.Value_NumberValue{
																			NumberValue: 21.5,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/up/normalized?devEUI=%s", baseURL, types.MustEUI64(registeredDeviceID.DevEui)),
							},
							{
								Name: "UplinkMessage/UnregisteredDevice",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: &unregisteredDeviceID,
									Up: &ttnpb.ApplicationUp_UplinkMessage{
										UplinkMessage: &ttnpb.ApplicationUplink{
											SessionKeyId: []byte{0x22},
											FPort:        42,
											FCnt:         42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
										},
									},
								},
								OK: false,
							},
							{
								Name: "JoinAccept",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_JoinAccept{
										JoinAccept: &ttnpb.ApplicationJoinAccept{
											SessionKeyId: []byte{0x22},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/join?joinEUI=%s", baseURL, types.MustEUI64(registeredDeviceID.JoinEui)),
							},
							{
								Name: "DownlinkMessage/Ack",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_DownlinkAck{
										DownlinkAck: &ttnpb.ApplicationDownlink{
											SessionKeyId: []byte{0x22},
											FCnt:         42,
											FPort:        42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/down/ack", baseURL),
							},
							{
								Name: "DownlinkMessage/Nack",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_DownlinkNack{
										DownlinkNack: &ttnpb.ApplicationDownlink{
											SessionKeyId: []byte{0x22},
											FCnt:         42,
											FPort:        42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/down/nack", baseURL),
							},
							{
								Name: "DownlinkMessage/Sent",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_DownlinkSent{
										DownlinkSent: &ttnpb.ApplicationDownlink{
											SessionKeyId: []byte{0x22},
											FCnt:         42,
											FPort:        42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/down/sent", baseURL),
							},
							{
								Name: "DownlinkMessage/Queued",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_DownlinkQueued{
										DownlinkQueued: &ttnpb.ApplicationDownlink{
											SessionKeyId: []byte{0x22},
											FCnt:         42,
											FPort:        42,
											FrmPayload:   []byte{0x1, 0x2, 0x3},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/down/queued", baseURL),
							},
							{
								Name: "DownlinkMessage/QueueInvalidated",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
										DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
											Downlinks: []*ttnpb.ApplicationDownlink{
												{
													SessionKeyId: []byte{0x22},
													FCnt:         42,
													FPort:        42,
													FrmPayload:   []byte{0x1, 0x2, 0x3},
												},
											},
											LastFCntDown: 42,
											SessionKeyId: []byte{0x22},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/down/invalidated", baseURL),
							},
							{
								Name: "DownlinkMessage/Failed",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_DownlinkFailed{
										DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
											Downlink: &ttnpb.ApplicationDownlink{
												SessionKeyId: []byte{0x22},
												FCnt:         42,
												FPort:        42,
												FrmPayload:   []byte{0x1, 0x2, 0x3},
											},
											Error: &ttnpb.ErrorDetails{
												Name: "test",
											},
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/down/failed", baseURL),
							},
							{
								Name: "LocationSolved",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_LocationSolved{
										LocationSolved: &ttnpb.ApplicationLocation{
											Location: &ttnpb.Location{
												Latitude:  10,
												Longitude: 20,
												Altitude:  30,
											},
											Service: "test",
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/location", baseURL),
							},
							{
								Name: "ServiceData",
								Message: &ttnpb.ApplicationUp{
									EndDeviceIds: registeredDeviceID,
									Up: &ttnpb.ApplicationUp_ServiceData{
										ServiceData: &ttnpb.ApplicationServiceData{
											Data: &structpb.Struct{
												Fields: map[string]*structpb.Value{
													"battery": {
														Kind: &structpb.Value_NumberValue{
															NumberValue: 42.0,
														},
													},
												},
											},
											Service: "test",
										},
									},
								},
								OK:  true,
								URL: fmt.Sprintf("%s/service/data", baseURL),
							},
						} {
							t.Run(tc.Name, func(t *testing.T) {
								a := assertions.New(t)
								err := as.Publish(ctx, tc.Message)
								if !a.So(err, should.BeNil) {
									t.FailNow()
								}
								var req *http.Request
								select {
								case req = <-sinkCh:
									if !tc.OK {
										t.Fatalf("Did not expect message but received: %v", req)
									}
								case <-time.After(timeout):
									if !tc.OK {
										return
									}
									t.Fatal("Expected message but nothing received")
								}
								a.So(req.URL.String(), should.Equal, tc.URL)
								a.So(req.Header.Get("Authorization"), should.Equal, longAuthHeader)
								a.So(req.Header.Get("Content-Type"), should.Equal, "application/json")
								a.So(req.Header.Get("X-Downlink-Apikey"), should.Equal, "foo.secret")
								a.So(req.Header.Get("X-Downlink-Push"), should.Equal,
									"https://example.com/api/v3/as/applications/foo-app/webhooks/foo-hook/devices/foo-device/down/push",
								)
								a.So(req.Header.Get("X-Downlink-Replace"), should.Equal,
									"https://example.com/api/v3/as/applications/foo-app/webhooks/foo-hook/devices/foo-device/down/replace", //nolint:lll
								)
								a.So(req.Header.Get("X-Tts-Domain"), should.Equal, "example.com")
								actualBody, err := stdio.ReadAll(req.Body)
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
		})
	}

	//nolint:paralleltest
	t.Run("Downstream", func(t *testing.T) {
		is, isAddr, closeIS := mockis.New(ctx)
		defer closeIS()
		is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey,
			ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
			ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE)
		conf := &component.Config{
			ServiceBase: config.ServiceBase{
				GRPC: config.GRPC{
					Listen:                      ":0",
					AllowInsecureForCredentials: true,
				},
				Cluster: cluster.Config{
					IdentityServer: isAddr,
				},
			},
		}
		c := componenttest.NewComponent(t, conf)
		io := mock.NewServer(c)
		testSink := mocksink.New(nil)
		w, err := web.NewWebhooks(ctx, io, registry, testSink, downlinks)
		if err != nil {
			t.Fatalf("Unexpected error %v", err)
		}
		c.RegisterWeb(w)
		componenttest.StartComponent(t, c)
		defer c.Close()

		mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

		//nolint:paralleltest
		t.Run("Authorization", func(t *testing.T) {
			for _, tc := range []struct {
				Name       string
				ID         *ttnpb.ApplicationIdentifiers
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
					ID:         &ttnpb.ApplicationIdentifiers{ApplicationId: "--invalid-id"},
					Key:        "invalid key",
					ExpectCode: http.StatusBadRequest,
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					a := assertions.New(t)
					url := fmt.Sprintf("/api/v3/as/applications/%s/webhooks/%s/devices/%s/down/replace",
						tc.ID.ApplicationId, registeredWebhookID, registeredDeviceID.DeviceId,
					)
					body := bytes.NewReader([]byte(`{"downlinks":[]}`))
					req := httptest.NewRequest(http.MethodPost, url, body)
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Key))
					res := httptest.NewRecorder()
					c.ServeHTTP(res, req)
					a.So(res.Code, should.Equal, tc.ExpectCode)
					downlinks, err := io.DownlinkQueueList(ctx, registeredDeviceID)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					a.So(downlinks, should.Resemble, []*ttnpb.ApplicationDownlink{})
				})
			}
		})
	})

	//nolint:paralleltest
	t.Run("Paused", func(t *testing.T) {
		_, ctx := test.New(t)

		// Create an active webhook.
		_, err := registry.Set(ctx, ids, nil,
			func(_ *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
				return &ttnpb.ApplicationWebhook{
						Ids:           ids,
						BaseUrl:       "https://myapp.com/api/ttn/v3/{/appID,devID}",
						Format:        "json",
						Paused:        false,
						UplinkMessage: &ttnpb.ApplicationWebhook_Message{Path: "/"},
					},
					[]string{
						"ids.application_ids",
						"ids.webhook_id",
						"base_url",
						"format",
						"paused",
						"uplink_message",
						"field_mask",
					}, nil
			})
		if err != nil {
			t.Fatalf("Failed to set webhook in registry: %s", err)
		}

		sinkCh := make(chan *http.Request, 1)
		testSink := mocksink.New(sinkCh)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		c := componenttest.NewComponent(t, &component.Config{})
		componenttest.StartComponent(t, c)
		defer c.Close()

		as := mock.NewServer(c)
		_, err = web.NewWebhooks(ctx, as, registry, testSink, downlinks)
		if err != nil {
			t.Fatalf("Unexpected error %v", err)
		}

		message := &ttnpb.ApplicationUp{
			EndDeviceIds: registeredDeviceID,
			Up: &ttnpb.ApplicationUp_UplinkMessage{
				UplinkMessage: &ttnpb.ApplicationUplink{
					SessionKeyId: []byte{0x11},
					FPort:        42,
					FCnt:         42,
					FrmPayload:   []byte{0x1, 0x2, 0x3},
				},
			},
		}

		a := assertions.New(t)
		err = as.Publish(ctx, message)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		var req *http.Request
		select {
		case req = <-sinkCh:
		case <-time.After(timeout):
			t.Fatal("Expected message but nothing received")
		}

		actualBody, err := stdio.ReadAll(req.Body)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		expectedBody, err := formatters.JSON.FromUp(message)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(actualBody, should.Resemble, expectedBody)

		// Pause the webhook.
		_, err = registry.Set(ctx, ids, nil,
			func(_ *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
				return &ttnpb.ApplicationWebhook{
						Ids:     ids,
						BaseUrl: "https://myapp.com/api/ttn/v3/{/appID,devID}",
						Format:  "json",
						Paused:  true,
					},
					[]string{
						"ids.application_ids",
						"ids.webhook_id",
						"base_url",
						"format",
						"paused",
						"field_mask",
					}, nil
			})
		if err != nil {
			t.Fatalf("Failed to set webhook in registry: %s", err)
		}

		err = as.Publish(ctx, message)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		select {
		case req = <-sinkCh:
			t.Fatalf("Did not expect message but received: %v", req)
		case <-time.After(timeout):
			// Webhook was not received.
		}
	})
}
