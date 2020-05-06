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

package webmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

type logEntryChannel chan log.Entry

func (ch logEntryChannel) HandleLog(e log.Entry) error {
	ch <- e
	return nil
}

func (ch logEntryChannel) Expect(t *testing.T, f func(log.Entry)) {
	select {
	case e := <-ch:
		f(e)
	case <-time.After(time.Second):
		t.Fatal("Missing log entry")
	}
}

func TestLog(t *testing.T) {
	ch := make(logEntryChannel, 10)
	m := Log(&log.Logger{
		Handler: ch,
	})

	t.Run("Normal Request", func(t *testing.T) {
		a := assertions.New(t)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.So(log.FromContext(r.Context()), should.NotEqual, log.Noop)
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, r)
		ch.Expect(t, func(e log.Entry) {
			a.So(e.Level(), should.Equal, log.InfoLevel)
			a.So(e.Message(), should.Equal, "Request handled")
			fields := e.Fields().Fields()
			for _, key := range []string{"method", "url", "remote_addr", "request_id", "status", "duration", "response_size"} {
				a.So(fields, should.ContainKey, key)
			}
			a.So(fields["method"], should.Equal, http.MethodGet)
			a.So(fields["url"], should.Equal, "/")
			a.So(fields["status"], should.Equal, http.StatusOK)
			a.So(fields["response_size"], should.Equal, 0)
		})
	})

	t.Run("Client Error", func(t *testing.T) {
		a := assertions.New(t)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		})).ServeHTTP(rec, r)
		ch.Expect(t, func(e log.Entry) {
			a.So(e.Level(), should.Equal, log.InfoLevel)
			a.So(e.Message(), should.Equal, "Client error")
		})
	})

	t.Run("Server Error", func(t *testing.T) {
		a := assertions.New(t)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})).ServeHTTP(rec, r)
		ch.Expect(t, func(e log.Entry) {
			a.So(e.Level(), should.Equal, log.ErrorLevel)
			a.So(e.Message(), should.Equal, "Server error")
		})
	})

	t.Run("Rich Error", func(t *testing.T) {
		a := assertions.New(t)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			webhandlers.NotFound(w, r)
		})).ServeHTTP(rec, r)
		ch.Expect(t, func(e log.Entry) {
			fields := e.Fields().Fields()
			a.So(fields, should.ContainKey, "error")
		})
	})
}
