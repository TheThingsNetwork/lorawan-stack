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
	"fmt"
	"sort"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Client is the client model in the database.
type Client struct {
	bun.BaseModel `bun:"table:clients,alias:cli"`

	Model
	SoftDelete

	ClientID string `bun:"client_id,notnull"`

	Name        string `bun:"name,nullzero"`
	Description string `bun:"description,nullzero"`

	Attributes []*Attribute `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	ContactInfo []*ContactInfo `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	AdministrativeContactID *string  `bun:"administrative_contact_id,type:uuid"`
	AdministrativeContact   *Account `bun:"rel:belongs-to,join:administrative_contact_id=id"`

	TechnicalContactID *string  `bun:"technical_contact_id,type:uuid"`
	TechnicalContact   *Account `bun:"rel:belongs-to,join:technical_contact_id=id"`

	ClientSecret string `bun:"client_secret,nullzero"`

	RedirectURIs       []string `bun:"redirect_uris,array,nullzero"`
	LogoutRedirectURIs []string `bun:"logout_redirect_uris,array,nullzero"`

	State            int    `bun:"state,notnull"`
	StateDescription string `bun:"state_description,nullzero"`

	SkipAuthorization bool `bun:"skip_authorization,notnull"`
	Endorsed          bool `bun:"endorsed,notnull"`

	Grants []int `bun:"grants,array,nullzero"`
	Rights []int `bun:"rights,array,nullzero"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Client) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func clientToPB(m *Client, fieldMask ...string) (*ttnpb.Client, error) {
	pb := &ttnpb.Client{
		Ids: &ttnpb.ClientIdentifiers{
			ClientId: m.ClientID,
		},

		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
		DeletedAt: ttnpb.ProtoTime(m.DeletedAt),

		Name:        m.Name,
		Description: m.Description,

		Secret: m.ClientSecret,

		RedirectUris:       m.RedirectURIs,
		LogoutRedirectUris: m.LogoutRedirectURIs,

		State:            ttnpb.State(m.State),
		StateDescription: m.StateDescription,

		SkipAuthorization: m.SkipAuthorization,
		Endorsed:          m.Endorsed,

		Grants: convertIntSlice[int, ttnpb.GrantType](m.Grants),
		Rights: convertIntSlice[int, ttnpb.Right](m.Rights),
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

	if m.AdministrativeContact != nil {
		pb.AdministrativeContact = m.AdministrativeContact.GetOrganizationOrUserIdentifiers()
	}
	if m.TechnicalContact != nil {
		pb.TechnicalContact = m.TechnicalContact.GetOrganizationOrUserIdentifiers()
	}

	if len(fieldMask) == 0 {
		return pb, nil
	}

	res := &ttnpb.Client{}
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

type clientStore struct {
	*baseStore
}

func newClientStore(baseStore *baseStore) *clientStore {
	return &clientStore{
		baseStore: baseStore,
	}
}

func (s *clientStore) CreateClient(
	ctx context.Context, pb *ttnpb.Client,
) (*ttnpb.Client, error) {
	ctx, span := tracer.Start(ctx, "CreateClient", trace.WithAttributes(
		attribute.String("client_id", pb.GetIds().GetClientId()),
	))
	defer span.End()

	clientModel := &Client{
		ClientID:    pb.GetIds().GetClientId(),
		Name:        pb.Name,
		Description: pb.Description,

		ClientSecret: pb.Secret,

		RedirectURIs:       pb.RedirectUris,
		LogoutRedirectURIs: pb.LogoutRedirectUris,

		State:            int(pb.State),
		StateDescription: pb.StateDescription,

		SkipAuthorization: pb.SkipAuthorization,
		Endorsed:          pb.Endorsed,

		Grants: convertIntSlice[ttnpb.GrantType, int](pb.Grants),
		Rights: convertIntSlice[ttnpb.Right, int](pb.Rights),
	}

	if contact := pb.AdministrativeContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		clientModel.AdministrativeContact = account
		clientModel.AdministrativeContactID = &account.ID
	}
	if contact := pb.TechnicalContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		clientModel.TechnicalContact = account
		clientModel.TechnicalContactID = &account.ID
	}

	_, err := s.DB.NewInsert().
		Model(clientModel).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	if len(pb.Attributes) > 0 {
		clientModel.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "client", clientModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.ContactInfo) > 0 {
		clientModel.ContactInfo, err = s.replaceContactInfo(
			ctx, nil, pb.ContactInfo, "client", clientModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	pb, err = clientToPB(clientModel)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (*clientStore) selectWithFields(q *bun.SelectQuery, fieldMask store.FieldMask) (*bun.SelectQuery, error) {
	if fieldMask == nil {
		q = q.ExcludeColumn()
	} else {
		columns := []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"client_id",
		}
		for _, f := range fieldMask.TopLevel() {
			switch f {
			default:
				return nil, fmt.Errorf("unknown field %q", f)
			case "ids", "created_at", "updated_at", "deleted_at":
				// Always selected.
			case "name", "description",
				"redirect_uris", "logout_redirect_uris",
				"state", "state_description",
				"skip_authorization", "endorsed",
				"grants", "rights":
				// Proto name equals model name.
				columns = append(columns, f)
			case "secret":
				columns = append(columns, "client_secret")
			case "attributes":
				q = q.Relation("Attributes")
			case "contact_info":
				q = q.Relation("ContactInfo")
			case "administrative_contact":
				q = q.Relation("AdministrativeContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			case "technical_contact":
				q = q.Relation("TechnicalContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			}
		}
		q = q.Column(columns...)
	}
	return q, nil
}

func (s *clientStore) CountClients(ctx context.Context) (uint64, error) {
	selectQuery := s.DB.NewSelect().
		Model(&Client{}).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(selectWithContext(ctx))

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return 0, wrapDriverError(err)
	}

	return uint64(count), nil
}

func (s *clientStore) listClientsBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.Client, error) {
	models := []*Client{}
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
		Apply(selectWithOrderFromContext(ctx, "client_id", map[string]string{
			"client_id":  "client_id",
			"name":       "name",
			"created_at": "created_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	selectQuery, err = s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.Client, len(models))
	for i, model := range models {
		pb, err := clientToPB(model, fieldMask...)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (*clientStore) selectWithID(
	ctx context.Context, ids ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.client_id = ?", ids[0])
		default:
			return q.Where("?TableAlias.client_id IN (?)", bun.In(ids))
		}
	}
}

func (s *clientStore) FindClients(
	ctx context.Context, ids []*ttnpb.ClientIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.Client, error) {
	ctx, span := tracer.Start(ctx, "FindClients", trace.WithAttributes(
		attribute.StringSlice("client_ids", idStrings(ids...)),
	))
	defer span.End()

	return s.listClientsBy(ctx, s.selectWithID(ctx, idStrings(ids...)...), fieldMask)
}

func (s *clientStore) getClientModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) (*Client, error) {
	model := &Client{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	selectQuery, err := s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, wrapDriverError(err)
	}

	return model, nil
}

func (s *clientStore) GetClient(
	ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.Client, error) {
	ctx, span := tracer.Start(ctx, "GetClient", trace.WithAttributes(
		attribute.String("client_id", id.GetClientId()),
	))
	defer span.End()

	model, err := s.getClientModelBy(
		ctx, s.selectWithID(ctx, id.GetClientId()), fieldMask,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrClientNotFound.WithAttributes(
				"client_id", id.GetClientId(),
			)
		}
		return nil, err
	}
	pb, err := clientToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *clientStore) updateClientModel( //nolint:gocyclo
	ctx context.Context, model *Client, pb *ttnpb.Client, fieldMask store.FieldMask,
) (err error) {
	columns := store.FieldMask{"updated_at"}

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
				ctx, model.Attributes, pb.Attributes, "client", model.ID,
			)
			if err != nil {
				return err
			}

		case "contact_info":
			model.ContactInfo, err = s.replaceContactInfo(
				ctx, model.ContactInfo, pb.ContactInfo, "client", model.ID,
			)
			if err != nil {
				return err
			}

		case "administrative_contact":
			if contact := pb.AdministrativeContact; contact != nil {
				account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
				if err != nil {
					return err
				}
				model.AdministrativeContact = account
				model.AdministrativeContactID = &account.ID
			} else {
				model.AdministrativeContact = nil
				model.AdministrativeContactID = nil
			}
			columns = append(columns, "administrative_contact_id")

		case "technical_contact":
			if contact := pb.TechnicalContact; contact != nil {
				account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
				if err != nil {
					return err
				}
				model.TechnicalContact = account
				model.TechnicalContactID = &account.ID
			} else {
				model.TechnicalContact = nil
				model.TechnicalContactID = nil
			}
			columns = append(columns, "technical_contact_id")

		case "secret":
			model.ClientSecret = pb.Secret
			columns = append(columns, "client_secret")

		case "redirect_uris":
			model.RedirectURIs = pb.RedirectUris
			columns = append(columns, "redirect_uris")

		case "logout_redirect_uris":
			model.LogoutRedirectURIs = pb.LogoutRedirectUris
			columns = append(columns, "logout_redirect_uris")

		case "state":
			model.State = int(pb.State)
			columns = append(columns, "state")

		case "state_description":
			model.StateDescription = pb.StateDescription
			columns = append(columns, "state_description")

		case "skip_authorization":
			model.SkipAuthorization = pb.SkipAuthorization
			columns = append(columns, "skip_authorization")

		case "grants":
			model.Grants = convertIntSlice[ttnpb.GrantType, int](pb.Grants)
			columns = append(columns, "grants")

		case "rights":
			model.Rights = convertIntSlice[ttnpb.Right, int](pb.Rights)
			columns = append(columns, "rights")
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *clientStore) UpdateClient(
	ctx context.Context, pb *ttnpb.Client, fieldMask store.FieldMask,
) (*ttnpb.Client, error) {
	ctx, span := tracer.Start(ctx, "UpdateClient", trace.WithAttributes(
		attribute.String("client_id", pb.GetIds().GetClientId()),
	))
	defer span.End()

	model, err := s.getClientModelBy(
		ctx, s.selectWithID(ctx, pb.GetIds().GetClientId()), fieldMask,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrClientNotFound.WithAttributes(
				"client_id", pb.GetIds().GetClientId(),
			)
		}
		return nil, err
	}

	if err = s.updateClientModel(ctx, model, pb, fieldMask); err != nil {
		return nil, err
	}

	// Convert the result to protobuf.
	updatedPB, err := clientToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *clientStore) DeleteClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteClient", trace.WithAttributes(
		attribute.String("client_id", id.GetClientId()),
	))
	defer span.End()

	model, err := s.getClientModelBy(ctx, s.selectWithID(ctx, id.GetClientId()), store.FieldMask{"ids"})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrClientNotFound.WithAttributes(
				"client_id", id.GetClientId(),
			)
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *clientStore) RestoreClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error {
	ctx, span := tracer.Start(ctx, "RestoreClient", trace.WithAttributes(
		attribute.String("client_id", id.GetClientId()),
	))
	defer span.End()

	model, err := s.getClientModelBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithID(ctx, id.GetClientId()),
		store.FieldMask{"ids"},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrClientNotFound.WithAttributes(
				"client_id", id.GetClientId(),
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
		return wrapDriverError(err)
	}

	return nil
}

func (s *clientStore) PurgeClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error {
	ctx, span := tracer.Start(ctx, "PurgeClient", trace.WithAttributes(
		attribute.String("client_id", id.GetClientId()),
	))
	defer span.End()

	model, err := s.getClientModelBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithID(ctx, id.GetClientId()),
		store.FieldMask{"attributes", "contact_info"},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrClientNotFound.WithAttributes(
				"client_id", id.GetClientId(),
			)
		}
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "client", model.ID)
		if err != nil {
			return err
		}
	}

	if len(model.ContactInfo) > 0 {
		_, err = s.replaceContactInfo(ctx, model.ContactInfo, nil, "client", model.ID)
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

	return nil
}
