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
	"fmt"
	"runtime/trace"

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
}

func getMaxCounter(addressBlock types.EUI64Prefix) int64 {
	return int64(^(uint64(0)) >> (addressBlock.Length))
}

var (
	errMaxGlobalEUILimitReached = errors.DefineInvalidArgument("global_eui_limit_reached", "global eui limit from address block reached")
	errAppDevEUILimitReached    = errors.DefineInvalidArgument("application_dev_eui_limit_reached", "application issued DevEUI limit ({dev_eui_limit}) reached")
)

func (s *euiStore) incrementApplicationDevEUICounter(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, applicationLimit int) error {
	var appModel Application
	// Check if application exists.
	query := s.query(WithoutSoftDeleted(ctx), Application{}, withApplicationID(ids.GetApplicationId()))
	if err := query.First(&appModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errNotFoundForID(ids)
		}
		return err
	}
	// If app DevEUI limit not configured, skip this step.
	if applicationLimit == 0 {
		return nil
	}
	// Atomically check and increment DevEUI counter if its below threshold.
	result := s.query(ctx, Application{}).
		Where(Application{ApplicationID: appModel.ApplicationID}).
		Where(`"applications"."dev_eui_counter" < ?`, applicationLimit).
		Update("dev_eui_counter", gorm.Expr("dev_eui_counter + ?", 1))
	if err := result.Error; err != nil {
		return err
	}
	// If application DevEUI counter not updated, the limit was reached.
	if result.RowsAffected == 0 {
		return errAppDevEUILimitReached.WithAttributes("dev_eui_limit", fmt.Sprint(applicationLimit))
	}
	return nil
}

func (s *euiStore) issueDevEUIAddressFromBlock(ctx context.Context) (*types.EUI64, error) {
	var devEUIResult types.EUI64
	var euiBlock EUIBlock
	for {
		// Get current value from DevEUI db.
		query := s.query(ctx, EUIBlock{}).Where(EUIBlock{Type: "dev_eui"})
		if err := query.First(&euiBlock).Error; err != nil {
			return nil, err
		}
		// Check if DevEUI block limit exists.
		result := s.query(ctx, EUIBlock{}).
			Where(`"eui_blocks"."type"='dev_eui' AND "eui_blocks"."current_counter" <= "eui_blocks"."end_counter"`).
			Update("current_counter", gorm.Expr("current_counter + ?", 1))
		if err := result.Error; err != nil {
			return nil, err
		}
		if result.RowsAffected == 0 {
			return nil, errMaxGlobalEUILimitReached.New()
		}

		devEUIResult.UnmarshalNumber(euiBlock.StartEUI.toPB().MarshalNumber() | uint64(euiBlock.CurrentCounter))
		deviceQuery := s.query(ctx, EndDevice{}).Where(EndDevice{
			DevEUI: eui(&devEUIResult),
		})
		// Return the address assigned if no existing device uses the DevEUI
		if err := deviceQuery.First(&EndDevice{}).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return &devEUIResult, nil
			}
			return nil, err
		}
	}
}

// CreateEUIBlock configures the identity server with a new block of EUI addresses to be issued.
func (s *euiStore) CreateEUIBlock(ctx context.Context, euiType string, block types.EUI64Prefix, initCounterValue int64) (err error) {
	defer trace.StartRegion(ctx, "create eui block").End()

	var currentAddressBlock EUIBlock
	query := s.query(ctx, EUIBlock{}).Where(EUIBlock{Type: euiType})
	// Check if there is already an address block of same EUI type configured.
	err = query.First(&currentAddressBlock).Error
	if err == nil {
		// If same address block already configured, skip initialization.
		if block.EUI64.Equal(*currentAddressBlock.StartEUI.toPB()) {
			return nil
		}
		// If a different block configured, update the block in the database.
		return query.Select("start_eui", "end_counter", "current_counter").Save(
			&EUIBlock{
				StartEUI:       *eui(&block.EUI64),
				MaxCounter:     getMaxCounter(block),
				CurrentCounter: initCounterValue,
			},
		).Error
		// If no block found, create a new block in the database.
	} else if gorm.IsRecordNotFoundError(err) {
		return s.query(ctx, EUIBlock{}).Save(
			&EUIBlock{
				Type:           euiType,
				StartEUI:       *eui(&block.EUI64),
				MaxCounter:     getMaxCounter(block),
				CurrentCounter: initCounterValue,
			},
		).Error
	}
	return err
}

func (s *euiStore) IssueDevEUIForApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, applicationLimit int) (*types.EUI64, error) {
	defer trace.StartRegion(ctx, "assign DevEUI address to application").End()
	// Check if max DevEUI per application reached.
	err := s.incrementApplicationDevEUICounter(ctx, ids, applicationLimit)
	if err != nil {
		return nil, err
	}
	// Issue an unused DevEUI address.
	devEUIAddress, err := s.issueDevEUIAddressFromBlock(ctx)
	if err != nil {
		return nil, err
	}
	return devEUIAddress, nil
}
