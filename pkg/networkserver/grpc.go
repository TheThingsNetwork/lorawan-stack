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

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GenerateDevAddr returns a device address assignment in the device address
// range of the network server.
func (ns *NetworkServer) GenerateDevAddr(ctx context.Context, req *emptypb.Empty) (*ttnpb.GenerateDevAddrResponse, error) {
	devAddr := ns.newDevAddr(ctx)
	return &ttnpb.GenerateDevAddrResponse{DevAddr: devAddr.Bytes()}, nil
}

func (ns *NetworkServer) GetDefaultMACSettings(ctx context.Context, req *ttnpb.GetDefaultMACSettingsRequest) (*ttnpb.MACSettings, error) {
	fps, err := ns.FrequencyPlansStore(ctx)
	if err != nil {
		return nil, err
	}
	fp, phy, err := FrequencyPlanAndBand(req.FrequencyPlanId, req.LorawanPhyVersion, fps)
	if err != nil {
		return nil, err
	}
	classBTimeout := mac.DeviceClassBTimeout(nil, ns.defaultMACSettings)
	classCTimeout := mac.DeviceClassCTimeout(nil, ns.defaultMACSettings)
	adrMargin := mac.DeviceADRMargin(nil, ns.defaultMACSettings)
	statusTimePeriodicity := mac.DeviceStatusTimePeriodicity(nil, ns.defaultMACSettings)
	statusCountPeriodicity := mac.DeviceStatusCountPeriodicity(nil, ns.defaultMACSettings)
	settings := &ttnpb.MACSettings{
		ClassBTimeout:                durationpb.New(classBTimeout),
		PingSlotPeriodicity:          mac.DeviceDefaultPingSlotPeriodicity(nil, ns.defaultMACSettings),
		PingSlotDataRateIndex:        mac.DeviceDefaultPingSlotDataRateIndexValue(nil, phy, ns.defaultMACSettings),
		PingSlotFrequency:            &ttnpb.ZeroableFrequencyValue{Value: mac.DeviceDefaultPingSlotFrequency(nil, phy, ns.defaultMACSettings)},
		BeaconFrequency:              &ttnpb.ZeroableFrequencyValue{Value: mac.DeviceDefaultBeaconFrequency(nil, phy, ns.defaultMACSettings)},
		ClassCTimeout:                durationpb.New(classCTimeout),
		Rx1Delay:                     &ttnpb.RxDelayValue{Value: mac.DeviceDefaultRX1Delay(nil, phy, ns.defaultMACSettings)},
		Rx1DataRateOffset:            &ttnpb.DataRateOffsetValue{Value: mac.DeviceDefaultRX1DataRateOffset(nil, ns.defaultMACSettings)},
		Rx2DataRateIndex:             &ttnpb.DataRateIndexValue{Value: mac.DeviceDefaultRX2DataRateIndex(nil, phy, ns.defaultMACSettings)},
		Rx2Frequency:                 &ttnpb.FrequencyValue{Value: mac.DeviceDefaultRX2Frequency(nil, phy, ns.defaultMACSettings)},
		MaxDutyCycle:                 &ttnpb.AggregatedDutyCycleValue{Value: mac.DeviceDefaultMaxDutyCycle(nil, ns.defaultMACSettings)},
		Supports_32BitFCnt:           &ttnpb.BoolValue{Value: mac.DeviceSupports32BitFCnt(nil, ns.defaultMACSettings)},
		UseAdr:                       &ttnpb.BoolValue{Value: mac.DeviceUseADR(nil, ns.defaultMACSettings, phy)},
		AdrMargin:                    &wrapperspb.FloatValue{Value: adrMargin},
		ResetsFCnt:                   &ttnpb.BoolValue{Value: mac.DeviceResetsFCnt(nil, ns.defaultMACSettings)},
		StatusTimePeriodicity:        durationpb.New(statusTimePeriodicity),
		StatusCountPeriodicity:       &wrapperspb.UInt32Value{Value: statusCountPeriodicity},
		DesiredRx1Delay:              &ttnpb.RxDelayValue{Value: mac.DeviceDesiredRX1Delay(nil, phy, ns.defaultMACSettings)},
		DesiredRx1DataRateOffset:     &ttnpb.DataRateOffsetValue{Value: mac.DeviceDesiredRX1DataRateOffset(nil, ns.defaultMACSettings)},
		DesiredRx2DataRateIndex:      &ttnpb.DataRateIndexValue{Value: mac.DeviceDesiredRX2DataRateIndex(nil, phy, fp, ns.defaultMACSettings)},
		DesiredRx2Frequency:          &ttnpb.FrequencyValue{Value: mac.DeviceDesiredRX2Frequency(nil, phy, fp, ns.defaultMACSettings)},
		DesiredMaxDutyCycle:          &ttnpb.AggregatedDutyCycleValue{Value: mac.DeviceDesiredMaxDutyCycle(nil, ns.defaultMACSettings)},
		DesiredAdrAckLimitExponent:   mac.DeviceDesiredADRAckLimitExponent(nil, phy, ns.defaultMACSettings),
		DesiredAdrAckDelayExponent:   mac.DeviceDesiredADRAckDelayExponent(nil, phy, ns.defaultMACSettings),
		DesiredPingSlotDataRateIndex: mac.DeviceDesiredPingSlotDataRateIndexValue(nil, phy, fp, ns.defaultMACSettings),
		DesiredPingSlotFrequency:     &ttnpb.ZeroableFrequencyValue{Value: mac.DeviceDesiredPingSlotFrequency(nil, phy, fp, ns.defaultMACSettings)},
		DesiredBeaconFrequency:       &ttnpb.ZeroableFrequencyValue{Value: mac.DeviceDesiredBeaconFrequency(nil, phy, ns.defaultMACSettings)},
		DesiredMaxEirp:               &ttnpb.DeviceEIRPValue{Value: lorawan.Float32ToDeviceEIRP(mac.DeviceDesiredMaxEIRP(nil, phy, fp, ns.defaultMACSettings))},
		UplinkDwellTime:              mac.DeviceUplinkDwellTime(nil, phy, ns.defaultMACSettings),
		DownlinkDwellTime:            mac.DeviceDownlinkDwellTime(nil, phy, ns.defaultMACSettings),
		ScheduleDownlinks:            &ttnpb.BoolValue{Value: mac.DeviceScheduleDownlinks(nil, ns.defaultMACSettings)},
		Relay:                        mac.DeviceDefaultRelayParameters(nil, ns.defaultMACSettings),
		DesiredRelay:                 mac.DeviceDesiredRelayParameters(nil, ns.defaultMACSettings),
	}
	return settings, nil
}

// GetNetID returns the NetID of the Network Server.
func (ns *NetworkServer) GetNetID(ctx context.Context, _ *emptypb.Empty) (*ttnpb.GetNetIDResponse, error) {
	return &ttnpb.GetNetIDResponse{
		NetId: ns.netID(ctx).Bytes(),
	}, nil
}

// GetDeviceAddressPrefixes return the configured device address prefixes of the Network Server.
func (ns *NetworkServer) GetDeviceAddressPrefixes(
	ctx context.Context, _ *emptypb.Empty,
) (*ttnpb.GetDeviceAdressPrefixesResponse, error) {
	output := &ttnpb.GetDeviceAdressPrefixesResponse{}

	prefixes := ns.devAddrPrefixes(ctx)

	for _, devAddrPrefix := range prefixes {
		output.DevAddrPrefixes = append(output.DevAddrPrefixes, devAddrPrefix.Bytes())
	}

	return output, nil
}
