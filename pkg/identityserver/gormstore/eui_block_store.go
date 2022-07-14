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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// GetEUIStore returns an EUIStore on the given db (or transaction).
func GetEUIStore(db *gorm.DB) store.EUIStore {
	return &euiStore{baseStore: newStore(db)}
}

type euiStore struct {
	*baseStore
}

func getMaxCounter(addressBlock types.EUI64Prefix) int64 {
	return int64(^(uint64(0)) >> (addressBlock.Length))
}

var (
	errMaxGlobalEUILimitReached = errors.DefineFailedPrecondition(
		"global_eui_limit_reached",
		"global eui limit from address block reached",
	)
	errAppDevEUILimitReached = errors.DefineFailedPrecondition(
		"application_dev_eui_limit_reached",
		"application issued DevEUI limit ({dev_eui_limit}) reached",
	)
)

func (s *euiStore) incrementApplicationDevEUICounter(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers, applicationLimit int,
) error {
	var appModel Application
	// Check if application exists.
	query := s.query(ctx, Application{}, withApplicationID(ids.GetApplicationId()))
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

const atomicReadAndUpdateDevEUI = `
UPDATE "eui_blocks" SET "current_counter" = "current_counter" + 1
WHERE "start_eui" = (
	SELECT "start_eui"
	FROM "eui_blocks"
	WHERE "current_counter"<="end_counter" AND "type"='dev_eui'
	LIMIT 1
) RETURNING "start_eui","current_counter"`

func (s *euiStore) issueDevEUIAddressFromBlock(ctx context.Context) (*types.EUI64, error) {
	var euiResult []struct {
		StartEui       EUI64
		CurrentCounter int64
	}
	for {
		// Atomically update valid block and return the values.
		if err := s.query(ctx, EUIBlock{}).
			Raw(atomicReadAndUpdateDevEUI).
			Scan(&euiResult).
			Error; err != nil {
			return nil, err
		}
		if len(euiResult) == 0 {
			return nil, errMaxGlobalEUILimitReached.New()
		}
		var devEUIResult types.EUI64
		devEUIResult.UnmarshalNumber(
			euiResult[0].StartEui.toPB().MarshalNumber() | uint64(euiResult[0].CurrentCounter-1),
		)
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

// CreateUIBlock creates the block of appropriate type in IS db.
func (s *euiStore) CreateEUIBlock(
	ctx context.Context, prefix types.EUI64Prefix, initCounter int64, euiType string,
) error {
	defer trace.StartRegion(ctx, "create eui block").End()
	var block EUIBlock
	if err := s.query(ctx, EUIBlock{}).
		Where(EUIBlock{Type: "dev_eui", StartEUI: *eui(&prefix.EUI64)}).
		First(&block).Error; err != nil {
		// If block is not found, create it
		if gorm.IsRecordNotFoundError(err) {
			euiBlock := &EUIBlock{
				Type:           euiType,
				StartEUI:       *eui(&prefix.EUI64),
				MaxCounter:     getMaxCounter(prefix),
				CurrentCounter: initCounter,
			}
			return s.query(ctx, EUIBlock{}).Create(&euiBlock).Error
		}
		return err
	}
	return nil
}

// IssueDevEUIForApplication issues DevEUI address from the configured DevEUI block.
func (s *euiStore) IssueDevEUIForApplication(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers, applicationLimit int,
) (*types.EUI64, error) {
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
