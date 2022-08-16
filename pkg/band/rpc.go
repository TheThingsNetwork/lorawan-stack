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

package band

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/exp/maps"
)

// GetPhyVersions returns the list of supported phy versions for the given band.
func GetPhyVersions(ctx context.Context, req *ttnpb.GetPhyVersionsRequest) (*ttnpb.GetPhyVersionsResponse, error) {
	var res *ttnpb.GetPhyVersionsResponse
	if req.BandId != "" {
		versions, ok := All[req.BandId]
		if !ok {
			return nil, errBandNotFound.WithAttributes("id", req.BandId)
		}
		vs := make([]ttnpb.PHYVersion, 0, len(versions))
		for version := range versions {
			vs = append(vs, version)
		}
		res = &ttnpb.GetPhyVersionsResponse{
			VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
				{
					BandId:      req.BandId,
					PhyVersions: vs,
				},
			},
		}
	} else {
		versionInfo := make([]*ttnpb.GetPhyVersionsResponse_VersionInfo, 0, len(All))
		for bandID, versions := range All {
			vs := make([]ttnpb.PHYVersion, 0, len(versions))
			for version := range versions {
				vs = append(vs, version)
			}
			versionInfo = append(versionInfo, &ttnpb.GetPhyVersionsResponse_VersionInfo{
				BandId:      bandID,
				PhyVersions: vs,
			})
		}
		res = &ttnpb.GetPhyVersionsResponse{
			VersionInfo: versionInfo,
		}
	}
	return res, nil
}

// ListBands returns the list of supported bands.
func ListBands(ctx context.Context, req *ttnpb.ListBandsRequest) (*ttnpb.ListBandsResponse, error) {
	filteredVersions := make(map[string]map[ttnpb.PHYVersion]Band)

	if req.BandId != "" {
		versions, ok := All[req.BandId]
		if !ok {
			return nil, errBandNotFound.WithAttributes("id", req.BandId)
		}

		filteredVersions[req.BandId] = versions
	}

	if len(filteredVersions) == 0 {
		filteredVersions = maps.Clone(All)
	}

	if req.PhyVersion != ttnpb.PHYVersion_PHY_UNKNOWN {
		for bandID, versions := range filteredVersions {
			version, ok := versions[req.PhyVersion]
			if !ok {
				delete(filteredVersions, bandID)
				continue
			}
			filteredVersions[bandID] = make(map[ttnpb.PHYVersion]Band)
			filteredVersions[bandID][req.PhyVersion] = version
		}
	}

	res := &ttnpb.ListBandsResponse{
		Descriptions: make(map[string]*ttnpb.ListBandsResponse_VersionedBandDescription),
	}

	for bandID, versions := range filteredVersions {
		versionedBandDescription := &ttnpb.ListBandsResponse_VersionedBandDescription{
			Band: make(map[string]*ttnpb.BandDescription),
		}

		for PHYVersion, band := range versions {
			versionedBandDescription.Band[PHYVersion.String()] = band.convertToBandDescription()
		}

		res.Descriptions[bandID] = versionedBandDescription
	}

	return res, nil
}

// convertToBandDescription parses a band into a ttnpb.BandDescription
func (band Band) convertToBandDescription() *ttnpb.BandDescription {
	bandDescription := &ttnpb.BandDescription{
		Id: band.ID,
		Beacon: &ttnpb.BandDescription_Beacon{
			DataRateIndex:    band.Beacon.DataRateIndex,
			CodingRate:       band.Beacon.CodingRate,
			InvertedPolarity: band.Beacon.InvertedPolarity,
		},
		PingSlotFrequency:      nil,
		MaxUplinkChannels:      uint32(band.MaxUplinkChannels),
		UplinkChannels:         make([]*ttnpb.BandDescription_Channel, 0, len(band.UplinkChannels)),
		MaxDownlinkChannels:    uint32(band.MaxDownlinkChannels),
		DownlinkChannels:       make([]*ttnpb.BandDescription_Channel, 0, len(band.DownlinkChannels)),
		SubBands:               make([]*ttnpb.BandDescription_SubBandParameters, 0, len(band.SubBands)),
		DataRates:              make(map[uint32]*ttnpb.BandDescription_BandDataRate),
		FreqMultiplier:         band.FreqMultiplier,
		ImplementsCfList:       band.ImplementsCFList,
		CfListType:             band.CFListType,
		ReceiveDelay_1:         types.DurationProto(band.ReceiveDelay1),
		ReceiveDelay_2:         types.DurationProto(band.ReceiveDelay2),
		JoinAcceptDelay_1:      types.DurationProto(band.JoinAcceptDelay1),
		JoinAcceptDelay_2:      types.DurationProto(band.JoinAcceptDelay2),
		MaxFcntGap:             uint64(band.MaxFCntGap),
		SupportsDynamicAdr:     band.SupportsDynamicADR,
		AdrAckLimit:            band.ADRAckLimit,
		MinRetransmitTimeout:   types.DurationProto(band.MinRetransmitTimeout),
		MaxRetransmitTimeout:   types.DurationProto(band.MaxRetransmitTimeout),
		TxOffset:               band.TxOffset,
		MaxAdrDataRateIndex:    band.MaxADRDataRateIndex,
		TxParamSetupReqSupport: band.TxParamSetupReqSupport,
		DefaultMaxEirp:         band.DefaultMaxEIRP,
		LoraCodingRate:         band.LoRaCodingRate,
		DefaultRx2Parameters: &ttnpb.BandDescription_Rx2Parameters{
			DataRateIndex: band.DefaultRx2Parameters.DataRateIndex,
			Frequency:     band.DefaultRx2Parameters.Frequency,
		},
		BootDwellTime: &ttnpb.BandDescription_DwellTime{},
	}

	for _, channel := range band.UplinkChannels {
		bandDescription.UplinkChannels = append(bandDescription.UplinkChannels, &ttnpb.BandDescription_Channel{
			Frequency:   channel.Frequency,
			MinDataRate: channel.MinDataRate,
			MaxDataRate: channel.MaxDataRate,
		})
	}

	for _, channel := range band.DownlinkChannels {
		bandDescription.DownlinkChannels = append(bandDescription.DownlinkChannels, &ttnpb.BandDescription_Channel{
			Frequency:   channel.Frequency,
			MinDataRate: channel.MinDataRate,
			MaxDataRate: channel.MaxDataRate,
		})
	}

	for _, subbands := range band.SubBands {
		bandDescription.SubBands = append(bandDescription.SubBands, &ttnpb.BandDescription_SubBandParameters{
			MinFrequency: subbands.MinFrequency,
			MaxFrequency: subbands.MaxFrequency,
			DutyCycle:    subbands.DutyCycle,
			MaxEirp:      subbands.MaxEIRP,
		})
	}

	for index, datarate := range band.DataRates {
		bandDescription.DataRates[uint32(index)] = &ttnpb.BandDescription_BandDataRate{
			Rate: &ttnpb.DataRate{
				Modulation: datarate.Rate.Modulation,
			},
		}
	}

	if band.PingSlotFrequency != nil {
		bandDescription.PingSlotFrequency = &types.UInt64Value{
			Value: *band.PingSlotFrequency,
		}
	}

	if band.BootDwellTime.Uplinks != nil {
		bandDescription.BootDwellTime.Uplinks = &types.BoolValue{
			Value: *band.BootDwellTime.Uplinks,
		}
	}

	if band.BootDwellTime.Downlinks != nil {
		bandDescription.BootDwellTime.Downlinks = &types.BoolValue{
			Value: *band.BootDwellTime.Downlinks,
		}
	}

	return bandDescription
}
