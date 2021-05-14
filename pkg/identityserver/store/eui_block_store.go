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
	"sync"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func GetEUIStore(db *gorm.DB) EUIStore {
	return &euiStore{store: newStore(db)}
}

type euiStore struct {
	*store
	devEUIMu sync.Mutex
}

func getMaxAddress(addressBlock types.EUI64Prefix) *EUI64 {
	var maxAddress types.EUI64
	maxAddress64 := addressBlock.EUI64.MarshalNumber() | (^(uint64(0)) >> (64 - addressBlock.Length))
	maxAddress.UnmarshalNumber(maxAddress64)
	return eui(&maxAddress)
}

var (
	errMaxGlobalEUILimitReached = errors.DefineInternal("global_eui_limit_reached", "global eui limit from address block reached")
	errAppDevEUILimitReached    = errors.DefineInternal("application_dev_eui_limit_reached", "application dev eui limit reached")
)

func (s *euiStore) assignDevEUIAddress(ctx context.Context, devEUIBlock *EUIBlock) (*types.EUI64, error) {
	var addrFound bool
	var devEUI types.EUI64
	// Loop until an unused address is found or limit reached.
	for !addrFound {
		// Check if the global DevEUI address block limit is reached.
		if devEUIBlock.CurrentAddress.toPB().Equal(*devEUIBlock.EndAddress.toPB()) {
			return nil, errMaxGlobalEUILimitReached
		}
		// Check if the current DevEUI address is already used by any device.
		query := s.query(ctx, EndDevice{}, withDevEUI(*devEUIBlock.CurrentAddress))
		if err := query.First(EndDevice{}).Error; err != nil {
			// If the address is not used, assign the address.
			if !gorm.IsRecordNotFoundError(err) {
				return nil, err
			}
			addrFound = true
			devEUI = *devEUIBlock.CurrentAddress.toPB()
		}
		// Increment the global DevEUI counter to the next address.
		currentAddress := devEUIBlock.CurrentAddress.toPB()
		currentAddress.Inc()
		devEUIBlock.CurrentAddress = eui(currentAddress)
	}
	// Update database counter entry.
	err := s.query(ctx, devEUIBlock).Select("current_address").Save(devEUIBlock).Error
	if err != nil {
		return nil, err
	}
	// Return the address assigned
	return &devEUI, nil
}

// CreateEUIBlock configures the identity server with a new block of EUI addresses to be issued.
func (s *euiStore) CreateEUIBlock(ctx context.Context, euiType string, block string) (err error) {
	defer trace.StartRegion(ctx, "create eui block").End()

	var addressBlock types.EUI64Prefix
	err = addressBlock.UnmarshalConfigString(block)
	if err != nil {
		return err
	}

	var currentAddressBlock EUIBlock
	query := s.query(ctx, EUIBlock{}).Where(EUIBlock{Type: euiType})
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
		return s.query(ctx, EUIBlock{}).Save(
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

func (s *euiStore) IssueDevEUIForApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, maxAddressPerApp int) (*types.EUI64, error) {
	defer trace.StartRegion(ctx, "assign dev eui address to application").End()
	s.devEUIMu.Lock()
	defer s.devEUIMu.Unlock()
	// Check if max DevEUI per application reached.
	query := s.query(ctx, Application{}, withApplicationID(ids.GetApplicationId()))
	var appModel Application
	if err := query.First(&appModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(ids)
		}
		return nil, err
	}
	if appModel.DevEUIAddressCounter == maxAddressPerApp {
		return nil, errAppDevEUILimitReached
	}
	// Check if DevEUI block limit reached.
	var currentAddressBlock EUIBlock
	query = s.query(ctx, EUIBlock{}).Where(EUIBlock{Type: "dev_eui"})
	err := query.First(&currentAddressBlock).Error
	if err != nil {
		return nil, err
	}
	// Issue an unused DevEUI address.
	devEUIAddress, err := s.assignDevEUIAddress(ctx, &currentAddressBlock)
	if err != nil {
		return nil, err
	}
	// Increment App DevEUI counter
	appModel.DevEUIAddressCounter++
	if err = s.updateEntity(ctx, &appModel, "dev_eui_counter"); err != nil {
		return nil, err
	}
	return devEUIAddress, nil
}
