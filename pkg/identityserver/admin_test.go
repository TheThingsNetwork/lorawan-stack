// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/identityserver/email/mock"
	"go.thethings.network/lorawan-stack/pkg/identityserver/email/templates"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	errshould "go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var _ ttnpb.IsAdminServer = new(adminService)

func TestAdminSettings(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)
	defer is.store.Settings.Set(newTestSettings())

	ctx := newTestCtx(newTestUsers()["alice"].UserIdentifiers)

	resp, err := is.adminService.GetSettings(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	a.So(resp, should.EqualFieldsWithIgnores(SettingsGeneratedFields...), newTestSettings())

	// Modify settings.
	_, err = is.adminService.UpdateSettings(ctx, &ttnpb.UpdateSettingsRequest{
		Settings: ttnpb.IdentityServerSettings{
			IdentityServerSettings_UserRegistrationFlow: ttnpb.IdentityServerSettings_UserRegistrationFlow{
				SkipValidation: true,
			},
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"allowed_emails", "user_registration.skip_validation"},
		},
	})
	a.So(err, should.BeNil)

	resp, err = is.GetSettings(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	a.So(resp.AllowedEmails, should.HaveLength, 0)
	a.So(resp.IdentityServerSettings_UserRegistrationFlow.SkipValidation, should.BeTrue)
}

