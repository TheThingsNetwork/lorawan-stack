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
	"reflect"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestFieldIsZero(t *testing.T) {
	t.Parallel()
	for v, paths := range map[interface{ FieldIsZero(string) bool }][]string{
		&ADRSettings{}:                                          ADRSettingsFieldPathsNested,
		(*ADRSettings)(nil):                                     ADRSettingsFieldPathsNested,
		&ADRSettings_StaticMode{}:                               ADRSettings_StaticModeFieldPathsNested,
		(*ADRSettings_StaticMode)(nil):                          ADRSettings_StaticModeFieldPathsNested,
		&ADRSettings_DynamicMode{}:                              ADRSettings_DynamicModeFieldPathsNested,
		(*ADRSettings_DynamicMode)(nil):                         ADRSettings_DynamicModeFieldPathsNested,
		&ADRSettings_DynamicMode_ChannelSteeringSettings{}:      ADRSettings_DynamicMode_ChannelSteeringSettingsFieldPathsNested, // nolint: lll
		(*ADRSettings_DynamicMode_ChannelSteeringSettings)(nil): ADRSettings_DynamicMode_ChannelSteeringSettingsFieldPathsNested, // nolint: lll
		&ADRAckDelayExponentValue{}:                             ADRAckDelayExponentValueFieldPathsNested,
		(*ADRAckDelayExponentValue)(nil):                        ADRAckDelayExponentValueFieldPathsNested,
		&ADRAckLimitExponentValue{}:                             ADRAckLimitExponentValueFieldPathsNested,
		(*ADRAckLimitExponentValue)(nil):                        ADRAckLimitExponentValueFieldPathsNested,
		&AggregatedDutyCycleValue{}:                             AggregatedDutyCycleValueFieldPathsNested,
		(*AggregatedDutyCycleValue)(nil):                        AggregatedDutyCycleValueFieldPathsNested,
		&ApplicationDownlink_ClassBC{}:                          ApplicationDownlink_ClassBCFieldPathsNested,
		(*ApplicationDownlink_ClassBC)(nil):                     ApplicationDownlink_ClassBCFieldPathsNested,
		&ApplicationDownlink{}:                                  ApplicationDownlinkFieldPathsNested,
		(*ApplicationDownlink)(nil):                             ApplicationDownlinkFieldPathsNested,
		&CFList{}:                                               CFListFieldPathsNested,
		(*CFList)(nil):                                          CFListFieldPathsNested,
		&DataRateIndexValue{}:                                   DataRateIndexValueFieldPathsNested,
		(*DataRateIndexValue)(nil):                              DataRateIndexValueFieldPathsNested,
		&DeviceEIRPValue{}:                                      DeviceEIRPValueFieldPathsNested,
		(*DeviceEIRPValue)(nil):                                 DeviceEIRPValueFieldPathsNested,
		&DLSettings{}:                                           DLSettingsFieldPathsNested,
		(*DLSettings)(nil):                                      DLSettingsFieldPathsNested,
		&EndDeviceAuthenticationCode{}:                          EndDeviceAuthenticationCodeFieldPathsNested,
		(*EndDeviceAuthenticationCode)(nil):                     EndDeviceAuthenticationCodeFieldPathsNested,
		&EndDeviceIdentifiers{}:                                 EndDeviceIdentifiersFieldPathsNested,
		(*EndDeviceIdentifiers)(nil):                            EndDeviceIdentifiersFieldPathsNested,
		&EndDeviceVersionIdentifiers{}:                          EndDeviceVersionIdentifiersFieldPathsNested,
		(*EndDeviceVersionIdentifiers)(nil):                     EndDeviceVersionIdentifiersFieldPathsNested,
		&EndDevice{}:                                            EndDeviceFieldPathsNested,
		(*EndDevice)(nil):                                       EndDeviceFieldPathsNested,
		&FCtrl{}:                                                FCtrlFieldPathsNested,
		(*FCtrl)(nil):                                           FCtrlFieldPathsNested,
		&FHDR{}:                                                 FHDRFieldPathsNested,
		(*FHDR)(nil):                                            FHDRFieldPathsNested,
		&JoinAcceptPayload{}:                                    JoinAcceptPayloadFieldPathsNested,
		(*JoinAcceptPayload)(nil):                               JoinAcceptPayloadFieldPathsNested,
		&JoinRequestPayload{}:                                   JoinRequestPayloadFieldPathsNested,
		(*JoinRequestPayload)(nil):                              JoinRequestPayloadFieldPathsNested,
		&JoinRequest{}:                                          JoinRequestFieldPathsNested,
		(*JoinRequest)(nil):                                     JoinRequestFieldPathsNested,
		&MACParameters{}:                                        MACParametersFieldPathsNested,
		(*MACParameters)(nil):                                   MACParametersFieldPathsNested,
		&MACPayload{}:                                           MACPayloadFieldPathsNested,
		(*MACPayload)(nil):                                      MACPayloadFieldPathsNested,
		&MACSettings{}:                                          MACSettingsFieldPathsNested,
		(*MACSettings)(nil):                                     MACSettingsFieldPathsNested,
		&MACState_JoinAccept{}:                                  MACState_JoinAcceptFieldPathsNested,
		(*MACState_JoinAccept)(nil):                             MACState_JoinAcceptFieldPathsNested,
		&MACState_JoinRequest{}:                                 MACState_JoinRequestFieldPathsNested,
		(*MACState_JoinRequest)(nil):                            MACState_JoinRequestFieldPathsNested,
		&MACState{}:                                             MACStateFieldPathsNested,
		(*MACState)(nil):                                        MACStateFieldPathsNested,
		&MessagePayloadFormatters{}:                             MessagePayloadFormattersFieldPathsNested,
		(*MessagePayloadFormatters)(nil):                        MessagePayloadFormattersFieldPathsNested,
		&Message{}:                                              MessageFieldPathsNested,
		(*Message)(nil):                                         MessageFieldPathsNested,
		&MHDR{}:                                                 MHDRFieldPathsNested,
		(*MHDR)(nil):                                            MHDRFieldPathsNested,
		&Picture_Embedded{}:                                     Picture_EmbeddedFieldPathsNested,
		(*Picture_Embedded)(nil):                                Picture_EmbeddedFieldPathsNested,
		&Picture{}:                                              PictureFieldPathsNested,
		(*Picture)(nil):                                         PictureFieldPathsNested,
		&PingSlotPeriodValue{}:                                  PingSlotPeriodValueFieldPathsNested,
		(*PingSlotPeriodValue)(nil):                             PingSlotPeriodValueFieldPathsNested,
		&RejoinRequestPayload{}:                                 RejoinRequestPayloadFieldPathsNested,
		(*RejoinRequestPayload)(nil):                            RejoinRequestPayloadFieldPathsNested,
		&RootKeys{}:                                             RootKeysFieldPathsNested,
		(*RootKeys)(nil):                                        RootKeysFieldPathsNested,
		&RxDelayValue{}:                                         RxDelayValueFieldPathsNested,
		(*RxDelayValue)(nil):                                    RxDelayValueFieldPathsNested,
		&SetEndDeviceRequest{}:                                  EndDeviceFieldPathsNested,
		(*SetEndDeviceRequest)(nil):                             EndDeviceFieldPathsNested,
		&UpdateEndDeviceRequest{}:                               EndDeviceFieldPathsNested,
		(*UpdateEndDeviceRequest)(nil):                          EndDeviceFieldPathsNested,
		&RelayParameters{}:                                      RelayParametersFieldPathsNested,
		(*RelayParameters)(nil):                                 RelayParametersFieldPathsNested,
		&ServedRelayParameters{}:                                ServedRelayParametersFieldPathsNested,
		(*ServedRelayParameters)(nil):                           ServedRelayParametersFieldPathsNested,
		&ServingRelayParameters{}:                               ServingRelayParametersFieldPathsNested,
		(*ServingRelayParameters)(nil):                          ServingRelayParametersFieldPathsNested,
		&RelaySecondChannel{}:                                   RelaySecondChannelFieldPathsNested,
		(*RelaySecondChannel)(nil):                              RelaySecondChannelFieldPathsNested,
		&RelayEndDeviceDynamicMode{}:                            RelayEndDeviceDynamicModeFieldPathsNested,
		(*RelayEndDeviceDynamicMode)(nil):                       RelayEndDeviceDynamicModeFieldPathsNested,
		&RelayForwardLimits{}:                                   RelayForwardLimitsFieldPathsNested,
		(*RelayForwardLimits)(nil):                              RelayForwardLimitsFieldPathsNested,
		&ServingRelayForwardingLimits{}:                         ServingRelayForwardingLimitsFieldPathsNested,
		(*ServingRelayForwardingLimits)(nil):                    ServingRelayForwardingLimitsFieldPathsNested,
	} {
		for _, p := range paths {
			v, p := v, p
			test.RunSubtest(t, test.SubtestConfig{
				Name: fmt.Sprintf("%T(%s)/%s", v, func() string {
					if reflect.ValueOf(v).IsZero() {
						return "nil"
					}
					return "{}"
				}(), p),
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
