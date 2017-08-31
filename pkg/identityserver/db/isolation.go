// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import "database/sql"

// IsolationLevel is the level of isolation to use for transactions
type IsolationLevel int

const (
	// LevelSnapshot is the snapshot level
	LevelSnapshot IsolationLevel = IsolationLevel(sql.LevelSnapshot)

	// LevelSerializable is the serializable snapshot level
	LevelSerializable = IsolationLevel(sql.LevelSerializable)
)

var (
	// Snapshot sets the transaction isolation level to Snapshot
	Snapshot = Isolation(LevelSnapshot)

	// Serializable sets the transaction isolation level to Serializable
	Serializable = Isolation(LevelSerializable)
)

// Isolation sets the transaction isolation level
func Isolation(level IsolationLevel) TxOption {
	return func(opts *sql.TxOptions) {
		opts.Isolation = sql.IsolationLevel(level)
	}
}
