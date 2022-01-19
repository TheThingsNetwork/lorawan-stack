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

package interop_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockComponent struct {
	ctx context.Context
}

func (c *mockComponent) Context() context.Context {
	return c.ctx
}

func (c *mockComponent) RateLimiter() ratelimit.Interface {
	return &ratelimit.NoopRateLimiter{}
}

func (c *mockComponent) HTTPClient(context.Context, ...httpclient.Option) (*http.Client, error) {
	return http.DefaultClient, nil
}

type mockTarget struct {
	JoinRequestFunc    func(context.Context, *interop.JoinReq) (*interop.JoinAns, error)
	AppSKeyRequestFunc func(context.Context, *interop.AppSKeyReq) (*interop.AppSKeyAns, error)
	HomeNSRequestFunc  func(context.Context, *interop.HomeNSReq) (*interop.TTIHomeNSAns, error)
}

func (m mockTarget) JoinRequest(ctx context.Context, req *interop.JoinReq) (*interop.JoinAns, error) {
	if m.JoinRequestFunc != nil {
		return m.JoinRequestFunc(ctx, req)
	}
	panic("JoinRequest called but not registered")
}

func (m mockTarget) AppSKeyRequest(ctx context.Context, req *interop.AppSKeyReq) (*interop.AppSKeyAns, error) {
	if m.AppSKeyRequestFunc != nil {
		return m.AppSKeyRequestFunc(ctx, req)
	}
	panic("AppSKeyRequest called but not registered")
}

func (m mockTarget) HomeNSRequest(ctx context.Context, req *interop.HomeNSReq) (*interop.TTIHomeNSAns, error) {
	if m.HomeNSRequestFunc != nil {
		return m.HomeNSRequestFunc(ctx, req)
	}
	panic("HomeNSRequest called but not registered")
}

