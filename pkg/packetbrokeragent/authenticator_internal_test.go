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

package packetbrokeragent

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"golang.org/x/oauth2"
)

type tokenSourceFunc func() (*oauth2.Token, error)

func (f tokenSourceFunc) Token() (*oauth2.Token, error) {
	return f()
}

func TestOAuth2(t *testing.T) {
	a, ctx := test.New(t)

	authenticator := &oauth2Authenticator{
		tokenSource: tokenSourceFunc(func() (*oauth2.Token, error) {
			return &oauth2.Token{
				AccessToken: "eyJhbGciOiJFZERTQSIsImtpZCI6InByb2QtMjAyMC0xMi0yOCIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjgyNTk1MjAsImh0dHBzOi8vaWFtLnBhY2tldGJyb2tlci5uZXQvY2xhaW1zIjp7Im5zIjpbeyJuaWQiOjE5LCJ0aWQiOiJ0dGkifV0sInJpZ2h0cyI6WzAsNCw3LDgsOSwxMCwxMV19LCJpYXQiOjE2MjgyNTU5MjAsImlzcyI6Imh0dHBzOi8vaWFtLnBhY2tldGJyb2tlci5uZXQiLCJqdGkiOiIwMUZDRFNWN1hURVg0NlZRUjlLMTJFMVFFMiJ9.AA",
				TokenType:   "bearer",
				Expiry:      time.Date(2021, 8, 6, 14, 19, 0, 0, time.UTC),
			}, nil
		}),
	}
	tenantID, err := authenticator.AuthInfo(ctx)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(tenantID, should.Resemble, &ttnpb.PacketBrokerNetworkIdentifier{
		NetId:    0x000013,
		TenantId: "tti",
	})
}
