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
	"reflect"
	"runtime/trace"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetClientStore returns an ClientStore on the given db (or transaction).
func GetClientStore(db *gorm.DB) ClientStore {
	return &clientStore{store: newStore(db)}
}

type clientStore struct {
	*store
}

// selectClientFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectClientFields(ctx context.Context, query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query.Preload("Attributes")
	}
	var clientColumns []string
	var notFoundPaths []string
	for _, path := range ttnpb.TopLevelFields(fieldMask.Paths) {
		switch path {
		case "ids", "created_at", "updated_at":
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		default:
			if columns, ok := clientColumnNames[path]; ok {
				clientColumns = append(clientColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(cleanFields(append(append(modelColumns, "client_id"), clientColumns...)...))
}

func (s *clientStore) CreateClient(ctx context.Context, cli *ttnpb.Client) (*ttnpb.Client, error) {
	defer trace.StartRegion(ctx, "create client").End()
	cliModel := Client{
		ClientID: cli.ClientID, // The ID is not mutated by fromPB.
	}
	cliModel.fromPB(cli, nil)
	if err := s.createEntity(ctx, &cliModel); err != nil {
		return nil, err
	}
	var cliProto ttnpb.Client
	cliModel.toPB(&cliProto, nil)
	return &cliProto, nil
}

func (s *clientStore) FindClients(ctx context.Context, ids []*ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Client, error) {
	defer trace.StartRegion(ctx, "find clients").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetClientID()
	}
	query := s.query(ctx, Client{}, withClientID(idStrings...))
	query = selectClientFields(ctx, query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(Client{}))
		query = query.Limit(limit).Offset(offset)
	}
	var cliModels []Client
	query = query.Find(&cliModels)
	setTotal(ctx, uint64(len(cliModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	cliProtos := make([]*ttnpb.Client, len(cliModels))
	for i, cliModel := range cliModels {
		cliProto := &ttnpb.Client{}
		cliModel.toPB(cliProto, fieldMask)
		cliProtos[i] = cliProto
	}
	return cliProtos, nil
}

func (s *clientStore) GetClient(ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Client, error) {
	defer trace.StartRegion(ctx, "get client").End()
	query := s.query(ctx, Client{}, withClientID(id.GetClientID()))
	query = selectClientFields(ctx, query, fieldMask)
	var cliModel Client
	if err := query.First(&cliModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id)
		}
		return nil, err
	}
	cliProto := &ttnpb.Client{}
	cliModel.toPB(cliProto, fieldMask)
	return cliProto, nil
}

func (s *clientStore) UpdateClient(ctx context.Context, cli *ttnpb.Client, fieldMask *types.FieldMask) (updated *ttnpb.Client, err error) {
	defer trace.StartRegion(ctx, "update client").End()
	query := s.query(ctx, Client{}, withClientID(cli.GetClientID()))
	query = selectClientFields(ctx, query, fieldMask)
	var cliModel Client
	if err = query.First(&cliModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(cli.ClientIdentifiers)
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes := cliModel.Attributes
	columns := cliModel.fromPB(cli, fieldMask)
	if err = s.updateEntity(ctx, &cliModel, columns...); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(oldAttributes, cliModel.Attributes) {
		if err = s.replaceAttributes(ctx, "client", cliModel.ID, oldAttributes, cliModel.Attributes); err != nil {
			return nil, err
		}
	}
	updated = &ttnpb.Client{}
	cliModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *clientStore) DeleteClient(ctx context.Context, id *ttnpb.ClientIdentifiers) error {
	defer trace.StartRegion(ctx, "delete client").End()
	return s.deleteEntity(ctx, id)
}
