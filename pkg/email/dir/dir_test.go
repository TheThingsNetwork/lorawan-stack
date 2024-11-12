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

package dir

import (
	"os"
	"path/filepath"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMailDir(t *testing.T) {
	a, ctx := test.New(t)
	tempDir := t.TempDir()

	mailer, err := New(ctx, email.Config{
		SenderConfig: email.SenderConfig{
			SenderName:    "Sender Name",
			SenderAddress: "sender@example.com",
		},
		Network: email.NetworkConfig{
			// Not used
		},
	}, tempDir)
	a.So(err, should.BeNil)

	err = mailer.Send(&email.Message{
		TemplateName:     ttnpb.GetNotificationTypeString(ttnpb.NotificationType_UNKNOWN),
		RecipientName:    "John Doe",
		RecipientAddress: "john.doe@example.com",
		Subject:          "Email Subject",
		HTMLBody:         "<h1>Title</h1><p>Body</p>",
		TextBody:         "Title\n-----\n\nBody",
	})
	a.So(err, should.BeNil)

	entries, err := os.ReadDir(tempDir)
	if a.So(err, should.BeNil) && a.So(entries, should.HaveLength, 1) {
		data, err := os.ReadFile(filepath.Join(tempDir, entries[0].Name()))
		a.So(err, should.BeNil)
		a.So(string(data), should.ContainSubstring, "<h1>Title</h1><p>Body</p>")
		a.So(string(data), should.ContainSubstring, "Title\n-----\n\nBody")
	}
}
