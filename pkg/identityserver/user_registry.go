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

package identityserver

import (
	"context"
	"runtime/trace"
	"strings"
	"time"
	"unicode"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/emails"
	gormstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/validate"
)

var (
	evtCreateUser = events.Define(
		"user.create", "create user",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateUser = events.Define(
		"user.update", "update user",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_INFO),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteUser = events.Define(
		"user.delete", "delete user",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreUser = events.Define(
		"user.restore", "restore user",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeUser = events.Define(
		"user.purge", "purge user",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateUserIncorrectPassword = events.Define(
		"user.update.incorrect_password", "update user failure: incorrect password",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var (
	errUserRegistrationDisabled  = errors.DefineInvalidArgument("user_registration_disabled", "user registration disabled")
	errInvitationTokenRequired   = errors.DefineInvalidArgument("invitation_token_required", "invitation token required")
	errInvitationTokenExpired    = errors.DefineInvalidArgument("invitation_token_expired", "invitation token expired")
	errPasswordStrengthMinLength = errors.DefineInvalidArgument("password_strength_min_length", "need at least `{n}` characters")
	errPasswordStrengthMaxLength = errors.DefineInvalidArgument("password_strength_max_length", "need at most `{n}` characters")
	errPasswordStrengthUppercase = errors.DefineInvalidArgument("password_strength_uppercase", "need at least `{n}` uppercase letter(s)")
	errPasswordStrengthDigits    = errors.DefineInvalidArgument("password_strength_digits", "need at least `{n}` digit(s)")
	errPasswordStrengthSpecial   = errors.DefineInvalidArgument("password_strength_special", "need at least `{n}` special character(s)")
	errPasswordEqualsOld         = errors.DefineInvalidArgument("password_equals_old", "must not equal old password")
	errPasswordContainsUserID    = errors.DefineInvalidArgument("password_contains_user_id", "must not contain user ID")
	errCommonPassword            = errors.DefineInvalidArgument("common_password", "must not be too common")
	errAdminsPurgeUsers          = errors.DefinePermissionDenied("admins_purge_users", "users may only be purged by admins")
)

// Source: https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/10-million-password-list-top-10000.txt
// Filtered for passwords that are at least 8 characters long, and contain both numbers and letters.
var commonPasswords = []string{
	"1qaz2wsx", "trustno1", "1234qwer", "q1w2e3r4t5", "qwer1234", "q1w2e3r4", "1q2w3e4r", "jordan23", "abcd1234",
	"password1", "qwerty123", "1q2w3e4r5t", "rush2112", "passw0rd", "1qazxsw2", "blink182", "12qwaszx", "asdf1234",
	"1232323q", "12345qwert", "123456789a", "suckballz1", "qwerty12", "zaq12wsx", "ncc1701d", "hello123", "michael1",
	"123456789q", "123qweasd", "charlie1", "a1b2c3d4", "password123", "oso123aljg", "123qweasdzxc", "letmein1",
	"1234abcd", "qazwsx123", "mustang1", "freedom1", "fuckyou2", "1qaz2wsx3edc", "welcome1", "123qwe123", "wrinkle1",
	"access14", "babylon5", "yankees1", "q1w2e3r4t5y6", "jessica1", "ncc1701e", "super123", "letmein2", "a1234567",
	"gn56gn56", "matthew1", "anthony1", "satan666", "1q2w3e4r5t6y", "fuckyou1", "shaney14", "qwerty12345", "1234567a",
	"1a2b3c4d", "ailcreated5240", "william1", "1234567q", "zaq1xsw2", "zxcv1234", "formula1", "a1s2d3f4", "thunder1",
	"heather1", "chelsea1", "123456qwerty", "1234567890q", "richard1", "qwerty123456", "asshole1", "qwert123",
	"scooter1", "ncc1701a", "pa55word", "patrick1", "gateway1", "cowboys1", "agent007", "porsche9", "diamond1",
	"assword1", "1qaz1qaz", "pokemon1", "123456789z", "front242", "apollo13", "gordon24", "brandon1", "arsenal1",
	"123456aa", "raiders1", "ojdlg123aljg", "jackson1", "fordf150", "pa55w0rd", "melissa1", "kcj9wx5n", "happy123",
	"football1", "abc12345", "1qa2ws3ed", "rangers1", "p0015123", "nwo4life", "phoenix1", "pass1234", "chester1",
	"jasmine1", "r2d2c3po", "chicken1", "marino13", "apple123", "samsung1", "1x2zkg8w", "test1234", "a123456789",
	"america1", "12345678q", "qazwsx12", "qwerty1234", "montgom240", "12qw34er", "123qwerty", "1q2w3e4r5", "superman1",
	"zxcvbnm1", "james007", "12345qwe", "zxasqw12", "gfhjkm123", "packers1", "newpass6", "charles1", "12345678a",
	"shannon1", "madison1", "izdec0211", "nokia6300", "chicago1", "florida1", "baseball1", "123qq123", "1234567890a",
	"50spanks", "password2", "digital1", "123456qw", "z1x2c3v4", "jasnel12", "q2w3e4r5", "lineage2", "fuckoff1",
	"newyork1", "fishing1", "dragon12", "wg8e3wjf", "rebecca1", "ferrari1", "monster1", "crystal1", "winston1",
	"monkey12", "jackson5", "1234asdf", "panther1", "green123", "1a2s3d4f", "123456qwe", "gandalf1", "devil666",
	"9293709b13", "rainbow6", "qazwsxedc123", "scorpio1", "iverson3", "bulldog1", "master12", "ood123654", "dolphin1",
	"a12345678", "pussy123", "tiger123", "summer99", "playboy1", "michael2", "killer12", "iloveyou2", "zxcvbnm123",
	"pool6123", "mazdarx7", "hawaii50", "gabriel1", "1z2x3c4v", "yankees2", "tiffany1", "nascar24", "mazda626",
	"asdfgh01", "123456789s", "just4fun", "cameron1", "andyod22", "password12", "james123", "drummer1", "qwerty11",
	"qweasd123", "broncos1", "zxcasdqwe123", "soccer12", "soccer10", "qwert12345", "pumpkin1", "porsche1", "noname123",
	"death666", "12qw12qw", "angel123", "123456ru", "pufunga7782", "iloveyou1", "david123", "yamahar1", "spencer1",
	"marcius2", "ghbdtn123", "cygnusx1", "buddy123", "zachary1", "qwe123qwe", "mustang6", "jackass1", "ghhh47hj7649",
	"1234zxcv", "vikings1", "penguin1", "assword123", "12345qwerty", "shadow12", "private1", "nokian73", "hallo123",
	"cbr900rr", "asdqwe123", "warrior1", "nirvana1", "money123", "marines1", "cricket1", "chris123", "bubba123",
	"f00tball", "peaches1", "nokia6233", "maxwell1", "mash4077", "spartan1", "q123456789", "power123", "genesis1",
	"favorite6", "dodgers1", "awesome1", "12345qaz", "trouble1", "testing1", "summer69", "segblue2", "p0o9i8u7",
	"gsxr1000", "austin31", "23skidoo", "123qwert", "12345qwer", "12345abc", "123456789m", "voyager1", "sammy123",
	"rainbow1", "perfect1", "pantera1", "p4ssw0rd", "johnson1", "dragon69", "blue1234", "123456789qwe", "sabrina1",
	"q1234567", "ncc74656", "natasha1", "destiny1", "1qazzaq1", "1qazxsw23edc", "123456qqq", "123456789d", "stephen1",
	"liverpool1", "killer123", "buffalo1", "7777777a", "1passwor", "therock1", "success1", "password9", "eclipse1",
	"charlie2", "1qw23er4", "1q1q1q1q", "1234rewq", "weare138", "vanessa1", "patches1", "password99", "forever1",
	"captain1", "bubbles1",
}

func (is *IdentityServer) validatePasswordStrength(ctx context.Context, username, password string) error {
	requirements := is.configFromContext(ctx).UserRegistration.PasswordRequirements
	if len(password) < requirements.MinLength {
		return errPasswordStrengthMinLength.WithAttributes("n", requirements.MinLength)
	}
	if len(password) > requirements.MaxLength {
		return errPasswordStrengthMaxLength.WithAttributes("n", requirements.MaxLength)
	}
	var uppercase, digits, special int
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			uppercase++
		case unicode.IsDigit(r):
			digits++
		case !unicode.IsLetter(r) && !unicode.IsNumber(r):
			special++
		}
	}
	if uppercase < requirements.MinUppercase {
		return errPasswordStrengthUppercase.WithAttributes("n", requirements.MinUppercase)
	}
	if digits < requirements.MinDigits {
		return errPasswordStrengthDigits.WithAttributes("n", requirements.MinDigits)
	}
	if special < requirements.MinSpecial {
		return errPasswordStrengthSpecial.WithAttributes("n", requirements.MinSpecial)
	}
	if requirements.RejectUserID && strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return errPasswordContainsUserID.New()
	}
	if requirements.RejectCommon {
		for _, reject := range commonPasswords {
			if strings.ToLower(password) == reject {
				return errCommonPassword.New()
			}
		}
	}
	return nil
}

func (is *IdentityServer) createUser(ctx context.Context, req *ttnpb.CreateUserRequest) (usr *ttnpb.User, err error) {
	createdByAdmin := is.IsAdmin(ctx)
	config := is.configFromContext(ctx)

	if err = blacklist.Check(ctx, req.User.GetIds().GetUserId()); err != nil {
		return nil, err
	}
	if req.InvitationToken == "" && config.UserRegistration.Invitation.Required && !createdByAdmin {
		return nil, errInvitationTokenRequired.New()
	}

	if err := validate.Email(req.User.PrimaryEmailAddress); err != nil {
		return nil, err
	}
	if err := validateContactInfo(req.User.ContactInfo); err != nil {
		return nil, err
	}

	if !createdByAdmin {
		if !config.UserRegistration.Enabled {
			return nil, errUserRegistrationDisabled.New()
		}
		req.User.PrimaryEmailAddressValidatedAt = nil
		req.User.RequirePasswordUpdate = false
		if config.UserRegistration.AdminApproval.Required {
			req.User.State = ttnpb.State_STATE_REQUESTED
			req.User.StateDescription = "admin approval required"
		} else {
			req.User.State = ttnpb.State_STATE_APPROVED
		}
		req.User.Admin = false
		req.User.TemporaryPassword = ""
		req.User.TemporaryPasswordCreatedAt = nil
		req.User.TemporaryPasswordExpiresAt = nil
		cleanContactInfo(req.User.ContactInfo)
	}

	var primaryEmailAddressFound bool
	for _, contactInfo := range req.User.ContactInfo {
		if contactInfo.ContactMethod == ttnpb.ContactMethod_CONTACT_METHOD_EMAIL && contactInfo.Value == req.User.PrimaryEmailAddress {
			primaryEmailAddressFound = true
			if contactInfo.ValidatedAt != nil {
				req.User.PrimaryEmailAddressValidatedAt = contactInfo.ValidatedAt
				break
			}
		}
	}
	if !primaryEmailAddressFound {
		req.User.ContactInfo = append(req.User.ContactInfo, &ttnpb.ContactInfo{
			ContactMethod: ttnpb.ContactMethod_CONTACT_METHOD_EMAIL,
			Value:         req.User.PrimaryEmailAddress,
			ValidatedAt:   req.User.PrimaryEmailAddressValidatedAt,
		})
	}

	if err := is.validatePasswordStrength(ctx, req.User.GetIds().GetUserId(), req.User.Password); err != nil {
		return nil, err
	}
	hashedPassword, err := auth.Hash(ctx, req.User.Password)
	if err != nil {
		return nil, err
	}
	req.User.Password = hashedPassword
	req.User.PasswordUpdatedAt = ttnpb.ProtoTimePtr(time.Now())

	if req.User.ProfilePicture != nil {
		if err = is.processUserProfilePicture(ctx, req.User); err != nil {
			return nil, err
		}
	}
	defer func() { is.setFullProfilePictureURL(ctx, usr) }()

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if req.InvitationToken != "" {
			invitationToken, err := gormstore.GetInvitationStore(db).GetInvitation(ctx, req.InvitationToken)
			if err != nil {
				return err
			}
			if expiresAt := ttnpb.StdTime(invitationToken.ExpiresAt); expiresAt != nil && expiresAt.Before(time.Now()) {
				return errInvitationTokenExpired.New()
			}
		}

		usr, err = gormstore.GetUserStore(db).CreateUser(ctx, req.User)
		if err != nil {
			return err
		}

		if len(req.User.ContactInfo) > 0 {
			usr.ContactInfo, err = gormstore.GetContactInfoStore(db).SetContactInfo(ctx, usr.GetIds(), req.User.ContactInfo)
			if err != nil {
				return err
			}
		}

		if req.InvitationToken != "" {
			if err = gormstore.GetInvitationStore(db).SetInvitationAcceptedBy(ctx, req.InvitationToken, usr.GetIds()); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if usr.State == ttnpb.State_STATE_REQUESTED {
		err = is.SendAdminsEmail(ctx, func(data emails.Data) email.MessageData {
			data.Entity.Type, data.Entity.ID = "user", usr.GetIds().GetUserId()
			return &emails.UserRequested{
				Data: data,
			}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send user requested email")
		}
	}

	// TODO: Send welcome email (https://github.com/TheThingsNetwork/lorawan-stack/issues/72).

	if _, err := is.requestContactInfoValidation(ctx, req.User.GetIds().GetEntityIdentifiers()); err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send contact info validations")
	}

	usr.Password = "" // Create doesn't have a FieldMask, so we need to manually remove the password.
	events.Publish(evtCreateUser.NewWithIdentifiersAndData(ctx, req.User.GetIds(), nil))
	return usr, nil
}

func (is *IdentityServer) getUser(ctx context.Context, req *ttnpb.GetUserRequest) (usr *ttnpb.User, err error) {
	req.FieldMask = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = rights.RequireUser(ctx, *req.GetUserIds(), ttnpb.Right_RIGHT_USER_INFO); err != nil {
		if err := is.RequireAuthenticated(ctx); err != nil {
			return nil, err
		}
		if !ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicUserFields...) {
			return nil, err
		}
		defer func() { usr = usr.PublicSafe() }()
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "profile_picture") {
		if is.configFromContext(ctx).ProfilePicture.UseGravatar {
			if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "primary_email_address") {
				req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "primary_email_address")
				defer func() {
					if usr != nil {
						usr.PrimaryEmailAddress = ""
					}
				}()
			}
			defer func() { fillGravatar(ctx, usr) }()
		}
		defer func() { is.setFullProfilePictureURL(ctx, usr) }()
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usr, err = gormstore.GetUserStore(db).GetUser(ctx, req.GetUserIds(), req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			usr.ContactInfo, err = gormstore.GetContactInfoStore(db).GetContactInfo(ctx, usr.GetIds())
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return usr, nil
}

func (is *IdentityServer) listUsers(ctx context.Context, req *ttnpb.ListUsersRequest) (users *ttnpb.Users, err error) {
	req.FieldMask = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = is.RequireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	users = &ttnpb.Users{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		users.Users, err = gormstore.GetUserStore(db).FindUsers(paginateCtx, nil, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

var (
	errUpdateUserPasswordRequest = errors.DefineInvalidArgument("password_in_update", "can not update password with regular user update request")
	errUpdateUserAdminField      = errors.DefinePermissionDenied("user_update_admin_field", "only admins can update the `{field}` field")
)

func (is *IdentityServer) setFullProfilePictureURL(ctx context.Context, usr *ttnpb.User) {
	bucketURL := is.configFromContext(ctx).ProfilePicture.BucketURL
	if bucketURL == "" {
		return
	}
	bucketURL = strings.TrimSuffix(bucketURL, "/") + "/"
	if usr != nil && usr.ProfilePicture != nil {
		for size, file := range usr.ProfilePicture.Sizes {
			if !strings.Contains(file, "://") {
				usr.ProfilePicture.Sizes[size] = bucketURL + strings.TrimPrefix(file, "/")
			}
		}
	}
}

func (is *IdentityServer) updateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (usr *ttnpb.User, err error) {
	if err = rights.RequireUser(ctx, *req.User.GetIds(), ttnpb.Right_RIGHT_USER_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}
	updatedByAdmin := is.IsAdmin(ctx)

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "primary_email_address") {
		if err := validate.Email(req.User.PrimaryEmailAddress); err != nil {
			return nil, err
		}
	}
	if err := validateContactInfo(req.User.ContactInfo); err != nil {
		return nil, err
	}

	if !updatedByAdmin {
		for _, path := range req.FieldMask.Paths {
			switch path {
			case "primary_email_address_validated_at",
				"require_password_update",
				"state", "state_description", "admin",
				"temporary_password", "temporary_password_created_at", "temporary_password_expires_at":
				return nil, errUpdateUserAdminField.WithAttributes("field", path)
			}
		}
		req.User.PrimaryEmailAddressValidatedAt = nil
		cleanContactInfo(req.User.ContactInfo)
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state_description") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "state_description")
			req.User.StateDescription = ""
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "temporary_password") {
		hashedTemporaryPassword, err := auth.Hash(ctx, req.User.TemporaryPassword)
		if err != nil {
			return nil, err
		}
		req.User.TemporaryPassword = hashedTemporaryPassword
		now := time.Now()
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "temporary_password_created_at") {
			req.User.TemporaryPasswordCreatedAt = ttnpb.ProtoTimePtr(now)
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "temporary_password_created_at")
		}
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "temporary_password_expires_at") {
			req.User.TemporaryPasswordExpiresAt = ttnpb.ProtoTimePtr(now.Add(36 * time.Hour))
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "temporary_password_expires_at")
		}
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "profile_picture") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "profile_picture") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "profile_picture")
		}
		if req.User.ProfilePicture != nil {
			if err = is.processUserProfilePicture(ctx, req.User); err != nil {
				return nil, err
			}
		}
		defer func() { is.setFullProfilePictureURL(ctx, usr) }()
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		updatingContactInfo := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info")
		var contactInfo []*ttnpb.ContactInfo
		updatingPrimaryEmailAddress := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "primary_email_address")
		if updatingContactInfo || updatingPrimaryEmailAddress {
			if updatingContactInfo {
				contactInfo, err = gormstore.GetContactInfoStore(db).SetContactInfo(ctx, req.User.GetIds(), req.User.ContactInfo)
				if err != nil {
					return err
				}
			}
			if updatingPrimaryEmailAddress {
				if !updatingContactInfo {
					contactInfo, err = gormstore.GetContactInfoStore(db).GetContactInfo(ctx, req.User.GetIds())
					if err != nil {
						return err
					}
				}
				if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "primary_email_address_validated_at") {
					for _, contactInfo := range contactInfo {
						if contactInfo.ContactMethod == ttnpb.ContactMethod_CONTACT_METHOD_EMAIL && contactInfo.Value == req.User.PrimaryEmailAddress {
							req.User.PrimaryEmailAddressValidatedAt = contactInfo.ValidatedAt
							req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "primary_email_address_validated_at")
							break
						}
					}
				}
			}
		}
		usr, err = gormstore.GetUserStore(db).UpdateUser(ctx, req.User, req.FieldMask)
		if err != nil {
			return err
		}
		if updatingContactInfo {
			usr.ContactInfo = contactInfo
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateUser.NewWithIdentifiersAndData(ctx, req.User.GetIds(), req.FieldMask.GetPaths()))

	// TODO: Send emails (https://github.com/TheThingsNetwork/lorawan-stack/issues/72).
	// - If primary email address changed
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state") {
		err = is.SendUserEmail(ctx, req.User.GetIds(), func(data emails.Data) email.MessageData {
			data.SetEntity(req)
			return &emails.EntityStateChanged{
				Data:             data,
				State:            strings.ToLower(strings.TrimPrefix(usr.State.String(), "STATE_")),
				StateDescription: usr.StateDescription,
			}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send state change notification email")
		}
	}

	return usr, nil
}

