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
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestEUIStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()
	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Application{}, &EUIBlock{})
		euiStore := GetEUIStore(db)
		s := newStore(db)

		created := &Application{ApplicationID: "test-app"}
		s.createEntity(ctx, created)

		// Test block creation.
		var startEui types.EUI64Prefix
		var devEUIBlock EUIBlock

		startEui.UnmarshalConfigString("70B3D57ED0000000/36")
		err := euiStore.CreateEUIBlock(ctx, "dev_eui", startEui, 0)
		a.So(err, should.BeNil)

		query := s.query(ctx, EUIBlock{}).Where(EUIBlock{
			Type: "dev_eui",
		})
		query.First(&devEUIBlock)
		a.So(devEUIBlock.StartEUI.toPB().String(), should.Equal, "70B3D57ED0000000")
		a.So(devEUIBlock.CurrentCounter, should.Equal, 0)
		a.So(devEUIBlock.MaxCounter, should.Equal, 268435455)

		// Test DevEUI issuing.
		devEUI, err := euiStore.IssueDevEUIForApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}, 3)
		a.So(err, should.BeNil)
		if a.So(devEUI, should.NotBeNil) {
			a.So(devEUI.String(), should.Equal, "70B3D57ED0000000")
		}
		query = s.query(ctx, EUIBlock{}).Where(EUIBlock{
			Type: "dev_eui",
		})
		err = query.First(&devEUIBlock).Error
		a.So(err, should.BeNil)
		a.So(devEUIBlock.CurrentCounter, should.Equal, 1)

		// Test application DevEUI limit reached.
		devEUI, err = euiStore.IssueDevEUIForApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}, 1)
		if a.So(err, should.NotBeNil) {
			a.So(devEUI, should.BeNil)
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		// Test global DevEUI block limit reached.
		query.First(&devEUIBlock)
		devEUIBlock.CurrentCounter = devEUIBlock.MaxCounter + 1
		s.query(ctx, EUIBlock{}).Save(&devEUIBlock)
		devEUI, err = euiStore.IssueDevEUIForApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}, 100)
		if a.So(err, should.NotBeNil) {
			a.So(devEUI, should.BeNil)
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	})
}
