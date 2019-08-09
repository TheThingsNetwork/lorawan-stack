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

package interop_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestJoinServerFQDN(t *testing.T) {
	for _, tc := range []struct {
		JoinEUI  types.EUI64
		Expected string
	}{
		{
			JoinEUI:  types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00},
			Expected: "0.0.0.0.0.0.0.d.e.7.5.d.3.b.0.7.joineuis.lora-alliance.org",
		},
	} {
		a := assertions.New(t)
		a.So(JoinServerFQDN(tc.JoinEUI, LoRaAllianceJoinEUIDomain), should.Equal, tc.Expected)
	}
}

func TestGetAppSKey(t *testing.T) {
	makeSessionKeyRequest := func() *ttnpb.SessionKeyRequest {
		return &ttnpb.SessionKeyRequest{
			JoinEUI:      types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00},
			DevEUI:       types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			SessionKeyID: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc},
		}
	}

	for _, tc := range []struct {
		Name                 string
		NewServer            func(*testing.T) *httptest.Server
		NewFallbackTLSConfig func() *tls.Config
		NewClientConfig      func(fqdn string, port uint32) (config.InteropClient, func() error)
		AsID                 string
		Request              *ttnpb.SessionKeyRequest
		ResponseAssertion    func(*testing.T, *ttnpb.AppSKeyResponse) bool
		ErrorAssertion       func(*testing.T, error) bool
	}{
		{
			Name: "Backend Interfaces 1.0/MICFailed",
			NewServer: func(t *testing.T) *httptest.Server {
				return newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := assertions.New(t)
					a.So(r.Method, should.Equal, http.MethodPost)

					b, err := ioutil.ReadAll(r.Body)
					a.So(err, should.BeNil)
					a.So(string(b), should.Equal, `{"ProtocolVersion":"1.0","TransactionID":0,"MessageType":"AppSKeyReq","SenderID":"test-as","ReceiverID":"70B3D57ED0000000","DevEUI":"0102030405060708","SessionKeyID":"016BFA7BAD4756346A674981E75CDBDC"}
`)
					a.So(r.Body.Close(), should.BeNil)

					_, err = w.Write([]byte(`{
  "Result": {
    "ResultCode": "MICFailed"
  }
}`))
					a.So(err, should.BeNil)
				}))
			},
			NewFallbackTLSConfig: func() *tls.Config { return nil },
			NewClientConfig: func(fqdn string, port uint32) (config.InteropClient, func() error) {
				confDir := test.Must(ioutil.TempDir("", "lorawan-stack-js-interop-test")).(string)
				confPath := filepath.Join(confDir, InteropClientConfigurationName)
				js1Path := filepath.Join(confDir, "test-js-1.yaml")
				js2Path := filepath.Join(confDir, "foo", "test-js-2.yaml")
				js3Path := filepath.Join(confDir, "test-js-3.yaml")

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "testdata"), 0755))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientCertPath), ClientCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientKeyPath), ClientKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerCertPath), ServerCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerKeyPath), ServerKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, RootCAPath), RootCA, 0644))

				rel := func(path string) string {
					return test.Must(filepath.Rel(confDir, path)).(string)
				}

				test.MustMultiple(ioutil.WriteFile(confPath, []byte(fmt.Sprintf(`join-servers:
   - file: %s
     join-euis:
        - 0000000000000000/0
        - 70b3d57ed0001000/52

   - file: %s
     join-euis:
        - 70b3d57ed0000000/40

   - file: %s
     join-euis:
        - 70b3d57ed0000000/39
        - 70b3d83ed0000000/30`,
					rel(js1Path),
					rel(js2Path),
					rel(js3Path),
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js1Path, []byte(fmt.Sprintf(`fqdn: test-js.fqdn
port: 12345
protocol: BI1.1
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   SomeHeader: Some foo bar
   TestHeader: baz`,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "foo"), 0755))
				test.MustMultiple(ioutil.WriteFile(js2Path, []byte(fmt.Sprintf(`fqdn: %s
port: %d
protocol: BI1.0
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   Authorization: Custom foo bar
   TestHeader: baz`,
					fqdn,
					port,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js3Path, []byte(`dns: invalid.dns
path: test-path
protocol: BI1.1`), 0644))

				return config.InteropClient{
						Directory: confDir,
					}, func() error {
						return os.RemoveAll(confDir)
					}
			},
			AsID:    "test-as",
			Request: makeSessionKeyRequest(),
			ResponseAssertion: func(t *testing.T, resp *ttnpb.AppSKeyResponse) bool {
				return assertions.New(t).So(resp, should.BeNil)
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrMIC)
			},
		},

		{
			Name: "Backend Interfaces 1.1/MICFailed",
			NewServer: func(t *testing.T) *httptest.Server {
				return newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := assertions.New(t)
					a.So(r.Method, should.Equal, http.MethodPost)

					b, err := ioutil.ReadAll(r.Body)
					a.So(err, should.BeNil)
					a.So(string(b), should.Equal, `{"ProtocolVersion":"1.1","TransactionID":0,"MessageType":"AppSKeyReq","SenderID":"test-as","ReceiverID":"70B3D57ED0000000","DevEUI":"0102030405060708","SessionKeyID":"016BFA7BAD4756346A674981E75CDBDC"}
`)
					a.So(r.Body.Close(), should.BeNil)

					_, err = w.Write([]byte(`{
  "Result": {
    "ResultCode": "MICFailed"
  }
}`))
					a.So(err, should.BeNil)
				}))
			},
			NewFallbackTLSConfig: func() *tls.Config { return nil },
			NewClientConfig: func(fqdn string, port uint32) (config.InteropClient, func() error) {
				confDir := test.Must(ioutil.TempDir("", "lorawan-stack-js-interop-test")).(string)
				confPath := filepath.Join(confDir, InteropClientConfigurationName)
				js1Path := filepath.Join(confDir, "test-js-1.yaml")
				js2Path := filepath.Join(confDir, "foo", "test-js-2.yaml")
				js3Path := filepath.Join(confDir, "test-js-3.yaml")

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "testdata"), 0755))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientCertPath), ClientCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientKeyPath), ClientKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerCertPath), ServerCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerKeyPath), ServerKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, RootCAPath), RootCA, 0644))

				rel := func(path string) string {
					return test.Must(filepath.Rel(confDir, path)).(string)
				}

				test.MustMultiple(ioutil.WriteFile(confPath, []byte(fmt.Sprintf(`join-servers:
   - file: %s
     join-euis:
        - 0000000000000000/0
        - 70b3d57ed0001000/52

   - file: %s
     join-euis:
        - 70b3d57ed0000000/40

   - file: %s
     join-euis:
        - 70b3d57ed0000000/39
        - 70b3d83ed0000000/30`,
					rel(js1Path),
					rel(js2Path),
					rel(js3Path),
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js1Path, []byte(fmt.Sprintf(`fqdn: test-js.fqdn
port: 12345
protocol: BI1.0
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   SomeHeader: Some foo bar
   TestHeader: baz`,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "foo"), 0755))
				test.MustMultiple(ioutil.WriteFile(js2Path, []byte(fmt.Sprintf(`fqdn: %s
port: %d
protocol: BI1.1
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   Authorization: Custom foo bar
   TestHeader: baz`,
					fqdn,
					port,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js3Path, []byte(`dns: invalid.dns
path: test-path
protocol: BI1.0`), 0644))

				return config.InteropClient{
						Directory: confDir,
					}, func() error {
						return os.RemoveAll(confDir)
					}
			},
			AsID:    "test-as",
			Request: makeSessionKeyRequest(),
			ResponseAssertion: func(t *testing.T, resp *ttnpb.AppSKeyResponse) bool {
				return assertions.New(t).So(resp, should.BeNil)
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrMIC)
			},
		},

		{
			Name: "Backend Interfaces 1.0/Success",
			NewServer: func(t *testing.T) *httptest.Server {
				return newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := assertions.New(t)
					a.So(r.Method, should.Equal, http.MethodPost)

					b, err := ioutil.ReadAll(r.Body)
					a.So(err, should.BeNil)
					a.So(string(b), should.Equal, `{"ProtocolVersion":"1.0","TransactionID":0,"MessageType":"AppSKeyReq","SenderID":"test-as","ReceiverID":"70B3D57ED0000000","DevEUI":"0102030405060708","SessionKeyID":"016BFA7BAD4756346A674981E75CDBDC"}
`)
					a.So(r.Body.Close(), should.BeNil)

					_, err = w.Write([]byte(`{
  "ProtocolVersion": "1.0",
  "TransactionID": 0,
  "MessageType": "AppSKeyAns",
  "ReceiverToken": "01",
  "SenderID": "70B3D57ED0000000",
  "ReceiverID": "test-as",
  "Result": {
    "ResultCode": "Success"
  },
  "Lifetime": 0,
  "AppSKey": {
    "KEKLabel": "as:010042",
    "AESKey": "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69"
  },
  "SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC"
}`))
					a.So(err, should.BeNil)
				}))
			},
			NewFallbackTLSConfig: func() *tls.Config { return nil },
			NewClientConfig: func(fqdn string, port uint32) (config.InteropClient, func() error) {
				confDir := test.Must(ioutil.TempDir("", "lorawan-stack-js-interop-test")).(string)
				confPath := filepath.Join(confDir, InteropClientConfigurationName)
				js1Path := filepath.Join(confDir, "test-js-1.yaml")
				js2Path := filepath.Join(confDir, "foo", "test-js-2.yaml")
				js3Path := filepath.Join(confDir, "test-js-3.yaml")

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "testdata"), 0755))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientCertPath), ClientCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientKeyPath), ClientKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerCertPath), ServerCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerKeyPath), ServerKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, RootCAPath), RootCA, 0644))

				rel := func(path string) string {
					return test.Must(filepath.Rel(confDir, path)).(string)
				}

				test.MustMultiple(ioutil.WriteFile(confPath, []byte(fmt.Sprintf(`join-servers:
   - file: %s
     join-euis:
        - 0000000000000000/0
        - 70b3d57ed0001000/52

   - file: %s
     join-euis:
        - 70b3d57ed0000000/40

   - file: %s
     join-euis:
        - 70b3d57ed0000000/39
        - 70b3d83ed0000000/30`,
					rel(js1Path),
					rel(js2Path),
					rel(js3Path),
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js1Path, []byte(fmt.Sprintf(`fqdn: test-js.fqdn
port: 12345
protocol: BI1.1
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   SomeHeader: Some foo bar
   TestHeader: baz`,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "foo"), 0755))
				test.MustMultiple(ioutil.WriteFile(js2Path, []byte(fmt.Sprintf(`fqdn: %s
port: %d
protocol: BI1.0
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   Authorization: Custom foo bar
   TestHeader: baz`,
					fqdn,
					port,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js3Path, []byte(`dns: invalid.dns
path: test-path
protocol: BI1.1`), 0644))

				return config.InteropClient{
						Directory: confDir,
					}, func() error {
						return os.RemoveAll(confDir)
					}
			},
			AsID:    "test-as",
			Request: makeSessionKeyRequest(),
			ResponseAssertion: func(t *testing.T, resp *ttnpb.AppSKeyResponse) bool {
				return assertions.New(t).So(resp, should.Resemble, &ttnpb.AppSKeyResponse{
					AppSKey: ttnpb.KeyEnvelope{
						KEKLabel:     "as:010042",
						EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69},
					},
				})
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name: "Backend Interfaces 1.1/Success",
			NewServer: func(t *testing.T) *httptest.Server {
				return newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := assertions.New(t)
					a.So(r.Method, should.Equal, http.MethodPost)

					b, err := ioutil.ReadAll(r.Body)
					a.So(err, should.BeNil)
					a.So(string(b), should.Equal, `{"ProtocolVersion":"1.1","TransactionID":0,"MessageType":"AppSKeyReq","SenderID":"test-as","ReceiverID":"70B3D57ED0000000","DevEUI":"0102030405060708","SessionKeyID":"016BFA7BAD4756346A674981E75CDBDC"}
`)
					a.So(r.Body.Close(), should.BeNil)

					_, err = w.Write([]byte(`{
  "ProtocolVersion": "1.1",
  "TransactionID": 0,
  "MessageType": "AppSKeyAns",
  "ReceiverToken": "01",
  "SenderID": "70B3D57ED0000000",
  "ReceiverID": "test-as",
  "Result": {
    "ResultCode": "Success"
  },
  "Lifetime": 0,
  "AppSKey": {
    "KEKLabel": "as:010042",
    "AESKey": "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69"
  },
  "SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC"
}`))
					a.So(err, should.BeNil)
				}))
			},
			NewFallbackTLSConfig: func() *tls.Config { return nil },
			NewClientConfig: func(fqdn string, port uint32) (config.InteropClient, func() error) {
				confDir := test.Must(ioutil.TempDir("", "lorawan-stack-js-interop-test")).(string)
				confPath := filepath.Join(confDir, InteropClientConfigurationName)
				js1Path := filepath.Join(confDir, "test-js-1.yaml")
				js2Path := filepath.Join(confDir, "foo", "test-js-2.yaml")
				js3Path := filepath.Join(confDir, "test-js-3.yaml")

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "testdata"), 0755))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientCertPath), ClientCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientKeyPath), ClientKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerCertPath), ServerCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerKeyPath), ServerKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, RootCAPath), RootCA, 0644))

				rel := func(path string) string {
					return test.Must(filepath.Rel(confDir, path)).(string)
				}

				test.MustMultiple(ioutil.WriteFile(confPath, []byte(fmt.Sprintf(`join-servers:
   - file: %s
     join-euis:
        - 0000000000000000/0
        - 70b3d57ed0001000/52

   - file: %s
     join-euis:
        - 70b3d57ed0000000/40

   - file: %s
     join-euis:
        - 70b3d57ed0000000/39
        - 70b3d83ed0000000/30`,
					rel(js1Path),
					rel(js2Path),
					rel(js3Path),
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js1Path, []byte(fmt.Sprintf(`fqdn: test-js.fqdn
port: 12345
protocol: BI1.0
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   SomeHeader: Some foo bar
   TestHeader: baz`,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "foo"), 0755))
				test.MustMultiple(ioutil.WriteFile(js2Path, []byte(fmt.Sprintf(`fqdn: %s
port: %d
protocol: BI1.1
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   Authorization: Custom foo bar
   TestHeader: baz`,
					fqdn,
					port,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js3Path, []byte(`dns: invalid.dns
path: test-path
protocol: BI1.0`), 0644))

				return config.InteropClient{
						Directory: confDir,
					}, func() error {
						return os.RemoveAll(confDir)
					}
			},
			AsID:    "test-as",
			Request: makeSessionKeyRequest(),
			ResponseAssertion: func(t *testing.T, resp *ttnpb.AppSKeyResponse) bool {
				return assertions.New(t).So(resp, should.Resemble, &ttnpb.AppSKeyResponse{
					AppSKey: ttnpb.KeyEnvelope{
						KEKLabel:     "as:010042",
						EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69},
					},
				})
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			ctx = log.NewContext(ctx, test.GetLogger(t))

			srv := tc.NewServer(t)
			defer srv.Close()

			host := strings.Split(test.Must(url.Parse(srv.URL)).(*url.URL).Host, ":")
			if len(host) != 2 {
				t.Fatalf("Invalid server host: %s", host)
			}

			conf, flush := tc.NewClientConfig(host[0], uint32(test.Must(strconv.ParseUint(host[1], 10, 32)).(uint64)))
			defer flush()

			cl, err := NewClient(ctx, conf, tc.NewFallbackTLSConfig())
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to create new client: %s", err)
			}

			res, err := cl.GetAppSKey(ctx, tc.AsID, tc.Request)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(tc.ResponseAssertion(t, res), should.BeTrue)
			} else if err != nil {
				t.Errorf("Received unexpected error: %v", errors.Stack(err))
			}
		})
	}
}

