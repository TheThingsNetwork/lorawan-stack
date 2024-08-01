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

package webmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func TestPeer(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	m := Peer()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	m(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		peer, ok := peer.FromContext(r.Context())
		a.So(ok, should.BeTrue)
		a.So(peer.Addr.String(), should.Equal, "192.0.2.1:1234")
	})).ServeHTTP(rec, r)
}

func TestPeerTLS(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	m := Peer()

	handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		peer, ok := peer.FromContext(r.Context())
		a.So(ok, should.BeTrue)
		a.So(peer.AuthInfo, should.HaveSameTypeAs, credentials.TLSInfo{})
	})

	server := httptest.NewTLSServer(m(handler))
	defer server.Close()

	client := server.Client()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make a GET request: %v", err)
	}
	defer resp.Body.Close()
}
