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

import "testing"

func TestExecTx(t *testing.T) {
	getInstance(t).Transact(func(tx *Tx) error {
		testExec(t, tx)
		return nil
	})
}

func TestNamedExecTx(t *testing.T) {
	getInstance(t).Transact(func(tx *Tx) error {
		testNamedExec(t, tx)
		return nil
	})
}

func TestSelectTx(t *testing.T) {
	getInstance(t).Transact(func(tx *Tx) error {
		testSelect(t, tx)
		return nil
	})
}

func TestNamedSelectTx(t *testing.T) {
	getInstance(t).Transact(func(tx *Tx) error {
		testNamedSelect(t, tx)
		return nil
	})
}

func TestSelectOneTx(t *testing.T) {
	getInstance(t).Transact(func(tx *Tx) error {
		testSelectOne(t, tx)
		return nil
	})
}

func TestNamedSelectOneTx(t *testing.T) {
	getInstance(t).Transact(func(tx *Tx) error {
		testNamedSelectOne(t, tx)
		return nil
	})
}
