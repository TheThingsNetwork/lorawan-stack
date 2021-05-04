// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func getMaxAddress(addressBlock types.EUI64Prefix) *EUI64 {
	var maxAddress types.EUI64
	maxAddress64 := addressBlock.EUI64.MarshalNumber() | (^(uint64(0)) >> (64 - addressBlock.Length))
	maxAddress.UnmarshalNumber(maxAddress64)
	return eui(&maxAddress)
}

// CreateEUIBlock configures the identity server with a new block of EUI addresses to be issued.
func CreateEUIBlock(ctx context.Context, db *gorm.DB, euiType string, block string) (err error) {
	defer trace.StartRegion(ctx, "create eui block").End()

	var addressBlock types.EUI64Prefix
	err = addressBlock.UnmarshalConfigString(block)
	if err != nil {
		return err
	}

	var currentAddressBlock EUIBlock
	query := db.Where(EUIBlock{Type: euiType})
	// Check if there is already an address block of same EUI type configured.
	err = query.First(&currentAddressBlock).Error
	if err == nil {
		// If same address block already configured, skip initialization.
		if addressBlock.EUI64.Equal(*currentAddressBlock.StartAddress.toPB()) {
			return nil
		}
		// If a different block configured, update the block in the database.
		return query.Select("start_address", "end_address", "current_address").Save(
			&EUIBlock{
				StartAddress:   eui(&addressBlock.EUI64),
				EndAddress:     getMaxAddress(addressBlock),
				CurrentAddress: eui(&addressBlock.EUI64),
			},
		).Error
		// If no block found, create a new block in the database.
	} else if gorm.IsRecordNotFoundError(err) {
		return db.Save(
			&EUIBlock{
				Type:           euiType,
				StartAddress:   eui(&addressBlock.EUI64),
				EndAddress:     getMaxAddress(addressBlock),
				CurrentAddress: eui(&addressBlock.EUI64),
			},
		).Error
	}
	return err
}
