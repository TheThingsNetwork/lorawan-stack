// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	needEventCompatible                      = "This assertion requires a Event-compatible comparison type (you provided %T)."
	needEventDefinitionDataClosureCompatible = "This assertion requires an EventDefinitionDataClosure-compatible comparison type (you provided %T)."
)

func ShouldResembleEvent(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	ee, ok := expected[0].(events.Event)
	if !ok {
		return fmt.Sprintf(needEventCompatible, expected[0])
	}
	ae, ok := actual.(events.Event)
	if !ok {
		return fmt.Sprintf(needEventCompatible, actual)
	}
	ep, err := events.Proto(ee)
	if s := assertions.ShouldBeNil(err); s != success {
		return s
	}
	ap, err := events.Proto(ae)
	if s := assertions.ShouldBeNil(err); s != success {
		return s
	}
	ap.Time = time.Time{}
	ep.Time = time.Time{}
	return ShouldResemble(ap, ep)
}

func ShouldResembleEventDefinitionDataClosure(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	ed, ok := expected[0].(events.DefinitionDataClosure)
	if !ok {
		return fmt.Sprintf(needEventDefinitionDataClosureCompatible, expected[0])
	}
	ad, ok := actual.(events.DefinitionDataClosure)
	if !ok {
		return fmt.Sprintf(needEventDefinitionDataClosureCompatible, actual)
	}
	ctx := context.Background()
	ids := &ttnpb.EndDeviceIdentifiers{
		DeviceID: "test-dev",
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "test-app",
		},
	}
	return ShouldResembleEvent(ad(ctx, ids), ed(ctx, ids))
}

func ShouldResembleEventDefinitionDataClosures(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	eds, ok := expected[0].([]events.DefinitionDataClosure)
	if !ok {
		return fmt.Sprintf(needEventDefinitionDataClosureCompatible, expected[0])
	}
	ads, ok := actual.([]events.DefinitionDataClosure)
	if !ok {
		return fmt.Sprintf(needEventDefinitionDataClosureCompatible, actual)
	}
	ctx := context.Background()
	ids := &ttnpb.EndDeviceIdentifiers{
		DeviceID: "test-dev",
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "test-app",
		},
	}
	if s := assertions.ShouldHaveLength(ads, len(eds)); s != success {
		return s
	}
	for i, ad := range ads {
		if s := ShouldResembleEvent(ad(ctx, ids), eds[i](ctx, ids)); s != success {
			return fmt.Sprintf("Mismatch in event definition %d: %s", i, s)
		}
	}
	return success
}
