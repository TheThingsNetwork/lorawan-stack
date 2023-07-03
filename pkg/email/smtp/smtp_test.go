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

package smtp

import (
	"net"
	"testing"

	"github.com/emersion/go-smtp"
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestSMTP(t *testing.T) {
	a := assertions.New(t)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	bkd := &backend{
		messages: make(chan *message, 1),
	}
	server := smtp.NewServer(bkd)

	go server.Serve(lis)

	smtpAddress := lis.Addr().String()

	ctx := test.Context()
	ctx = log.NewContext(ctx, test.GetLogger(t))

	smtp, err := New(
		ctx,
		email.Config{
			SenderName:    "Unit Test",
			SenderAddress: "unit@test.local",
		},
		Config{
			Address: smtpAddress,
		},
	)
	a.So(err, should.BeNil)

	email := &email.Message{
		TemplateName:     "test",
		RecipientName:    "John Doe",
		RecipientAddress: "john.doe@example.com",
		Subject:          "Testing SMTP",
		HTMLBody:         "<h1>Testing SMTP</h1><p>We are testing SMTP</p>",
		TextBody:         "****************\r\nTesting SMTP\r\n****************\r\n\r\nWe are testing SMTP",
	}

	err = smtp.Send(email)
	a.So(err, should.BeNil)

	received := <-bkd.messages

	a.So(received.Sender, should.Equal, "unit@test.local")
	a.So(received.Recipients, should.Contain, email.RecipientAddress)

	dataString := string(received.Data)

	a.So(dataString, should.ContainSubstring, email.Subject)
	a.So(dataString, should.ContainSubstring, email.HTMLBody)
	a.So(dataString, should.ContainSubstring, email.TextBody)
}
