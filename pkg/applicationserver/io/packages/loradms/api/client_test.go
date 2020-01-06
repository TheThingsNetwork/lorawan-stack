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
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/loradms/v1/api"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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

func withClient(t *testing.T, opts []api.Option, f func(*testing.T, <-chan *http.Request, chan<- *http.Response, chan<- error, *api.Client)) {
	reqChan := make(chan *http.Request, 5)
	respChan := make(chan *http.Response, 5)
	errChan := make(chan error, 5)
	cl, err := api.New(&http.Client{
		Transport: chanRoundTrip(reqChan, respChan, errChan),
	}, opts...)
	if !assertions.New(t).So(err, should.BeNil) {
		t.FailNow()
	}
	f(t, reqChan, respChan, errChan, cl)
}

func TestNoAuth(t *testing.T) {
	withClient(t, nil,
		func(t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			a := assertions.New(t)

			respChan <- &http.Response{
				Body: ioutil.NopCloser(bytes.NewBufferString("")),
			}
			errChan <- nil

			resp, err := cl.Do("foo", "bar", "baz", http.MethodGet, nil)
			req := <-reqChan
			a.So(resp, should.NotBeNil)
			a.So(err, should.BeNil)
			a.So(req.Header, should.NotContainKey, "Authorization")
		})
}

func TestAuth(t *testing.T) {
	withClient(t, []api.Option{api.WithToken("foobar")},
		func(t *testing.T, reqChan <-chan *http.Request, respChan chan<- *http.Response, errChan chan<- error, cl *api.Client) {
			a := assertions.New(t)

			respChan <- &http.Response{
				Body: ioutil.NopCloser(bytes.NewBufferString("")),
			}
			errChan <- nil

			resp, err := cl.Do("foo", "bar", "baz", http.MethodGet, nil)
			req := <-reqChan
			a.So(resp, should.NotBeNil)
			a.So(err, should.BeNil)
			a.So(req.Header, should.ContainKey, "Authorization")
			a.So(req.Header["Authorization"], should.Resemble, []string{"foobar"})
		})
}
