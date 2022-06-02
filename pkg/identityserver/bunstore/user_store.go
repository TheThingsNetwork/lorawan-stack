// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"context"
	"sort"
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// User is the user model in the database.
type User struct {
	bun.BaseModel `bun:"table:users,alias:usr,select:user_accounts"`

	Model
	SoftDelete

	Account EmbeddedAccount `bun:"embed:account_"`

	Name        string `bun:"name,nullzero"`
	Description string `bun:"description,nullzero"`

	Attributes []*Attribute `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	ContactInfo []*ContactInfo `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	PrimaryEmailAddress            string     `bun:"primary_email_address,notnull"`
	PrimaryEmailAddressValidatedAt *time.Time `bun:"primary_email_address_validated_at"`

	Password              string    `bun:"password,notnull"`
	PasswordUpdatedAt     time.Time `bun:"password_updated_at,notnull"`
	RequirePasswordUpdate bool      `bun:"require_password_update,notnull"`

	State            int    `bun:"state,notnull"`
	StateDescription string `bun:"state_description,nullzero"`

	Admin bool `bun:"admin,notnull"`

	TemporaryPassword          string     `bun:"temporary_password,nullzero"`
	TemporaryPasswordCreatedAt *time.Time `bun:"temporary_password_created_at"`
	TemporaryPasswordExpiresAt *time.Time `bun:"temporary_password_expires_at"`

	ProfilePictureID *string  `bun:"profile_picture_id"`
	ProfilePicture   *Picture `bun:"rel:belongs-to,join:profile_picture_id=id"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func userToPB(m *User, fieldMask ...string) (*ttnpb.User, error) {
	pb := &ttnpb.User{
		Ids: &ttnpb.UserIdentifiers{
			UserId: m.Account.UID,
		},

		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
		DeletedAt: ttnpb.ProtoTime(m.DeletedAt),

		Name:        m.Name,
		Description: m.Description,

		PrimaryEmailAddress:            m.PrimaryEmailAddress,
		PrimaryEmailAddressValidatedAt: ttnpb.ProtoTime(m.PrimaryEmailAddressValidatedAt),

		Password:              m.Password,
		PasswordUpdatedAt:     ttnpb.ProtoTimePtr(m.PasswordUpdatedAt),
		RequirePasswordUpdate: m.RequirePasswordUpdate,

		State:            ttnpb.State(m.State),
		StateDescription: m.StateDescription,

		Admin: m.Admin,

		TemporaryPassword:          m.TemporaryPassword,
		TemporaryPasswordCreatedAt: ttnpb.ProtoTime(m.TemporaryPasswordCreatedAt),
		TemporaryPasswordExpiresAt: ttnpb.ProtoTime(m.TemporaryPasswordExpiresAt),
	}

	if len(m.Attributes) > 0 {
		pb.Attributes = make(map[string]string, len(m.Attributes))
		for _, a := range m.Attributes {
			pb.Attributes[a.Key] = a.Value
		}
	}

	if len(m.ContactInfo) > 0 {
		pb.ContactInfo = make([]*ttnpb.ContactInfo, len(m.ContactInfo))
		for i, contactInfo := range m.ContactInfo {
			pb.ContactInfo[i] = contactInfoToPB(contactInfo)
		}
		sort.Sort(contactInfoProtoSlice(pb.ContactInfo))
	}

	if m.ProfilePicture != nil {
		picture, err := pictureToPB(m.ProfilePicture)
		if err != nil {
			return nil, err
		}
		pb.ProfilePicture = picture
	}

	if len(fieldMask) == 0 {
		return pb, nil
	}

	res := &ttnpb.User{}
	if err := res.SetFields(pb, fieldMask...); err != nil {
		return nil, err
	}

	// Set fields that are always present.
	res.Ids = pb.Ids
	res.CreatedAt = pb.CreatedAt
	res.UpdatedAt = pb.UpdatedAt
	res.DeletedAt = pb.DeletedAt

	return res, nil
}

type userStore struct {
	*baseStore
}

func newUserStore(baseStore *baseStore) *userStore {
	return &userStore{
		baseStore: baseStore,
	}
}

func (s *userStore) CreateUser(ctx context.Context, pb *ttnpb.User) (*ttnpb.User, error) {
	ctx, span := tracer.Start(ctx, "CreateUser", trace.WithAttributes(
		attribute.String("user_id", pb.GetIds().GetUserId()),
	))
	defer span.End()

	userModel := &User{
		Account: EmbeddedAccount{
			UID: pb.GetIds().GetUserId(),
		},
		Name:                           pb.Name,
		Description:                    pb.Description,
		PrimaryEmailAddress:            pb.PrimaryEmailAddress,
		PrimaryEmailAddressValidatedAt: ttnpb.StdTime(pb.PrimaryEmailAddressValidatedAt),
		Password:                       pb.Password,
		PasswordUpdatedAt:              ttnpb.StdTimeOrZero(pb.PasswordUpdatedAt),
		RequirePasswordUpdate:          pb.RequirePasswordUpdate,
		State:                          int(pb.State),
		StateDescription:               pb.StateDescription,
		Admin:                          pb.Admin,
		TemporaryPassword:              pb.TemporaryPassword,
		TemporaryPasswordCreatedAt:     ttnpb.StdTime(pb.TemporaryPasswordCreatedAt),
		TemporaryPasswordExpiresAt:     ttnpb.StdTime(pb.TemporaryPasswordExpiresAt),
	}

	if pb.ProfilePicture != nil {
		picture, err := pictureFromPB(ctx, pb.ProfilePicture)
		if err != nil {
			return nil, err
		}
		userModel.ProfilePicture = picture

		_, err = s.DB.NewInsert().
			Model(userModel.ProfilePicture).
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}

		userModel.ProfilePictureID = &userModel.ProfilePicture.ID
	}

	// Run user+account creation in a transaction if we're not already in one.
	err := s.transact(ctx, func(ctx context.Context, tx bun.IDB) error {
		_, err := tx.NewInsert().
			Model(userModel).
			Exec(ctx)
		if err != nil {
			return err
		}

		accountModel := &Account{
			UID:         pb.GetIds().GetUserId(),
			AccountType: "user",
			AccountID:   userModel.ID,
		}
		_, err = tx.NewInsert().
			Model(accountModel).
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, wrapDriverError(err)
	}

	if len(pb.Attributes) > 0 {
		userModel.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "user", userModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.ContactInfo) > 0 {
		userModel.ContactInfo, err = s.replaceContactInfo(
			ctx, nil, pb.ContactInfo, "user", userModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	pb, err = userToPB(userModel)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (*userStore) selectWithFields(fieldMask store.FieldMask) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		// TODO: Instead of selecting everything, select only columns that are part of the field mask.
		q = q.
			ExcludeColumn().
			Column("account_uid")
		if fieldMask.Contains("attributes") {
			q = q.Relation("Attributes")
		}
		if fieldMask.Contains("contact_info") {
			q = q.Relation("ContactInfo")
		}
		if fieldMask.Contains("profile_picture") {
			q = q.Relation("ProfilePicture")
		}
		return q
	}
}

func (s *userStore) listUsersBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.User, error) {
	models := []*User{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "account_uid", map[string]string{
			"user_id":               "account_uid",
			"name":                  "name",
			"primary_email_address": "primary_email_address",
			"state":                 "state",
			"admin":                 "admin",
			"created_at":            "created_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx)).
		Apply(s.selectWithFields(fieldMask))

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.User, len(models))
	for i, model := range models {
		pb, err := userToPB(model, fieldMask...)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *userStore) FindUsers(
	ctx context.Context, ids []*ttnpb.UserIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.User, error) {
	ctx, span := tracer.Start(ctx, "FindUsers", trace.WithAttributes(
		attribute.StringSlice("user_ids", idStrings(ids...)),
	))
	defer span.End()

	return s.listUsersBy(ctx, selectWithEmbeddedAccountUIDs(ctx, ids...), fieldMask)
}

func (s *userStore) ListAdmins(
	ctx context.Context, fieldMask store.FieldMask,
) ([]*ttnpb.User, error) {
	ctx, span := tracer.Start(ctx, "ListAdmins")
	defer span.End()

	return s.listUsersBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("?TableAlias.admin = true")
	}, fieldMask)
}

func (s *userStore) getUserModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) (*User, error) {
	model := &User{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by).
		Apply(s.selectWithFields(fieldMask))

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, wrapDriverError(err)
	}

	return model, nil
}

func (s *userStore) GetUser(
	ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.User, error) {
	ctx, span := tracer.Start(ctx, "GetUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(ctx, selectWithEmbeddedAccountUID(ctx, id), fieldMask)
	if err != nil {
		return nil, err
	}
	pb, err := userToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (*userStore) selectWithPrimaryEmailAddress(
	ctx context.Context, primaryEmailAddress string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		return q.Where("LOWER(?TableAlias.primary_email_address) = LOWER(?)", primaryEmailAddress)
	}
}

func (s *userStore) GetUserByPrimaryEmailAddress(
	ctx context.Context, primaryEmailAddress string, fieldMask store.FieldMask,
) (*ttnpb.User, error) {
	ctx, span := tracer.Start(ctx, "GetUserByPrimaryEmailAddress", trace.WithAttributes(
		attribute.String("primary_email_address", primaryEmailAddress),
	))
	defer span.End()

	model, err := s.getUserModelBy(
		ctx, s.selectWithPrimaryEmailAddress(ctx, primaryEmailAddress), fieldMask,
	)
	if err != nil {
		return nil, err
	}
	pb, err := userToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *userStore) UpdateUser( //nolint:gocyclo
	ctx context.Context, pb *ttnpb.User, fieldMask store.FieldMask,
) (*ttnpb.User, error) {
	ctx, span := tracer.Start(ctx, "UpdateUser", trace.WithAttributes(
		attribute.String("user_id", pb.GetIds().GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(ctx, selectWithEmbeddedAccountUID(ctx, pb.GetIds()), fieldMask)
	if err != nil {
		return nil, err
	}

	columns := []string{"updated_at"}

	for _, field := range fieldMask {
		switch field {
		case "name":
			model.Name = pb.Name
			columns = append(columns, "name")

		case "description":
			model.Description = pb.Description
			columns = append(columns, "description")

		case "attributes":
			model.Attributes, err = s.replaceAttributes(
				ctx, model.Attributes, pb.Attributes, "user", model.ID,
			)
			if err != nil {
				return nil, err
			}

		case "contact_info":
			model.ContactInfo, err = s.replaceContactInfo(
				ctx, model.ContactInfo, pb.ContactInfo, "user", model.ID,
			)
			if err != nil {
				return nil, err
			}

		case "primary_email_address":
			model.PrimaryEmailAddress = pb.PrimaryEmailAddress
			columns = append(columns, "primary_email_address")

		case "primary_email_address_validated_at":
			model.PrimaryEmailAddressValidatedAt = ttnpb.StdTime(pb.PrimaryEmailAddressValidatedAt)
			columns = append(columns, "primary_email_address_validated_at")

		case "password":
			model.Password = pb.Password
			columns = append(columns, "password")

		case "password_updated_at":
			model.PasswordUpdatedAt = ttnpb.StdTimeOrZero(pb.PasswordUpdatedAt)
			columns = append(columns, "password_updated_at")

		case "require_password_update":
			model.RequirePasswordUpdate = pb.RequirePasswordUpdate
			columns = append(columns, "require_password_update")

		case "state":
			model.State = int(pb.State)
			columns = append(columns, "state")

		case "state_description":
			model.StateDescription = pb.StateDescription
			columns = append(columns, "state_description")

		case "admin":
			model.Admin = pb.Admin
			columns = append(columns, "admin")

		case "temporary_password":
			model.TemporaryPassword = pb.TemporaryPassword
			columns = append(columns, "temporary_password")

		case "temporary_password_created_at":
			model.TemporaryPasswordCreatedAt = ttnpb.StdTime(pb.TemporaryPasswordCreatedAt)
			columns = append(columns, "temporary_password_created_at")

		case "temporary_password_expires_at":
			model.TemporaryPasswordExpiresAt = ttnpb.StdTime(pb.TemporaryPasswordExpiresAt)
			columns = append(columns, "temporary_password_expires_at")

		case "profile_picture":
			if model.ProfilePicture != nil {
				_, err = s.DB.NewDelete().
					Model(model.ProfilePicture).
					WherePK().
					Exec(ctx)
				if err != nil {
					return nil, wrapDriverError(err)
				}
			}
			if pb.ProfilePicture != nil {
				model.ProfilePicture, err = pictureFromPB(ctx, pb.ProfilePicture)
				if err != nil {
					return nil, err
				}

				_, err = s.DB.NewInsert().
					Model(model.ProfilePicture).
					Exec(ctx)
				if err != nil {
					return nil, wrapDriverError(err)
				}

				model.ProfilePictureID = &model.ProfilePicture.ID
				columns = append(columns, "profile_picture_id")
			} else {
				model.ProfilePicture = nil
				model.ProfilePictureID = nil
				columns = append(columns, "profile_picture_id")
			}
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the result to protobuf.
	updatedPB, err := userToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *userStore) DeleteUser(ctx context.Context, id *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(ctx, selectWithEmbeddedAccountUID(ctx, id), nil)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	if model.ProfilePictureID != nil {
		_, err = s.DB.NewDelete().
			Model((*Picture)(nil)).
			Where("id = ?", *model.ProfilePictureID).
			Exec(ctx)
		if err != nil {
			return wrapDriverError(err)
		}
	}

	return nil
}

func (s *userStore) RestoreUser(ctx context.Context, id *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.Start(ctx, "RestoreUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(store.WithSoftDeleted(ctx, true), selectWithEmbeddedAccountUID(ctx, id), nil)
	if err != nil {
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	if model.ProfilePictureID != nil {
		_, err = s.DB.NewUpdate().
			Model((*Picture)(nil)).
			WhereAllWithDeleted().
			Where("id = ?", *model.ProfilePictureID).
			Set("deleted_at = NULL").
			Exec(ctx)
		if err != nil {
			return wrapDriverError(err)
		}
	}

	return nil
}

func (s *userStore) PurgeUser(ctx context.Context, id *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.Start(ctx, "PurgeUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(
		store.WithSoftDeleted(ctx, false),
		selectWithEmbeddedAccountUID(ctx, id),
		store.FieldMask{"attributes", "contact_info"},
	)
	if err != nil {
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "user", model.ID)
		if err != nil {
			return err
		}
	}

	if len(model.ContactInfo) > 0 {
		_, err = s.replaceContactInfo(ctx, model.ContactInfo, nil, "user", model.ID)
		if err != nil {
			return err
		}
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	// Instead of purging, we soft-delete the profile picture,
	// so that a cleanup process can clean up the storage bucket.
	if model.ProfilePictureID != nil {
		_, err = s.DB.NewDelete().
			Model((*Picture)(nil)).
			Where("id = ?", *model.ProfilePictureID).
			Exec(ctx)
		if err != nil {
			return wrapDriverError(err)
		}
	}

	return nil
}
