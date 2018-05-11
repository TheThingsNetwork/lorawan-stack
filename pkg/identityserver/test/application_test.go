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

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func app() *ttnpb.Application {
	return &ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "demo-app"},
		Description:            "Demo application",
	}
}

func TestShouldBeApplication(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeApplication(app(), app()), should.Equal, success)

	modified := app()
	modified.CreatedAt = time.Now()

	a.So(ShouldBeApplication(modified, app()), should.NotEqual, success)
}

func TestShouldBeApplicationIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeApplicationIgnoringAutoFields(app(), app()), should.Equal, success)

	modified := app()
	modified.Description = "foo"

	a.So(ShouldBeApplicationIgnoringAutoFields(modified, app()), should.NotEqual, success)
}
