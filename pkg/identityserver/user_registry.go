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
	"bytes"
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/picture"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtCreateUser = events.Define("user.create", "Create user")
	evtUpdateUser = events.Define("user.update", "Update user")
	evtDeleteUser = events.Define("user.delete", "Delete user")

	evtUpdateUserIncorrectPassword = events.Define("user.update.incorrect_password", "Incorrect password for user update")
)

var (
	errInvitationTokenRequired = errors.DefineUnauthenticated("invitation_token_required", "invitation token required")
	errInvitationTokenExpired  = errors.DefineUnauthenticated("invitation_token_expired", "invitation token expired")
)

const maxProfilePictureStoredDimensions = 1024

func (is *IdentityServer) preprocessUserProfilePicture(usr *ttnpb.User) (err error) {
	if usr.ProfilePicture == nil {
		return
	}
	var original string
	if len(usr.ProfilePicture.Sizes) > 0 {
		original = usr.ProfilePicture.Sizes[0]
		if original == "" {
			var max uint32
			for size, url := range usr.ProfilePicture.Sizes {
				if size > max {
					max = size
					original = url
				}
			}
		}
	}
	if usr.ProfilePicture.Embedded != nil && len(usr.ProfilePicture.Embedded.Data) > 0 {
		usr.ProfilePicture, err = picture.MakeSquare(bytes.NewBuffer(usr.ProfilePicture.Embedded.Data), maxProfilePictureStoredDimensions)
		if err != nil {
			return err
		}
		// TODO: Upload to blob store, set original to path (https://github.com/TheThingsIndustries/lorawan-stack/issues/393).
	}
	if original != "" {
		usr.ProfilePicture.Sizes = map[uint32]string{0: original}
	} else {
		usr.ProfilePicture.Sizes = nil
	}
	// TODO: Schedule background processing (https://github.com/TheThingsIndustries/lorawan-stack/issues/393).
	return
}

func (is *IdentityServer) createUser(ctx context.Context, req *ttnpb.CreateUserRequest) (usr *ttnpb.User, err error) {
	createdByAdmin := is.UniversalRights(ctx).IncludesAll(ttnpb.RIGHT_USER_ALL)

	if err = blacklist.Check(ctx, req.UserID); err != nil {
		return nil, err
	}
	if req.InvitationToken == "" && is.configFromContext(ctx).UserRegistration.Invitation.Required && !createdByAdmin {
		return nil, errInvitationTokenRequired
	}

	var primaryEmailAddressFound bool
	for _, contactInfo := range req.User.ContactInfo {
		if !createdByAdmin {
			contactInfo.ValidatedAt = nil
		}
		if contactInfo.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL && contactInfo.Value == req.User.PrimaryEmailAddress {
			primaryEmailAddressFound = true
		}
	}
	if !primaryEmailAddressFound {
		req.User.ContactInfo = append(req.User.ContactInfo, &ttnpb.ContactInfo{
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         req.User.PrimaryEmailAddress,
		})
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

	if err = is.preprocessUserProfilePicture(&req.User); err != nil {
		return nil, err
	}

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
			cleanContactInfo(req.ContactInfo)
			usr.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, usr.EntityIdentifiers(), req.ContactInfo)
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

	// TODO: Send welcome email (https://github.com/TheThingsIndustries/lorawan-stack/issues/1395).

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
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_INFO); err != nil {
		if hasOnlyAllowedFields(topLevelFields(req.FieldMask.Paths), ttnpb.PublicUserFields) {
			defer func() { usr = usr.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usr, err = store.GetUserStore(db).GetUser(ctx, &req.UserIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if hasField(req.FieldMask.Paths, "contact_info") {
			usr.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, usr.EntityIdentifiers())
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

func (is *IdentityServer) updateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (usr *ttnpb.User, err error) {
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	updatedByAdmin := is.UniversalRights(ctx).IncludesAll(ttnpb.RIGHT_USER_ALL)

	for _, path := range req.FieldMask.Paths {
		switch path {
		case "password", "password_updated_at":
			return nil, errUpdateUserPasswordRequest
		}
	}

	if !updatedByAdmin {
		for _, path := range req.FieldMask.Paths {
			switch path {
			case "primary_email_address",
				"require_password_update",
				"state", "admin",
				"temporary_password", "temporary_password_created_at", "temporary_password_expires_at":
				return nil, errUpdateUserAdminField.WithAttributes("field", path)
			}
		}

		for _, contactInfo := range req.User.ContactInfo {
			contactInfo.ValidatedAt = nil
		}
	}

	if err = is.preprocessUserProfilePicture(&req.User); err != nil {
		return nil, err
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if hasField(req.FieldMask.Paths, "primary_email_address") {
			// TODO: if updating primary_email_address, get existing contact info and set primary_email_address_validated_at
			// depending on existing contact info. Until then, the primary email address can only be updated by admins.
			req.PrimaryEmailAddressValidatedAt = nil
			req.FieldMask.Paths = append(req.FieldMask.Paths, "primary_email_address_validated_at")
		}
		usr, err = store.GetUserStore(db).UpdateUser(ctx, &req.User, &req.FieldMask)
		if err != nil {
			return err
		}
		if hasField(req.FieldMask.Paths, "contact_info") {
			cleanContactInfo(req.ContactInfo)
			usr.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, usr.EntityIdentifiers(), req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateUser(ctx, req.UserIdentifiers, req.FieldMask.Paths))

	// TODO: Send emails (https://github.com/TheThingsIndustries/lorawan-stack/issues/1395).
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
		valid, err := auth.Password(usr.Password).Validate(req.Old)
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
			valid, err = auth.Password(usr.TemporaryPassword).Validate(req.Old)
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
	// TODO: Send password update email (https://github.com/TheThingsIndustries/lorawan-stack/issues/1395).
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
		if usr.TemporaryPasswordExpiresAt.After(time.Now()) {
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
	// TODO: Send temporary password email (https://github.com/TheThingsIndustries/lorawan-stack/issues/1395).
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
