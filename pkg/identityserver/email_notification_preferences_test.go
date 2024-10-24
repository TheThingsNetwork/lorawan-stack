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

package identityserver

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestEmailNotificationPreferences(t *testing.T) {
	p := &storetest.Population{}
	admin := p.NewUser()
	admin.Admin = true
	admin.EmailNotificationPreferences = &ttnpb.EmailNotificationPreferences{
		Types: []ttnpb.NotificationType{
			ttnpb.NotificationType_API_KEY_CREATED,
		},
	}
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.Email.Provider = "dir"
		tempDir := t.TempDir()
		is.config.Email.Dir = tempDir

		reg := ttnpb.NewUserAccessClient(cc)

		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIds: admin.GetIds(),
			Name:    "api-key-name",
			Rights:  []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
		}, adminCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		entries, err := os.ReadDir(tempDir)
		if a.So(err, should.BeNil) && a.So(entries, should.HaveLength, 1) {
			data, err := os.ReadFile(filepath.Join(tempDir, entries[0].Name()))
			fmt.Printf("/////////////data: %v\n", string(data))
			a.So(err, should.BeNil)
			a.So(string(data), should.ContainSubstring, "<p>A new API key has just been created for your user")
		}
	})
}
