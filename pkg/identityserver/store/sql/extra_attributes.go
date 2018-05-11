// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package sql

import (
	"fmt"
	"strings"

	"github.com/satori/go.uuid"
	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
)

type extraAttributesStore struct {
	storer
	entity     string
	foreignKey string
}

func newExtraAttributesStore(store storer, entity string) *extraAttributesStore {
	return &extraAttributesStore{
		storer:     store,
		entity:     entity,
		foreignKey: fmt.Sprintf("%s_id", entity),
	}
}

func (s *extraAttributesStore) loadAttributes(q db.QueryContext, entityID uuid.UUID, attributer store.Attributer) error {
	// Fill the application from all specified namespaces.
	for _, namespace := range attributer.Namespaces() {
		m := make(map[string]interface{})
		err := q.SelectOne(
			&m,
			fmt.Sprintf("SELECT * FROM %s_%ss WHERE %s = $1", namespace, s.entity, s.foreignKey),
			entityID)
		if !db.IsNoRows(err) && err != nil {
			return err
		}

		err = attributer.Fill(namespace, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *extraAttributesStore) storeAttributes(q db.QueryContext, entityID uuid.UUID, attributer store.Attributer) error {
	for _, namespace := range attributer.Namespaces() {
		m := attributer.Attributes(namespace)
		values := make([]interface{}, 0, len(m)+1)
		keys := make([]string, 0, len(m)+1)
		colonKeys := make([]string, 0, len(m)+1)

		values = append(values, entityID)
		keys = append(keys, s.foreignKey)
		colonKeys = append(colonKeys, "$1")

		for k, v := range m {
			values = append(values, v)
			keys = append(keys, k)
			colonKeys = append(colonKeys, fmt.Sprintf("$%v", len(values)))
		}

		query := fmt.Sprintf(
			`UPSERT
				INTO %s_%ss (%s)
				VALUES (%s)
				RETURNING *`,
			namespace,
			s.entity,
			strings.Join(keys, ", "),
			strings.Join(colonKeys, ", "))

		r := make(map[string]interface{})
		err := q.SelectOne(r, query, values...)
		if !db.IsNoRows(err) && err != nil {
			return err
		}
	}

	return nil
}
