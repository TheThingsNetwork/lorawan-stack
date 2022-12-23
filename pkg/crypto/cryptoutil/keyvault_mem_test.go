// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package cryptoutil_test

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/hex"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMemKeyVault(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	plaintext, _ := hex.DecodeString("00112233445566778899AABBCCDDEEFF")
	kek, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F")
	ciphertext, _ := hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")

	genericPlainText := []byte("thisisabigsecret")
	key, _ := hex.DecodeString("00112233445566778899AABBCCDDEEFF")

	kv := cryptoutil.NewMemKeyVault(map[string][]byte{
		"kek1": kek,
		"key1": key,
		"cert1": []byte(`-----BEGIN CERTIFICATE-----
MIIBfDCCASKgAwIBAgIBBDAKBggqhkjOPQQDAjAkMRAwDgYDVQQKEwdBY21lIENv
MRAwDgYDVQQDEwdSb290IENBMCAXDTE5MDgwNzExMzYxOFoYDzIxMTkwNzE0MTEz
NjE4WjAyMRAwDgYDVQQKEwdBY21lIENvMR4wHAYDVQQDDBVjbGllbnRfYXV0aF90
ZXN0X2NlcnQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASN/fNk9eVz+yx5O3tj
MXSjrV95e+T3wkXLL6z+PSDMzNSMSRrv5bNM8RGL24xCMRGezWpcb/0Mkt79DGLS
vziEozUwMzAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYIKwYBBQUHAwIwDAYD
VR0TAQH/BAIwADAKBggqhkjOPQQDAgNIADBFAiEAylB8RCRTv3FJYonJkfKTVOMN
cr7idt4xexCs+l8ALzMCIGBu4+S8YWGq9yQ4BL86Rcf7j7veXm57o6kjxU4F6V7x
-----END CERTIFICATE-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBVXljefOUPY++0sovcF0dboOLEJz4eZ9DoUE8o9Y7GHoAoGCCqGSM49
AwEHoUQDQgAEjf3zZPXlc/sseTt7YzF0o61feXvk98JFyy+s/j0gzMzUjEka7+Wz
TPERi9uMQjERns1qXG/9DJLe/Qxi0r84hA==
-----END EC PRIVATE KEY-----
`),
		"client": []byte(`-----BEGIN CERTIFICATE-----
MIIEbzCCAtegAwIBAgIRANiI/y0X2PqYWt0LYz7EcvkwDQYJKoZIhvcNAQELBQAw
gZsxHjAcBgNVBAoTFW1rY2VydCBkZXZlbG9wbWVudCBDQTE4MDYGA1UECwwvam9o
YW5ASm9oYW5zLU1hY0Jvb2stUHJvLmxvY2FsIChKb2hhbiBTdG9ra2luZykxPzA9
BgNVBAMMNm1rY2VydCBqb2hhbkBKb2hhbnMtTWFjQm9vay1Qcm8ubG9jYWwgKEpv
aGFuIFN0b2traW5nKTAeFw0yMjEyMTQxMzE0MzBaFw0yNTAzMTQxMzE0MzBaMGMx
JzAlBgNVBAoTHm1rY2VydCBkZXZlbG9wbWVudCBjZXJ0aWZpY2F0ZTE4MDYGA1UE
Cwwvam9oYW5ASm9oYW5zLU1hY0Jvb2stUHJvLmxvY2FsIChKb2hhbiBTdG9ra2lu
ZykwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC8CUD5bBqUZKPU0a5m
X8HKN+WMiBqHnAOC2b3DG66LpE67G1iQSskpnyaC3HfjR4sbN7fvujq67KM9CCSp
N+HrJsolW9lob3E7NPqrC9fp1o8k6e13JeQH3HNogkcEd2xBcmOm/pOp1zXUQfMo
2sqsaESmy7++cAyMITBK627dZVh3nTpNGtPuQSQuzO57EVxM9eJnYKJK6+Qyt+qg
kxod+uHsCHT6avV7oITD5m8hYueEwGTk5X5yvuOtyQgjhMGgS4wypyRahIXuUymd
bNA0656ze4+XwBYISuJtCMONveYgjVfSAzF0qWe9WwwYjNvl65aI1V77ar5U6kMl
yPCBAgMBAAGjZTBjMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcD
AgYIKwYBBQUHAwEwHwYDVR0jBBgwFoAUhzWfRIDn/IZkQruKVvTVZtIAVIQwEQYD
VR0RBAowCIIGMDAwMDEzMA0GCSqGSIb3DQEBCwUAA4IBgQA/aqDqz5gpYN5G9foP
mouFDJlNm4sWaDxMtCweaOaI5K7golON7vXU1OYs8/mZijeiWwWI5Lvp6iI2oFVW
mPf4DdyCUXzDQUHlBP5UvbEPDosRR2+hwLUDmGTmCd2SJ00Gqi6tIJvSdDReYyMx
fNu9SweSiN31adVO4yLxvkDbsvRNySlmJtsN0zZOlf5EO1AjBgFjShCV6VXzRZVB
/E+hhZLYVQHftdXsZwoZsWrrz7NIzI3LQWXIhdM85A6JxCj5/X7QCKfC5w5bDsFE
WsGApwMYMPIU03Pi74xImho/oHVhJl5P6VKWiry/odczSnUeR8JGPKLI74xowz5x
SFgCNkcRoEBSvFc/RtcarHln0bTWdpHn+M17mz0bGm18hmqJk869xXVTGdBV0IMt
VFD2heIHIPCuQeK72HiYBuPf4zUDzdJBM0bws9jg4tG7C4UQ/G9nKijCTTSBZ3mh
PfSiNlBbUVD+Kr9jdJ6olE+ptRkrkwHs3fS3GczUap6iJoQ=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFBzCCA2+gAwIBAgIQYgrUieWsg08BHkDDcf5LxjANBgkqhkiG9w0BAQsFADCB
mzEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMTgwNgYDVQQLDC9qb2hh
bkBKb2hhbnMtTWFjQm9vay1Qcm8ubG9jYWwgKEpvaGFuIFN0b2traW5nKTE/MD0G
A1UEAww2bWtjZXJ0IGpvaGFuQEpvaGFucy1NYWNCb29rLVByby5sb2NhbCAoSm9o
YW4gU3Rva2tpbmcpMB4XDTIyMTIxNDEyNDUwNVoXDTMyMTIxNDEyNDUwNVowgZsx
HjAcBgNVBAoTFW1rY2VydCBkZXZlbG9wbWVudCBDQTE4MDYGA1UECwwvam9oYW5A
Sm9oYW5zLU1hY0Jvb2stUHJvLmxvY2FsIChKb2hhbiBTdG9ra2luZykxPzA9BgNV
BAMMNm1rY2VydCBqb2hhbkBKb2hhbnMtTWFjQm9vay1Qcm8ubG9jYWwgKEpvaGFu
IFN0b2traW5nKTCCAaIwDQYJKoZIhvcNAQEBBQADggGPADCCAYoCggGBAK1yiSEd
Oa+aDu/YOFvDfE0uYv9jpncI9Cy3gN4D1M0+w1gk4i7LavxJOJF/OEv2pY5dnQoX
ucYQEtZvxeUY3TSSE6ak4DNSvryOytn7wvvRpvnnbwRNKu8UkrdJT1gotobY8JUt
q4exzzaqPhD4rOVWbDfoTkO4++qpecMgtkLJ7lPTmYZOlht4Zan4xEYfleBJvEAh
g2UU5UbYI8q3DcREP3sl/V5S82NmrCgFwJY+2T2hRZkt2HXG4e7wKvc6TTMFnOyy
SAIqZfv2WVBo+haLOJkXWhzU3bExyhbsrWTDZEeloZK74w/tSv9dz6QGn6RlVb1O
6TuG8fl29tXrx1TIF22fgwK29D5e2xlHyCioQyxXOdt2kvrlLSKq48d+rMB6PiBD
TccISOCTkAxzITACSncBzu1Y3u8F8hdDtGYgtXhcvTn5pbzvEvOsg5tgqrVwzj4y
ZbE40+ZNf0pz1//OtZhrbapoS/kwwMFy+rSfly6LInrqvQQQqEgtjrEJ3wIDAQAB
o0UwQzAOBgNVHQ8BAf8EBAMCAgQwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4E
FgQUhzWfRIDn/IZkQruKVvTVZtIAVIQwDQYJKoZIhvcNAQELBQADggGBACRJqdQ3
OIE51KUrtVe07OrUpBirfRM7puIvrB/+MEN2fFOUblQ+B7UKyzHOeH/UmPnSaR1W
AkdwT0fx3r9rf9tVadUpNlwEk/pDT7aQQCw7KAdVUoEPFcyBD0xVvNpNaWozeub1
g2JIUnLd2XTvygbLRAUQ/+I9WBquMDyTYtgDgp5ShzsWwnYGuQjQyJb7UV/Qddf2
ohFwxsIea8C/QEkMgkJDTVLn6JNKFlriYPLi0TXE0KQbN5f187j9caeoFKahyxZA
Br5UbQI6ACWzdTn32lSQTs0RGTnOWq2hFPLYB045XsdfTcUFk3l3ivnlZ103HSDq
+EDPC6I/Hhjwsn8WGcQmnlcSGL7RSRNckhg4ruCjNSI0TuJ2wI3tTpGwPHEBUgHh
sMxzjZqT4RBSF7I+9Ki/yufqWNsrovKufhOrj5FnJdqMOs1Pm0fcRb7LCeuuWc0p
pveJZu/2DmCUNMGVQKew7mrNnUTtFgvDja3/BS+OxZ/p7pqmvpXsvkAvzA==
-----END CERTIFICATE-----
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC8CUD5bBqUZKPU
0a5mX8HKN+WMiBqHnAOC2b3DG66LpE67G1iQSskpnyaC3HfjR4sbN7fvujq67KM9
CCSpN+HrJsolW9lob3E7NPqrC9fp1o8k6e13JeQH3HNogkcEd2xBcmOm/pOp1zXU
QfMo2sqsaESmy7++cAyMITBK627dZVh3nTpNGtPuQSQuzO57EVxM9eJnYKJK6+Qy
t+qgkxod+uHsCHT6avV7oITD5m8hYueEwGTk5X5yvuOtyQgjhMGgS4wypyRahIXu
UymdbNA0656ze4+XwBYISuJtCMONveYgjVfSAzF0qWe9WwwYjNvl65aI1V77ar5U
6kMlyPCBAgMBAAECggEAa7vtkzqh+/WxfGTqxFMG6EKgbaUpdhsoU9dHhzscBXwN
c9yWII4ItaUu3nlM41aBWAXTiDGuJp0gZf59asrO0Pk3hrIaXWDEgoS3PjsZ6St6
dk7lNIfsH6jqIq3J3MBDsTfF6s8fcYcRm1xx4i2BQ8i11M8WPBlcxwjY74P20Ded
ZgQb0B3KVTdrZvNHH/uL4S/fa/E5MHCFw4p1pT2hhHLzumxK+mwXnkKXEiiPBNtG
4QRwixsGhJkE+aDMc14H3q1EltTBSpq3twD1sfl+uDnIo83aEp6An4/AJt+1eouF
3Jsdl6oaZLVIEvxqMBzP9Fjqif/HfIc8VypVFoB8AQKBgQDu5rwbBth2AHSRxROv
ahYNzAydnxOc6ma4H7ax7LAZk0qKCmxahuXAuXGSouYaJr1kRmBw+2CwU8u84lHa
nnr92qk2fa/+P6lDADIHK2dATIqEXViuMzzZCugQYFUwrDlnvxH1HQ4qY/jOYZ6M
b9HDNrasahBEDHPu0uiNVeUaoQKBgQDJfosAXHemxRCPq3Lw7R2dU1KSVe91w2Sv
JSwtzZn3EkN54r3LQFUGU8L1A7i3jeF9g8VGRFkO5GUudBq1QNf5GH8H1zHuQ9hE
KoX6YnuWIJ9UWEKR+eZB+R9Tn6jAZsALRNV7G+9EJZMpVXzogzuxSQJnDKOsp3Rc
SPU5pDDp4QKBgGq82nR01Ye7YlmypL3t9xaJAWX3Kgsky2oeeUD7kB6NKXONfqXf
uY0nDbBHaelrP5kqvHIeTi/Z8KBeudWkky0SYiH/e/9rsBNIZhG/+azHxeen0TRb
nicW8WJHuCg7+pX4z2wlZCvaaNLE2NLELwM6UdmstcHBkpa00sQ7CVahAoGBAKa/
BRMwcohdjt4GWWGOKLLYkH2vhjJjl7/luFDTU/YGdDa68KvyOiq5SJ5xDP1B+fhg
AvKqfzT2x9EQnkWfOtvWbNG1QYnXNXL76dISjAnqR1CKldSuBOJV4pnWh9VpcsYg
mbZ+oJw5qDZNm8fjSpPlQoq7B/xKu93fNqkT+rKhAoGAHFBfgS7WAi6Snz0LreMb
JRirtVNS4U9qx4x9yvLXrwbqHg4wlOLNbu5DAXxVCy2tLjTT1faIG1idIYLYGSw2
6XIkwt9xoA7LJ5IIebRuAKNpINxu2IlscVLly/nXnpnNnEwvwWadQBc5dRC3mEPp
BQApUt5UZZkvwrTGD5Ez4KM=
-----END PRIVATE KEY-----
`),
	})
	ks := crypto.NewKeyService(kv)

	// Existing KEK.
	{
		actual, err := ks.Wrap(test.Context(), plaintext, "kek1")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, ciphertext)
	}
	{
		actual, err := ks.Unwrap(test.Context(), ciphertext, "kek1")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, plaintext)
	}

	// Non-existing KEK.
	{
		_, err := ks.Wrap(test.Context(), plaintext, "kek2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}
	{
		_, err := ks.Unwrap(test.Context(), ciphertext, "kek2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Existing Key.
	{
		encrypted, err := ks.Encrypt(test.Context(), genericPlainText, "key1")
		a.So(err, should.BeNil)

		expectedCiphertextLen := len(genericPlainText) + 12 + 16
		a.So(len(encrypted), should.Equal, expectedCiphertextLen)

		actual, err := ks.Decrypt(test.Context(), encrypted, "key1")
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, genericPlainText)
	}

	// Non-existing Key.
	{
		_, err := ks.Encrypt(test.Context(), genericPlainText, "key2")
		a.So(errors.IsNotFound(err), should.BeTrue)

		_, err = ks.Decrypt(test.Context(), []uint8{}, "key2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Export existing certificate.
	{
		cert, err := ks.ServerCertificate(test.Context(), "cert1")
		a.So(err, should.BeNil)
		if a.So(len(cert.Certificate), should.Equal, 1) {
			a.So(len(cert.Certificate[0]), should.Equal, 384)
		}
		a.So(cert.PrivateKey, should.HaveSameTypeAs, &ecdsa.PrivateKey{})
	}

	// Export non-existing certificate.
	{
		_, err := ks.ServerCertificate(test.Context(), "cert2")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Export existing client certificate.
	{
		cert, err := ks.ClientCertificate(test.Context())
		a.So(err, should.BeNil)
		a.So(len(cert.Certificate), should.Equal, 2)
		a.So(cert.PrivateKey, should.HaveSameTypeAs, &rsa.PrivateKey{})
	}

	// Hashing with valid key
	{
		hash1, err := ks.HMACHash(test.Context(), []byte{0x01, 0x02}, "key1")
		a.So(err, should.BeNil)
		a.So(hash1, should.NotBeNil)

		hash2, err := ks.HMACHash(test.Context(), []byte{0x01, 0x02}, "key1")
		a.So(err, should.BeNil)
		a.So(hash2, should.NotBeNil)
		a.So(subtle.ConstantTimeCompare(hash1, hash2), should.Equal, 1)
	}

	// Hashing with invalid key
	{
		hash, err := ks.HMACHash(test.Context(), []byte{0x01, 0x02}, "key2")
		a.So(errors.IsNotFound(err), should.BeTrue)
		a.So(hash, should.BeNil)
	}

	// Hashing with corrupted value key
	{
		hash1, err := ks.HMACHash(test.Context(), []byte{0x01, 0x02}, "key1")
		a.So(err, should.BeNil)
		a.So(hash1, should.NotBeNil)

		hash2, err := ks.HMACHash(test.Context(), []byte{0x01, 0x02, 0x03}, "key1")
		a.So(err, should.BeNil)
		a.So(hash2, should.NotBeNil)
		a.So(subtle.ConstantTimeCompare(hash1, hash2), should.BeZeroValue)
	}
}
