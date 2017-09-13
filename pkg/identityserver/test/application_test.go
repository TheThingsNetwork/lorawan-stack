// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	ttn_types "github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func app() *types.DefaultApplication {
	return &types.DefaultApplication{
		ID:          "demo-app",
		Description: "Demo application",
		EUIs: []types.AppEUI{
			types.AppEUI(ttn_types.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1})),
			types.AppEUI(ttn_types.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})),
		},
		APIKeys: []types.ApplicationAPIKey{
			types.ApplicationAPIKey{
				Name: "test-key",
				Key:  "123",
				Rights: []types.Right{
					types.Right("bar"),
					types.Right("foo"),
				},
			},
		},
	}
}

func TestShouldBeApplication(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeApplication(app(), app()), should.Equal, success)

	modified := app()
	modified.Created = time.Now()

	a.So(ShouldBeApplication(modified, app()), should.NotEqual, success)
}

func TestShouldBeApplicationIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeApplicationIgnoringAutoFields(app(), app()), should.Equal, success)

	modified := app()
	modified.EUIs = append(modified.EUIs, app().EUIs[0])

	a.So(ShouldBeApplicationIgnoringAutoFields(modified, app()), should.NotEqual, success)
}
