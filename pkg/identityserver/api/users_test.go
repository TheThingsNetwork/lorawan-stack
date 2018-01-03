// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package api_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/api"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/mock"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestUser(t *testing.T) {
	a := assertions.New(t)
	g := getGRPC(t)

	user := ttnpb.User{
		UserIdentifier: ttnpb.UserIdentifier{"daniel"},
		Password:       "12345",
		Email:          "foo@bar.com",
		Name:           "hi",
	}

	ctx := claims.NewContext(context.Background(), &auth.Claims{
		EntityID:  user.UserID,
		EntityTyp: auth.EntityUser,
		Source:    auth.Token,
		Rights:    ttnpb.AllUserRights,
	})

	// can't create an account using a not allowed email
	user.Email = "foo@foo.com"
	_, err := g.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
		User: user,
	})
	a.So(err, should.NotBeNil)
	a.So(ErrNotAllowedEmail.Describes(err), should.BeTrue)
	user.Email = "foo@bar.com"

	// can't create account using a blacklisted id
	for _, id := range settings.BlacklistedIDs {
		user.UserID = id
		_, err = g.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
			User: user,
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}
	user.UserID = "daniel"

	// create the account
	_, err = g.CreateUser(context.Background(), &ttnpb.CreateUserRequest{
		User: user,
	})
	a.So(err, should.BeNil)

	// can't retrieve profile without proper claims
	found, err := g.GetUser(context.Background(), &pbtypes.Empty{})
	a.So(found, should.BeNil)
	a.So(err, should.Equal, ErrNotAuthorized)

	// check that response doesnt include password within
	found, err = g.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.Name, should.Equal, user.Name)
	a.So(found.Password, should.HaveLength, 0)
	a.So(found.Email, should.Equal, user.Email)
	a.So(found.ValidatedAt.IsZero(), should.BeTrue)
	if settings.IdentityServerSettings_UserRegistrationFlow.AdminApproval {
		a.So(found.State, should.Equal, ttnpb.STATE_PENDING)
	} else {
		a.So(found.State, should.Equal, ttnpb.STATE_APPROVED)
	}

	// extract the validation token from the email and validate the user account
	data, ok := mock.Data().(map[string]interface{})
	if a.So(ok, should.BeTrue) && a.So(data["token"], should.NotBeEmpty) {
		token := data["token"].(string)

		_, err := g.ValidateUserEmail(context.Background(), &ttnpb.ValidateUserEmailRequest{
			Token: token,
		})
		a.So(err, should.BeNil)

		found, err := g.GetUser(ctx, &pbtypes.Empty{})
		a.So(err, should.BeNil)
		a.So(found.ValidatedAt.IsZero(), should.BeFalse)
	}

	// try to update the user password providing a wrong old password
	_, err = g.UpdateUserPassword(ctx, &ttnpb.UpdateUserPasswordRequest{
		New: "heheh",
	})
	a.So(err, should.NotBeNil)
	a.So(ErrPasswordsDoNotMatch.Describes(err), should.BeTrue)

	_, err = g.UpdateUserPassword(ctx, &ttnpb.UpdateUserPasswordRequest{
		Old: user.Password,
		New: "heheh",
	})
	a.So(err, should.BeNil)

	// generate a new API key
	key, err := g.GenerateUserAPIKey(ctx, &ttnpb.GenerateUserAPIKeyRequest{
		Name:   "foo",
		Rights: ttnpb.AllUserRights,
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllUserRights)

	// update api key
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = g.UpdateUserAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// can't generate another API Key with the same name
	_, err = g.GenerateUserAPIKey(ctx, &ttnpb.GenerateUserAPIKeyRequest{
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := g.ListUserAPIKeys(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = g.RemoveUserAPIKey(ctx, &ttnpb.RemoveUserAPIKeyRequest{
		Name: key.Name,
	})

	keys, err = g.ListUserAPIKeys(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// update the user's email
	_, err = g.UpdateUser(ctx, &ttnpb.UpdateUserRequest{
		User: ttnpb.User{
			Email: "newfoo@bar.com",
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"email"},
		},
	})
	a.So(err, should.BeNil)

	// check that the field validated_at has been reset
	found, err = g.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.ValidatedAt.IsZero(), should.BeTrue)

	token := ""

	// extract the token from mail
	data, ok = mock.Data().(map[string]interface{})
	if a.So(ok, should.BeTrue) && a.So(data["token"], should.NotBeEmpty) {
		token = data["token"].(string)
	}
	a.So(token, should.NotBeEmpty)

	// request a new validation token
	_, err = g.RequestUserEmailValidation(ctx, &pbtypes.Empty{})

	// check that the old validation token doesnt work because we requested a new one
	_, err = g.ValidateUserEmail(context.Background(), &ttnpb.ValidateUserEmailRequest{
		Token: token,
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrValidationTokenNotFound.Describes(err), should.BeTrue)

	// and therefore the email isn't validated yet
	found, err = g.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.ValidatedAt.IsZero(), should.BeTrue)

	// get the latest sent validation token
	data, ok = mock.Data().(map[string]interface{})
	if a.So(ok, should.BeTrue) {
		token = data["token"].(string)
	}
	a.So(token, should.NotBeEmpty)

	// validate the email
	_, err = g.ValidateUserEmail(context.Background(), &ttnpb.ValidateUserEmailRequest{
		Token: token,
	})
	a.So(err, should.BeNil)

	found, err = g.GetUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(found.UserIdentifier.UserID, should.Equal, user.UserID)
	a.So(found.ValidatedAt.IsZero(), should.BeFalse)

	// create a fake authorized client to the user
	client := &ttnpb.Client{
		ClientIdentifier: ttnpb.ClientIdentifier{"test-client"},
		Description:      "description",
		Secret:           "secret",
		Grants:           []ttnpb.GrantType{ttnpb.GRANT_PASSWORD},
		Rights:           []ttnpb.Right{ttnpb.Right(1)},
		RedirectURI:      "foo.ttn.dev/oauth",
		Creator:          ttnpb.UserIdentifier{user.UserID},
	}
	err = store.Clients.Create(client)
	a.So(err, should.BeNil)

	refreshData := &types.RefreshData{
		RefreshToken: "123",
		UserID:       user.UserID,
		ClientID:     client.ClientID,
		CreatedAt:    time.Now(),
	}
	err = store.OAuth.SaveRefreshToken(refreshData)
	a.So(err, should.BeNil)

	accessData := &types.AccessData{
		AccessToken: "456",
		UserID:      user.UserID,
		ClientID:    client.ClientID,
		CreatedAt:   time.Now(),
		ExpiresIn:   3600,
	}
	err = store.OAuth.SaveAccessToken(accessData)
	a.So(err, should.BeNil)

	clients, err := g.ListAuthorizedClients(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(clients.Clients, should.HaveLength, 1) {
		cli := clients.Clients[0]

		a.So(cli.ClientID, should.Equal, client.ClientID)
		a.So(cli.Description, should.Equal, client.Description)
		a.So(cli.Secret, should.HaveLength, 0)
		a.So(cli.RedirectURI, should.HaveLength, 0)
		a.So(cli.Grants, should.BeEmpty)
	}

	_, err = g.RevokeAuthorizedClient(ctx, &ttnpb.ClientIdentifier{"non-existent-client"})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAuthorizedClientNotFound.Describes(err), should.BeTrue)

	_, err = g.RevokeAuthorizedClient(ctx, &ttnpb.ClientIdentifier{client.ClientID})
	a.So(err, should.BeNil)

	clients, err = g.ListAuthorizedClients(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(clients.Clients, should.HaveLength, 0)

	_, err = g.DeleteUser(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
}
