// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package webmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

type logEntryChannel chan log.Entry

func (ch logEntryChannel) HandleLog(e log.Entry) error {
	ch <- e
	return nil
}

func (ch logEntryChannel) Expect(t *testing.T, f func(*testing.T, log.Entry)) {
	select {
	case e := <-ch:
		f(t, e)
	case <-time.After(time.Second):
		t.Fatal("Missing log entry")
	}
}

func (ch logEntryChannel) ExpectNoLog(t *testing.T) {
	select {
	case <-ch:
		t.Fatal("Expected no log entry but received one")
	case <-time.After(100 * time.Millisecond):
	}
}

func verifySuccess(t *testing.T, e log.Entry) {
	a := assertions.New(t)
	a.So(e.Level(), should.Equal, log.InfoLevel)
	a.So(e.Message(), should.Equal, "Request handled")
	fields := e.Fields().Fields()
	for _, key := range []string{"http.method", "http.path", "peer.address", "request_id", "http.status", "duration"} {
		a.So(fields, should.ContainKey, key)
	}
	a.So(fields["http.method"], should.Equal, http.MethodGet)
	a.So(fields["http.path"], should.Equal, "/")
	a.So(fields["http.status"], should.Equal, http.StatusOK)
}

func verifyClientError(t *testing.T, e log.Entry) {
	a := assertions.New(t)
	a.So(e.Level(), should.Equal, log.InfoLevel)
	a.So(e.Message(), should.Equal, "Client error")
}

func verifyServerError(t *testing.T, e log.Entry) {
	a := assertions.New(t)
	a.So(e.Level(), should.Equal, log.ErrorLevel)
	a.So(e.Message(), should.Equal, "Server error")
}

func TestLog(t *testing.T) {
	a := assertions.New(t)
	ch := make(logEntryChannel, 10)
	m := Log(&log.Logger{
		Handler: ch,
	}, []string{"/path", "/ignorepath"})

	for _, tc := range []struct {
		name     string
		path     string
		setReq   func(r *http.Request)
		validate func(t *testing.T)
		handler  func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "NormalSuccess",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			path: "/",
			validate: func(t *testing.T) {
				ch.Expect(t, func(t *testing.T, e log.Entry) {
					verifySuccess(t, e)
					assertions.New(t).So(e.Fields().Fields()["peer.address"], should.Equal, "192.0.2.1:1234")
				})
			},
		},
		{
			name: "RealIP",
			setReq: func(r *http.Request) {
				r.Header.Set("X-Real-Ip", "12.34.56.78")
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			path: "/",
			validate: func(t *testing.T) {
				ch.Expect(t, func(t *testing.T, e log.Entry) {
					verifySuccess(t, e)
					assertions.New(t).So(e.Fields().Fields()["peer.real_ip"], should.Equal, "12.34.56.78")
				})
			},
		},
		{
			name: "IgnoreSuccess",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			path:     "/path",
			validate: ch.ExpectNoLog,
		},
		{
			name: "IgnoreSuccess",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			path:     "/ignorepath",
			validate: ch.ExpectNoLog,
		},
		{
			name: "ClientError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			path: "/",
			validate: func(t *testing.T) {
				ch.Expect(t, verifyClientError)
			},
		},
		{
			name: "ClientErrorWithIgnore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			path: "/path",
			validate: func(t *testing.T) {
				ch.Expect(t, verifyClientError)
			},
		},
		{
			name: "ServerError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			path: "/",
			validate: func(t *testing.T) {
				ch.Expect(t, verifyServerError)
			},
		},
		{
			name: "ServerErrorWithIgnore",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			path: "/path",
			validate: func(t *testing.T) {
				ch.Expect(t, verifyServerError)
			},
		},
		{
			name: "RichError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				webhandlers.NotFound(w, r)
			},
			path: "/path",
			validate: func(t *testing.T) {
				ch.Expect(t, func(t *testing.T, e log.Entry) {
					a := assertions.New(t)
					fields := e.Fields().Fields()
					a.So(fields, should.ContainKey, "error")
				})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.setReq != nil {
				tc.setReq(r)
			}
			rec := httptest.NewRecorder()
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				a.So(log.FromContext(r.Context()), should.NotEqual, log.Noop)
				tc.handler(w, r)
			})).ServeHTTP(rec, r)
			tc.validate(t)
		})
	}
}
