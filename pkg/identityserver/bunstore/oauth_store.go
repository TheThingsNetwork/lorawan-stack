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
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// ClientAuthorization is the OAuth client authorization model in the database.
type ClientAuthorization struct {
	bun.BaseModel `bun:"table:client_authorizations,alias:oca"`

	Model

	Client   *Client `bun:"rel:belongs-to,join:client_id=id"`
	ClientID string  `bun:"client_id,notnull"`

	User   *User  `bun:"rel:belongs-to,join:user_id=id"`
	UserID string `bun:"user_id,notnull"`

	Rights []int `bun:"rights,array,nullzero"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *ClientAuthorization) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func clientAuthorizationToPB(
	m *ClientAuthorization, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) (*ttnpb.OAuthClientAuthorization, error) {
	pb := &ttnpb.OAuthClientAuthorization{
		UserIds:   userIDs,
		ClientIds: clientIDs,
		Rights:    convertIntSlice[int, ttnpb.Right](m.Rights),
		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
	}
	if pb.UserIds == nil && m.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{
			UserId: m.User.Account.UID,
		}
	}
	if pb.ClientIds == nil && m.Client != nil {
		pb.ClientIds = &ttnpb.ClientIdentifiers{
			ClientId: m.Client.ClientID,
		}
	}
	return pb, nil
}

// AuthorizationCode is the OAuth authorization code model in the database.
type AuthorizationCode struct {
	bun.BaseModel `bun:"table:authorization_codes,alias:oac"`

	Model

	Client   *Client `bun:"rel:belongs-to,join:client_id=id"`
	ClientID string  `bun:"client_id,notnull"`

	User   *User  `bun:"rel:belongs-to,join:user_id=id"`
	UserID string `bun:"user_id,notnull"`

	UserSession   *UserSession `bun:"rel:belongs-to,join:user_session_id=id"`
	UserSessionID string       `bun:"user_session_id,nullzero"`

	Rights []int `bun:"rights,array,nullzero"`

	Code        string `bun:"code,notnull"`
	RedirectURI string `bun:"redirect_uri,nullzero"`
	State       string `bun:"state,nullzero"`

	ExpiresAt *time.Time `bun:"expires_at"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *AuthorizationCode) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func authorizationCodeToPB(
	m *AuthorizationCode, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) (*ttnpb.OAuthAuthorizationCode, error) {
	pb := &ttnpb.OAuthAuthorizationCode{
		UserIds:       userIDs,
		UserSessionId: m.UserSessionID,
		ClientIds:     clientIDs,
		Rights:        convertIntSlice[int, ttnpb.Right](m.Rights),
		Code:          m.Code,
		RedirectUri:   m.RedirectURI,
		State:         m.State,
		CreatedAt:     ttnpb.ProtoTimePtr(m.CreatedAt),
		ExpiresAt:     ttnpb.ProtoTime(m.ExpiresAt),
	}
	if pb.UserIds == nil && m.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{
			UserId: m.User.Account.UID,
		}
	}
	if pb.ClientIds == nil && m.Client != nil {
		pb.ClientIds = &ttnpb.ClientIdentifiers{
			ClientId: m.Client.ClientID,
		}
	}
	return pb, nil
}

