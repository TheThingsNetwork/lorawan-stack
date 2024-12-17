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
	"encoding/json"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	ConsolePreferences           json.RawMessage `bun:"console_preferences"`
	UniversalRights              []int           `bun:"universal_rights,array,nullzero"`
	EmailNotificationPreferences []int           `bun:"email_notification_preferences,array,nullzero"`
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

		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		DeletedAt: ttnpb.ProtoTime(m.DeletedAt),

		Name:        m.Name,
		Description: m.Description,

		PrimaryEmailAddress:            m.PrimaryEmailAddress,
		PrimaryEmailAddressValidatedAt: ttnpb.ProtoTime(m.PrimaryEmailAddressValidatedAt),

		Password:              m.Password,
		PasswordUpdatedAt:     timestamppb.New(m.PasswordUpdatedAt),
		RequirePasswordUpdate: m.RequirePasswordUpdate,

		State:            ttnpb.State(m.State),
		StateDescription: m.StateDescription,

		Admin: m.Admin,

		TemporaryPassword:          m.TemporaryPassword,
		TemporaryPasswordCreatedAt: ttnpb.ProtoTime(m.TemporaryPasswordCreatedAt),
		TemporaryPasswordExpiresAt: ttnpb.ProtoTime(m.TemporaryPasswordExpiresAt),

		ConsolePreferences: &ttnpb.UserConsolePreferences{},
		EmailNotificationPreferences: &ttnpb.EmailNotificationPreferences{
			Types: convertIntSlice[int, ttnpb.NotificationType](m.EmailNotificationPreferences),
		},
		UniversalRights: convertIntSlice[int, ttnpb.Right](m.UniversalRights),
	}

	if len(m.Attributes) > 0 {
		pb.Attributes = make(map[string]string, len(m.Attributes))
		for _, a := range m.Attributes {
			pb.Attributes[a.Key] = a.Value
		}
	}

	if m.ProfilePicture != nil {
		picture, err := pictureToPB(m.ProfilePicture)
		if err != nil {
			return nil, err
		}
		pb.ProfilePicture = picture
	}

	if len(m.ConsolePreferences) > 0 {
		if err := jsonpb.TTN().Unmarshal(m.ConsolePreferences, pb.ConsolePreferences); err != nil {
			return nil, err
		}
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
	ctx, span := tracer.StartFromContext(ctx, "CreateUser", trace.WithAttributes(
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
		PrimaryEmailAddressValidatedAt: cleanTimePtr(ttnpb.StdTime(pb.PrimaryEmailAddressValidatedAt)),
		Password:                       pb.Password,
		PasswordUpdatedAt:              cleanTime(ttnpb.StdTimeOrZero(pb.PasswordUpdatedAt)),
		RequirePasswordUpdate:          pb.RequirePasswordUpdate,
		State:                          int(pb.State),
		StateDescription:               pb.StateDescription,
		Admin:                          pb.Admin,
		TemporaryPassword:              pb.TemporaryPassword,
		TemporaryPasswordCreatedAt:     cleanTimePtr(ttnpb.StdTime(pb.TemporaryPasswordCreatedAt)),
		TemporaryPasswordExpiresAt:     cleanTimePtr(ttnpb.StdTime(pb.TemporaryPasswordExpiresAt)),
		EmailNotificationPreferences:   convertIntSlice[ttnpb.NotificationType, int](pb.EmailNotificationPreferences.GetTypes()), // nolint:lll
		UniversalRights:                convertIntSlice[ttnpb.Right, int](pb.UniversalRights),
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
			return nil, storeutil.WrapDriverError(err)
		}

		userModel.ProfilePictureID = &userModel.ProfilePicture.ID
	}

	if pb.ConsolePreferences != nil {
		b, err := jsonpb.TTN().Marshal(pb.ConsolePreferences)
		if err != nil {
			return nil, err
		}
		userModel.ConsolePreferences = b
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
		return nil, storeutil.WrapDriverError(err)
	}
	if len(pb.Attributes) > 0 {
		userModel.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "user", userModel.ID,
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

func (*userStore) selectWithFields(q *bun.SelectQuery, fieldMask store.FieldMask) (*bun.SelectQuery, error) {
	if fieldMask == nil {
		q = q.ExcludeColumn().
			Column("account_uid") // NOTE: not selected by default because it's not in the table, only in the view.
	} else {
		columns := []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"account_uid",
		}
		for _, f := range fieldMask.TopLevel() {
			switch f {
			default:
				return nil, fmt.Errorf("unknown field %q", f)
			case "ids", "created_at", "updated_at", "deleted_at":
				// Always selected.
			case "name", "description",
				"primary_email_address", "primary_email_address_validated_at",
				"password", "password_updated_at", "require_password_update",
				"state", "state_description",
				"admin",
				"temporary_password", "temporary_password_created_at", "temporary_password_expires_at":
				// Proto name equals model name.
				columns = append(columns, f)
			case "console_preferences",
				"console_preferences.console_theme",
				"console_preferences.dashboard_layouts",
				"console_preferences.sort_by":
				columns = append(columns, "console_preferences")
			case "email_notification_preferences", "email_notification_preferences.types":
				columns = append(columns, "email_notification_preferences")
			case "universal_rights":
				columns = append(columns, "universal_rights")
			case "attributes":
				q = q.Relation("Attributes")
			case "administrative_contact":
				q = q.Relation("AdministrativeContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			case "technical_contact":
				q = q.Relation("TechnicalContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			case "profile_picture":
				q = q.Relation("ProfilePicture")
			}
		}
		q = q.Column(columns...)
	}
	return q, nil
}

func (s *userStore) CountUsers(ctx context.Context) (uint64, error) {
	selectQuery := s.newSelectModel(ctx, &User{})

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return 0, storeutil.WrapDriverError(err)
	}

	return uint64(count), nil
}

func (s *userStore) listUsersBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.User, error) {
	models := []*User{}
	selectQuery := newSelectModels(ctx, s.DB, &models).Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
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
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	selectQuery, err = s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
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

func (*userStore) selectWithID(
	_ context.Context, ids ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.account_uid = ?", ids[0])
		default:
			return q.Where("?TableAlias.account_uid IN (?)", bun.In(ids))
		}
	}
}

