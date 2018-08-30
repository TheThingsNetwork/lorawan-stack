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

package ttnpb_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRights(t *testing.T) {
	var nilRights *ttnpb.Rights
	someAppRights := ttnpb.RightsFrom(
		ttnpb.RIGHT_APPLICATION_INFO,
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)

	t.Run("Sorted", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.Sorted().GetRights(), should.BeEmpty)
		a.So(someAppRights.Sorted().GetRights(), should.Resemble, []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_DEVICES_READ,
			ttnpb.RIGHT_APPLICATION_INFO,
		})
	})
	t.Run("Unique", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.Unique().GetRights(), should.BeEmpty)
		a.So(ttnpb.RightsFrom(
			ttnpb.RIGHT_GATEWAY_INFO,
			ttnpb.RIGHT_GATEWAY_LOCATION_READ,
			ttnpb.RIGHT_GATEWAY_STATUS_READ,
			ttnpb.RIGHT_GATEWAY_LOCATION_READ,
		).Unique().GetRights(), should.HaveLength, 3)
	})
	t.Run("Union", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.Union(nilRights).GetRights(), should.BeEmpty)
		a.So(someAppRights.Union(nilRights).GetRights(), should.HaveLength, 2)
		a.So(nilRights.Union(someAppRights).GetRights(), should.HaveLength, 2)
		a.So(someAppRights.Union(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO)).GetRights(), should.HaveLength, 2)
	})
	t.Run("Sub", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.Sub(nilRights).GetRights(), should.BeEmpty)
		a.So(nilRights.Sub(someAppRights).GetRights(), should.BeEmpty)
		a.So(someAppRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO)).GetRights(), should.HaveLength, 1)
		a.So(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO).Sub(someAppRights).GetRights(), should.BeEmpty)
	})
	t.Run("Intersect", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.Intersect(nilRights).GetRights(), should.BeEmpty)
		a.So(someAppRights.Intersect(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO)).GetRights(), should.Contain, ttnpb.RIGHT_APPLICATION_INFO)
		a.So(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO).Intersect(someAppRights).GetRights(), should.Contain, ttnpb.RIGHT_APPLICATION_INFO)
	})
	t.Run("Implied", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.Implied().GetRights(), should.BeEmpty)
		a.So(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_ALL).Implied().GetRights(), should.Contain, ttnpb.RIGHT_APPLICATION_DELETE)
		a.So(ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_ALL).Implied().GetRights(), should.Contain, ttnpb.RIGHT_GATEWAY_DELETE)
		a.So(ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_ALL).Implied().GetRights(), should.Contain, ttnpb.RIGHT_ORGANIZATION_DELETE)
		a.So(ttnpb.RightsFrom(ttnpb.RIGHT_USER_ALL).Implied().GetRights(), should.Contain, ttnpb.RIGHT_USER_DELETE)
	})
	t.Run("IncludesAll", func(t *testing.T) {
		a := assertions.New(t)
		a.So(nilRights.IncludesAll(), should.BeTrue)
		a.So(
			ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC).IncludesAll(ttnpb.RIGHT_APPLICATION_INFO),
			should.BeTrue,
		)
		a.So(
			ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO).IncludesAll(ttnpb.RIGHT_APPLICATION_INFO, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC),
			should.BeFalse,
		)
	})
}