func TestAdminInvitations(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	ctx := newTestCtx(newTestUsers()["alice"].UserIdentifiers)
	email := "bar@baz.com"

	_, err := is.adminService.SendInvitation(ctx, &ttnpb.SendInvitationRequest{Email: email})
	a.So(err, should.BeNil)

	// Gather the token to register an account.
	token := ""

	invitation, ok := mock.Data().(*templates.Invitation)
	if a.So(ok, should.BeTrue) {
		token = invitation.Token
		a.So(invitation.Token, should.NotBeEmpty)
	}

	invitations, err := is.adminService.ListInvitations(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	if a.So(invitations.Invitations, should.HaveLength, 1) {
		i := invitations.Invitations[0]
		a.So(i.Email, should.Equal, email)
		a.So(i.IssuedAt.IsZero(), should.BeFalse)
		a.So(i.ExpiresAt.IsZero(), should.BeFalse)
	}

	// Use invitation.
	settings, err := is.store.Settings.Get()
	a.So(err, should.BeNil)
	defer func() {
		settings.IdentityServerSettings_UserRegistrationFlow.InvitationOnly = false
		is.store.Settings.Set(*settings)
	}()

	settings.IdentityServerSettings_UserRegistrationFlow.InvitationOnly = true
	err = is.store.Settings.Set(*settings)
	a.So(err, should.BeNil)

	user := ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{
			UserID: "invitation-user",
			Email:  email,
		},
		Password: "lol",
		Name:     "HI",
	}

	_, err = is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{User: user})
	a.So(err, should.NotBeNil)
	a.So(err, errshould.Describe, ErrInvitationTokenMissing)

	_, err = is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
		User:            user,
		InvitationToken: token,
	})
	a.So(err, should.BeNil)
	defer is.store.Users.Delete(user.UserIdentifiers)

	// Check user was created.
	found, err := is.adminService.GetUser(ctx, &ttnpb.UserIdentifiers{UserID: user.UserID})
	a.So(err, should.BeNil)
	a.So(found.UserID, should.Equal, user.UserID)
	a.So(found.Password, should.BeEmpty)

	// Check invitation was used.
	invitations, err = is.adminService.ListInvitations(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	a.So(invitations.Invitations, should.HaveLength, 0)

	// Can't send invitation to an already used email address.
	_, err = is.adminService.SendInvitation(ctx, &ttnpb.SendInvitationRequest{Email: email})
	a.So(err, should.NotBeNil)
	a.So(err, errshould.Describe, ErrEmailAddressAlreadyUsed)

	// Send a new invitation but revoke it later.
	email = "bar@bazqux.com"
	_, err = is.adminService.SendInvitation(ctx, &ttnpb.SendInvitationRequest{Email: email})
	a.So(err, should.BeNil)

	invitations, err = is.adminService.ListInvitations(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	if a.So(invitations.Invitations, should.HaveLength, 1) {
		i := invitations.Invitations[0]
		a.So(i.Email, should.Equal, email)
		a.So(i.IssuedAt.IsZero(), should.BeFalse)
		a.So(i.ExpiresAt.IsZero(), should.BeFalse)
	}

	_, err = is.adminService.DeleteInvitation(ctx, &ttnpb.DeleteInvitationRequest{Email: email})
	a.So(err, should.BeNil)

	invitations, err = is.adminService.ListInvitations(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	a.So(invitations.Invitations, should.HaveLength, 0)
}

func TestAdminUsers(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	ctx := newTestCtx(newTestUsers()["alice"].UserIdentifiers)
	user := newTestUsers()["bob"]

	// Reset password.
	found, err := is.store.Users.GetByID(user.UserIdentifiers, is.specializers.User)
	a.So(err, should.BeNil)

	old := found.GetUser().Password

	{
		resp, err := is.adminService.ResetUserPassword(ctx, &ttnpb.UserIdentifiers{UserID: user.UserID})
		a.So(err, should.BeNil)
		a.So(resp.Password, should.NotBeEmpty)

		data, ok := mock.Data().(*templates.PasswordReset)
		if a.So(ok, should.BeTrue) {
			a.So(data.Password, should.NotBeEmpty)
			a.So(data.Password, should.Equal, resp.Password)
		}

		found, err = is.store.Users.GetByID(user.UserIdentifiers, is.specializers.User)
		a.So(err, should.BeNil)
		a.So(old, should.NotEqual, found.GetUser().Password)
		a.So(found.GetUser().RequirePasswordUpdate, should.BeTrue)
	}

	// Make user admin.
	_, err = is.adminService.UpdateUser(ctx, &ttnpb.UpdateUserRequest{
		User: ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: user.UserID},
			Admin:           true,
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"admin"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.store.Users.GetByID(user.UserIdentifiers, is.specializers.User)
	a.So(err, should.BeNil)
	a.So(found.GetUser().Admin, should.BeTrue)

	// Delete user.
	user.UserID = "tmp-user"
	user.Email = "fofofo@bar.com"
	err = is.store.Users.Create(user)
	a.So(err, should.BeNil)

	_, err = is.adminService.DeleteUser(ctx, &ttnpb.UserIdentifiers{UserID: user.UserID})
	a.So(err, should.BeNil)

	ddata, ok := mock.Data().(*templates.AccountDeleted)
	if a.So(ok, should.BeTrue) {
		a.So(ddata.UserID, should.Equal, user.UserID)
	}

	_, err = is.store.Users.GetByID(user.UserIdentifiers, is.specializers.User)
	a.So(err, should.NotBeNil)
	a.So(err, errshould.Describe, store.ErrUserNotFound)

	// List approved users.
	resp, err := is.adminService.ListUsers(ctx, &ttnpb.ListUsersRequest{
		ListUsersRequest_FilterState: &ttnpb.ListUsersRequest_FilterState{State: ttnpb.STATE_APPROVED},
	})
	a.So(err, should.BeNil)
	if a.So(resp.Users, should.HaveLength, 2) { // Second user is the admin default user.
		for _, user := range resp.Users {
			if user.UserIdentifiers.UserID == newTestUsers()["alice"].UserIdentifiers.UserID {
				a.So(user, should.EqualFieldsWithIgnores(UserGeneratedFields...), newTestUsers()["alice"])
			}
		}
	}
}

func TestAdminClients(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	ctx := newTestCtx(newTestUsers()["alice"].UserIdentifiers)
	client := newTestClient()

	found, err := is.adminService.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientID: client.ClientID})
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(ClientGeneratedFields...), client)

	clients, err := is.adminService.ListClients(ctx, &ttnpb.ListClientsRequest{
		ListClientsRequest_FilterState: &ttnpb.ListClientsRequest_FilterState{State: ttnpb.STATE_PENDING},
	})
	a.So(err, should.BeNil)
	a.So(clients.Clients, should.HaveLength, 0)

	client.ClientID = "tmp-client"

	err = is.store.Clients.Create(client)
	a.So(err, should.BeNil)

	_, err = is.adminService.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientID: client.ClientID})
	a.So(err, should.BeNil)

	data, ok := mock.Data().(*templates.ClientDeleted)
	if a.So(ok, should.BeTrue) {
		a.So(data.ClientID, should.Equal, "tmp-client")
	}
}
