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

package component_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type mockInterop struct {
}

func (m mockInterop) RegisterInterop(s *interop.Server) {
	s.RegisterJS(m)
}

func (m mockInterop) JoinRequest(ctx context.Context, req *interop.JoinReq) (*interop.JoinAns, error) {
	ansHeader, err := req.AnswerHeader()
	if err != nil {
		return nil, err
	}
	return &interop.JoinAns{
		JsNsMessageHeader: ansHeader,
		Result:            interop.ResultSuccess,
	}, nil
}

func TestInteropTLS(t *testing.T) {
	a := assertions.New(t)

	config := &component.Config{
		ServiceBase: config.ServiceBase{
			TLS: config.TLS{
				RootCA:      "testdata/serverca.pem",
				Certificate: "testdata/servercert.pem",
				Key:         "testdata/serverkey.pem",
			},
			Interop: config.Interop{
				ListenTLS: ":9188",
				SenderClientCAs: map[string]string{
					"000001": "testdata/clientca.pem",
				},
			},
		},
	}

	mockInterop := &mockInterop{}
	c := component.MustNew(test.GetLogger(t), config)
	c.RegisterInterop(mockInterop)

	test.Must(nil, c.Start())
	defer c.Close()

	certPool := x509.NewCertPool()
	certContent, err := ioutil.ReadFile("testdata/serverca.pem")
	a.So(err, should.BeNil)
	certPool.AppendCertsFromPEM(certContent)
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: certPool,
			GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair("testdata/clientcert.pem", "testdata/clientkey.pem")
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
		}},
	}

	// Correct SenderID.
	{
		req := &interop.JoinReq{
			NsJsMessageHeader: interop.NsJsMessageHeader{
				MessageHeader: interop.MessageHeader{
					MessageType:     interop.MessageTypeJoinReq,
					ProtocolVersion: "1.1",
				},
				SenderID:   types.NetID{0x0, 0x0, 0x1},
				ReceiverID: types.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			MACVersion: interop.MACVersion(ttnpb.MAC_V1_0_3),
		}
		buf, err := json.Marshal(req)
		a.So(err, should.BeNil)
		res, err := client.Post("https://localhost:9188", "application/json", bytes.NewReader(buf))
		a.So(err, should.BeNil)
		a.So(res.StatusCode, should.Equal, http.StatusOK)
	}

	// Wrong SenderID.
	{
		req := &interop.JoinReq{
			NsJsMessageHeader: interop.NsJsMessageHeader{
				MessageHeader: interop.MessageHeader{
					MessageType:     interop.MessageTypeJoinReq,
					ProtocolVersion: "1.1",
				},
				SenderID:   types.NetID{0x0, 0x0, 0x2},
				ReceiverID: types.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
			MACVersion: interop.MACVersion(ttnpb.MAC_V1_0_3),
		}
		buf, err := json.Marshal(req)
		a.So(err, should.BeNil)
		res, err := client.Post("https://localhost:9188", "application/json", bytes.NewReader(buf))
		a.So(err, should.BeNil)
		a.So(res.StatusCode, should.Equal, http.StatusBadRequest)
		var msg interop.ErrorMessage
		if !a.So(json.NewDecoder(res.Body).Decode(&msg), should.BeNil) {
			t.FailNow()
		}
		a.So(msg.Result, should.Equal, interop.ResultUnknownSender)
	}
}
