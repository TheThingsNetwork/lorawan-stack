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

package gatewayserver_test

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestGetFrequencyPlan(t *testing.T) {
	a := assertions.New(t)

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	gs := gatewayserver.New(c, gatewayserver.Config{
		FileFrequencyPlansStore: dir,
	})

	fp, err := gs.GetFrequencyPlan(context.Background(), &ttnpb.GetFrequencyPlanRequest{FrequencyPlanID: "EU_863_870"})
	a.So(err, should.BeNil)
	a.So(fp.BandID, should.Equal, "EU_863_870")
	a.So(len(fp.Channels), should.Equal, 8)

	_, err = gs.GetFrequencyPlan(context.Background(), &ttnpb.GetFrequencyPlanRequest{FrequencyPlanID: "FP_THAT_DOES_NOT_EXIST"})
	a.So(err, should.NotBeNil)

	defer gs.Close()
}
