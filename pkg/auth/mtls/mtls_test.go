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
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/mtls"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	//go:embed testdata/2222222222222222.pem
	validStoreGatewayCertificatePEM []byte

	//go:embed testdata/rootCA.pem
	contextRootCAPEM []byte

	//go:embed testdata/3333333333333333.pem
	validContextGatewayCertificatePEM []byte
)

func TestTLSCertVerification(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	// Certificate:
	// 	Data:
	// 	Version: 1 (0x0)
	// 	Serial Number:
	// 			bb:1c:b7:84:aa:f9:45:84
	// 	Signature Algorithm: sha256WithRSAEncryption
	// 	Issuer: C=NL, ST=Utrecht, L=Utrecht, O=The Test Network, OU=test, CN=test client, emailAddress=test@test
	// 	Validity
	// 			Not Before: Aug 16 12:35:43 2022 GMT
	// 			Not After : Aug 16 12:35:43 2023 GMT
	// 	Subject: C=NL, ST=Utrecht, L=Utrecht, O=The Test Network, OU=test, CN=test client, emailAddress=test@test
	invalidTestCertPEM := `-----BEGIN CERTIFICATE-----
MIIFlDCCA3wCCQC7HLeEqvlFhDANBgkqhkiG9w0BAQsFADCBizELMAkGA1UEBhMC
TkwxEDAOBgNVBAgMB1V0cmVjaHQxEDAOBgNVBAcMB1V0cmVjaHQxGTAXBgNVBAoM
EFRoZSBUZXN0IE5ldHdvcmsxDTALBgNVBAsMBHRlc3QxFDASBgNVBAMMC3Rlc3Qg
Y2xpZW50MRgwFgYJKoZIhvcNAQkBFgl0ZXN0QHRlc3QwHhcNMjIwODE2MTIzNTQz
WhcNMjMwODE2MTIzNTQzWjCBizELMAkGA1UEBhMCTkwxEDAOBgNVBAgMB1V0cmVj
aHQxEDAOBgNVBAcMB1V0cmVjaHQxGTAXBgNVBAoMEFRoZSBUZXN0IE5ldHdvcmsx
DTALBgNVBAsMBHRlc3QxFDASBgNVBAMMC3Rlc3QgY2xpZW50MRgwFgYJKoZIhvcN
AQkBFgl0ZXN0QHRlc3QwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDQ
Cq6H2nD+IxmicHd6iAcsMSsdTyEfI+OhsYbQUiDzhYQoSxLNBABu5GKvf3YhG1xf
tON1YEar4WXDVrSz+03MBC5yeLAIlz+kZLrzVb1AYLTYajwI3WmDyBuaXPqfakFt
94dghk6WW7ONDx+JzO4doSS8n2O2CCIqtrEKsUb9/8NAKNHkgV8GiXImICcvCcEP
8TPz+H5rxSUAUPnq2HXJAc+HIpt+1A/i2aRXhplTBA1gAWJbfLe+YZbjSVWNN0OG
DXEJj1c3zSnhtuH98ze0mtHiW1FF1YpVDK0m0zXB/WqPupT0XcY1U+UKd4CApVqs
EWxohpmTgVReW9MnWGCPSv8dz1VQPVfyKf/KlOk7TEawtFZ/4o1+52vNcxKY/zG5
kSNsYQXyNvjKNLboqdSKOcbREAISAJgiJZSmq42PgCEnUnrlolqUO9l550NiKTGE
XxUH7Pcfwx16KHSJEslP19szt0ZJI+CvN/Z8TM+Px+duJ/cbnaRgP1NHX1R7Duq2
2xSD9JEXQlJmWuoMj6HbbPStT7Uffb5hKwrwfR1EKoBv7+dUgDt/4IY8pZjasQES
oVEZG2b+QKpf0UFIrCVNm9MfmqVPphUgLhLtxf3wpycDrU4oGSaDq5MD8/IUNo4R
9eDXUGfO/roG8Qq0Avwss5gQ+KyQ0/Ke+LqW1rwN4wIDAQABMA0GCSqGSIb3DQEB
CwUAA4ICAQAkyYsuZMgZXlufh9dsXdtrPO60br66oisfi5uTqH5ozfnEtLOPdz8W
Z+eE+DRoAJzdtR7MsYV73RN6n1Is3Iy7YAByPnrY2ElO0AKmTv33muedzfIvgAeN
WfQOt/S1vrbJT8EfhmP/vNy0/NTHwoOWsKBB54eciBFZAeYZBo3xjE7JiZkv/g/j
yMDX3T7GnTuptNKQlpEk6xAPGZg5iPlzNbnlhl57/UA0tSzlh5aW9II8Gep7WsE7
XyGRJbhRfbAxp35cio9dNoX5Vxyad/hgO5qf1h/1AQve015o8R3G8B6d15SJCaGs
xIMxzzaVkBrRLu4Eg0z/RRW0P45zo6QgGMTvi0KW2q41ijjTeeHcrzMFJ8ljV0k1
wN0jcOETtUm8TL8uAWa86/k7e7Qv0xOyhbodUSmxDVqFVuJO9FDXBTNqhXTx4xdd
QqbdkgVF+vf0tWUKvx/8J1psvkUsMvzXxX8xud2zhdPsEMXWdgrgYgTpp2/u6azM
779WLAsXJIOxWMnXsoJClZB+6PTL+/Q2hYlvy35QVtvDvxPuVuqk67IV9cM7ot6G
fMoN3frX7CfJ+Mz1JARSnKFzD5VH1+9gMtLg7lbRXmnQCOV0yntylb+yCTeTALwD
enSyC2URWEsszHuPDCO9J0KAdbMbyIgq6w7as6ZeE1z90YC8H3Y8OA==
-----END CERTIFICATE-----`

	fetcher := fetch.FromFilesystem("testdata/store")
	caStore, err := mtls.NewCAStore(ctx, fetcher)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	// Invalid (wrong CN and expired)
	block, _ := pem.Decode([]byte(invalidTestCertPEM))
	a.So(block, should.NotBeNil)
	a.So(block.Type, should.Equal, "CERTIFICATE")
	invalidTestCert, err := x509.ParseCertificate(block.Bytes)
	a.So(err, should.BeNil)
	err = caStore.Verify(ctx, mtls.ClientTypeUnspecified, "1111111111111111", invalidTestCert)
	a.So(errors.IsInvalidArgument(err), should.BeTrue)

	// Valid
	block, _ = pem.Decode(validStoreGatewayCertificatePEM)
	a.So(block, should.NotBeNil)
	validStoreGatewayCertificate, err := x509.ParseCertificate(block.Bytes)
	a.So(err, should.BeNil)
	err = caStore.Verify(ctx, mtls.ClientTypeUnspecified, "2222222222222222", validStoreGatewayCertificate)
	a.So(err, should.BeNil)
	ctx = mtls.NewContextWithClientCertificate(ctx, validStoreGatewayCertificate)
	cert := mtls.ClientCertificateFromContext(ctx)
	a.So(cert, should.Resemble, validStoreGatewayCertificate)

	// Invalid (CA is not in store)
	block, _ = pem.Decode(validContextGatewayCertificatePEM)
	a.So(block, should.NotBeNil)
	validContextGatewayCertificate, err := x509.ParseCertificate(block.Bytes)
	a.So(err, should.BeNil)
	err = caStore.Verify(ctx, mtls.ClientTypeUnspecified, "3333333333333333", validContextGatewayCertificate)
	a.So(errors.IsInvalidArgument(err), should.BeTrue)

	// Valid (CA is in context)
	rootCACtx := mtls.AppendRootCAsToContext(ctx, contextRootCAPEM)
	err = caStore.Verify(rootCACtx, mtls.ClientTypeUnspecified, "3333333333333333", validContextGatewayCertificate)
	a.So(err, should.BeNil)
}
