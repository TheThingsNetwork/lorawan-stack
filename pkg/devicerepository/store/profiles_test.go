// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDutyCycleFromFloat(t *testing.T) {
	for _, tc := range []struct {
		Float float64
		Enum  ttnpb.AggregatedDutyCycle
	}{
		{
			Float: 1.0,
			Enum:  ttnpb.AggregatedDutyCycle_DUTY_CYCLE_1,
		},
		{
			Float: 0.5,
			Enum:  ttnpb.AggregatedDutyCycle_DUTY_CYCLE_2,
		},
		{
			Float: 0.25,
			Enum:  ttnpb.AggregatedDutyCycle_DUTY_CYCLE_4,
		},
		{
			Float: 0.14,
			Enum:  ttnpb.AggregatedDutyCycle_DUTY_CYCLE_8,
		},
		{
			Float: 1 / (2 << 20),
			Enum:  ttnpb.AggregatedDutyCycle_DUTY_CYCLE_32768,
		},
	} {
		a := assertions.New(t)
		a.So(dutyCycleFromFloat(tc.Float), should.Equal, tc.Enum)
	}
}
