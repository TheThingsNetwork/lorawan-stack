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
	"path"
	"runtime/trace"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

var (
	evtCreateUser = events.Define(
		"user.create", "create user",
		ttnpb.RIGHT_USER_INFO,
	)
	evtUpdateUser = events.Define(
		"user.update", "update user",
		ttnpb.RIGHT_USER_INFO,
	)
	evtDeleteUser = events.Define(
		"user.delete", "delete user",
		ttnpb.RIGHT_USER_INFO,
	)
	evtUpdateUserIncorrectPassword = events.Define(
		"user.update.incorrect_password", "update user failure: incorrect password",
		ttnpb.RIGHT_USER_INFO,
	)
)

var (
	errInvitationTokenRequired   = errors.DefineUnauthenticated("invitation_token_required", "invitation token required")
	errInvitationTokenExpired    = errors.DefineUnauthenticated("invitation_token_expired", "invitation token expired")
	errPasswordStrengthMinLength = errors.DefineInvalidArgument("password_strength_min_length", "need at least `{n}` characters")
	errPasswordStrengthMaxLength = errors.DefineInvalidArgument("password_strength_max_length", "need at most `{n}` characters")
	errPasswordStrengthUppercase = errors.DefineInvalidArgument("password_strength_uppercase", "need at least `{n}` uppercase letter(s)")
	errPasswordStrengthDigits    = errors.DefineInvalidArgument("password_strength_digits", "need at least `{n}` digit(s)")
	errPasswordStrengthSpecial   = errors.DefineInvalidArgument("password_strength_special", "need at least `{n}` special character(s)")
)

func (is *IdentityServer) validatePasswordStrength(ctx context.Context, password string) error {
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
		case r >= 'A' && r <= 'Z':
			uppercase++
		case r >= '0' && r <= '9':
			digits++
		case r == '!' || r == '%' || r == '@' || r == '#' || r == '$' || r == '&' || r == '*':
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
	return nil
}

func (is *IdentityServer) createUser(ctx context.Context, req *ttnpb.CreateUserRequest) (usr *ttnpb.User, err error) {
	createdByAdmin := is.IsAdmin(ctx)

	if err = blacklist.Check(ctx, req.UserID); err != nil {
		return nil, err
	}
	if req.InvitationToken == "" && is.configFromContext(ctx).UserRegistration.Invitation.Required && !createdByAdmin {
		return nil, errInvitationTokenRequired
	}

	if err := validate.Email(req.User.PrimaryEmailAddress); err != nil {
		return nil, err
	}
	if err := validateContactInfo(req.User.ContactInfo); err != nil {
		return nil, err
	}

	if !createdByAdmin {
		req.User.PrimaryEmailAddressValidatedAt = nil
		cleanContactInfo(req.User.ContactInfo)
	}
	var primaryEmailAddressFound bool
	for _, contactInfo := range req.User.ContactInfo {
		if contactInfo.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL && contactInfo.Value == req.User.PrimaryEmailAddress {
			primaryEmailAddressFound = true
			if contactInfo.ValidatedAt != nil {
				req.User.PrimaryEmailAddressValidatedAt = contactInfo.ValidatedAt
				break
			}
		}
	}
	if !primaryEmailAddressFound {
		req.User.ContactInfo = append(req.User.ContactInfo, &ttnpb.ContactInfo{
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         req.User.PrimaryEmailAddress,
			ValidatedAt:   req.User.PrimaryEmailAddressValidatedAt,
		})
	}

	if err := is.validatePasswordStrength(ctx, req.User.Password); err != nil {
		return nil, err
	}
	hashedPassword, err := auth.Hash(req.User.Password)
	if err != nil {
		return nil, err
	}
	req.User.Password = string(hashedPassword)
	req.User.PasswordUpdatedAt = time.Now()

	if !createdByAdmin {
		if is.configFromContext(ctx).UserRegistration.AdminApproval.Required {
			req.User.State = ttnpb.STATE_REQUESTED
		} else {
			req.User.State = ttnpb.STATE_APPROVED
		}
		req.User.Admin = false
	}

	if req.User.ProfilePicture != nil {
		if err = is.processUserProfilePicture(ctx, &req.User); err != nil {
			return nil, err
		}
	}
	defer func() { is.setFullProfilePictureURL(ctx, usr) }()

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if req.InvitationToken != "" {
			invitationToken, err := store.GetInvitationStore(db).GetInvitation(ctx, req.InvitationToken)
			if err != nil {
				return err
			}
			if !invitationToken.ExpiresAt.IsZero() && invitationToken.ExpiresAt.Before(time.Now()) {
				return errInvitationTokenExpired
			}
		}

		usr, err = store.GetUserStore(db).CreateUser(ctx, &req.User)
		if err != nil {
			return err
		}

		if len(req.ContactInfo) > 0 {
			usr.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, usr.UserIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}

		if req.InvitationToken != "" {
			if err = store.GetInvitationStore(db).SetInvitationAcceptedBy(ctx, req.InvitationToken, &usr.UserIdentifiers); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// TODO: Send welcome email (https://github.com/TheThingsNetwork/lorawan-stack/issues/72).

	if _, err := is.requestContactInfoValidation(ctx, req.UserIdentifiers.EntityIdentifiers()); err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send contact info validations")
	}

	usr.Password = "" // Create doesn't have a FieldMask, so we need to manually remove the password.
	events.Publish(evtCreateUser(ctx, req.UserIdentifiers, nil))
	return usr, nil
}

