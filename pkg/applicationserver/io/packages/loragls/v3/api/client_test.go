// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

const authHeader = "Authorization"

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func chanRoundTrip(reqChan chan<- *http.Request, respChan <-chan *http.Response, errChan <-chan error) http.RoundTripper {
	return roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			reqChan <- req
			return <-respChan, <-errChan
		},
	)
}

func withClient(ctx context.Context, t *testing.T, opts []api.Option, f func(context.Context, *testing.T, <-chan *http.Request, chan<- *http.Response, chan<- error, *api.Client)) {
	reqChan := make(chan *http.Request, 5)
	respChan := make(chan *http.Response, 5)
	errChan := make(chan error, 5)
	cl, err := api.New(&http.Client{
		Transport: chanRoundTrip(reqChan, respChan, errChan),
	}, opts...)
	if !assertions.New(t).So(err, should.BeNil) {
		t.FailNow()
	}
	f(ctx, t, reqChan, respChan, errChan, cl)
}

func TestNoAuth(t *testing.T) {
	withClient(test.Context(), t, nil,
		func(ctx context.Context, t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			a := assertions.New(t)

			respChan <- &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("")),
			}
			errChan <- nil

			resp, err := cl.Do(ctx, "v3", "foo", "bar", http.MethodGet, nil)
			req := <-reqChan
			a.So(resp, should.NotBeNil)
			a.So(err, should.BeNil)
			a.So(req.Header, should.NotContainKey, authHeader)
		})
}

func TestAuth(t *testing.T) {
	withClient(test.Context(), t, []api.Option{api.WithToken("foobar")},
		func(ctx context.Context, t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			a := assertions.New(t)

			respChan <- &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("")),
			}
			errChan <- nil

			resp, err := cl.Do(ctx, "v3", "foo", "bar", http.MethodGet, nil)
			req := <-reqChan
			a.So(resp, should.NotBeNil)
			a.So(err, should.BeNil)
			if a.So(req.Header, should.ContainKey, authHeader) {
				a.So(req.Header[authHeader], should.Resemble, []string{"foobar"})
			}
		})
}

var (
	singleFrameRequest = api.BuildSingleFrameRequest(test.Context(), []*ttnpb.RxMetadata{
		{
			GatewayIds: &ttnpb.GatewayIdentifiers{
				GatewayId: "gtw1",
			},
			Location: &ttnpb.Location{
				Latitude:  123.4,
				Longitude: 234.5,
				Altitude:  345,
			},
			Rssi: 567.8,
			Snr:  678.9,
		},
		{
			GatewayIds: &ttnpb.GatewayIdentifiers{
				GatewayId: "gtw2",
			},
			Location: &ttnpb.Location{
				Latitude:  234.5,
				Longitude: 345.6,
				Altitude:  456,
			},
			FineTimestamp: 890,
			Rssi:          678.9,
			Snr:           789.1,
		},
	})
	singleFrameResponse = api.LocationSolverResponse{
		Result: &api.LocationSolverResult{
			UsedGateways: 2,
			HDOP:         float64Ptr(123.4),
			Algorithm:    api.Algorithm_RSSITDOA,
			Location: api.Location{
				Latitude:  123.4,
				Longitude: 456.7,
				Tolerance: 123,
			},
		},
	}

	multiFrameRequest = api.BuildMultiFrameRequest(test.Context(), [][]*ttnpb.RxMetadata{
		{
			{
				GatewayIds: &ttnpb.GatewayIdentifiers{
					GatewayId: "gtw1",
				},
				Location: &ttnpb.Location{
					Latitude:  123.4,
					Longitude: 234.5,
					Altitude:  345,
				},
				Rssi: 567.8,
				Snr:  678.9,
			},
			{
				GatewayIds: &ttnpb.GatewayIdentifiers{
					GatewayId: "gtw2",
				},
				Location: &ttnpb.Location{
					Latitude:  234.5,
					Longitude: 345.6,
					Altitude:  456,
				},
				FineTimestamp: 890,
				Rssi:          678.9,
				Snr:           789.1,
			},
		},
		{
			{
				GatewayIds: &ttnpb.GatewayIdentifiers{
					GatewayId: "gtw1",
				},
				Location: &ttnpb.Location{
					Latitude:  123.4,
					Longitude: 234.5,
					Altitude:  345,
				},
				Rssi: 890.1,
				Snr:  910.1,
			},
			{
				GatewayIds: &ttnpb.GatewayIdentifiers{
					GatewayId: "gtw2",
				},
				Location: &ttnpb.Location{
					Latitude:  234.5,
					Longitude: 345.6,
					Altitude:  456,
				},
				FineTimestamp: 910,
				Rssi:          789.1,
				Snr:           890.1,
			},
		},
	})
	multiFrameResponse = api.LocationSolverResponse{
		Result: &api.LocationSolverResult{
			UsedGateways: 2,
			HDOP:         float64Ptr(345.6),
			Algorithm:    api.Algorithm_RSSITDOA,
			Location: api.Location{
				Latitude:  234.5,
				Longitude: 678.9,
				Tolerance: 234,
			},
		},
	}

	gnssRequest = &api.GNSSRequest{
		Payload: []byte{0x01, 0x02, 0x03},
	}
	gnssResponse = api.GNSSLocationSolverResponse{
		Result: &api.GNSSLocationSolverResult{
			LLH:      []float64{123.4, 456.8, 567.9},
			Accuracy: 678.8,
		},
	}

	wifiRequest = &api.WiFiRequest{
		LoRaWAN: []api.TDOAUplink{
			{
				GatewayID: "gtw1",
				RSSI:      123.4,
				SNR:       234.5,
				TDOA:      234,
				AntennaLocation: api.AntennaLocation{
					Latitude:  234.5,
					Longitude: 678.9,
					Altitude:  345.7,
				},
			},
			{
				GatewayID: "gtw2",
				RSSI:      234.5,
				SNR:       345.7,
				TDOA:      456,
				AntennaLocation: api.AntennaLocation{
					Latitude:  345.5,
					Longitude: 567.9,
					Altitude:  789.7,
				},
			},
		},
		WiFiAccessPoints: []api.AccessPoint{
			{
				MACAddress:     "00:02:01:53:8B:50",
				SignalStrength: -20,
			},
			{
				MACAddress:     "00:02:13:44:55:20",
				SignalStrength: -80,
			},
		},
	}
	wifiResponse = api.WiFiLocationSolverResponse{
		Result: &api.WiFiLocationSolverResult{
			Latitude:         234.5,
			Longitude:        678.9,
			Altitude:         123.4,
			Accuracy:         23,
			Algorithm:        "Wifi",
			GatewaysReceived: 2,
			GatewaysUsed:     2,
		},
	}
)

