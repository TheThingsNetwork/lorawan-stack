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
	"sync"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type accumulator struct {
	m sync.Map
}

func (acc *accumulator) Add(v interface{}) {
	acc.m.Store(v, struct{}{})
}

func (acc *accumulator) Range(f func(v interface{})) {
	acc.m.Range(func(k, _ interface{}) bool {
		f(k)
		return true
	})
}

func (acc *accumulator) Reset() {
	acc.Range(acc.m.Delete)
}

func newAccumulator(vs ...interface{}) *accumulator {
	var acc accumulator
	for _, v := range vs {
		acc.Add(v)
	}
	return &acc
}

type metadataAccumulator struct {
	accumulator
}

func (acc *metadataAccumulator) Accumulated() (md []*ttnpb.RxMetadata) {
	md = make([]*ttnpb.RxMetadata, 0, accumulationCapacity)
	acc.accumulator.Range(func(k interface{}) {
		md = append(md, k.(*ttnpb.RxMetadata))
	})
	return
}

func (acc *metadataAccumulator) Add(mds ...*ttnpb.RxMetadata) {
	for _, md := range mds {
		acc.accumulator.Add(md)
	}
}

func newMetadataAccumulator(mds ...*ttnpb.RxMetadata) *metadataAccumulator {
	var acc metadataAccumulator
	acc.Add(mds...)
	return &acc
}
