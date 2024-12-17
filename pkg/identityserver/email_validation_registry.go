// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/email/templates"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errValidationRequestForbidden = errors.DefinePermissionDenied(
	"validation_request_forbidden",
	"validation request forbidden because it was already sent, wait `{retry_interval}` before retrying",
)

type emailValidationRegistry struct {
	ttnpb.UnimplementedEmailValidationRegistryServer

	*IdentityServer
}

func (is *IdentityServer) refreshEmailValidation(
	ctx context.Context, usrID *ttnpb.UserIdentifiers, opts ...ttnpb.EmailValidationOverwrite,
) (*ttnpb.EmailValidation, error) {
	ttl := is.configFromContext(ctx).UserRegistration.ContactInfoValidation.TokenTTL
	expires := time.Now().Add(ttl)
	retryInterval := is.configFromContext(ctx).UserRegistration.ContactInfoValidation.RetryInterval

	var validation *ttnpb.EmailValidation
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		validation, err = st.GetRefreshableEmailValidation(ctx, usrID, retryInterval)
		if err != nil {
			return err
		}
		validation.ExpiresAt = timestamppb.New(expires)

		for _, opt := range opts {
			opt(validation)
		}

		if err = st.RefreshEmailValidation(ctx, validation); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return validation, err
}

func (is *IdentityServer) requestEmailValidation(
	ctx context.Context, usrID *ttnpb.UserIdentifiers, opts ...ttnpb.EmailValidationOverwrite,
) (*ttnpb.EmailValidation, error) {
	var validation *ttnpb.EmailValidation
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		// Attempts to find and refresh an existing validation. If it doesn't exist, validation is `nil`.
		validation, err = is.refreshEmailValidation(ctx, usrID)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if validation == nil {
			id, err := auth.GenerateID(ctx)
			if err != nil {
				return err
			}
			token, err := auth.GenerateKey(ctx)
			if err != nil {
				return err
			}
			usr, err := st.GetUser(ctx, usrID, []string{"primary_email_address"})
			if err != nil {
				return err
			}
			validation = &ttnpb.EmailValidation{
				Id:      id,
				Token:   token,
				Address: usr.PrimaryEmailAddress,
				ExpiresAt: timestamppb.New(
					time.Now().Add(is.configFromContext(ctx).UserRegistration.ContactInfoValidation.TokenTTL),
				),
			}

			for _, opt := range opts {
				opt(validation)
			}

			validation, err = st.CreateEmailValidation(ctx, validation)
			if err != nil {
				// Only one validation can exist at a time, so if it already exists, return a forbidden error.
				if errors.IsAlreadyExists(err) {
					return errValidationRequestForbidden.WithAttributes(
						"retry_interval",
						is.configFromContext(ctx).UserRegistration.ContactInfoValidation.RetryInterval,
					)
				}
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Prepare validateData outside of goroutine to avoid issues with range variable or races with unsetting the Token.
	validateData := &templates.ValidateData{
		EntityIdentifiers: usrID.GetEntityIdentifiers(),
		ID:                validation.Id,
		Token:             validation.Token,
		TTL:               time.Until(validation.ExpiresAt.AsTime()),
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"email", validation.Address,
		"Id", validation.Id,
	)).Info("Sending validation email")
	go is.SendTemplateEmailToUsers( // nolint:errcheck
		is.Component.FromRequestContext(ctx),
		ttnpb.GetNotificationTypeString(ttnpb.NotificationType_VALIDATE),
		func(_ context.Context, data email.TemplateData) (email.TemplateData, error) {
			validateData.TemplateData = data
			return validateData, nil
		},
		&ttnpb.User{PrimaryEmailAddress: validation.Address},
	)
	validation.Token = "" // Unset tokens after sending emails
	return validation, nil
}

func (evr *emailValidationRegistry) validateEmail(
	ctx context.Context,
	req *ttnpb.ValidateEmailRequest,
) (*emptypb.Empty, error) {
	err := evr.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		validation, err := st.GetEmailValidation(ctx, &ttnpb.EmailValidation{Id: req.Id, Token: req.Token})
		if err != nil {
			return err
		}
		return st.ExpireEmailValidation(ctx, validation)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (evr *emailValidationRegistry) RequestValidation(
	ctx context.Context,
	usrID *ttnpb.UserIdentifiers,
) (*ttnpb.EmailValidation, error) {
	err := rights.RequireUser(ctx, usrID, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}
	return evr.requestEmailValidation(ctx, usrID)
}

func (evr *emailValidationRegistry) Validate(
	ctx context.Context,
	req *ttnpb.ValidateEmailRequest,
) (*emptypb.Empty, error) {
	return evr.validateEmail(ctx, req)
}