func TestClient(t *testing.T) {
	withClient(test.Context(), t, nil,
		func(ctx context.Context, t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			for _, tc := range []struct {
				name          string
				request       any
				response      any
				do            func(ctx context.Context, a *assertions.Assertion)
				assertRequest func(t *testing.T, a *assertions.Assertion, req *http.Request)
			}{
				{
					name:     "SingleFrameRequest",
					request:  singleFrameRequest,
					response: singleFrameResponse,
					do: func(ctx context.Context, a *assertions.Assertion) {
						resp, err := cl.SolveSingleFrame(ctx, singleFrameRequest)
						if a.So(err, should.BeNil) {
							a.So(resp.LocationSolverResponse, should.Resemble, singleFrameResponse)
						}
					},
					assertRequest: func(t *testing.T, a *assertions.Assertion, req *http.Request) {
						a.So(req.URL.Path, should.Equal, "/api/v1/solve/singleframe")

						request := &api.SingleFrameRequest{}
						a.So(json.NewDecoder(req.Body).Decode(request), should.BeNil)
						a.So(request, should.Resemble, singleFrameRequest)
					},
				},
				{
					name:     "MultiFrameRequest",
					request:  multiFrameRequest,
					response: multiFrameResponse,
					do: func(ctx context.Context, a *assertions.Assertion) {
						resp, err := cl.SolveMultiFrame(ctx, multiFrameRequest)
						if a.So(err, should.BeNil) {
							a.So(resp.LocationSolverResponse, should.Resemble, multiFrameResponse)
						}
					},
					assertRequest: func(t *testing.T, a *assertions.Assertion, req *http.Request) {
						a.So(req.URL.Path, should.Equal, "/api/v1/solve/multiframe")

						request := &api.MultiFrameRequest{}
						a.So(json.NewDecoder(req.Body).Decode(request), should.BeNil)
						a.So(request, should.Resemble, multiFrameRequest)
					},
				},
				{
					name:     "GNSSRequest",
					request:  gnssRequest,
					response: gnssResponse,
					do: func(ctx context.Context, a *assertions.Assertion) {
						resp, err := cl.SolveGNSS(ctx, gnssRequest)
						if a.So(err, should.BeNil) {
							a.So(resp.GNSSLocationSolverResponse, should.Resemble, gnssResponse)
						}
					},
					assertRequest: func(t *testing.T, a *assertions.Assertion, req *http.Request) {
						a.So(req.URL.Path, should.Equal, "/api/v1/solve/gnss_lr1110_singleframe")

						request := &api.GNSSRequest{}
						a.So(json.NewDecoder(req.Body).Decode(request), should.BeNil)
						a.So(request, should.Resemble, gnssRequest)
					},
				},
				{
					name:     "WiFiRequest",
					request:  wifiRequest,
					response: wifiResponse,
					do: func(ctx context.Context, a *assertions.Assertion) {
						resp, err := cl.SolveWiFi(ctx, wifiRequest)
						if a.So(err, should.BeNil) {
							a.So(resp.WiFiLocationSolverResponse, should.Resemble, wifiResponse)
						}
					},
					assertRequest: func(t *testing.T, a *assertions.Assertion, req *http.Request) {
						a.So(req.URL.Path, should.Equal, "/api/v1/solve/loraWifi")

						request := &api.WiFiRequest{}
						a.So(json.NewDecoder(req.Body).Decode(request), should.BeNil)
						a.So(request, should.Resemble, wifiRequest)
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					a := assertions.New(t)

					b := bytes.NewBuffer(nil)
					a.So(json.NewEncoder(b).Encode(tc.response), should.BeNil)

					respChan <- &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(b),
					}
					errChan <- nil

					tc.do(ctx, a)

					req := <-reqChan
					if a.So(req, should.NotBeNil) {
						tc.assertRequest(t, a, req)
					}
				})
			}
		})
}

func float64Ptr(f float64) *float64 {
	return &f
}
