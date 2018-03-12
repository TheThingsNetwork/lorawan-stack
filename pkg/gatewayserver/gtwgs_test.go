// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
	gs, err := gatewayserver.New(c, &gatewayserver.Config{
		LocalFrequencyPlansStore: dir,
	})
	a.So(err, should.BeNil)

	fp, err := gs.GetFrequencyPlan(context.Background(), &ttnpb.FrequencyPlanRequest{FrequencyPlanID: "EU_863_870"})
	a.So(err, should.BeNil)
	a.So(fp.BandID, should.Equal, "EU_863_870")
	a.So(len(fp.Channels), should.Equal, 8)

	_, err = gs.GetFrequencyPlan(context.Background(), &ttnpb.FrequencyPlanRequest{FrequencyPlanID: "FP_THAT_DOES_NOT_EXIST"})
	a.So(err, should.NotBeNil)

	defer gs.Close()
}
