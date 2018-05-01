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

package test_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/fetch"
	"github.com/TheThingsNetwork/ttn/pkg/frequencyplans"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func ExampleFrequencyPlansStore() {
	store, err := test.NewFrequencyPlansStore()
	if err != nil {
		panic(err)
	}

	defer store.Destroy()

	config := config.ServiceBase{
		FrequencyPlans: config.FrequencyPlans{
			StoreDirectory: store.Directory(),
		},
	}

	component, err := component.New(log.Default, &component.Config{
		ServiceBase: config,
	})
	if err != nil {
		panic(err)
	}

	gs, err := gatewayserver.New(component, gatewayserver.Config{})
	if err != nil {
		panic(err)
	}

	gs.Start()
}

func TestFrequencyPlans(t *testing.T) {
	a := assertions.New(t)

	s, err := test.NewFrequencyPlansStore()
	a.So(err, should.BeNil)

	a.So(s.Directory(), should.NotEqual, "")

	fp := frequencyplans.NewStore(fetch.FromFilesystem(s.Directory()))

	euFP, err := fp.GetByID(test.EUFrequencyPlanID)
	a.So(err, should.BeNil)
	a.So(euFP.BandID, should.Equal, "EU_863_870")

	err = s.Destroy()
	a.So(err, should.BeNil)
}
