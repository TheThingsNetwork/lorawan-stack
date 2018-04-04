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

package db

import "database/sql"

// IsolationLevel is the level of isolation to use for transactions.
type IsolationLevel int

const (
	// LevelSnapshot is the snapshot level.
	LevelSnapshot IsolationLevel = IsolationLevel(sql.LevelSnapshot)

	// LevelSerializable is the serializable snapshot level.
	LevelSerializable = IsolationLevel(sql.LevelSerializable)
)

var (
	// Snapshot sets the transaction isolation level to Snapshot.
	Snapshot = Isolation(LevelSnapshot)

	// Serializable sets the transaction isolation level to Serializable.
	Serializable = Isolation(LevelSerializable)
)

// Isolation sets the transaction isolation level.
func Isolation(level IsolationLevel) TxOption {
	return func(opts *sql.TxOptions) {
		opts.Isolation = sql.IsolationLevel(level)
	}
}
