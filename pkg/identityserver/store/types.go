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

package store

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// EUI64 adds methods on a types.EUI64 so that it can be stored in an SQL database.
type EUI64 types.EUI64

var zeroEUI64 EUI64

// Value returns the value to store in the database.
func (eui EUI64) Value() (driver.Value, error) {
	return types.EUI64(eui).String(), nil
}

// Scan reads the value from the database into the EUI.
func (eui *EUI64) Scan(src interface{}) (err error) {
	var dto types.EUI64
	switch src := src.(type) {
	case []byte:
		err = dto.UnmarshalText(src)
	case string:
		err = dto.UnmarshalText([]byte(src))
	case nil:
		*eui = EUI64{}
	default:
		err = fmt.Errorf("cannot convert %T to EUI64", src)
	}
	if err != nil {
		return
	}
	*eui = EUI64(dto)
	return nil
}

// Grants adds methods on a []ttnpb.GrantType so that it can be stored in an SQL database.
type Grants []ttnpb.GrantType

// Value returns the value to store in the database.
func (g Grants) Value() (driver.Value, error) {
	ints := make([]int64, len(g))
	for i, grant := range g {
		ints[i] = int64(grant)
	}
	return pq.Int64Array(ints).Value()
}

// Scan reads the value from the database into the Grants.
func (g *Grants) Scan(src interface{}) error {
	var ints pq.Int64Array
	err := ints.Scan(src)
	if err != nil {
		return err
	}
	grants := make(Grants, len(ints))
	for i, grant := range ints {
		grants[i] = ttnpb.GrantType(grant)
	}
	*g = grants
	return nil
}

// Rights adds methods on a ttnpb.Rights so that it can be stored in an SQL database.
type Rights ttnpb.Rights

// Value returns the value to store in the database.
func (r Rights) Value() (driver.Value, error) {
	ints := make([]int64, len(r.Rights))
	for i, right := range r.Rights {
		ints[i] = int64(right)
	}
	return pq.Int64Array(ints).Value()
}

// Scan reads the value from the database into the Rights.
func (r *Rights) Scan(src interface{}) error {
	var ints pq.Int64Array
	err := ints.Scan(src)
	if err != nil {
		return err
	}
	rights := make([]ttnpb.Right, len(ints))
	for i, right := range ints {
		rights[i] = ttnpb.Right(right)
	}
	*r = Rights{Rights: rights}
	return nil
}
