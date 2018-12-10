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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestAccumulator(t *testing.T) {
	a := assertions.New(t)

	acc := newMetadataAccumulator()
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

	acc = newMetadataAccumulator(vals...)
	a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals)
	acc.Reset()
	a.So(acc.Accumulated(), should.BeEmpty)

	acc.Add(vals[0], vals[1], vals[2])
	a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals[:3])

	for i := 2; i < len(vals); i++ {
		acc.Add(vals[i])
		a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals[:i+1])
	}
	a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals)

	acc.Reset()
	a.So(acc.Accumulated(), should.BeEmpty)
}
