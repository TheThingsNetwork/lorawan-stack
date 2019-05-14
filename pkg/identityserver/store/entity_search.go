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
	return &entitySearch{db: db}
}

type entitySearch struct {
	db *gorm.DB
}

func (s *entitySearch) FindEntities(ctx context.Context, req *ttnpb.SearchEntitiesRequest, entityType string) ([]ttnpb.Identifiers, error) {
	defer trace.StartRegion(ctx, "find entities").End()

	table := entityType + "s"
	db := s.db.Scopes(withContext(ctx)).Table(table)
	idField := fmt.Sprintf("%s_id", entityType)
	if entityType == "user" || entityType == "organization" {
		idField = "accounts.uid"
		db = db.Joins(fmt.Sprintf("JOIN accounts ON accounts.account_type = ? AND accounts.account_id = %s.id", table), entityType)
	}
	db = db.Select(fmt.Sprintf("%s AS id", idField)).Where(fmt.Sprintf("%s.deleted_at IS NULL", table))
	if req.IDContains != "" {
		db = db.Where(fmt.Sprintf("%s LIKE ?", idField), "%"+req.IDContains+"%") // TODO: Escape wildcards (https://github.com/TheThingsNetwork/lorawan-stack/issues/73).
	}
	if req.NameContains != "" {
		db = db.Where("name LIKE ?", "%"+req.NameContains+"%") // TODO: Escape wildcards (https://github.com/TheThingsNetwork/lorawan-stack/issues/73).
	}
	if req.DescriptionContains != "" {
		db = db.Where("description LIKE ?", "%"+req.DescriptionContains+"%") // TODO: Escape wildcards (https://github.com/TheThingsNetwork/lorawan-stack/issues/73).
	}
	if len(req.AttributesContain) > 0 {
		sub := s.db.Scopes(withContext(ctx)).Table("attributes").Select("entity_id").Where("entity_type = ?", entityType)
		for key, value := range req.AttributesContain {
			sub = sub.Where("key = ? AND value LIKE ?", key, "%"+value+"%") // TODO: Escape wildcards (https://github.com/TheThingsNetwork/lorawan-stack/issues/73).
		}
		db = db.Where(fmt.Sprintf("%s.id IN (?)", table), sub.QueryExpr())
	}

	var entities []struct {
		ID string
	}
	if err := db.Scan(&entities).Error; err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, nil
	}

	identifiers := make([]ttnpb.Identifiers, len(entities))
	for i, entity := range entities {
		switch entityType {
		case "application":
			identifiers[i] = ttnpb.ApplicationIdentifiers{ApplicationID: entity.ID}
		case "client":
			identifiers[i] = ttnpb.ClientIdentifiers{ClientID: entity.ID}
		case "gateway":
			identifiers[i] = ttnpb.GatewayIdentifiers{GatewayID: entity.ID}
		case "organization":
			identifiers[i] = ttnpb.OrganizationIdentifiers{OrganizationID: entity.ID}
		case "user":
			identifiers[i] = ttnpb.UserIdentifiers{UserID: entity.ID}
		}
	}

	return identifiers, nil
}
