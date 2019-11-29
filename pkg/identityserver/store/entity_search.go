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

package store

import (
	"context"
	"fmt"
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetEntitySearch returns an EntitySearch on the given db (or transaction).
func GetEntitySearch(db *gorm.DB) EntitySearch {
	return &entitySearch{store: newStore(db)}
}

type entitySearch struct {
	*store
}

type metaFields interface {
	GetIDContains() string
	GetNameContains() string
	GetDescriptionContains() string
	GetAttributesContain() map[string]string
}

func (s *entitySearch) queryMetaFields(ctx context.Context, query *gorm.DB, entityType string, req metaFields) *gorm.DB {
	if v := req.GetIDContains(); v != "" {
		switch entityType {
		case "organization", "user":
			query = query.Where(`"accounts"."uid" LIKE ?`, "%"+v+"%")
		default:
			query = query.Where(fmt.Sprintf(`"%[1]ss"."%[1]s_id" LIKE ?`, entityType), "%"+v+"%")
		}
	}
	if dbKind, ok := query.Get("db:kind"); ok && dbKind == "PostgreSQL" {
		language := "english"
		if v := req.GetNameContains(); v != "" {
			query = query.Where(fmt.Sprintf("to_tsvector('%[1]s', name) @@ to_tsquery('%[1]s', ?)", language), v)
		}
		if v := req.GetDescriptionContains(); v != "" {
			query = query.Where(fmt.Sprintf("to_tsvector('%[1]s', description) @@ to_tsquery('%[1]s', ?)", language), v)
		}
	} else {
		if v := req.GetNameContains(); v != "" {
			query = query.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", v))
		}
		if v := req.GetDescriptionContains(); v != "" {
			query = query.Where("description ILIKE ?", fmt.Sprintf("%%%s%%", v))
		}
	}
	if kv := req.GetAttributesContain(); len(kv) > 0 {
		sub := s.query(ctx, &Attribute{}).Select("entity_id").Where("entity_type = ?", entityType)
		for k, v := range kv {
			sub = sub.Where("key = ? AND value ILIKE ?", k, fmt.Sprintf("%%%s%%", v))
		}
		query = query.Where(fmt.Sprintf(`"%ss"."id" IN (?)`, entityType), sub.QueryExpr())
	}
	return query
}

func (s *entitySearch) FindEntities(ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchEntitiesRequest, entityType string) ([]ttnpb.Identifiers, error) {
	defer trace.StartRegion(ctx, "find entities").End()

	query := s.query(ctx, modelForEntityType(entityType))
	switch entityType {
	case "organization":
		query = query.
			Joins(`JOIN "accounts" ON "accounts"."account_type" = 'organization' AND "accounts"."account_id" = "organizations"."id"`).
			Select(`"accounts"."uid" AS "friendly_id"`)
	case "user":
		query = query.
			Joins(`JOIN "accounts" ON "accounts"."account_type" = 'user' AND "accounts"."account_id" = "users"."id"`).
			Select(`"accounts"."uid" AS "friendly_id"`)
	default:
		query = query.
			Select(fmt.Sprintf(`"%[1]ss"."%[1]s_id" AS "friendly_id"`, entityType))
	}

	if member != nil {
		membershipsQuery := (&membershipStore{store: s.store}).queryMemberships(ctx, member, entityType, true).Select("entity_id").QueryExpr()
		if entityType == "organization" {
			query = query.Where(`"accounts"."account_type" = ? AND "accounts"."account_id" IN (?)`, entityType, membershipsQuery)
		} else {
			query = query.Where(fmt.Sprintf(`"%[1]ss"."id" IN (?)`, entityType), membershipsQuery)
		}
	}

	query = s.queryMetaFields(ctx, query, entityType, req)

	query = query.Order(`"friendly_id"`)
	page := query
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}
	var results []struct {
		FriendlyID string
	}
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 && (offset > 0 || len(results) == int(limit)) {
		countTotal(ctx, query)
	} else {
		setTotal(ctx, uint64(len(results)))
	}
	identifiers := make([]ttnpb.Identifiers, len(results))
	for i, result := range results {
		identifiers[i] = buildIdentifiers(entityType, result.FriendlyID)
	}
	return identifiers, nil
}
