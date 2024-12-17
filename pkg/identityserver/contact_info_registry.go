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

var (
	errContactNoCollaborator = errors.DefineInvalidArgument(
		"contact_no_collaborator", "contact `{contact}` is not a collaborator",
	)
	errNoValidationNeeded = errors.DefineInvalidArgument(
		"no_validation_needed", "no validation needed for this contact info",
	)
	errValidationsAlreadySent = errors.DefineAlreadyExists(
		"validations_already_sent",
		"validations for this contact info already sent, wait `{retry_interval}` before retrying",
	)
)

// getContactsFromEntity fetches the administrative and technical contacts from the provided entity and returns a slice
// of ContactInfo pointers. The usage of this function should be restricted to the replacement of read operations
// related to the deprecated `ContactInfo`.
func getContactsFromEntity[
	X interface {
		GetAdministrativeContact() *ttnpb.OrganizationOrUserIdentifiers
		GetTechnicalContact() *ttnpb.OrganizationOrUserIdentifiers
	},
](
	ctx context.Context, id X, st store.Store,
) ([]*ttnpb.ContactInfo, error) {
	orgMask := []string{"administrative_contact", "technical_contact"}
	contacts := make([]*ttnpb.ContactInfo, 0, 2)

	fn := func(id *ttnpb.OrganizationOrUserIdentifiers, isAdminContact bool) error {
		usrID := id.GetUserIds()
		var contactType ttnpb.ContactType
		if !isAdminContact {
			contactType = ttnpb.ContactType_CONTACT_TYPE_TECHNICAL
		}

		if orgID := id.GetOrganizationIds(); orgID != nil {
			org, err := st.GetOrganization(ctx, orgID, orgMask)
			if err != nil {
				return err
			}

			if isAdminContact {
				usrID = org.AdministrativeContact.GetUserIds()
			} else {
				usrID = org.TechnicalContact.GetUserIds()
			}
		}

		usr, err := st.GetUser(ctx, usrID, []string{"primary_email_address"})
		if err != nil {
			return err
		}

		contacts = append(contacts, &ttnpb.ContactInfo{
			ContactMethod: ttnpb.ContactMethod_CONTACT_METHOD_EMAIL,
			ContactType:   contactType,
			ValidatedAt:   usr.PrimaryEmailAddressValidatedAt,
			Value:         usr.PrimaryEmailAddress,
		})

		return nil
	}

	if err := fn(id.GetAdministrativeContact(), true); err != nil {
		return nil, err
	}
	if err := fn(id.GetTechnicalContact(), false); err != nil {
		return nil, err
	}

	return contacts, nil
}

// validateContactInfoRestrictions fetches the auth info from the context and validates if the caller ID matches the
// provided `ids` in the parameters. The usage of this function should be restricted to testing the administrative and
// technical contacts in methods belonging to each entity registry.
func (is *IdentityServer) validateContactInfoRestrictions(
	ctx context.Context, ids ...*ttnpb.OrganizationOrUserIdentifiers,
) error {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return err
	}
	callerID := authInfo.GetOrganizationOrUserIdentifiers()
	if is.configFromContext(ctx).CollaboratorRights.SetOthersAsContacts || authInfo.IsAdmin {
		return nil
	}

	for _, id := range ids {
		if id == nil {
			continue
		}
		if callerID.EntityType() != id.EntityType() || callerID.IDString() != id.IDString() {
			return store.ErrContactInfoRestricted.New()
		}
	}
	return nil
}

func validateCollaboratorEqualsContact(collaborator, contact *ttnpb.OrganizationOrUserIdentifiers) error {
	if contact == nil {
		return nil
	}
	if collaborator.EntityType() != contact.EntityType() || collaborator.IDString() != contact.IDString() {
		return errContactNoCollaborator.WithAttributes("contact", contact.IDString())
	}
	return nil
}

func validateContactIsCollaborator(
	ctx context.Context,
	st store.Store,
	contact *ttnpb.OrganizationOrUserIdentifiers,
	entity *ttnpb.EntityIdentifiers,
) error {
	if contact == nil {
		return nil
	}
	_, err := st.GetMember(ctx, contact, entity)
	if err != nil {
		if errors.IsNotFound(err) {
			return errContactNoCollaborator.WithAttributes("contact", contact.IDString())
		}
		return err
	}
	return nil
}

