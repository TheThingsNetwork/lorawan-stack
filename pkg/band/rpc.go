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
	"maps"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GetPhyVersions returns the list of supported phy versions for the given band.
func GetPhyVersions(_ context.Context, req *ttnpb.GetPhyVersionsRequest) (*ttnpb.GetPhyVersionsResponse, error) {
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
func ListBands(_ context.Context, req *ttnpb.ListBandsRequest) (*ttnpb.ListBandsResponse, error) {
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
			versionedBandDescription.Band[PHYVersion.String()] = band.BandDescription()
		}

		res.Descriptions[bandID] = versionedBandDescription
	}

	return res, nil
}

// BandDescription parses a band into a ttnpb.BandDescription.
func (b Band) BandDescription() *ttnpb.BandDescription {
	bandDescription := &ttnpb.BandDescription{
		Id: b.ID,
		Beacon: &ttnpb.BandDescription_Beacon{
			DataRateIndex: b.Beacon.DataRateIndex,
			CodingRate:    b.Beacon.CodingRate,
			Frequencies:   b.Beacon.Frequencies,
		},
		PingSlotFrequencies:    b.PingSlotFrequencies,
		MaxUplinkChannels:      uint32(b.MaxUplinkChannels),
		UplinkChannels:         make([]*ttnpb.BandDescription_Channel, 0, len(b.UplinkChannels)),
		MaxDownlinkChannels:    uint32(b.MaxDownlinkChannels),
		DownlinkChannels:       make([]*ttnpb.BandDescription_Channel, 0, len(b.DownlinkChannels)),
		SubBands:               make([]*ttnpb.BandDescription_SubBandParameters, 0, len(b.SubBands)),
		DataRates:              make(map[uint32]*ttnpb.BandDescription_BandDataRate),
		FreqMultiplier:         b.FreqMultiplier,
		ImplementsCfList:       b.ImplementsCFList,
		CfListType:             b.CFListType,
		ReceiveDelay_1:         durationpb.New(b.ReceiveDelay1),
		ReceiveDelay_2:         durationpb.New(b.ReceiveDelay2),
		JoinAcceptDelay_1:      durationpb.New(b.JoinAcceptDelay1),
		JoinAcceptDelay_2:      durationpb.New(b.JoinAcceptDelay2),
		MaxFcntGap:             uint64(b.MaxFCntGap),
		SupportsDynamicAdr:     b.SupportsDynamicADR,
		AdrAckLimit:            b.ADRAckLimit,
		MinRetransmitTimeout:   durationpb.New(b.MinRetransmitTimeout),
		MaxRetransmitTimeout:   durationpb.New(b.MaxRetransmitTimeout),
		TxOffset:               b.TxOffset,
		MaxAdrDataRateIndex:    b.MaxADRDataRateIndex,
		TxParamSetupReqSupport: b.TxParamSetupReqSupport,
		DefaultMaxEirp:         b.DefaultMaxEIRP,
		DefaultRx2Parameters: &ttnpb.BandDescription_Rx2Parameters{
			DataRateIndex: b.DefaultRx2Parameters.DataRateIndex,
			Frequency:     b.DefaultRx2Parameters.Frequency,
		},
		BootDwellTime: &ttnpb.BandDescription_DwellTime{},
		Relay:         &ttnpb.BandDescription_RelayParameters{},
	}

	if b.SharedParameters.RelayForwardDelay != 0 {
		bandDescription.RelayForwardDelay = durationpb.New(b.SharedParameters.RelayForwardDelay)
	}
	if b.SharedParameters.RelayReceiveDelay != 0 {
		bandDescription.RelayReceiveDelay = durationpb.New(b.SharedParameters.RelayReceiveDelay)
	}

	for _, channel := range b.UplinkChannels {
		bandDescription.UplinkChannels = append(bandDescription.UplinkChannels, &ttnpb.BandDescription_Channel{
			Frequency:   channel.Frequency,
			MinDataRate: channel.MinDataRate,
			MaxDataRate: channel.MaxDataRate,
		})
	}

	for _, channel := range b.DownlinkChannels {
		bandDescription.DownlinkChannels = append(bandDescription.DownlinkChannels, &ttnpb.BandDescription_Channel{
			Frequency:   channel.Frequency,
			MinDataRate: channel.MinDataRate,
			MaxDataRate: channel.MaxDataRate,
		})
	}

	for _, subbands := range b.SubBands {
		bandDescription.SubBands = append(bandDescription.SubBands, &ttnpb.BandDescription_SubBandParameters{
			MinFrequency: subbands.MinFrequency,
			MaxFrequency: subbands.MaxFrequency,
			DutyCycle:    subbands.DutyCycle,
			MaxEirp:      subbands.MaxEIRP,
		})
	}

	for index, dataRate := range b.DataRates {
		bandDescription.DataRates[uint32(index)] = &ttnpb.BandDescription_BandDataRate{
			Rate: &ttnpb.DataRate{
				Modulation: dataRate.Rate.Modulation,
			},
		}
	}

	if b.BootDwellTime.Uplinks != nil {
		bandDescription.BootDwellTime.Uplinks = &wrapperspb.BoolValue{
			Value: *b.BootDwellTime.Uplinks,
		}
	}

	if b.BootDwellTime.Downlinks != nil {
		bandDescription.BootDwellTime.Downlinks = &wrapperspb.BoolValue{
			Value: *b.BootDwellTime.Downlinks,
		}
	}

	for _, ch := range b.Relay.WORChannels {
		bandDescription.Relay.WorChannels = append(
			bandDescription.Relay.WorChannels, &ttnpb.BandDescription_RelayParameters_RelayWORChannel{
				Frequency:     ch.Frequency,
				AckFrequency:  ch.ACKFrequency,
				DataRateIndex: ch.DataRateIndex,
			},
		)
	}

	return bandDescription
}
