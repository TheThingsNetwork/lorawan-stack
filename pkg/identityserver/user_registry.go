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
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtCreateUser = events.Define("user.create", "Create user")
	evtUpdateUser = events.Define("user.update", "Update user")
	evtDeleteUser = events.Define("user.delete", "Delete user")
)

var (
	errInvitationTokenRequired = errors.DefineUnauthenticated("invitation_token_required", "invitation token required")
	errInvitationTokenExpired  = errors.DefineUnauthenticated("invitation_token_expired", "invitation token expired")
)

func (is *IdentityServer) createUser(ctx context.Context, req *ttnpb.CreateUserRequest) (usr *ttnpb.User, err error) {
	if err := blacklist.Check(ctx, req.UserID); err != nil {
		return nil, err
	}
	if req.InvitationToken == "" && is.configFromContext(ctx).UserInvitation.Required {
		return nil, errInvitationTokenRequired
	}

	hashedPassword, err := auth.Hash(req.User.Password)
	if err != nil {
		return nil, err
	}
	req.User.Password = string(hashedPassword)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		invitationStore := store.GetInvitationStore(db)
		if req.InvitationToken != "" {
			invitationToken, err := invitationStore.GetInvitation(ctx, req.InvitationToken)
			if err != nil {
				return err
			}
			if !invitationToken.ExpiresAt.IsZero() && invitationToken.ExpiresAt.Before(time.Now()) {
				return errInvitationTokenExpired
			}
		}

		usrStore := store.GetUserStore(db)
		usr, err = usrStore.CreateUser(ctx, &req.User)
		if err != nil {
			return err
		}

		if req.InvitationToken != "" {
			err = invitationStore.SetInvitationAcceptedBy(ctx, req.InvitationToken, &usr.UserIdentifiers)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	usr.Password = "" // Create doesn't have a FieldMask, so we need to manually remove the password.
	events.Publish(evtCreateUser(ctx, req.UserIdentifiers, nil))
	return usr, nil
}

func (is *IdentityServer) getUser(ctx context.Context, req *ttnpb.GetUserRequest) (usr *ttnpb.User, err error) {
	err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_INFO)
	if err != nil {
		return nil, err
	}
	// TODO: Filter FieldMask by Rights
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usrStore := store.GetUserStore(db)
		usr, err = usrStore.GetUser(ctx, &req.UserIdentifiers, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	return usr, nil
}

func (is *IdentityServer) updateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (usr *ttnpb.User, err error) {
	err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}
	// TODO: Filter FieldMask by Rights
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usrStore := store.GetUserStore(db)
		usr, err = usrStore.UpdateUser(ctx, &req.User, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateUser(ctx, req.UserIdentifiers, req.FieldMask.Paths))
	return usr, nil
}

var (
	errIncorrectPassword        = errors.DefineUnauthenticated("old_password", "incorrect old password")
	errTemporaryPasswordExpired = errors.DefineUnauthenticated("temporary_password_expired", "temporary password expired")
)

var (
	updatePasswordFieldMask    = &types.FieldMask{Paths: []string{"password"}}
	temporaryPasswordFieldMask = &types.FieldMask{Paths: []string{
		"password", "temporary_password", "temporary_password_created_at", "temporary_password_expires_at",
	}}
	updateTemporaryPasswordFieldMask = &types.FieldMask{Paths: []string{
		"temporary_password", "temporary_password_created_at", "temporary_password_expires_at",
	}}
)

func (is *IdentityServer) updateUserPassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*types.Empty, error) {
	err := rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_ALL)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := auth.Hash(req.New)
	if err != nil {
		return nil, err
	}
	updateMask := updatePasswordFieldMask
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usrStore := store.GetUserStore(db)
		usr, err := usrStore.GetUser(ctx, &req.UserIdentifiers, temporaryPasswordFieldMask)
		if err != nil {
			return err
		}
		valid, err := auth.Password(usr.Password).Validate(req.Old)
		if err != nil {
			return err
		}
		if !valid {
			if usr.TemporaryPassword == "" {
				return errIncorrectPassword
			}
			valid, err = auth.Password(usr.TemporaryPassword).Validate(req.Old)
			switch {
			case err != nil:
				return err
			case !valid:
				return errIncorrectPassword
			case usr.TemporaryPasswordExpiresAt.Before(time.Now()):
				return errTemporaryPasswordExpired
			}
			usr.TemporaryPassword, usr.TemporaryPasswordCreatedAt, usr.TemporaryPasswordExpiresAt = "", nil, nil
			updateMask = temporaryPasswordFieldMask
		}
		usr.Password = string(hashedPassword)
		usr, err = usrStore.UpdateUser(ctx, usr, updateMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateUser(ctx, req.UserIdentifiers, updateMask))
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
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usrStore := store.GetUserStore(db)
		usr, err := usrStore.GetUser(ctx, &req.UserIdentifiers, temporaryPasswordFieldMask)
		if err != nil {
			return err
		}
		if usr.TemporaryPasswordExpiresAt.After(time.Now()) {
			return errTemporaryPasswordStillValid
		}
		usr.TemporaryPassword = string(hashedTemporaryPassword)
		expires := now.Add(time.Hour)
		usr.TemporaryPasswordCreatedAt, usr.TemporaryPasswordExpiresAt = &now, &expires
		usr, err = usrStore.UpdateUser(ctx, usr, updateTemporaryPasswordFieldMask)
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
	// TODO: Send Email
	return ttnpb.Empty, nil
}

func (is *IdentityServer) deleteUser(ctx context.Context, ids *ttnpb.UserIdentifiers) (*types.Empty, error) {
	err := rights.RequireUser(ctx, *ids, ttnpb.RIGHT_USER_DELETE)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		usrStore := store.GetUserStore(db)
		err = usrStore.DeleteUser(ctx, ids)
		return err
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
