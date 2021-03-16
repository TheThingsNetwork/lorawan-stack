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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

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
				Body: ioutil.NopCloser(bytes.NewBufferString("")),
			}
			errChan <- nil

			resp, err := cl.Do(ctx, "foo", "bar", http.MethodGet, nil)
			req := <-reqChan
			a.So(resp, should.NotBeNil)
			a.So(err, should.BeNil)
			a.So(req.Header, should.NotContainKey, "Ocp-Apim-Subscription-Key")
		})
}

func TestAuth(t *testing.T) {
	withClient(test.Context(), t, []api.Option{api.WithToken("foobar")},
		func(ctx context.Context, t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			a := assertions.New(t)

			respChan <- &http.Response{
				Body: ioutil.NopCloser(bytes.NewBufferString("")),
			}
			errChan <- nil

			resp, err := cl.Do(ctx, "foo", "bar", http.MethodGet, nil)
			req := <-reqChan
			a.So(resp, should.NotBeNil)
			a.So(err, should.BeNil)
			if a.So(req.Header, should.ContainKey, "Ocp-Apim-Subscription-Key") {
				a.So(req.Header["Ocp-Apim-Subscription-Key"], should.Resemble, []string{"foobar"})
			}
		})
}

var (
	singleFrameRequest = api.BuildSingleFrameRequest([]*ttnpb.RxMetadata{
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "gtw1",
			},
			Location: &ttnpb.Location{
				Latitude:  123.4,
				Longitude: 234.5,
				Altitude:  345,
			},
			RSSI: 567.8,
			SNR:  678.9,
		},
		{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "gtw2",
			},
			Location: &ttnpb.Location{
				Latitude:  234.5,
				Longitude: 345.6,
				Altitude:  456,
			},
			FineTimestamp: 890,
			RSSI:          678.9,
			SNR:           789.1,
		},
	})

	singleFrameResponse = api.SingleFrameResponse{
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
)

func TestSingleFrame(t *testing.T) {
	withClient(test.Context(), t, nil,
		func(ctx context.Context, t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			a := assertions.New(t)

			b := bytes.NewBuffer(nil)
			a.So(json.NewEncoder(b).Encode(singleFrameResponse), should.BeNil)

			respChan <- &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(b),
			}
			errChan <- nil

			resp, err := cl.SolveSingleFrame(ctx, singleFrameRequest)
			req := <-reqChan
			if a.So(err, should.BeNil) {
				a.So(resp.SingleFrameResponse, should.Resemble, singleFrameResponse)
			}

			if a.So(req, should.NotBeNil) {
				request := &api.SingleFrameRequest{}
				a.So(json.NewDecoder(req.Body).Decode(request), should.BeNil)
				a.So(request, should.Resemble, singleFrameRequest)
			}
		})
}

func float64Ptr(f float64) *float64 {
	return &f
}
