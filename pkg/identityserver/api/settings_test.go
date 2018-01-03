// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package api_test

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSettings(t *testing.T) {
	a := assertions.New(t)
	g := getGRPC(t)
	defer store.Settings.Set(settings)

	user := testUsers()["alice"]

	ctx := claims.NewContext(context.Background(), &auth.Claims{
		EntityID:  user.UserID,
		EntityTyp: auth.EntityUser,
		Source:    auth.Token,
		Rights:    ttnpb.AllUserRights,
	})

	resp, err := g.GetSettings(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(resp, test.ShouldBeSettingsIgnoringAutoFields, settings)

	// modify settings
	_, err = g.UpdateSettings(ctx, &ttnpb.UpdateSettingsRequest{
		Settings: ttnpb.IdentityServerSettings{
			IdentityServerSettings_UserRegistrationFlow: ttnpb.IdentityServerSettings_UserRegistrationFlow{
				SelfRegistration: true,
				SkipValidation:   true,
			},
		},
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"allowed_emails", "user_registration.self_registration"},
		},
	})
	a.So(err, should.BeNil)

	resp, err = g.GetSettings(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(resp.AllowedEmails, should.HaveLength, 0)
	a.So(resp.IdentityServerSettings_UserRegistrationFlow.SelfRegistration, should.BeTrue)
}
