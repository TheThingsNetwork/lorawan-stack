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

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var now = time.Now()

func testSettings() *ttnpb.IdentityServerSettings {
	return &ttnpb.IdentityServerSettings{
		BlacklistedIDs:     []string{"admin"},
		AllowedEmails:      []string{},
		UpdatedAt:          now,
		ValidationTokenTTL: time.Duration(time.Hour),
	}
}

func TestShouldBeSettings(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeSettings(testSettings(), testSettings()), should.Equal, success)

	modified := testSettings()
	modified.ValidationTokenTTL = time.Duration(time.Hour * 2)

	a.So(ShouldBeSettings(modified, testSettings()), should.NotEqual, success)
}

func TestShouldBeSettingsIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeSettingsIgnoringAutoFields(testSettings(), testSettings()), should.Equal, success)

	modified := testSettings()
	modified.AllowedEmails = nil

	a.So(ShouldBeSettingsIgnoringAutoFields(modified, testSettings()), should.NotEqual, success)
}