func (is *IdentityServer) getUser(ctx context.Context, req *ttnpb.GetUserRequest) (usr *ttnpb.User, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_INFO); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths, ttnpb.PublicUserFields...) {
			defer func() { usr = usr.PublicSafe() }()
		} else {
			return nil, err
		}
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.Paths), "profile_picture") {
		if is.configFromContext(ctx).ProfilePicture.UseGravatar {
			if !ttnpb.HasAnyField(req.FieldMask.Paths, "primary_email_address") {
				req.FieldMask.Paths = append(req.FieldMask.Paths, "primary_email_address")
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
		usr, err = store.GetUserStore(db).GetUser(ctx, &req.UserIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			usr.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, usr.UserIdentifiers)
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

var (
	errUpdateUserPasswordRequest = errors.DefineInvalidArgument("password_in_update", "can not update password with regular user update request")
	errUpdateUserAdminField      = errors.DefinePermissionDenied("user_update_admin_field", "only admins can update the `{field}` field")
)

func (is *IdentityServer) setFullProfilePictureURL(ctx context.Context, usr *ttnpb.User) {
	bucketURL := is.configFromContext(ctx).ProfilePicture.BucketURL
	if bucketURL == "" {
		return
	}
	if usr != nil && usr.ProfilePicture != nil {
		for size, file := range usr.ProfilePicture.Sizes {
			if !strings.Contains(file, "://") {
				usr.ProfilePicture.Sizes[size] = path.Join(bucketURL, file)
			}
		}
	}
}

func (is *IdentityServer) updateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (usr *ttnpb.User, err error) {
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask.Paths, nil, getPaths)
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = updatePaths
	}
	updatedByAdmin := is.IsAdmin(ctx)

	if ttnpb.HasAnyField(req.FieldMask.Paths, "password", "password_updated_at") {
		return nil, errUpdateUserPasswordRequest
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "primary_email_address") {
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
				"state", "admin",
				"temporary_password", "temporary_password_created_at", "temporary_password_expires_at":
				return nil, errUpdateUserAdminField.WithAttributes("field", path)
			}
		}
		cleanContactInfo(req.User.ContactInfo)
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.Paths), "profile_picture") {
		if !ttnpb.HasAnyField(req.FieldMask.Paths, "profile_picture") {
			req.FieldMask.Paths = append(req.FieldMask.Paths, "profile_picture")
		}
		if req.User.ProfilePicture != nil {
			if err = is.processUserProfilePicture(ctx, &req.User); err != nil {
				return nil, err
			}
		}
		defer func() { is.setFullProfilePictureURL(ctx, usr) }()
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		updatingContactInfo := ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info")
		var contactInfo []*ttnpb.ContactInfo
		updatingPrimaryEmailAddress := ttnpb.HasAnyField(req.FieldMask.Paths, "primary_email_address")
		if updatingContactInfo || updatingPrimaryEmailAddress {
			if updatingContactInfo {
				contactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, req.User.UserIdentifiers, req.ContactInfo)
				if err != nil {
					return err
				}
				contactInfo = usr.ContactInfo
			}
			if updatingPrimaryEmailAddress {
				if !updatingContactInfo {
					contactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, req.User.UserIdentifiers)
					if err != nil {
						return err
					}
				}
				req.PrimaryEmailAddressValidatedAt = nil
				if !ttnpb.HasAnyField(req.FieldMask.Paths, "primary_email_address_validated_at") {
					req.FieldMask.Paths = append(req.FieldMask.Paths, "primary_email_address_validated_at")
				}
				for _, contactInfo := range contactInfo {
					if contactInfo.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL && contactInfo.Value == req.User.PrimaryEmailAddress {
						req.PrimaryEmailAddressValidatedAt = contactInfo.ValidatedAt
						break
					}
				}
			}
		}
		usr, err = store.GetUserStore(db).UpdateUser(ctx, &req.User, &req.FieldMask)
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
	events.Publish(evtUpdateUser(ctx, req.UserIdentifiers, req.FieldMask.Paths))

	// TODO: Send emails (https://github.com/TheThingsNetwork/lorawan-stack/issues/72).
	// - If user state changed (approved, rejected, flagged, suspended)
	// - If primary email address changed

	return usr, nil
}