func (is *IdentityServer) requestContactInfoValidation(
	ctx context.Context,
	ids *ttnpb.EntityIdentifiers,
) (*ttnpb.ContactInfoValidation, error) {
	// NOTE: This does NOT check auth. Internal use only.
	id, err := auth.GenerateID(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	ttl := is.configFromContext(ctx).UserRegistration.ContactInfoValidation.TokenTTL
	expires := now.Add(ttl)
	retryInterval := is.configFromContext(ctx).UserRegistration.ContactInfoValidation.RetryInterval

	var contactInfo []*ttnpb.ContactInfo
	resendValidation := make(map[string]*ttnpb.ContactInfoValidation)

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		contactInfo, err = st.GetContactInfo(ctx, ids)
		if err != nil {
			return err
		}

		validations, err := st.ListRefreshableValidations(ctx, ids, retryInterval)
		if err != nil {
			return err
		}
		for _, v := range validations {
			resendValidation[v.ContactInfo[0].Value] = v
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	emailValidations := make(map[string]*ttnpb.ContactInfoValidation)
	for _, info := range contactInfo {
		if info.ContactMethod == ttnpb.ContactMethod_CONTACT_METHOD_EMAIL && info.ValidatedAt == nil {
			validation, ok := emailValidations[info.Value]
			if !ok {
				key, err := auth.GenerateKey(ctx)
				if err != nil {
					return nil, err
				}
				validation = &ttnpb.ContactInfoValidation{
					Id:        id,
					Token:     key,
					Entity:    ids,
					CreatedAt: timestamppb.New(now),
					ExpiresAt: timestamppb.New(expires),
				}
				emailValidations[info.Value] = validation
			}
			validation.ContactInfo = append(validation.ContactInfo, info)
		}
	}
	if len(emailValidations) == 0 {
		return nil, errNoValidationNeeded.New()
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		for email, validation := range emailValidations {
			v, err := st.CreateValidation(ctx, validation)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					delete(emailValidations, email)
					continue
				}
				return err
			}
			log.FromContext(ctx).WithFields(log.Fields(
				"email", email,
				"token", v.Token,
			)).Info("Created email validation token")
			emailValidations[email] = v

			// Remove from resend list if a new validation was created.
			delete(resendValidation, email)
		}

		for email, validation := range resendValidation {
			if err := st.RefreshValidation(ctx, validation); err != nil {
				return err
			}
			emailValidations[email] = validation
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var pendingContactInfo []*ttnpb.ContactInfo
	for address, validation := range emailValidations {
		// Prepare validateData outside of goroutine to avoid issues with range variable or races with unsetting the Token.
		validateData := &templates.ValidateData{
			EntityIdentifiers: validation.Entity,
			ID:                validation.Id,
			Token:             validation.Token,
			TTL:               validation.ExpiresAt.AsTime().Sub(now),
		}
		go is.SendTemplateEmailToUsers( // nolint:errcheck
			is.FromRequestContext(ctx),
			ttnpb.GetNotificationTypeString(ttnpb.NotificationType_VALIDATE),
			func(_ context.Context, data email.TemplateData) (email.TemplateData, error) {
				validateData.TemplateData = data
				return validateData, nil
			},
			&ttnpb.User{PrimaryEmailAddress: address},
		)
		pendingContactInfo = append(pendingContactInfo, validation.ContactInfo...)
		validation.Token = "" // Unset tokens after sending emails
	}
	if len(pendingContactInfo) == 0 {
		return nil, errValidationsAlreadySent.WithAttributes("retry_interval", retryInterval)
	}

	return &ttnpb.ContactInfoValidation{
		Id:          id,
		Entity:      ids,
		ContactInfo: pendingContactInfo,
		CreatedAt:   timestamppb.New(now),
	}, nil
}

func (is *IdentityServer) validateContactInfo(
	ctx context.Context, req *ttnpb.ContactInfoValidation,
) (*emptypb.Empty, error) {
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		validation, err := st.GetValidation(ctx, req)
		if err != nil {
			return err
		}

		err = st.ValidateContactInfo(ctx, validation)
		if err != nil {
			return err
		}

		return st.ExpireValidation(ctx, validation)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

type contactInfoRegistry struct {
	ttnpb.UnimplementedContactInfoRegistryServer

	*IdentityServer
}

var errNoContactInfoForEntity = errors.DefineInvalidArgument("no_contact_info", "no contact info for this entity type")

func (cir *contactInfoRegistry) RequestValidation(
	ctx context.Context,
	ids *ttnpb.EntityIdentifiers,
) (*ttnpb.ContactInfoValidation, error) {
	var err error
	switch id := ids.GetIds().(type) {
	case *ttnpb.EntityIdentifiers_ApplicationIds:
		err = rights.RequireApplication(ctx, id.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC)
	case *ttnpb.EntityIdentifiers_ClientIds:
		err = rights.RequireClient(ctx, id.ClientIds, ttnpb.Right_RIGHT_CLIENT_SETTINGS_BASIC)
	case *ttnpb.EntityIdentifiers_GatewayIds:
		err = rights.RequireGateway(ctx, id.GatewayIds, ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC)
	case *ttnpb.EntityIdentifiers_OrganizationIds:
		err = rights.RequireOrganization(ctx, id.OrganizationIds, ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_BASIC)
	case *ttnpb.EntityIdentifiers_UserIds:
		err = rights.RequireUser(ctx, id.UserIds, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC)
	default:
		return nil, errNoContactInfoForEntity.New()
	}
	if err != nil {
		return nil, err
	}
	return cir.requestContactInfoValidation(ctx, ids)
}

func (cir *contactInfoRegistry) Validate(
	ctx context.Context,
	req *ttnpb.ContactInfoValidation,
) (*emptypb.Empty, error) {
	return cir.validateContactInfo(ctx, req)
}
