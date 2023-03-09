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

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

// EUIBlock is the EUI block model in the database.
type EUIBlock struct {
	bun.BaseModel `bun:"table:eui_blocks,alias:euib"`

	Model

	// TODO: The database schema allows NULL for the fields below. We should fix the schema.
	// https://github.com/TheThingsNetwork/lorawan-stack/issues/5613
	Type           string `bun:"type,notnull"`
	StartEUI       string `bun:"start_eui,notnull"`
	EndCounter     int64  `bun:"end_counter,notnull"`
	CurrentCounter int64  `bun:"current_counter,notnull"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *EUIBlock) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

type euiStore struct {
	*baseStore
	applicationStore *applicationStore
}

func newEUIStore(baseStore *baseStore) *euiStore {
	return &euiStore{
		baseStore:        baseStore,
		applicationStore: newApplicationStore(baseStore),
	}
}

func (s *euiStore) CreateEUIBlock(
	ctx context.Context, prefix types.EUI64Prefix, initCounter int64, euiType string,
) error {
	ctx, span := tracer.StartFromContext(ctx, "CreateEUIBlock")
	defer span.End()

	model := &EUIBlock{}
	selectQuery := s.newSelectModel(ctx, model).
		Where("type = ?", euiType).
		Where("LOWER(start_eui) = LOWER(?)", prefix.EUI64.String())

	err := selectQuery.Scan(ctx)
	if err == nil {
		return nil
	}

	err = storeutil.WrapDriverError(err)
	if !errors.IsNotFound(err) {
		return err
	}

	model = &EUIBlock{
		Type:           euiType,
		StartEUI:       prefix.EUI64.String(),
		EndCounter:     int64(^(uint64(0)) >> (prefix.Length)),
		CurrentCounter: initCounter,
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}

func (s *euiStore) IssueDevEUIForApplication(
	ctx context.Context, id *ttnpb.ApplicationIdentifiers, applicationLimit int,
) (*types.EUI64, error) {
	ctx, span := tracer.StartFromContext(ctx, "IssueDevEUIForApplication", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationId()),
	))
	defer span.End()

	euiBlockModel := &EUIBlock{}

	err := s.transact(ctx, func(ctx context.Context, tx bun.IDB) error {
		applicationModel, err := s.applicationStore.getApplicationModelBy(
			ctx,
			combineApply(
				s.applicationStore.selectWithID(ctx, id.GetApplicationId()),
				func(q *bun.SelectQuery) *bun.SelectQuery { return q.For("UPDATE") },
			),
			store.FieldMask{"dev_eui_counter"},
		)
		if err != nil {
			return err
		}

		if applicationLimit > 0 && applicationModel.DevEUICounter >= applicationLimit {
			return store.ErrApplicationDevEUILimitReached.WithAttributes(
				"dev_eui_limit", applicationLimit,
			)
		}

		_, err = s.DB.NewUpdate().
			Model(applicationModel).
			WherePK().
			Set("dev_eui_counter = dev_eui_counter + 1").
			Returning("dev_eui_counter").
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}

		if applicationLimit > 0 && applicationModel.DevEUICounter > applicationLimit {
			return store.ErrApplicationDevEUILimitReached.WithAttributes(
				"dev_eui_limit", applicationLimit,
			)
		}

		selectQuery := s.newSelectModel(ctx, euiBlockModel).
			For("UPDATE").
			Where("type = ?", "dev_eui").
			Where("current_counter <= end_counter").
			Order("created_at ASC").
			Limit(1)
		if err := selectQuery.Scan(ctx); err != nil {
			err = storeutil.WrapDriverError(err)
			if errors.IsNotFound(err) {
				return store.ErrNoEUIBlockAvailable.New()
			}
			return err
		}

		_, err = s.DB.NewUpdate().
			Model(euiBlockModel).
			WherePK().
			Set("current_counter = current_counter + 1").
			Returning("current_counter").
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}

		return nil
	})
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	var eui types.EUI64
	if err := eui.UnmarshalText([]byte(euiBlockModel.StartEUI)); err != nil {
		return nil, err
	}

	eui.UnmarshalNumber(eui.MarshalNumber() | uint64(euiBlockModel.CurrentCounter-1))

	return &eui, nil
}