func TestServer(t *testing.T) {
	authorizer := interop.Authorizer{}

	for _, tc := range []struct {
		Name              string
		JS                interop.JoinServer
		ClientTLSConfig   *tls.Config
		PacketBrokerToken bool
		RequestBody       interface{}
		ResponseAssertion func(*testing.T, *http.Response) bool
	}{
		{
			Name:            "ClientTLS/Empty",
			ClientTLSConfig: makeClientTLSConfig(),
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				return a.So(res.StatusCode, should.Equal, http.StatusBadRequest)
			},
		},
		{
			Name:            "ClientTLS/JoinReq/InvalidSenderID",
			ClientTLSConfig: makeClientTLSConfig(),
			RequestBody: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						MessageType:     interop.MessageTypeJoinReq,
						ProtocolVersion: interop.ProtocolV1_0,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x02},
					ReceiverID: interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_0_3),
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				if !a.So(res.StatusCode, should.Equal, http.StatusOK) {
					return false
				}
				var msg interop.ErrorMessage
				err := json.NewDecoder(res.Body).Decode(&msg)
				if !a.So(err, should.BeNil) {
					return false
				}
				return a.So(msg.Result.ResultCode, should.Equal, interop.ResultUnknownSender)
			},
		},
		{
			Name:            "ClientTLS/JoinReq/NotRegistered",
			ClientTLSConfig: makeClientTLSConfig(),
			RequestBody: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						MessageType:     interop.MessageTypeJoinReq,
						ProtocolVersion: interop.ProtocolV1_0,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x01},
					ReceiverID: interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_0_3),
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				if !a.So(res.StatusCode, should.Equal, http.StatusOK) {
					return false
				}
				var msg interop.ErrorMessage
				err := json.NewDecoder(res.Body).Decode(&msg)
				if !a.So(err, should.BeNil) {
					return false
				}
				return a.So(msg.Result.ResultCode, should.Equal, interop.ResultMalformedRequest)
			},
		},
		{
			Name: "ClientTLS/JoinReq/UnknownDevEUI",
			JS: mockTarget{
				JoinRequestFunc: func(ctx context.Context, req *interop.JoinReq) (*interop.JoinAns, error) {
					if err := authorizer.RequireAuthorized(ctx); err != nil {
						return nil, err
					}
					return nil, interop.ErrUnknownDevEUI.New()
				},
			},
			ClientTLSConfig: makeClientTLSConfig(),
			RequestBody: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						MessageType:     interop.MessageTypeJoinReq,
						ProtocolVersion: interop.ProtocolV1_0,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x01},
					ReceiverID: interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_0_3),
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				if !a.So(res.StatusCode, should.Equal, http.StatusOK) {
					return false
				}
				var msg interop.ErrorMessage
				err := json.NewDecoder(res.Body).Decode(&msg)
				if !a.So(err, should.BeNil) {
					return false
				}
				return a.So(msg.Result.ResultCode, should.Equal, interop.ResultUnknownDevEUI)
			},
		},
		{
			Name: "ClientTLS/JoinReq/Success",
			JS: mockTarget{
				JoinRequestFunc: func(ctx context.Context, req *interop.JoinReq) (*interop.JoinAns, error) {
					if err := authorizer.RequireAuthorized(ctx); err != nil {
						return nil, err
					}
					if err := authorizer.RequireNetID(ctx, types.NetID{0x0, 0x0, 0x1}); err != nil {
						return nil, err
					}
					if err := authorizer.RequireAddress(ctx, "localhost:4242"); err != nil {
						return nil, err
					}
					if !types.EUI64(req.DevEUI).Equal(types.EUI64{0x42, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) {
						return nil, interop.ErrUnknownDevEUI.New()
					}
					return &interop.JoinAns{
						JsNsMessageHeader: interop.JsNsMessageHeader{
							MessageHeader: interop.MessageHeader{
								ProtocolVersion: req.ProtocolVersion,
								MessageType:     interop.MessageTypeJoinAns,
								TransactionID:   req.TransactionID,
							},
							SenderID:   req.ReceiverID,
							ReceiverID: req.SenderID,
						},
						PHYPayload:   bytes.Repeat([]byte{0x42}, 42),
						FNwkSIntKey:  (*interop.KeyEnvelope)(test.DefaultFNwkSIntKeyEnvelopeWrapped),
						AppSKey:      (*interop.KeyEnvelope)(test.DefaultAppSKeyEnvelopeWrapped),
						SessionKeyID: bytes.Repeat([]byte{0x42}, 6),
						Result: interop.Result{
							ResultCode: interop.ResultSuccess,
						},
					}, nil
				},
			},
			ClientTLSConfig: makeClientTLSConfig(),
			RequestBody: &interop.JoinReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						MessageType:     interop.MessageTypeJoinReq,
						ProtocolVersion: interop.ProtocolV1_0,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x01},
					ReceiverID: interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				MACVersion: interop.MACVersion(ttnpb.MACVersion_MAC_V1_0_3),
				DevEUI:     interop.EUI64{0x42, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				if !a.So(res.StatusCode, should.Equal, http.StatusOK) {
					return false
				}
				var msg interop.JoinAns
				err := json.NewDecoder(res.Body).Decode(&msg)
				return a.So(err, should.BeNil) &&
					a.So(msg.Result.ResultCode, should.Equal, interop.ResultSuccess) &&
					a.So(msg.SenderID, should.Resemble, interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) &&
					a.So(msg.ReceiverID, should.Resemble, interop.NetID{0x0, 0x0, 0x01})
			},
		},
		{
			Name: "PacketBroker/HomeNSReq/Success",
			JS: mockTarget{
				HomeNSRequestFunc: func(ctx context.Context, req *interop.HomeNSReq) (*interop.TTIHomeNSAns, error) {
					if err := authorizer.RequireAuthorized(ctx); err != nil {
						return nil, err
					}
					if err := authorizer.RequireNetID(ctx, types.NetID{0x0, 0x0, 0x0}); err != nil {
						return nil, err
					}
					if err := authorizer.RequireAddress(ctx, "test.packetbroker.io"); err != nil {
						return nil, err
					}
					return &interop.TTIHomeNSAns{
						HomeNSAns: interop.HomeNSAns{
							JsNsMessageHeader: interop.JsNsMessageHeader{
								MessageHeader: interop.MessageHeader{
									ProtocolVersion: req.ProtocolVersion,
									MessageType:     interop.MessageTypeHomeNSAns,
									TransactionID:   req.TransactionID,
								},
								SenderID:     req.ReceiverID,
								ReceiverID:   req.SenderID,
								ReceiverNSID: req.SenderNSID,
							},
							Result: interop.Result{
								ResultCode: interop.ResultSuccess,
							},
							HNetID: interop.NetID{0x0, 0x0, 0x1},
							HNSID:  &interop.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0},
						},
					}, nil
				},
			},
			PacketBrokerToken: true,
			RequestBody: &interop.HomeNSReq{
				NsJsMessageHeader: interop.NsJsMessageHeader{
					MessageHeader: interop.MessageHeader{
						MessageType:     interop.MessageTypeHomeNSReq,
						ProtocolVersion: interop.ProtocolV1_1,
					},
					SenderID:   interop.NetID{0x0, 0x0, 0x0},
					ReceiverID: interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				DevEUI: interop.EUI64{0x42, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				if !a.So(res.StatusCode, should.Equal, http.StatusOK) {
					return false
				}
				var msg interop.HomeNSAns
				err := json.NewDecoder(res.Body).Decode(&msg)
				return a.So(err, should.BeNil) &&
					a.So(msg.Result.ResultCode, should.Equal, interop.ResultSuccess) &&
					a.So(msg.SenderID, should.Resemble, interop.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) &&
					a.So(msg.ReceiverID, should.Resemble, interop.NetID{0x0, 0x0, 0x0}) &&
					a.So(msg.HNetID, should.Resemble, interop.NetID{0x0, 0x0, 0x1}) &&
					a.So(msg.HNSID, should.Resemble, &interop.EUI64{0x42, 0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0})
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				pbIssuer, pbToken := makePacketBrokerTokenIssuer(ctx, "test.packetbroker.io")

				s, err := interop.NewServer(&mockComponent{ctx}, nil, config.InteropServer{
					SenderClientCA: config.SenderClientCA{
						Source:    "directory",
						Directory: "testdata",
					},
					PacketBroker: config.PacketBrokerInteropAuth{
						Enabled:     true,
						TokenIssuer: pbIssuer, // Subject is the NSID and is used as authorized address.
					},
					PublicTLSAddress: "https://localhost", // Used as Packet Broker token audience.
				})
				if !a.So(err, should.BeNil) {
					t.Fatal("Failed to instantiate interop server")
				}
				if tc.JS != nil {
					s.RegisterJS(tc.JS)
				}

				srv := newTLSServer(s)
				defer srv.Close()

				buf, err := json.Marshal(tc.RequestBody)
				if !a.So(err, should.BeNil) {
					t.Fatal("Failed to marshal request body")
				}
				req := test.Must(http.NewRequest(http.MethodPost, srv.URL, bytes.NewReader(buf))).(*http.Request)
				req.Header.Set("Content-Type", "application/json")

				client := srv.Client()
				if tc.ClientTLSConfig != nil {
					client.Transport.(*http.Transport).TLSClientConfig = tc.ClientTLSConfig
				}
				if tc.PacketBrokerToken {
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbToken("https://localhost")))
				}

				res, err := client.Do(req)
				if !a.So(err, should.BeNil) {
					t.Fatal("Request failed")
				}
				if !tc.ResponseAssertion(t, res) {
					t.FailNow()
				}
			},
		})
	}
}
