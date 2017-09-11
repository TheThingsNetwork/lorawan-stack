// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package helpers

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
)

// WriteAttributes is a helper that constructs the sql query to write the Attributer
// attributes of a certain namespace.
func WriteAttributes(attributer store.Attributer, namespace, baseTableName, primaryKeyName, id string) (string, []interface{}) {
	m := attributer.Attributes(namespace)
	values := make([]interface{}, 0, len(m)+1)
	keys := make([]string, 0, len(m)+1)
	colonKeys := make([]string, 0, len(m)+1)

	values = append(values, id)
	keys = append(keys, primaryKeyName)
	colonKeys = append(colonKeys, "$1")

	for k, v := range m {
		values = append(values, v)
		keys = append(keys, k)
		colonKeys = append(colonKeys, fmt.Sprintf("$%v", len(values)))
	}

	// TODO: upsert on conflict
	// UPSERT is not supported with RETURNING
	// See https://github.com/cockroachdb/cockroach/issues/6637
	query := fmt.Sprintf(
		`INSERT
			INTO %s_%s (%s)
			VALUES (%s)
			RETURNING *`,
		namespace,
		baseTableName,
		strings.Join(keys, ", "),
		strings.Join(colonKeys, ", "))

	return query, values
}