func (s *userStore) FindUsers(
	ctx context.Context, ids []*ttnpb.UserIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.User, error) {
	ctx, span := tracer.StartFromContext(ctx, "FindUsers", trace.WithAttributes(
		attribute.StringSlice("user_ids", idStrings(ids...)),
	))
	defer span.End()

	return s.listUsersBy(ctx, s.selectWithID(ctx, idStrings(ids...)...), fieldMask)
}

func (s *userStore) ListAdmins(
	ctx context.Context, fieldMask store.FieldMask,
) ([]*ttnpb.User, error) {
	ctx, span := tracer.StartFromContext(ctx, "ListAdmins")
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
	selectQuery := s.newSelectModel(ctx, model).Apply(by)

	selectQuery, err := s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	return model, nil
}

func (s *userStore) GetUser(
	ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.User, error) {
	ctx, span := tracer.StartFromContext(ctx, "GetUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(ctx, s.selectWithID(ctx, id.GetUserId()), fieldMask)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
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
		return q.Where("LOWER(?TableAlias.primary_email_address) = LOWER(?)", primaryEmailAddress)
	}
}

func (s *userStore) GetUserByPrimaryEmailAddress(
	ctx context.Context, primaryEmailAddress string, fieldMask store.FieldMask,
) (*ttnpb.User, error) {
	ctx, span := tracer.StartFromContext(ctx, "GetUserByPrimaryEmailAddress", trace.WithAttributes(
		attribute.String("primary_email_address", primaryEmailAddress),
	))
	defer span.End()

	model, err := s.getUserModelBy(
		ctx, s.selectWithPrimaryEmailAddress(ctx, primaryEmailAddress), fieldMask,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrUserNotFoundByPrimaryEmailAddress.New()
		}
		return nil, err
	}
	pb, err := userToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *userStore) updateUserModel( //nolint:gocyclo
	ctx context.Context, model *User, pb *ttnpb.User, fieldMask store.FieldMask,
) (err error) {
	columns := store.FieldMask{"updated_at"}

	consolePreferences := &ttnpb.UserConsolePreferences{}
	updateConsolePreferences := false

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(fieldMask), "console_preferences") && len(model.ConsolePreferences) > 0 {
		if err := jsonpb.TTN().Unmarshal(model.ConsolePreferences, consolePreferences); err != nil {
			return err
		}
	}

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
				return err
			}

		case "primary_email_address":
			model.PrimaryEmailAddress = pb.PrimaryEmailAddress
			columns = append(columns, "primary_email_address")

		case "primary_email_address_validated_at":
			model.PrimaryEmailAddressValidatedAt = cleanTimePtr(ttnpb.StdTime(pb.PrimaryEmailAddressValidatedAt))
			columns = append(columns, "primary_email_address_validated_at")

		case "password":
			model.Password = pb.Password
			columns = append(columns, "password")

		case "password_updated_at":
			model.PasswordUpdatedAt = cleanTime(ttnpb.StdTimeOrZero(pb.PasswordUpdatedAt))
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
			model.TemporaryPasswordCreatedAt = cleanTimePtr(ttnpb.StdTime(pb.TemporaryPasswordCreatedAt))
			columns = append(columns, "temporary_password_created_at")

		case "temporary_password_expires_at":
			model.TemporaryPasswordExpiresAt = cleanTimePtr(ttnpb.StdTime(pb.TemporaryPasswordExpiresAt))
			columns = append(columns, "temporary_password_expires_at")

		case "profile_picture":
			if model.ProfilePicture != nil {
				_, err = s.DB.NewDelete().
					Model(model.ProfilePicture).
					WherePK().
					Exec(ctx)
				if err != nil {
					return storeutil.WrapDriverError(err)
				}
			}
			if pb.ProfilePicture != nil {
				model.ProfilePicture, err = pictureFromPB(ctx, pb.ProfilePicture)
				if err != nil {
					return err
				}

				_, err = s.DB.NewInsert().
					Model(model.ProfilePicture).
					Exec(ctx)
				if err != nil {
					return storeutil.WrapDriverError(err)
				}

				model.ProfilePictureID = &model.ProfilePicture.ID
				columns = append(columns, "profile_picture_id")
			} else {
				model.ProfilePicture = nil
				model.ProfilePictureID = nil
				columns = append(columns, "profile_picture_id")
			}

		case "console_preferences":
			updateConsolePreferences = true
			consolePreferences = pb.ConsolePreferences
		case "console_preferences.console_theme":
			updateConsolePreferences = true
			consolePreferences.ConsoleTheme = pb.ConsolePreferences.ConsoleTheme
		case "console_preferences.dashboard_layouts":
			updateConsolePreferences = true
			consolePreferences.DashboardLayouts = pb.ConsolePreferences.GetDashboardLayouts()
		case "console_preferences.sort_by":
			updateConsolePreferences = true
			consolePreferences.SortBy = pb.ConsolePreferences.GetSortBy()
		case "console_preferences.tutorials":
			updateConsolePreferences = true
			consolePreferences.Tutorials = pb.ConsolePreferences.GetTutorials()
		case "universal_rights":
			model.UniversalRights = convertIntSlice[ttnpb.Right, int](pb.UniversalRights)
			columns = append(columns, "universal_rights")
		case "email_notification_preferences", "email_notification_preferences.types":
			model.EmailNotificationPreferences = convertIntSlice[ttnpb.NotificationType, int](pb.EmailNotificationPreferences.Types) // nolint:lll
			columns = append(columns, "email_notification_preferences")
		}
	}

	// By separating the update of the console preferences we can allow for partial updates.
	if updateConsolePreferences {
		b, err := jsonpb.TTN().Marshal(consolePreferences)
		if err != nil {
			return err
		}
		model.ConsolePreferences = b
		columns = append(columns, "console_preferences")
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}

