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

package sendgrid

import (
	"os"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestSendGrid(t *testing.T) {
	a := assertions.New(t)

	apiKey := os.Getenv("SENDGRID_API_KEY")

	sg, err := New(
		log.NewContext(test.Context(), test.GetLogger(t)),
		email.Config{
			SenderConfig: email.SenderConfig{
				SenderName:    "Unit Test",
				SenderAddress: "unit@test.local",
			},
		},
		Config{
			APIKey:      apiKey,
			SandboxMode: true,
		},
	)
	a.So(err, should.BeNil)

	err = sg.Send(&email.Message{
		TemplateName:     ttnpb.NotificationType_UNKNOWN,
		RecipientName:    "John Doe",
		RecipientAddress: "john.doe@example.com",
		Subject:          "Testing SendGrid",
		HTMLBody:         "<h1>Testing SendGrid</h1><p>We are testing SendGrid</p>",
		TextBody:         "****************\nTesting SendGrid\n****************\n\nWe are testing SendGrid",
	})

	if apiKey == "" {
		a.So(err, should.NotBeNil)
	} else {
		a.So(err, should.BeNil)
	}
}
