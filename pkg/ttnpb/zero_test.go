// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestFieldIsZero(t *testing.T) {
	for v, paths := range map[interface{ FieldIsZero(string) bool }][]string{
		&ADRAckDelayExponentValue{}:    ADRAckDelayExponentValueFieldPathsNested,
		&ADRAckLimitExponentValue{}:    ADRAckLimitExponentValueFieldPathsNested,
		&AggregatedDutyCycleValue{}:    AggregatedDutyCycleValueFieldPathsNested,
		&ApplicationDownlink_ClassBC{}: ApplicationDownlink_ClassBCFieldPathsNested,
		&ApplicationDownlink{}:         ApplicationDownlinkFieldPathsNested,
		&CFList{}:                      CFListFieldPathsNested,
		&DataRateIndexValue{}:          DataRateIndexValueFieldPathsNested,
		&DLSettings{}:                  DLSettingsFieldPathsNested,
		&EndDeviceAuthenticationCode{}: EndDeviceAuthenticationCodeFieldPathsNested,
		&EndDeviceIdentifiers{}:        EndDeviceIdentifiersFieldPathsNested,
		&EndDeviceVersionIdentifiers{}: EndDeviceVersionIdentifiersFieldPathsNested,
		&EndDevice{}:                   EndDeviceFieldPathsNested,
		&FCtrl{}:                       FCtrlFieldPathsNested,
		&FHDR{}:                        FHDRFieldPathsNested,
		&JoinAcceptPayload{}:           JoinAcceptPayloadFieldPathsNested,
		&JoinRequestPayload{}:          JoinRequestPayloadFieldPathsNested,
		&JoinRequest{}:                 JoinRequestFieldPathsNested,
		&MACParameters{}:               MACParametersFieldPathsNested,
		&MACPayload{}:                  MACPayloadFieldPathsNested,
		&MACSettings{}:                 MACSettingsFieldPathsNested,
		&MACState_JoinAccept{}:         MACState_JoinAcceptFieldPathsNested,
		&MACState{}:                    MACStateFieldPathsNested,
		&MessagePayloadFormatters{}:    MessagePayloadFormattersFieldPathsNested,
		&Message{}:                     MessageFieldPathsNested,
		&MHDR{}:                        MHDRFieldPathsNested,
		&Picture_Embedded{}:            Picture_EmbeddedFieldPathsNested,
		&Picture{}:                     PictureFieldPathsNested,
		&PingSlotPeriodValue{}:         PingSlotPeriodValueFieldPathsNested,
		&RejoinRequestPayload{}:        RejoinRequestPayloadFieldPathsNested,
		&RootKeys{}:                    RootKeysFieldPathsNested,
		&RxDelayValue{}:                RxDelayValueFieldPathsNested,
		&SetEndDeviceRequest{}:         EndDeviceFieldPathsNested,
		&UpdateEndDeviceRequest{}:      EndDeviceFieldPathsNested,
	} {
		for _, p := range paths {
			v, p := v, p
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%T/%s", v, p),
				Parallel: true,
				Func: func(_ context.Context, _ *testing.T, a *assertions.Assertion) {
					var ok bool
					if a.So(func() {
						ok = v.FieldIsZero(p)
					}, should.NotPanic) {
						a.So(ok, should.BeTrue)
					}
				},
			})
		}
	}
}
