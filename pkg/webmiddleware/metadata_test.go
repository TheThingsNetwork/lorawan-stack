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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"google.golang.org/grpc/metadata"
)

func TestMetadata(t *testing.T) {
	a := assertions.New(t)

	m := Metadata("authorization", "X-Request-Id")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer token")
	r.Header.Set("X-Request-Id", "XXX")

	rec := httptest.NewRecorder()
	m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md, _ := metadata.FromIncomingContext(r.Context())
		a.So(md.Get("authorization"), should.Resemble, []string{"Bearer token"})
		a.So(md.Get("x-request-id"), should.Resemble, []string{"XXX"})
	})).ServeHTTP(rec, r)
}