func TestHandleJoinRequest(t *testing.T) {
	makeJoinRequest := func() *ttnpb.JoinRequest {
		return &ttnpb.JoinRequest{
			SelectedMACVersion: ttnpb.MAC_V1_0_3,
			DevAddr:            types.DevAddr{0x01, 0x02, 0x03, 0x04},
			RxDelay:            ttnpb.RX_DELAY_5,
			Payload: &ttnpb.Message{
				Payload: &ttnpb.Message_JoinRequestPayload{
					JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI: types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00},
						DevEUI:  types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
					},
				},
			},
			RawPayload: []byte{0x00, 0x00, 0x00, 0x00, 0xd0, 0x7e, 0xd5, 0xb3, 0x70, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x38, 0x51, 0xf0, 0xb6},
		}
	}

	for _, tc := range []struct {
		Name                 string
		NewServer            func(*testing.T) *httptest.Server
		NewFallbackTLSConfig func() *tls.Config
		NewClientConfig      func(fqdn string, port uint32) (config.InteropClient, func() error)
		NetID                types.NetID
		Request              *ttnpb.JoinRequest
		ResponseAssertion    func(*testing.T, *ttnpb.JoinResponse) bool
		ErrorAssertion       func(*testing.T, error) bool
	}{
		{
			Name: "Backend Interfaces 1.0/Success",
			NewServer: func(t *testing.T) *httptest.Server {
				return newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := assertions.New(t)
					a.So(r.Method, should.Equal, http.MethodPost)

					b, err := ioutil.ReadAll(r.Body)
					a.So(err, should.BeNil)
					a.So(string(b), should.Equal, `{"ProtocolVersion":"1.0","TransactionID":0,"MessageType":"JoinReq","SenderID":"42FFFF","ReceiverID":"70B3D57ED0000000","SenderNSID":"42FFFF","MACVersion":"1.0.3","PHYPayload":"00000000D07ED5B370080706050403020100003851F0B6","DevEUI":"0102030405060708","DevAddr":"01020304","DLSettings":"00","RxDelay":5,"CFList":""}
`)
					a.So(r.Body.Close(), should.BeNil)

					_, err = w.Write([]byte(`{
  "ProtocolVersion": "1.0",
  "TransactionID": 0,
  "MessageType": "JoinAns",
  "ReceiverToken": "01",
  "SenderID": "70B3D57ED0000000",
  "ReceiverID": "000000",
  "ReceiverNSID": "000000",
  "PHYPayload": "204D675073BB4153B23653EFA82C1F3A49E19C2A8696C9A34BF492674779E4BEFA",
  "Result": {
    "ResultCode": "Success"
  },
  "Lifetime": 0,
  "NwkSKey": {
    "KEKLabel": "ns:000000",
    "AESKey": "EB56FE6681999F25D548CFEDD4A6528B331BB5ADE1CAF17F"
  },
  "AppSKey": {
    "KEKLabel": "as:010042",
    "AESKey": "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69"
  },
  "SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC"
}`))
					a.So(err, should.BeNil)
				}))
			},
			NewFallbackTLSConfig: func() *tls.Config { return nil },
			NewClientConfig: func(fqdn string, port uint32) (config.InteropClient, func() error) {
				confDir := test.Must(ioutil.TempDir("", "lorawan-stack-js-interop-test")).(string)
				confPath := filepath.Join(confDir, InteropClientConfigurationName)
				js1Path := filepath.Join(confDir, "test-js-1.yaml")
				js2Path := filepath.Join(confDir, "foo", "test-js-2.yaml")
				js3Path := filepath.Join(confDir, "test-js-3.yaml")

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "testdata"), 0755))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientCertPath), ClientCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientKeyPath), ClientKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerCertPath), ServerCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerKeyPath), ServerKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, RootCAPath), RootCA, 0644))

				rel := func(path string) string {
					return test.Must(filepath.Rel(confDir, path)).(string)
				}

				test.MustMultiple(ioutil.WriteFile(confPath, []byte(fmt.Sprintf(`join-servers:
   - file: %s
     join-euis:
        - 0000000000000000/0
        - 70b3d57ed0001000/52

   - file: %s
     join-euis:
        - 70b3d57ed0000000/40

   - file: %s
     join-euis:
        - 70b3d57ed0000000/39
        - 70b3d83ed0000000/30`,
					rel(js1Path),
					rel(js2Path),
					rel(js3Path),
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js1Path, []byte(fmt.Sprintf(`fqdn: test-js.fqdn
port: 12345
protocol: BI1.1
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   SomeHeader: Some foo bar
   TestHeader: baz`,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "foo"), 0755))
				test.MustMultiple(ioutil.WriteFile(js2Path, []byte(fmt.Sprintf(`fqdn: %s
port: %d
protocol: BI1.0
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   Authorization: Custom foo bar
   TestHeader: baz`,
					fqdn,
					port,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js3Path, []byte(`dns: invalid.dns
path: test-path
protocol: BI1.1`), 0644))

				return config.InteropClient{
						Directory: confDir,
					}, func() error {
						return os.RemoveAll(confDir)
					}
			},
			NetID:   types.NetID{0x42, 0xff, 0xff},
			Request: makeJoinRequest(),
			ResponseAssertion: func(t *testing.T, resp *ttnpb.JoinResponse) bool {
				return assertions.New(t).So(resp, should.Resemble, &ttnpb.JoinResponse{
					RawPayload: []byte{0x20, 0x4d, 0x67, 0x50, 0x73, 0xbb, 0x41, 0x53, 0xb2, 0x36, 0x53, 0xef, 0xa8, 0x2c, 0x1f, 0x3a, 0x49, 0xe1, 0x9c, 0x2a, 0x86, 0x96, 0xc9, 0xa3, 0x4b, 0xf4, 0x92, 0x67, 0x47, 0x79, 0xe4, 0xbe, 0xfa},
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "ns:000000",
							EncryptedKey: []byte{0xeb, 0x56, 0xfe, 0x66, 0x81, 0x99, 0x9f, 0x25, 0xd5, 0x48, 0xcf, 0xed, 0xd4, 0xa6, 0x52, 0x8b, 0x33, 0x1b, 0xb5, 0xad, 0xe1, 0xca, 0xf1, 0x7f},
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "as:010042",
							EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69},
						},
					},
				})
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name: "Backend Interfaces 1.1/Success",
			NewServer: func(t *testing.T) *httptest.Server {
				return newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := assertions.New(t)
					a.So(r.Method, should.Equal, http.MethodPost)

					b, err := ioutil.ReadAll(r.Body)
					a.So(err, should.BeNil)
					a.So(string(b), should.Equal, `{"ProtocolVersion":"1.1","TransactionID":0,"MessageType":"JoinReq","SenderID":"42FFFF","ReceiverID":"70B3D57ED0000000","SenderNSID":"42FFFF","MACVersion":"1.0.3","PHYPayload":"00000000D07ED5B370080706050403020100003851F0B6","DevEUI":"0102030405060708","DevAddr":"01020304","DLSettings":"00","RxDelay":5,"CFList":""}
`)
					a.So(r.Body.Close(), should.BeNil)

					_, err = w.Write([]byte(`{
  "ProtocolVersion": "1.1",
  "TransactionID": 0,
  "MessageType": "JoinAns",
  "ReceiverToken": "01",
  "SenderID": "70B3D57ED0000000",
  "ReceiverID": "000000",
  "ReceiverNSID": "000000",
  "PHYPayload": "204D675073BB4153B23653EFA82C1F3A49E19C2A8696C9A34BF492674779E4BEFA",
  "Result": {
    "ResultCode": "Success"
  },
  "Lifetime": 0,
  "NwkSKey": {
    "KEKLabel": "ns:000000",
    "AESKey": "EB56FE6681999F25D548CFEDD4A6528B331BB5ADE1CAF17F"
  },
  "AppSKey": {
    "KEKLabel": "as:010042",
    "AESKey": "2A195CC93CA54AD82CFB36C83D91450F3D2D523556F13E69"
  },
  "SessionKeyID": "016BFA7BAD4756346A674981E75CDBDC"
}`))
					a.So(err, should.BeNil)
				}))
			},
			NewFallbackTLSConfig: func() *tls.Config { return nil },
			NewClientConfig: func(fqdn string, port uint32) (config.InteropClient, func() error) {
				confDir := test.Must(ioutil.TempDir("", "lorawan-stack-js-interop-test")).(string)
				confPath := filepath.Join(confDir, InteropClientConfigurationName)
				js1Path := filepath.Join(confDir, "test-js-1.yaml")
				js2Path := filepath.Join(confDir, "foo", "test-js-2.yaml")
				js3Path := filepath.Join(confDir, "test-js-3.yaml")

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "testdata"), 0755))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientCertPath), ClientCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ClientKeyPath), ClientKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerCertPath), ServerCert, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, ServerKeyPath), ServerKey, 0644))
				test.MustMultiple(ioutil.WriteFile(filepath.Join(confDir, RootCAPath), RootCA, 0644))

				rel := func(path string) string {
					return test.Must(filepath.Rel(confDir, path)).(string)
				}

				test.MustMultiple(ioutil.WriteFile(confPath, []byte(fmt.Sprintf(`join-servers:
   - file: %s
     join-euis:
        - 0000000000000000/0
        - 70b3d57ed0001000/52

   - file: %s
     join-euis:
        - 70b3d57ed0000000/40

   - file: %s
     join-euis:
        - 70b3d57ed0000000/39
        - 70b3d83ed0000000/30`,
					rel(js1Path),
					rel(js2Path),
					rel(js3Path),
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js1Path, []byte(fmt.Sprintf(`fqdn: test-js.fqdn
port: 12345
protocol: BI1.0
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   SomeHeader: Some foo bar
   TestHeader: baz`,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(os.Mkdir(filepath.Join(confDir, "foo"), 0755))
				test.MustMultiple(ioutil.WriteFile(js2Path, []byte(fmt.Sprintf(`fqdn: %s
port: %d
protocol: BI1.1
tls:
   root-ca: %s
   certificate: %s
   key: %s
headers:
   Authorization: Custom foo bar
   TestHeader: baz`,
					fqdn,
					port,
					RootCAPath,
					ClientCertPath,
					ClientKeyPath,
				)), 0644))

				test.MustMultiple(ioutil.WriteFile(js3Path, []byte(`dns: invalid.dns
path: test-path
protocol: BI1.0`), 0644))

				return config.InteropClient{
						Directory: confDir,
					}, func() error {
						return os.RemoveAll(confDir)
					}
			},
			NetID:   types.NetID{0x42, 0xff, 0xff},
			Request: makeJoinRequest(),
			ResponseAssertion: func(t *testing.T, resp *ttnpb.JoinResponse) bool {
				return assertions.New(t).So(resp, should.Resemble, &ttnpb.JoinResponse{
					RawPayload: []byte{0x20, 0x4d, 0x67, 0x50, 0x73, 0xbb, 0x41, 0x53, 0xb2, 0x36, 0x53, 0xef, 0xa8, 0x2c, 0x1f, 0x3a, 0x49, 0xe1, 0x9c, 0x2a, 0x86, 0x96, 0xc9, 0xa3, 0x4b, 0xf4, 0x92, 0x67, 0x47, 0x79, 0xe4, 0xbe, 0xfa},
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: []byte{0x01, 0x6b, 0xfa, 0x7b, 0xad, 0x47, 0x56, 0x34, 0x6a, 0x67, 0x49, 0x81, 0xe7, 0x5c, 0xdb, 0xdc},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "ns:000000",
							EncryptedKey: []byte{0xeb, 0x56, 0xfe, 0x66, 0x81, 0x99, 0x9f, 0x25, 0xd5, 0x48, 0xcf, 0xed, 0xd4, 0xa6, 0x52, 0x8b, 0x33, 0x1b, 0xb5, 0xad, 0xe1, 0xca, 0xf1, 0x7f},
						},
						AppSKey: &ttnpb.KeyEnvelope{
							KEKLabel:     "as:010042",
							EncryptedKey: []byte{0x2a, 0x19, 0x5c, 0xc9, 0x3c, 0xa5, 0x4a, 0xd8, 0x2c, 0xfb, 0x36, 0xc8, 0x3d, 0x91, 0x45, 0x0f, 0x3d, 0x2d, 0x52, 0x35, 0x56, 0xf1, 0x3e, 0x69},
						},
					},
				})
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			ctx = log.NewContext(ctx, test.GetLogger(t))

			srv := tc.NewServer(t)
			defer srv.Close()

			host := strings.Split(test.Must(url.Parse(srv.URL)).(*url.URL).Host, ":")
			if len(host) != 2 {
				t.Fatalf("Invalid server host: %s", host)
			}

			conf, flush := tc.NewClientConfig(host[0], uint32(test.Must(strconv.ParseUint(host[1], 10, 32)).(uint64)))
			defer flush()

			cl, err := NewClient(ctx, conf, tc.NewFallbackTLSConfig())
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to create new client: %s", err)
			}

			res, err := cl.HandleJoinRequest(ctx, tc.NetID, tc.Request)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(tc.ResponseAssertion(t, res), should.BeTrue)
			} else if err != nil {
				t.Errorf("Received unexpected error: %v", errors.Stack(err))
			}
		})
	}
}
