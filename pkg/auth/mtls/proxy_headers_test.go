// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package mtls_test

import (
	"net/http"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/mtls"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestProxyHeaders(t *testing.T) {
	a := assertions.New(t)

	// Envoy Proxy
	envoyHeader := http.Header{
		"X-Forwarded-Client-Cert": []string{`Chain=-----BEGIN%20CERTIFICATE-----%0AMIIBjzCCATSgAwIBAgIQUH3tGMgZLwzWHitr8Kg25zAKBggqhkjOPQQDAjAPMQ0w%0ACwYDVQQKDARUZXN0MB4XDTIzMDkwNTEyNTIxMFoXDTIzMDkxMjEzNTIxMFowDzEN%0AMAsGA1UEAwwEZGVtbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKBy7%2Feqgv%2F5%0AgJHb%2FB1cbRYpOLsDpP6vlp6qGquDY%2F%2BACq99Bu4sllxT%2Bub541PXvvfXCcui4yBj%0AiR9xyvaalbKjcjBwMAkGA1UdEwQCMAAwHwYDVR0jBBgwFoAUKD2obgo6rRioKOdD%0AQGvwbg0tgrwwHQYDVR0OBBYEFIqKk%2FkMWZI98meIv2RWRYBoJdzBMA4GA1UdDwEB%0A%2FwQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDAjAKBggqhkjOPQQDAgNJADBGAiEA%0AwOstWm11taMjE%2F63l2HXVPK7G2oWKds%2FKI5ytVPGP%2FoCIQCXz3kDfe0lARdThVq4%0AFemdmA39J3S0eofp1H0w%2Fdk45A%3D%3D%0A-----END%20CERTIFICATE-----%0A`},
	}
	envoyCert, err := mtls.FromProxyHeaders(envoyHeader)
	a.So(err, should.BeNil)
	a.So(envoyCert, should.NotBeNil)
	a.So(envoyCert.Subject.CommonName, should.Equal, "demo")

	// Treafik
	traefikHeader := http.Header{
		"X-Forwarded-Tls-Client-Cert": []string{`MIIBjzCCATSgAwIBAgIQUH3tGMgZLwzWHitr8Kg25zAKBggqhkjOPQQDAjAPMQ0wCwYDVQQKDARUZXN0MB4XDTIzMDkwNTEyNTIxMFoXDTIzMDkxMjEzNTIxMFowDzENMAsGA1UEAwwEZGVtbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKBy7/eqgv/5gJHb/B1cbRYpOLsDpP6vlp6qGquDY/+ACq99Bu4sllxT+ub541PXvvfXCcui4yBjiR9xyvaalbKjcjBwMAkGA1UdEwQCMAAwHwYDVR0jBBgwFoAUKD2obgo6rRioKOdDQGvwbg0tgrwwHQYDVR0OBBYEFIqKk/kMWZI98meIv2RWRYBoJdzBMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDAjAKBggqhkjOPQQDAgNJADBGAiEAwOstWm11taMjE/63l2HXVPK7G2oWKds/KI5ytVPGP/oCIQCXz3kDfe0lARdThVq4FemdmA39J3S0eofp1H0w/dk45A==,MIIBXzCCAQWgAwIBAgIRAJTV+HL0k86QRsgfvXvHOUYwCgYIKoZIzj0EAwIwDzENMAsGA1UECgwEVGVzdDAeFw0yMzAxMjAwOTU3MzNaFw0zMzAxMjAxMDU3MjZaMA8xDTALBgNVBAoMBFRlc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQ2Bs4mdaA2EzE4QNEIpA4jE6bls8C+EzwhPr3krpyyo8Qz7dPqQSDcZ2zCPWhX6m+yZVfvexwFWPbKQdO68wIzo0IwQDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQoPahuCjqtGKgo50NAa/BuDS2CvDAOBgNVHQ8BAf8EBAMCAYYwCgYIKoZIzj0EAwIDSAAwRQIhAO20d0cqsAIKRqVyFjMPTyaMv0KyYWzdWsSoQJAi2Zf1AiBrjWSJrYDmuUGhkZc1xRBqb6Ku/EocEzkSQ72DFVwMPA==`},
	}
	traefikCert, err := mtls.FromProxyHeaders(traefikHeader)
	a.So(err, should.BeNil)
	a.So(traefikCert, should.NotBeNil)
	a.So(traefikCert.Subject.CommonName, should.Equal, "demo")
}
