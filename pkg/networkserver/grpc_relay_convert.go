// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func relayParametersFromConfiguration(conf *ttnpb.RelayConfiguration) *ttnpb.RelayParameters {
	if conf == nil {
		return nil
	}
	switch mode := conf.Mode.(type) {
	case *ttnpb.RelayConfiguration_Served_:
		served := &ttnpb.ServedRelayParameters{
			Backoff:         mode.Served.Backoff,
			SecondChannel:   mode.Served.SecondChannel,
			ServingDeviceId: mode.Served.ServingDeviceId,
		}
		switch mode := mode.Served.Mode.(type) {
		case *ttnpb.RelayConfiguration_Served_Always:
			served.Mode = &ttnpb.ServedRelayParameters_Always{
				Always: mode.Always,
			}
		case *ttnpb.RelayConfiguration_Served_Dynamic:
			served.Mode = &ttnpb.ServedRelayParameters_Dynamic{
				Dynamic: mode.Dynamic,
			}
		case *ttnpb.RelayConfiguration_Served_EndDeviceControlled:
			served.Mode = &ttnpb.ServedRelayParameters_EndDeviceControlled{
				EndDeviceControlled: mode.EndDeviceControlled,
			}
		case nil:
		default:
			panic(fmt.Sprintf("unknown mode %T", mode))
		}
		return &ttnpb.RelayParameters{
			Mode: &ttnpb.RelayParameters_Served{
				Served: served,
			},
		}
	case *ttnpb.RelayConfiguration_Serving_:
		return &ttnpb.RelayParameters{
			Mode: &ttnpb.RelayParameters_Serving{
				Serving: &ttnpb.ServingRelayParameters{
					SecondChannel:       mode.Serving.SecondChannel,
					DefaultChannelIndex: mode.Serving.DefaultChannelIndex,
					CadPeriodicity:      mode.Serving.CadPeriodicity,
					Limits:              mode.Serving.Limits,
				},
			},
		}
	case nil:
		return &ttnpb.RelayParameters{}
	default:
		panic(fmt.Sprintf("unknown mode %T", mode))
	}
}

func relayConfigurationFromParameters(params *ttnpb.RelayParameters) *ttnpb.RelayConfiguration {
	if params == nil {
		return nil
	}
	switch mode := params.Mode.(type) {
	case *ttnpb.RelayParameters_Served:
		served := &ttnpb.RelayConfiguration_Served{
			Backoff:         mode.Served.Backoff,
			SecondChannel:   mode.Served.SecondChannel,
			ServingDeviceId: mode.Served.ServingDeviceId,
		}
		switch mode := mode.Served.Mode.(type) {
		case *ttnpb.ServedRelayParameters_Always:
			served.Mode = &ttnpb.RelayConfiguration_Served_Always{
				Always: mode.Always,
			}
		case *ttnpb.ServedRelayParameters_Dynamic:
			served.Mode = &ttnpb.RelayConfiguration_Served_Dynamic{
				Dynamic: mode.Dynamic,
			}
		case *ttnpb.ServedRelayParameters_EndDeviceControlled:
			served.Mode = &ttnpb.RelayConfiguration_Served_EndDeviceControlled{
				EndDeviceControlled: mode.EndDeviceControlled,
			}
		case nil:
		default:
			panic(fmt.Sprintf("unknown mode %T", mode))
		}
		return &ttnpb.RelayConfiguration{
			Mode: &ttnpb.RelayConfiguration_Served_{
				Served: served,
			},
		}
	case *ttnpb.RelayParameters_Serving:
		return &ttnpb.RelayConfiguration{
			Mode: &ttnpb.RelayConfiguration_Serving_{
				Serving: &ttnpb.RelayConfiguration_Serving{
					SecondChannel:       mode.Serving.SecondChannel,
					DefaultChannelIndex: mode.Serving.DefaultChannelIndex,
					CadPeriodicity:      mode.Serving.CadPeriodicity,
					Limits:              mode.Serving.Limits,
				},
			},
		}
	case nil:
		return &ttnpb.RelayConfiguration{}
	default:
		panic(fmt.Sprintf("unknown mode %T", mode))
	}
}

func relayUplinkForwardingRuleFromConfiguration(
	rule *ttnpb.RelayConfigurationUplinkForwardingRule,
) *ttnpb.RelayUplinkForwardingRule {
	if rule == nil {
		return nil
	}
	return &ttnpb.RelayUplinkForwardingRule{
		Limits:    rule.Limits,
		LastWFCnt: rule.LastWFCnt,
		DeviceId:  rule.DeviceId,
	}
}

func relayConfigurationUplinkForwardingRuleFromUplinkForwardingRule(
	rule *ttnpb.RelayUplinkForwardingRule,
) *ttnpb.RelayConfigurationUplinkForwardingRule {
	if rule == nil {
		return nil
	}
	return &ttnpb.RelayConfigurationUplinkForwardingRule{
		Limits:    rule.Limits,
		LastWFCnt: rule.LastWFCnt,
		DeviceId:  rule.DeviceId,
	}
}