var (
	errIncorrectPassword        = errors.DefineUnauthenticated("old_password", "incorrect old password")
	errTemporaryPasswordExpired = errors.DefineUnauthenticated("temporary_password_expired", "temporary password expired")
)

var (
	updatePasswordFieldMask = &types.FieldMask{Paths: []string{
		"password", "password_updated_at", "require_password_update",
	}}
	temporaryPasswordFieldMask = &types.FieldMask{Paths: []string{
		"password", "password_updated_at", "require_password_update",
		"temporary_password", "temporary_password_created_at", "temporary_password_expires_at",
	}}
	updateTemporaryPasswordFieldMask = &types.FieldMask{Paths: []string{
		"temporary_password", "temporary_password_created_at", "temporary_password_expires_at",
	}}
)

func (is *IdentityServer) updateUserPassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*types.Empty, error) {
	if err := is.validatePasswordStrength(ctx, req.New); err != nil {
		return nil, err
	}
	hashedPassword, err := auth.Hash(req.New)
	if err != nil {
		return nil, err
	}
	updateMask := updatePasswordFieldMask
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		usr, err := store.GetUserStore(db).GetUser(ctx, &req.UserIdentifiers, temporaryPasswordFieldMask)
		if err != nil {
			return err
		}
		if req.RevokeAllAccess == true {
			sessionStore := store.GetUserSessionStore(db)
			sessions, err := sessionStore.FindSessions(ctx, &req.UserIdentifiers)
			if err != nil {
				return err
			}
			for _, session := range sessions {
				err = sessionStore.DeleteSession(ctx, &req.UserIdentifiers, session.SessionID)
			}
			oauthStore := store.GetOAuthStore(db)
			authorizations, err := oauthStore.ListAuthorizations(ctx, &req.UserIdentifiers)
			if err != nil {
				return err
			}

			for _, auth := range authorizations {
				tokens, err := oauthStore.ListAccessTokens(ctx, &auth.UserIDs, &auth.ClientIDs)
				if err != nil {
					return err
				}
				for _, token := range tokens {
					err = oauthStore.DeleteAccessToken(ctx, token.ID)
					if err != nil {
						return err
					}
				}
			}
		}
		region := trace.StartRegion(ctx, "validate old password")
		valid, err := auth.Password(usr.Password).Validate(req.Old)
		region.End()
		if err != nil {
			return err
		}
		if valid {
			if err := rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_ALL); err != nil {
				return err
			}
		} else {
			if usr.TemporaryPassword == "" {
				events.Publish(evtUpdateUserIncorrectPassword(ctx, req.UserIdentifiers, nil))
				return errIncorrectPassword
			}
			region := trace.StartRegion(ctx, "validate temporary password")
			valid, err = auth.Password(usr.TemporaryPassword).Validate(req.Old)
			region.End()
			switch {
			case err != nil:
				return err
			case !valid:
				events.Publish(evtUpdateUserIncorrectPassword(ctx, req.UserIdentifiers, nil))
				return errIncorrectPassword
			case usr.TemporaryPasswordExpiresAt.Before(time.Now()):
				events.Publish(evtUpdateUserIncorrectPassword(ctx, req.UserIdentifiers, nil))
				return errTemporaryPasswordExpired
			}
			usr.TemporaryPassword, usr.TemporaryPasswordCreatedAt, usr.TemporaryPasswordExpiresAt = "", nil, nil
			updateMask = temporaryPasswordFieldMask
		}
		usr.Password, usr.PasswordUpdatedAt, usr.RequirePasswordUpdate = string(hashedPassword), time.Now(), false
		usr, err = store.GetUserStore(db).UpdateUser(ctx, usr, updateMask)
		return err
	})
	if err != nil {
		return nil, err
	}

	events.Publish(evtUpdateUser(ctx, req.UserIdentifiers, updateMask))
	err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
		return &emails.PasswordChanged{Data: data}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send password change notification email")
	}
	return ttnpb.Empty, nil
}

