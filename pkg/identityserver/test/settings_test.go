// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
