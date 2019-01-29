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

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (is *IdentityServer) requestContactInfoValidation(ctx context.Context, ids *ttnpb.EntityIdentifiers) (*ttnpb.ContactInfoValidation, error) {
	id, err := auth.GenerateID(ctx)
	if err != nil {
		return nil, err
	}
	var contactInfo []*ttnpb.ContactInfo
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		contactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, ids)
		return err
	})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	emailValidationsExpireAt := now.Add(24 * time.Hour)
	emailValidations := make(map[string]*ttnpb.ContactInfoValidation)
	for _, info := range contactInfo {
		if info.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL && info.ValidatedAt == nil {
			validation, ok := emailValidations[info.Value]
			if !ok {
				key, err := auth.GenerateKey(ctx)
				if err != nil {
					return nil, err
				}
				validation = &ttnpb.ContactInfoValidation{
					ID:        id,
					Token:     key,
					Entity:    ids,
					CreatedAt: &now,
					ExpiresAt: &emailValidationsExpireAt,
				}
				emailValidations[info.Value] = validation
			}
			validation.ContactInfo = append(validation.ContactInfo, info)
		}
	}
	var pendingContactInfo []*ttnpb.ContactInfo
	if len(emailValidations) > 0 {
		err := is.withDatabase(ctx, func(db *gorm.DB) (err error) {
			for email, validation := range emailValidations {
				validation, err = store.GetContactInfoStore(db).CreateValidation(ctx, validation)
				if err != nil {
					return err
				}
				log.FromContext(ctx).WithFields(log.Fields(
					"email", email,
					"token", validation.Token,
				)).Info("Created email validation token")
				emailValidations[email] = validation
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		// TODO: Send validation email (https://github.com/TheThingsNetwork/lorawan-stack/issues/72).
		for _, validation := range emailValidations {
			pendingContactInfo = append(pendingContactInfo, validation.ContactInfo...)
			validation.Token = "" // Unset tokens after sending emails
		}
	}

	return &ttnpb.ContactInfoValidation{
		ID:          id,
		Entity:      ids,
		ContactInfo: pendingContactInfo,
		CreatedAt:   &now,
	}, nil
}

func (is *IdentityServer) validateContactInfo(ctx context.Context, req *ttnpb.ContactInfoValidation) (*types.Empty, error) {
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetContactInfoStore(db).Validate(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

type contactInfoRegistry struct {
	*IdentityServer
}

func (cir *contactInfoRegistry) RequestValidation(ctx context.Context, ids *ttnpb.EntityIdentifiers) (*ttnpb.ContactInfoValidation, error) {
	return cir.requestContactInfoValidation(ctx, ids)
}

func (cir *contactInfoRegistry) Validate(ctx context.Context, req *ttnpb.ContactInfoValidation) (*types.Empty, error) {
	return cir.validateContactInfo(ctx, req)
}
