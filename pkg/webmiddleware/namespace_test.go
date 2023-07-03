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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/log/handler/memory"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestNamespace(t *testing.T) {
	a := assertions.New(t)

	mem := memory.New()
	logger := log.NewLogger(mem)

	ctx := log.NewContext(test.Context(), logger)

	m := Namespace("test")
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.FromContext(r.Context()).Info("foobar")
		entry := mem.Entries[0]
		a.So(entry.Message(), should.Equal, "foobar")
		fields := entry.Fields().Fields()
		a.So(fields, should.HaveLength, 1)
		a.So(fields["namespace"], should.Equal, "test")
	})).ServeHTTP(rec, req)
}