// AccessToken is the OAuth access token model in the database.
type AccessToken struct {
	bun.BaseModel `bun:"table:access_tokens,alias:oat"`

	Model

	Client   *Client `bun:"rel:belongs-to,join:client_id=id"`
	ClientID string  `bun:"client_id,notnull"`

	User   *User  `bun:"rel:belongs-to,join:user_id=id"`
	UserID string `bun:"user_id,notnull"`

	UserSession   *UserSession `bun:"rel:belongs-to,join:user_session_id=id"`
	UserSessionID string       `bun:"user_session_id,nullzero"`

	Rights []int `bun:"rights,array,nullzero"`

	TokenID string `bun:"token_id,notnull"`

	Previous   *AccessToken `bun:"rel:belongs-to,join:previous_id=id"`
	PreviousID string       `bun:"previous_id,nullzero"`

	AccessToken  string `bun:"access_token,notnull"`
	RefreshToken string `bun:"refresh_token,notnull"`

	ExpiresAt *time.Time `bun:"expires_at"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *AccessToken) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func accessTokenToPB(
	m *AccessToken, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) (*ttnpb.OAuthAccessToken, error) {
	pb := &ttnpb.OAuthAccessToken{
		UserIds:       userIDs,
		UserSessionId: m.UserSessionID,
		ClientIds:     clientIDs,
		Id:            m.TokenID,
		AccessToken:   m.AccessToken,
		RefreshToken:  m.RefreshToken,
		Rights:        convertIntSlice[int, ttnpb.Right](m.Rights),
		CreatedAt:     ttnpb.ProtoTimePtr(m.CreatedAt),
		ExpiresAt:     ttnpb.ProtoTime(m.ExpiresAt),
	}
	if pb.UserIds == nil && m.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{
			UserId: m.User.Account.UID,
		}
	}
	if pb.ClientIds == nil && m.Client != nil {
		pb.ClientIds = &ttnpb.ClientIdentifiers{
			ClientId: m.Client.ClientID,
		}
	}
	return pb, nil
}

type oauthStore struct {
	*entityStore
}

func newOAuthStore(baseStore *baseStore) *oauthStore {
	return &oauthStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (*oauthStore) selectWithUserIDs(
	_ context.Context, uuid string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("user_id = ?", uuid)
	}
}

func (s *oauthStore) ListAuthorizations(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers,
) ([]*ttnpb.OAuthClientAuthorization, error) {
	ctx, span := tracer.Start(ctx, "ListAuthorizations", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	models := []*ClientAuthorization{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(s.selectWithUserIDs(ctx, userUUID))

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "created_at", map[string]string{
			"created_at": "created_at",
			"expires_at": "expires_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	// Include the OAuth client identifiers.
	selectQuery = selectQuery.
		Relation("Client", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("client_id")
		})

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.OAuthClientAuthorization, len(models))
	for i, model := range models {
		pb, err := clientAuthorizationToPB(model, userIDs, nil)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *oauthStore) getClientUUID(
	ctx context.Context, clientIDs *ttnpb.ClientIdentifiers,
) (string, error) {
	clientModel, err := s.getClientModelBy(
		ctx,
		s.clientStore.selectWithID(ctx, clientIDs.GetClientId()),
		store.FieldMask{"ids"},
	)
	if err != nil {
		return "", err
	}
	return clientModel.ID, nil
}

func (*oauthStore) selectWithClientIDs(
	_ context.Context, uuid string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("client_id = ?", uuid)
	}
}

func (s *oauthStore) GetAuthorization(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) (*ttnpb.OAuthClientAuthorization, error) {
	ctx, span := tracer.Start(ctx, "GetAuthorization", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
		attribute.String("client_id", clientIDs.GetClientId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	clientUUID, err := s.getClientUUID(ctx, clientIDs)
	if err != nil {
		return nil, err
	}

	model := &ClientAuthorization{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(s.selectWithUserIDs(ctx, userUUID)).
		Apply(s.selectWithClientIDs(ctx, clientUUID))

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return nil, store.ErrAuthorizationNotFound.WithAttributes(
				"user_id", userIDs.GetUserId(),
				"client_id", clientIDs.GetClientId(),
			)
		}
		return nil, err
	}

	pb, err := clientAuthorizationToPB(model, userIDs, clientIDs)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *oauthStore) Authorize(
	ctx context.Context, pb *ttnpb.OAuthClientAuthorization,
) (authorization *ttnpb.OAuthClientAuthorization, err error) {
	ctx, span := tracer.Start(ctx, "Authorize", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().GetUserId()),
		attribute.String("client_id", pb.GetClientIds().GetClientId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	clientUUID, err := s.getClientUUID(ctx, pb.GetClientIds())
	if err != nil {
		return nil, err
	}

	model := &ClientAuthorization{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(s.selectWithUserIDs(ctx, userUUID)).
		Apply(s.selectWithClientIDs(ctx, clientUUID))

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if !errors.IsNotFound(err) {
			return nil, err
		}
		model = &ClientAuthorization{
			ClientID: clientUUID,
			UserID:   userUUID,
			Rights:   convertIntSlice[ttnpb.Right, int](pb.Rights),
		}
		_, err = s.DB.NewInsert().
			Model(model).
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	} else {
		model.Rights = convertIntSlice[ttnpb.Right, int](pb.Rights)
		_, err = s.DB.NewUpdate().
			Model(model).
			Column("rights", "updated_at").
			WherePK().
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	pb, err = clientAuthorizationToPB(model, pb.GetUserIds(), pb.GetClientIds())
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *oauthStore) DeleteAuthorization(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) error {
	ctx, span := tracer.Start(ctx, "DeleteAuthorization", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
		attribute.String("client_id", clientIDs.GetClientId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return err
	}

	clientUUID, err := s.getClientUUID(ctx, clientIDs)
	if err != nil {
		return err
	}

	model := &ClientAuthorization{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(s.selectWithUserIDs(ctx, userUUID)).
		Apply(s.selectWithClientIDs(ctx, clientUUID))

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return store.ErrAuthorizationNotFound.WithAttributes(
				"user_id", userIDs.GetUserId(),
				"client_id", clientIDs.GetClientId(),
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

func (s *oauthStore) CreateAuthorizationCode(
	ctx context.Context, pb *ttnpb.OAuthAuthorizationCode,
) (*ttnpb.OAuthAuthorizationCode, error) {
	ctx, span := tracer.Start(ctx, "CreateAuthorizationCode", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().GetUserId()),
		attribute.String("client_id", pb.GetClientIds().GetClientId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	clientUUID, err := s.getClientUUID(ctx, pb.GetClientIds())
	if err != nil {
		return nil, err
	}

	model := &AuthorizationCode{
		ClientID:      clientUUID,
		UserID:        userUUID,
		UserSessionID: pb.UserSessionId,
		Rights:        convertIntSlice[ttnpb.Right, int](pb.Rights),
		Code:          pb.Code,
		RedirectURI:   pb.RedirectUri,
		State:         pb.State,
		ExpiresAt:     cleanTimePtr(ttnpb.StdTime(pb.ExpiresAt)),
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	pb, err = authorizationCodeToPB(model, pb.GetUserIds(), pb.GetClientIds())
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *oauthStore) GetAuthorizationCode(ctx context.Context, code string) (*ttnpb.OAuthAuthorizationCode, error) {
	ctx, span := tracer.Start(ctx, "GetAuthorizationCode")
	defer span.End()

	model := &AuthorizationCode{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithContext(ctx)).
		Where("code = ?", code)

	// Include the user identifiers.
	selectQuery = selectQuery.
		Relation("User", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("account_uid")
		})

	// Include the OAuth client identifiers.
	selectQuery = selectQuery.
		Relation("Client", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("client_id")
		})

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return nil, store.ErrAuthorizationCodeNotFound.New()
		}
		return nil, err
	}

	pb, err := authorizationCodeToPB(model, nil, nil)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *oauthStore) DeleteAuthorizationCode(ctx context.Context, code string) error {
	ctx, span := tracer.Start(ctx, "DeleteAuthorizationCode")
	defer span.End()

	model := &AuthorizationCode{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithContext(ctx)).
		Where("code = ?", code)

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return store.ErrAuthorizationCodeNotFound.New()
		}
		return err
	}

	_, err := s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *oauthStore) CreateAccessToken(
	ctx context.Context, pb *ttnpb.OAuthAccessToken, previousID string,
) (*ttnpb.OAuthAccessToken, error) {
	ctx, span := tracer.Start(ctx, "CreateAccessToken", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().GetUserId()),
		attribute.String("client_id", pb.GetClientIds().GetClientId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	clientUUID, err := s.getClientUUID(ctx, pb.GetClientIds())
	if err != nil {
		return nil, err
	}

	model := &AccessToken{
		ClientID:      clientUUID,
		UserID:        userUUID,
		UserSessionID: pb.UserSessionId,
		Rights:        convertIntSlice[ttnpb.Right, int](pb.Rights),
		TokenID:       pb.Id,
		PreviousID:    previousID,
		AccessToken:   pb.AccessToken,
		RefreshToken:  pb.RefreshToken,
		ExpiresAt:     cleanTimePtr(ttnpb.StdTime(pb.ExpiresAt)),
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	pb, err = accessTokenToPB(model, pb.GetUserIds(), pb.GetClientIds())
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *oauthStore) ListAccessTokens(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) ([]*ttnpb.OAuthAccessToken, error) {
	ctx, span := tracer.Start(ctx, "ListAccessTokens", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
		attribute.String("client_id", clientIDs.GetClientId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	clientUUID, err := s.getClientUUID(ctx, clientIDs)
	if err != nil {
		return nil, err
	}

	models := []*AccessToken{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(s.selectWithUserIDs(ctx, userUUID)).
		Apply(s.selectWithClientIDs(ctx, clientUUID))

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "created_at", map[string]string{
			"created_at": "created_at",
			"expires_at": "expires_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.OAuthAccessToken, len(models))
	for i, model := range models {
		pb, err := accessTokenToPB(model, userIDs, clientIDs)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *oauthStore) GetAccessToken(ctx context.Context, id string) (*ttnpb.OAuthAccessToken, error) {
	ctx, span := tracer.Start(ctx, "GetAccessToken")
	defer span.End()

	model := &AccessToken{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithContext(ctx)).
		Where("token_id = ?", id)

	// Include the user identifiers.
	selectQuery = selectQuery.
		Relation("User", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("account_uid")
		})

	// Include the OAuth client identifiers.
	selectQuery = selectQuery.
		Relation("Client", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("client_id")
		})

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return nil, store.ErrAccessTokenNotFound.WithAttributes(
				"access_token_id", id,
			)
		}
		return nil, err
	}

	pb, err := accessTokenToPB(model, nil, nil)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *oauthStore) DeleteAccessToken(ctx context.Context, id string) error {
	ctx, span := tracer.Start(ctx, "DeleteAccessToken")
	defer span.End()

	model := &AccessToken{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithContext(ctx)).
		Where("token_id = ?", id)

	if err := selectQuery.Scan(ctx); err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return store.ErrAccessTokenNotFound.WithAttributes(
				"access_token_id", id,
			)
		}
		return err
	}

	_, err := s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *oauthStore) DeleteUserAuthorizations(ctx context.Context, userIDs *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteUserAuthorizations", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(&ClientAuthorization{}).
		Where("user_id = ?", userUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	_, err = s.DB.NewDelete().
		Model(&AuthorizationCode{}).
		Where("user_id = ?", userUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	_, err = s.DB.NewDelete().
		Model(&AccessToken{}).
		Where("user_id = ?", userUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *oauthStore) DeleteClientAuthorizations(ctx context.Context, clientIDs *ttnpb.ClientIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteClientAuthorizations", trace.WithAttributes(
		attribute.String("client_id", clientIDs.GetClientId()),
	))
	defer span.End()

	clientUUID, err := s.getClientUUID(ctx, clientIDs)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(&ClientAuthorization{}).
		Where("client_id = ?", clientUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	_, err = s.DB.NewDelete().
		Model(&AuthorizationCode{}).
		Where("client_id = ?", clientUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	_, err = s.DB.NewDelete().
		Model(&AccessToken{}).
		Where("client_id = ?", clientUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}
