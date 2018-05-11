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

package networkserver

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
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
