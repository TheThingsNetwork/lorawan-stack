// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/satori/go.uuid"
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