func (s *userStore) UpdateUser(
	ctx context.Context, pb *ttnpb.User, fieldMask store.FieldMask,
) (*ttnpb.User, error) {
	ctx, span := tracer.StartFromContext(ctx, "UpdateUser", trace.WithAttributes(
		attribute.String("user_id", pb.GetIds().GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(ctx, s.selectWithID(ctx, pb.GetIds().GetUserId()), fieldMask)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrUserNotFound.WithAttributes(
				"user_id", pb.GetIds().GetUserId(),
			)
		}
		return nil, err
	}

	if s.updateUserModel(ctx, model, pb, fieldMask) != nil {
		return nil, err
	}

	// Convert the result to protobuf.
	updatedPB, err := userToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *userStore) DeleteUser(ctx context.Context, id *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(ctx, s.selectWithID(ctx, id.GetUserId()), store.FieldMask{"ids"})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	accountModel, err := s.getAccountModel(ctx, id.GetEntityIdentifiers().EntityType(), id.GetUserId())
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(accountModel).
		WherePK().
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	if model.ProfilePictureID != nil {
		_, err = s.DB.NewDelete().
			Model((*Picture)(nil)).
			Where("id = ?", *model.ProfilePictureID).
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userStore) RestoreUser(ctx context.Context, id *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.StartFromContext(ctx, "RestoreUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithID(ctx, id.GetUserId()),
		store.FieldMask{"ids"},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	accountModel, err := s.getAccountModel(
		store.WithSoftDeleted(ctx, true), id.GetEntityIdentifiers().EntityType(), id.GetUserId(),
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(accountModel).
		WherePK().
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	if model.ProfilePictureID != nil {
		_, err = s.DB.NewUpdate().
			Model((*Picture)(nil)).
			WhereAllWithDeleted().
			Where("id = ?", *model.ProfilePictureID).
			Set("deleted_at = NULL").
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userStore) PurgeUser(ctx context.Context, id *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.StartFromContext(ctx, "PurgeUser", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	model, err := s.getUserModelBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithID(ctx, id.GetUserId()),
		store.FieldMask{"attributes"},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "user", model.ID)
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
		return storeutil.WrapDriverError(err)
	}

	accountModel, err := s.getAccountModel(
		store.WithSoftDeleted(ctx, false), id.GetEntityIdentifiers().EntityType(), id.GetUserId(),
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserNotFound.WithAttributes(
				"user_id", id.GetUserId(),
			)
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(accountModel).
		WherePK().
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	// Instead of purging, we soft-delete the profile picture,
	// so that a cleanup process can clean up the storage bucket.
	if model.ProfilePictureID != nil {
		_, err = s.DB.NewDelete().
			Model((*Picture)(nil)).
			Where("id = ?", *model.ProfilePictureID).
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}
