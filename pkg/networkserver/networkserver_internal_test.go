// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

const (
	RecentUplinkCount    = recentUplinkCount
	MaxFCntGap           = maxFCntGap
	AccumulationCapacity = accumulationCapacity
)

func TestAccumulator(t *testing.T) {
	a := assertions.New(t)

	acc := &metadataAccumulator{newAccumulator()}
	a.So(func() { acc.Add() }, should.NotPanic)

	vals := []*ttnpb.RxMetadata{
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		nil,
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
	}

	a.So(acc.Accumulated(), should.BeEmpty)

	acc.Add(vals[0], vals[1], vals[2])
	for _, v := range vals[:3] {
		a.So(acc.Accumulated(), should.Contain, v)
	}

	for i := 2; i < len(vals); i++ {
		acc.Add(vals[i])
		for _, v := range vals[:i] {
			a.So(acc.Accumulated(), should.Contain, v)
		}
	}

	acc.Reset()
	a.So(acc.Accumulated(), should.BeEmpty)
}
