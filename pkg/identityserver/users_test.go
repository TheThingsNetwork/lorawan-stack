// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/templates"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.IsUserServer = new(userService)

func TestUser(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	user := ttnpb.User{
		UserIdentifier: ttnpb.UserIdentifier{"daniel"},
		Password:       "12345",
		Email:          "foo@bar.com",
		Name:           "hi",
	}

	ctx := testCtx(user.UserID)

	// can't create an account using a not allowed email
	user.Email = "foo@foo.com"
	_, err := is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
		User: user,
	})
	a.So(err, should.NotBeNil)
	a.So(ErrEmailAddressNotAllowed.Describes(err), should.BeTrue)
	user.Email = "foo@bar.com"

	// can't create account using a blacklisted id
	for _, id := range testSettings().BlacklistedIDs {
		user.UserID = id
		_, err = is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
			User: user,
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}
	user.UserID = "daniel"

	// create the account
	_, err = is.userService.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
		User: user,
	})
	a.So(err, should.BeNil)

	// can't retrieve profile without proper claims
	found, err := is.userService.GetUser(context.Background(), &pbtypes.Empty{})
	a.So(found, should.BeNil)
	a.So(err, should.NotBeNil)
	a.So(ErrNotAuthorized.Describes(err), should.BeTrue)

	// check that response doesnt include password within
	found, err = is.userService.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.Name, should.Equal, user.Name)
	a.So(found.Password, should.HaveLength, 0)
	a.So(found.Email, should.Equal, user.Email)
	a.So(found.ValidatedAt.IsZero(), should.BeTrue)
	if testSettings().IdentityServerSettings_UserRegistrationFlow.AdminApproval {
		a.So(found.State, should.Equal, ttnpb.STATE_PENDING)
	} else {
		a.So(found.State, should.Equal, ttnpb.STATE_APPROVED)
	}

	// extract the validation token from the email and validate the user account
	data, ok := mock.Data().(*templates.EmailValidation)
	if a.So(ok, should.BeTrue) && a.So(data.Token, should.NotBeEmpty) {
		token := data.Token

		_, err := is.userService.ValidateUserEmail(context.Background(), &ttnpb.ValidateUserEmailRequest{
			Token: token,
		})
		a.So(err, should.BeNil)

		found, err := is.userService.GetUser(ctx, &pbtypes.Empty{})
		a.So(err, should.BeNil)
		a.So(found.ValidatedAt.IsZero(), should.BeFalse)
	}

	// try to update the user password providing a wrong old password
	_, err = is.userService.UpdateUserPassword(ctx, &ttnpb.UpdateUserPasswordRequest{
		New: "heheh",
	})
	a.So(err, should.NotBeNil)
	a.So(ErrInvalidPassword.Describes(err), should.BeTrue)

	_, err = is.userService.UpdateUserPassword(ctx, &ttnpb.UpdateUserPasswordRequest{
		Old: user.Password,
		New: "heheh",
	})
	a.So(err, should.BeNil)

	// generate a new API key
	key, err := is.userService.GenerateUserAPIKey(ctx, &ttnpb.GenerateUserAPIKeyRequest{
		Name:   "foo",
		Rights: ttnpb.AllUserRights(),
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllUserRights())

	// update api key
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.userService.UpdateUserAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// can't generate another API Key with the same name
	_, err = is.userService.GenerateUserAPIKey(ctx, &ttnpb.GenerateUserAPIKeyRequest{
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := is.userService.ListUserAPIKeys(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.userService.RemoveUserAPIKey(ctx, &ttnpb.RemoveUserAPIKeyRequest{
		Name: key.Name,
	})

	keys, err = is.userService.ListUserAPIKeys(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// update the user's email
	_, err = is.userService.UpdateUser(ctx, &ttnpb.UpdateUserRequest{
		User: ttnpb.User{
			Email: "newfoo@bar.com",
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"email"},
		},
	})
	a.So(err, should.BeNil)

	// check that the field validated_at has been reset
	found, err = is.userService.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.ValidatedAt.IsZero(), should.BeTrue)

	token := ""

	// extract the token from mail
	data, ok = mock.Data().(*templates.EmailValidation)
	if a.So(ok, should.BeTrue) && a.So(data.Token, should.NotBeEmpty) {
		token = data.Token
	}
	a.So(token, should.NotBeEmpty)

	// request a new validation token
	_, err = is.userService.RequestUserEmailValidation(ctx, &pbtypes.Empty{})

	// check that the old validation token doesnt work because we requested a new one
	_, err = is.userService.ValidateUserEmail(context.Background(), &ttnpb.ValidateUserEmailRequest{
		Token: token,
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrValidationTokenNotFound.Describes(err), should.BeTrue)

	// and therefore the email isn't validated yet
	found, err = is.userService.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.ValidatedAt.IsZero(), should.BeTrue)

	// get the latest sent validation token
	data, ok = mock.Data().(*templates.EmailValidation)
	if a.So(ok, should.BeTrue) {
		token = data.Token
	}
	a.So(token, should.NotBeEmpty)

	// validate the email
	_, err = is.userService.ValidateUserEmail(context.Background(), &ttnpb.ValidateUserEmailRequest{
		Token: token,
	})
	a.So(err, should.BeNil)

	found, err = is.userService.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.ValidatedAt.IsZero(), should.BeFalse)

	_, err = is.userService.RevokeAuthorizedClient(ctx, &ttnpb.ClientIdentifier{"non-existent-client"})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAuthorizedClientNotFound.Describes(err), should.BeTrue)

	// create a fake authorized client to the user
	client := &ttnpb.Client{
		ClientIdentifier: ttnpb.ClientIdentifier{"bar-client"},
		Description:      "description",
		Secret:           "secret",
		Grants:           []ttnpb.GrantType{ttnpb.GRANT_PASSWORD},
		Rights:           []ttnpb.Right{},
		RedirectURI:      "foo.ttn.dev/oauth",
		Creator:          testUsers()["john-doe"].UserIdentifier,
	}
	client.Rights = append(client.Rights, ttnpb.AllUserRights()...)
	err = is.store.Clients.Create(client)
	a.So(err, should.BeNil)

	accessToken, err := auth.GenerateAccessToken("")
	a.So(err, should.BeNil)

	accessData := &store.AccessData{
		AccessToken: accessToken,
		UserID:      user.UserID,
		ClientID:    client.ClientID,
		CreatedAt:   time.Now().UTC(),
		ExpiresIn:   time.Duration(time.Hour),
		Scope:       oauth.Scope(client.Rights),
	}
	err = is.store.OAuth.SaveAccessToken(accessData)
	a.So(err, should.BeNil)

	refreshData := &store.RefreshData{
		RefreshToken: "123",
		UserID:       user.UserID,
		ClientID:     client.ClientID,
		CreatedAt:    time.Now(),
		Scope:        oauth.Scope(client.Rights),
	}
	err = is.store.OAuth.SaveRefreshToken(refreshData)
	a.So(err, should.BeNil)

	_, err = is.userService.DeleteUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
}
