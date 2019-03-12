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

package interop

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestServeHTTP(t *testing.T) {
	for _, tc := range []struct {
		Name              string
		JS                JoinServer
		hNS               HomeNetworkServer
		sNS               ServingNetworkServer
		fNS               ForwardingNetworkServer
		AS                ApplicationServer
		RequestBody       interface{}
		ResponseAssertion func(*testing.T, *http.Response) bool
	}{
		{
			Name: "Empty",
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				return a.So(res.StatusCode, should.Equal, http.StatusBadRequest)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			s, err := New(test.Context(), config.Interop{})
			if !a.So(err, should.BeNil) {
				t.Fatal("Could not create an interop instance")
			}
			if tc.JS != nil {
				s.RegisterJS(tc.JS)
			}
			if tc.hNS != nil {
				s.RegisterHNS(tc.hNS)
			}
			if tc.sNS != nil {
				s.RegisterSNS(tc.sNS)
			}
			if tc.AS != nil {
				s.RegisterAS(tc.AS)
			}
			buf, err := json.Marshal(tc.RequestBody)
			if !a.So(err, should.BeNil) {
				t.Fatal("Failed to marshal request body")
			}
			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(buf))
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)
			if !tc.ResponseAssertion(t, rec.Result()) {
				t.FailNow()
			}
		})
	}
}
