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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
		{
			Name: "JoinReq/InvalidSenderID",
			RequestBody: &JoinReq{
				NsJsMessageHeader: NsJsMessageHeader{
					MessageHeader: MessageHeader{
						MessageType:     MessageTypeJoinReq,
						ProtocolVersion: "1.1",
					},
					SenderID:   NetID{0x0, 0x0, 0x02},
					ReceiverID: EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				MACVersion: MACVersion(ttnpb.MAC_V1_0_3),
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				if !a.So(res.StatusCode, should.Equal, http.StatusBadRequest) {
					return false
				}
				var msg ErrorMessage
				err := json.NewDecoder(res.Body).Decode(&msg)
				return a.So(err, should.BeNil) && a.So(msg.Result, should.Equal, ResultUnknownSender)
			},
		},
		{
			Name: "JoinReq/NotRegistered",
			RequestBody: &JoinReq{
				NsJsMessageHeader: NsJsMessageHeader{
					MessageHeader: MessageHeader{
						MessageType:     MessageTypeJoinReq,
						ProtocolVersion: "1.1",
					},
					SenderID:   NetID{0x0, 0x0, 0x01},
					ReceiverID: EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				MACVersion: MACVersion(ttnpb.MAC_V1_0_3),
			},
			ResponseAssertion: func(t *testing.T, res *http.Response) bool {
				a := assertions.New(t)
				return a.So(res.StatusCode, should.Equal, http.StatusNotFound)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			s, err := New(test.Context(), config.Interop{
				SenderClientCAs: map[string]string{
					"000001": "testdata/clientca.pem",
				},
			})
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

			clientRootCAsPem, err := ioutil.ReadFile("testdata/clientca.pem")
			if err != nil {
				t.Fatalf("Failed to read CA: %v", err)
			}
			clientRootCAs := x509.NewCertPool()
			clientRootCAs.AppendCertsFromPEM(clientRootCAsPem)
			srv := httptest.NewUnstartedServer(s)
			srv.TLS = &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
				ClientCAs:  clientRootCAs,
			}
			srv.StartTLS()
			defer srv.Close()

			clientCert, err := tls.LoadX509KeyPair("testdata/clientcert.pem", "testdata/clientkey.pem")
			if !a.So(err, should.BeNil) {
				t.Fatal("Failed to load client certificate")
			}
			client := srv.Client()
			client.Transport.(*http.Transport).TLSClientConfig.Certificates = []tls.Certificate{clientCert}

			buf, err := json.Marshal(tc.RequestBody)
			if !a.So(err, should.BeNil) {
				t.Fatal("Failed to marshal request body")
			}
			res, err := client.Post(srv.URL, "application/json", bytes.NewReader(buf))
			if !a.So(err, should.BeNil) {
				t.Fatal("Request failed")
			}
			if !tc.ResponseAssertion(t, res) {
				t.FailNow()
			}
		})
	}
}
