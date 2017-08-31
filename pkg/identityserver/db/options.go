// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import "database/sql"

// TxOption is a transaction option
type TxOption func(*sql.TxOptions)

// ReadOnly enables/disables read only mode on transactions
func ReadOnly(on bool) TxOption {
	return func(opts *sql.TxOptions) {
		// TODO: enable this when Cockroach supports it
		// opts.ReadOnly = on
	}
}
