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

package smtp

import (
	"os"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestSMTP(t *testing.T) {
	a := assertions.New(t)

	smtpAddress := os.Getenv("SMTP_ADDRESS")
	if smtpAddress == "" {
		t.Skip("No SMTP server configured, skipping test")
	}

	ctx := test.Context()
	ctx = log.NewContext(ctx, test.GetLogger(t))

	smtp, err := New(
		ctx,
		email.Config{
			SenderName:    "Unit Test",
			SenderAddress: "unit@test.local",
		},
		Config{
			Address:  smtpAddress,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
		},
	)
	a.So(err, should.BeNil)

	err = smtp.Send(&email.Message{
		TemplateName:     "test",
		RecipientName:    "John Doe",
		RecipientAddress: "john.doe@example.com",
		Subject:          "Testing SMTP",
		HTMLBody:         "<h1>Testing SMTP</h1><p>We are testing SMTP</p>",
		TextBody:         "****************\nTesting SMTP\n****************\n\nWe are testing SMTP",
	})

	a.So(err, should.BeNil)
}
