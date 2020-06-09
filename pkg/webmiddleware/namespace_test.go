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

package webmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

type mockLogger struct {
	fields map[string]interface{}
}

func (l *mockLogger) Debug(msg string) {
}

func (l *mockLogger) Info(msg string) {
}

func (l *mockLogger) Warn(msg string) {
}

func (l *mockLogger) Error(msg string) {
}

func (l *mockLogger) Fatal(msg string) {
}

func (l *mockLogger) Debugf(msg string, v ...interface{}) {
}

func (l *mockLogger) Infof(msg string, v ...interface{}) {
}

func (l *mockLogger) Warnf(msg string, v ...interface{}) {
}

func (l *mockLogger) Errorf(msg string, v ...interface{}) {
}

func (l *mockLogger) Fatalf(msg string, v ...interface{}) {
}

func (l *mockLogger) WithField(k string, v interface{}) log.Interface {
	l.fields[k] = v
	return l
}

func (l *mockLogger) WithFields(kv log.Fielder) log.Interface {
	for k, v := range kv.Fields() {
		l.fields[k] = v
	}
	return l
}

func (l *mockLogger) WithError(_ error) log.Interface {
	return l
}

func TestNamespace(t *testing.T) {
	a := assertions.New(t)

	logger := &mockLogger{
		fields: make(map[string]interface{}),
	}
	ctx := log.NewContext(test.Context(), logger)

	m := Namespace("test")
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, ok := log.FromContext(r.Context()).(*mockLogger)
		if !ok {
			t.Fatal("Unexpected logger type")
		}
		a.So(logger.fields, should.HaveLength, 1)
		a.So(logger.fields["namespace"], should.Equal, "test")
	})).ServeHTTP(rec, req)
}
