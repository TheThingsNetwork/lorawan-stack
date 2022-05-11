// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package devicetemplates_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func validateTemplate(t *testing.T, tmpl *ttnpb.EndDeviceTemplate) {
	a := assertions.New(t)
	if !a.So(tmpl, should.NotBeNil) {
		t.FailNow()
	}

	dev := &ttnpb.EndDevice{}
	a.So(dev.SetFields(tmpl.EndDevice, tmpl.FieldMask.GetPaths()...), should.BeNil)
	a.So(dev, should.Resemble, tmpl.EndDevice)
}

func validateTemplates(t *testing.T, templates []*ttnpb.EndDeviceTemplate, count int) {
	a := assertions.New(t)

	if !a.So(len(templates), should.Equal, count) {
		t.FailNow()
	}

	for _, template := range templates {
		validateTemplate(t, template)
	}
}