var errTemporaryPasswordStillValid = errors.DefineInvalidArgument("temporary_password_still_valid", "previous temporary password still valid")

func (is *IdentityServer) createTemporaryPassword(ctx context.Context, req *ttnpb.CreateTemporaryPasswordRequest) (*types.Empty, error) {
	now := time.Now()
	temporaryPassword, err := auth.GenerateKey(ctx)
	if err != nil {
		return nil, err
	}
	hashedTemporaryPassword, err := auth.Hash(temporaryPassword)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		usr, err := store.GetUserStore(db).GetUser(ctx, &req.UserIdentifiers, temporaryPasswordFieldMask)
		if err != nil {
			return err
		}
		if usr.TemporaryPasswordExpiresAt != nil && usr.TemporaryPasswordExpiresAt.After(time.Now()) {
			return errTemporaryPasswordStillValid
		}
		usr.TemporaryPassword = string(hashedTemporaryPassword)
		expires := now.Add(time.Hour)
		usr.TemporaryPasswordCreatedAt, usr.TemporaryPasswordExpiresAt = &now, &expires
		usr, err = store.GetUserStore(db).UpdateUser(ctx, usr, updateTemporaryPasswordFieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"user_uid", unique.ID(ctx, req.UserIdentifiers),
		"temporary_password", temporaryPassword,
	)).Info("Created temporary password")
	events.Publish(evtUpdateUser(ctx, req.UserIdentifiers, updateTemporaryPasswordFieldMask))
	err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
		return &emails.TemporaryPassword{
			Data:              data,
			TemporaryPassword: temporaryPassword,
		}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send temporary password email")
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) deleteUser(ctx context.Context, ids *ttnpb.UserIdentifiers) (*types.Empty, error) {
	if err := rights.RequireUser(ctx, *ids, ttnpb.RIGHT_USER_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetUserStore(db).DeleteUser(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteUser(ctx, ids, nil))
	return ttnpb.Empty, nil
}

type userRegistry struct {
	*IdentityServer
}

func (ur *userRegistry) Create(ctx context.Context, req *ttnpb.CreateUserRequest) (*ttnpb.User, error) {
	return ur.createUser(ctx, req)
}

func (ur *userRegistry) Get(ctx context.Context, req *ttnpb.GetUserRequest) (*ttnpb.User, error) {
	return ur.getUser(ctx, req)
}

func (ur *userRegistry) Update(ctx context.Context, req *ttnpb.UpdateUserRequest) (*ttnpb.User, error) {
	return ur.updateUser(ctx, req)
}

func (ur *userRegistry) UpdatePassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*types.Empty, error) {
	return ur.updateUserPassword(ctx, req)
}

func (ur *userRegistry) CreateTemporaryPassword(ctx context.Context, req *ttnpb.CreateTemporaryPasswordRequest) (*types.Empty, error) {
	return ur.createTemporaryPassword(ctx, req)
}

func (ur *userRegistry) Delete(ctx context.Context, req *ttnpb.UserIdentifiers) (*types.Empty, error) {
	return ur.deleteUser(ctx, req)
}