var (
	errIncorrectPassword        = errors.DefineUnauthenticated("old_password", "incorrect old password")
	errTemporaryPasswordExpired = errors.DefineUnauthenticated("temporary_password_expired", "temporary password expired")
)

var (
	updatePasswordFieldMask = &pbtypes.FieldMask{Paths: []string{
		"password", "password_updated_at", "require_password_update",
	}}
	temporaryPasswordFieldMask = &pbtypes.FieldMask{Paths: []string{
		"password", "password_updated_at", "require_password_update",
		"temporary_password", "temporary_password_created_at", "temporary_password_expires_at",
	}}
	updateTemporaryPasswordFieldMask = &pbtypes.FieldMask{Paths: []string{
		"temporary_password", "temporary_password_created_at", "temporary_password_expires_at",
	}}
)

func (is *IdentityServer) updateUserPassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*pbtypes.Empty, error) {
	if err := is.validatePasswordStrength(ctx, req.GetUserIds().GetUserId(), req.New); err != nil {
		return nil, err
	}
	if req.Old == req.New {
		return nil, errPasswordEqualsOld.New()
	}
	hashedPassword, err := auth.Hash(ctx, req.New)
	if err != nil {
		return nil, err
	}
	updateMask := updatePasswordFieldMask
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		usr, err := gormstore.GetUserStore(db).GetUser(ctx, req.GetUserIds(), temporaryPasswordFieldMask)
		if err != nil {
			return err
		}
		region := trace.StartRegion(ctx, "validate old password")
		valid, err := auth.Validate(usr.Password, req.Old)
		region.End()
		if err != nil {
			return err
		}
		if valid {
			// TODO: Add when 2FA is enabled (https://github.com/TheThingsNetwork/lorawan-stack/issues/2)
			// if err := rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_ALL); err != nil {
			//	return err
			// }
		} else {
			if usr.TemporaryPassword == "" {
				events.Publish(evtUpdateUserIncorrectPassword.NewWithIdentifiersAndData(ctx, req.GetUserIds(), nil))
				return errIncorrectPassword.New()
			}
			trace.WithRegion(ctx, "validate temporary password", func() {
				valid, err = auth.Validate(usr.TemporaryPassword, req.Old)
			})
			if err != nil {
				return err
			}
			if !valid {
				events.Publish(evtUpdateUserIncorrectPassword.NewWithIdentifiersAndData(ctx, req.GetUserIds(), nil))
				return errIncorrectPassword.New()
			}
			if temporaryPasswordExpiresAt := ttnpb.StdTime(usr.TemporaryPasswordExpiresAt); temporaryPasswordExpiresAt != nil && temporaryPasswordExpiresAt.Before(time.Now()) {
				events.Publish(evtUpdateUserIncorrectPassword.NewWithIdentifiersAndData(ctx, req.GetUserIds(), nil))
				return errTemporaryPasswordExpired.New()
			}
			usr.TemporaryPassword, usr.TemporaryPasswordCreatedAt, usr.TemporaryPasswordExpiresAt = "", nil, nil
			updateMask = temporaryPasswordFieldMask
		}
		if req.RevokeAllAccess {
			sessionStore := gormstore.GetUserSessionStore(db)
			sessions, err := sessionStore.FindSessions(ctx, req.GetUserIds())
			if err != nil {
				return err
			}
			for _, session := range sessions {
				err = sessionStore.DeleteSession(ctx, req.GetUserIds(), session.SessionId)
				if err != nil {
					return err
				}
			}
			oauthStore := gormstore.GetOAuthStore(db)
			authorizations, err := oauthStore.ListAuthorizations(ctx, req.GetUserIds())
			if err != nil {
				return err
			}
			for _, auth := range authorizations {
				tokens, err := oauthStore.ListAccessTokens(ctx, auth.UserIds, auth.ClientIds)
				if err != nil {
					return err
				}
				for _, token := range tokens {
					err = oauthStore.DeleteAccessToken(ctx, token.Id)
					if err != nil {
						return err
					}
				}
			}
		}
		now := time.Now()
		usr.Password, usr.PasswordUpdatedAt, usr.RequirePasswordUpdate = hashedPassword, ttnpb.ProtoTimePtr(now), false
		usr, err = gormstore.GetUserStore(db).UpdateUser(ctx, usr, updateMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateUser.NewWithIdentifiersAndData(ctx, req.GetUserIds(), updateMask))
	err = is.SendUserEmail(ctx, req.GetUserIds(), func(data emails.Data) email.MessageData {
		return &emails.PasswordChanged{Data: data}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send password change notification email")
	}
	return ttnpb.Empty, nil
}

var errTemporaryPasswordStillValid = errors.DefineInvalidArgument("temporary_password_still_valid", "previous temporary password still valid")

func (is *IdentityServer) createTemporaryPassword(ctx context.Context, req *ttnpb.CreateTemporaryPasswordRequest) (*pbtypes.Empty, error) {
	temporaryPassword, err := auth.GenerateKey(ctx)
	if err != nil {
		return nil, err
	}
	hashedTemporaryPassword, err := auth.Hash(ctx, temporaryPassword)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	ttl := time.Hour
	expires := now.Add(ttl)
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		usr, err := gormstore.GetUserStore(db).GetUser(ctx, req.GetUserIds(), temporaryPasswordFieldMask)
		if err != nil {
			return err
		}
		if temporaryPasswordExpiresAt := ttnpb.StdTime(usr.TemporaryPasswordExpiresAt); temporaryPasswordExpiresAt != nil && temporaryPasswordExpiresAt.After(time.Now()) {
			return errTemporaryPasswordStillValid.New()
		}
		usr.TemporaryPassword = hashedTemporaryPassword
		usr.TemporaryPasswordCreatedAt, usr.TemporaryPasswordExpiresAt = ttnpb.ProtoTimePtr(now), ttnpb.ProtoTimePtr(expires)
		usr, err = gormstore.GetUserStore(db).UpdateUser(ctx, usr, updateTemporaryPasswordFieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"user_uid", unique.ID(ctx, req.GetUserIds()),
		"temporary_password", temporaryPassword,
	)).Info("Created temporary password")
	events.Publish(evtUpdateUser.NewWithIdentifiersAndData(ctx, req.GetUserIds(), updateTemporaryPasswordFieldMask))
	err = is.SendUserEmail(ctx, req.GetUserIds(), func(data emails.Data) email.MessageData {
		return &emails.TemporaryPassword{
			Data:              data,
			TemporaryPassword: temporaryPassword,
			TTL:               ttl,
		}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send temporary password email")
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) deleteUser(ctx context.Context, ids *ttnpb.UserIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireUser(ctx, *ids, ttnpb.Right_RIGHT_USER_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		err := gormstore.GetUserStore(db).DeleteUser(ctx, ids)
		if err != nil {
			return err
		}
		// Also delete the the user's sessions to enforce logouts.
		return gormstore.GetUserSessionStore(db).DeleteAllUserSessions(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteUser.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreUser(ctx context.Context, ids *ttnpb.UserIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireUser(store.WithSoftDeleted(ctx, false), *ids, ttnpb.Right_RIGHT_USER_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		usrStore := gormstore.GetUserStore(db)
		usr, err := usrStore.GetUser(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		deletedAt := ttnpb.StdTime(usr.DeletedAt)
		if deletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*deletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		return usrStore.RestoreUser(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreUser.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeUser(ctx context.Context, ids *ttnpb.UserIdentifiers) (*pbtypes.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeUsers.New()
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		err := gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		// delete related API keys before purging the user
		err = gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		err = gormstore.GetMembershipStore(db).DeleteAccountMembers(ctx, ids.GetOrganizationOrUserIdentifiers())
		if err != nil {
			return err
		}
		err = gormstore.GetOAuthStore(db).DeleteUserAuthorizations(ctx, ids)
		if err != nil {
			return err
		}
		err = gormstore.GetUserSessionStore(db).DeleteAllUserSessions(ctx, ids)
		if err != nil {
			return err
		}
		return gormstore.GetUserStore(db).PurgeUser(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeUser.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

type userRegistry struct {
	*IdentityServer
}

func (ur *userRegistry) Create(ctx context.Context, req *ttnpb.CreateUserRequest) (*ttnpb.User, error) {
	return ur.createUser(ctx, req)
}

func (ur *userRegistry) List(ctx context.Context, req *ttnpb.ListUsersRequest) (*ttnpb.Users, error) {
	return ur.listUsers(ctx, req)
}

func (ur *userRegistry) Get(ctx context.Context, req *ttnpb.GetUserRequest) (*ttnpb.User, error) {
	return ur.getUser(ctx, req)
}

func (ur *userRegistry) Update(ctx context.Context, req *ttnpb.UpdateUserRequest) (*ttnpb.User, error) {
	return ur.updateUser(ctx, req)
}

func (ur *userRegistry) UpdatePassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*pbtypes.Empty, error) {
	return ur.updateUserPassword(ctx, req)
}

func (ur *userRegistry) CreateTemporaryPassword(ctx context.Context, req *ttnpb.CreateTemporaryPasswordRequest) (*pbtypes.Empty, error) {
	return ur.createTemporaryPassword(ctx, req)
}

func (ur *userRegistry) Delete(ctx context.Context, req *ttnpb.UserIdentifiers) (*pbtypes.Empty, error) {
	return ur.deleteUser(ctx, req)
}

func (ur *userRegistry) Restore(ctx context.Context, req *ttnpb.UserIdentifiers) (*pbtypes.Empty, error) {
	return ur.restoreUser(ctx, req)
}

func (ur *userRegistry) Purge(ctx context.Context, req *ttnpb.UserIdentifiers) (*pbtypes.Empty, error) {
	return ur.purgeUser(ctx, req)
}
