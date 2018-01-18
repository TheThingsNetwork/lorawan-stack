// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.IsAdminServer = new(adminService)

func TestSettings(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)
	defer is.store.Settings.Set(testSettings())

	ctx := testCtx()

	resp, err := is.adminService.GetSettings(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(resp, test.ShouldBeSettingsIgnoringAutoFields, testSettings())

	// modify settings
	_, err = is.adminService.UpdateSettings(ctx, &ttnpb.UpdateSettingsRequest{
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

	resp, err = is.GetSettings(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	a.So(resp.AllowedEmails, should.HaveLength, 0)
	a.So(resp.IdentityServerSettings_UserRegistrationFlow.SelfRegistration, should.BeTrue)
	a.So(resp.IdentityServerSettings_UserRegistrationFlow.SkipValidation, should.BeFalse)
}
