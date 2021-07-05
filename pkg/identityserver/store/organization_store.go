// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"reflect"
	"runtime/trace"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetOrganizationStore returns an OrganizationStore on the given db (or transaction).
func GetOrganizationStore(db *gorm.DB) OrganizationStore {
	return &organizationStore{store: newStore(db)}
}

type organizationStore struct {
	*store
}

// selectOrganizationFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectOrganizationFields(ctx context.Context, query *gorm.DB, fieldMask *pbtypes.FieldMask) *gorm.DB {
	if len(fieldMask.GetPaths()) == 0 {
		return query.Preload("Attributes").Select([]string{"accounts.uid", "organizations.*"})
	}
	var organizationColumns []string
	var notFoundPaths []string
	organizationColumns = append(organizationColumns, "organizations.deleted_at", "accounts.uid")
	for _, column := range modelColumns {
		organizationColumns = append(organizationColumns, "organizations."+column)
	}
	for _, path := range ttnpb.TopLevelFields(fieldMask.GetPaths()) {
		switch path {
		case "ids", "created_at", "updated_at", "deleted_at":
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		default:
			if columns, ok := organizationColumnNames[path]; ok {
				organizationColumns = append(organizationColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(organizationColumns)
}

func (s *organizationStore) CreateOrganization(ctx context.Context, org *ttnpb.Organization) (*ttnpb.Organization, error) {
	defer trace.StartRegion(ctx, "create organization").End()
	orgModel := Organization{
		Account: Account{UID: org.OrganizationId}, // The ID is not mutated by fromPB.
	}
	orgModel.fromPB(org, nil)
	if err := s.createEntity(ctx, &orgModel); err != nil {
		return nil, err
	}
	var orgProto ttnpb.Organization
	orgModel.toPB(&orgProto, nil)
	return &orgProto, nil
}

func (s *organizationStore) FindOrganizations(ctx context.Context, ids []*ttnpb.OrganizationIdentifiers, fieldMask *pbtypes.FieldMask) ([]*ttnpb.Organization, error) {
	defer trace.StartRegion(ctx, "find organizations").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetOrganizationId()
	}
	query := s.query(ctx, Organization{}, withOrganizationID(idStrings...))
	query = selectOrganizationFields(ctx, query, fieldMask)
	query = query.Order(orderFromContext(ctx, "organizations", `"accounts"."uid"`, "ASC"))
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(Organization{}))
		query = query.Limit(limit).Offset(offset)
	}
	var orgModels []organizationWithUID
	query = query.Find(&orgModels)
	setTotal(ctx, uint64(len(orgModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	orgProtos := make([]*ttnpb.Organization, len(orgModels))
	for i, orgModel := range orgModels {
		orgProto := &ttnpb.Organization{}
		orgModel.toPB(orgProto, fieldMask)
		orgProtos[i] = orgProto
	}
	return orgProtos, nil
}

func (s *organizationStore) GetOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers, fieldMask *pbtypes.FieldMask) (*ttnpb.Organization, error) {
	defer trace.StartRegion(ctx, "get organization").End()
	query := s.query(ctx, Organization{}, withOrganizationID(id.GetOrganizationId()))
	query = selectOrganizationFields(ctx, query, fieldMask)
	var orgModel organizationWithUID
	if err := query.First(&orgModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id)
		}
		return nil, err
	}
	orgProto := &ttnpb.Organization{}
	orgModel.toPB(orgProto, fieldMask)
	return orgProto, nil
}

func (s *organizationStore) UpdateOrganization(ctx context.Context, org *ttnpb.Organization, fieldMask *pbtypes.FieldMask) (updated *ttnpb.Organization, err error) {
	defer trace.StartRegion(ctx, "update organization").End()
	query := s.query(ctx, Organization{}, withOrganizationID(org.GetOrganizationId()))
	query = selectOrganizationFields(ctx, query, fieldMask)
	var orgModel organizationWithUID
	if err = query.First(&orgModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(org.OrganizationIdentifiers)
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes := orgModel.Attributes
	columns := orgModel.fromPB(org, fieldMask)
	if err = s.updateEntity(ctx, &orgModel.Organization, columns...); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(oldAttributes, orgModel.Attributes) {
		if err = s.replaceAttributes(ctx, "organization", orgModel.ID, oldAttributes, orgModel.Attributes); err != nil {
			return nil, err
		}
	}
	updated = &ttnpb.Organization{}
	orgModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *organizationStore) DeleteOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) (err error) {
	defer trace.StartRegion(ctx, "delete organization").End()
	return s.deleteEntity(ctx, id)
}

func (s *organizationStore) RestoreOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) (err error) {
	defer trace.StartRegion(ctx, "restore organization").End()
	return s.restoreEntity(ctx, id)
}

func (s *organizationStore) PurgeOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) (err error) {
	defer trace.StartRegion(ctx, "purge organization").End()

	query := s.query(ctx, Organization{}, withSoftDeleted(), withOrganizationID(id.GetOrganizationId()))
	query = selectOrganizationFields(ctx, query, nil)
	var orgModel organizationWithUID
	if err = query.First(&orgModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errNotFoundForID(id)
		}
		return err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return err
	}
	if len(orgModel.Attributes) > 0 {
		if err := s.replaceAttributes(ctx, "organization", orgModel.ID, orgModel.Attributes, nil); err != nil {
			return err
		}
	}

	err = s.purgeEntity(ctx, id)
	if err != nil {
		return err
	}
	// Purge account after purging organization because it is necessary for organization query
	return s.query(ctx, Account{}, withSoftDeleted()).Where(Account{
		UID:         id.IDString(),
		AccountType: id.EntityType(),
	}).Delete(Account{}).Error
}
