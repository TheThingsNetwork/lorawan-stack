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

package band_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type serializableBeacon struct {
	DataRateIndex ttnpb.DataRateIndex
	CodingRate    string
	Frequencies   []uint64
}

type serializableDataRate struct {
	Rate              *ttnpb.DataRate
	MaxMACPayloadSize map[bool]uint16
}

func (dr serializableDataRate) MarshalJSON() ([]byte, error) {
	drBytes, err := jsonpb.TTN().Marshal(dr.Rate)
	if err != nil {
		return nil, err
	}
	sizeBytes, err := jsonpb.TTN().Marshal(dr.MaxMACPayloadSize)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]json.RawMessage{
		"Rate":              drBytes,
		"MaxMACPayloadSize": sizeBytes,
	})
}

func (dr *serializableDataRate) UnmarshalJSON(b []byte) error {
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if err := jsonpb.TTN().Unmarshal(m["Rate"], &dr.Rate); err != nil {
		return err
	}
	return jsonpb.TTN().Unmarshal(m["MaxMACPayloadSize"], &dr.MaxMACPayloadSize)
}

func makeDataRates(m map[ttnpb.DataRateIndex]band.DataRate) map[ttnpb.DataRateIndex]serializableDataRate {
	sm := make(map[ttnpb.DataRateIndex]serializableDataRate)
	for idx, dr := range m {
		sm[idx] = serializableDataRate{
			Rate: dr.Rate,
			MaxMACPayloadSize: map[bool]uint16{
				false: dr.MaxMACPayloadSize(false),
				true:  dr.MaxMACPayloadSize(true),
			},
		}
	}
	return sm
}

type serializableBand struct {
	ID string

	Beacon              serializableBeacon
	PingSlotFrequencies []uint64

	MaxUplinkChannels uint8
	UplinkChannels    []band.Channel

	MaxDownlinkChannels uint8
	DownlinkChannels    []band.Channel

	SubBands []band.SubBandParameters

	DataRates        map[ttnpb.DataRateIndex]serializableDataRate
	StrictCodingRate bool

	FreqMultiplier   uint64
	ImplementsCFList bool
	CFListType       ttnpb.CFListType

	ReceiveDelay1 time.Duration
	ReceiveDelay2 time.Duration

	JoinAcceptDelay1 time.Duration
	JoinAcceptDelay2 time.Duration
	MaxFCntGap       uint

	SupportsDynamicADR   bool
	ADRAckLimit          ttnpb.ADRAckLimitExponent
	ADRAckDelay          ttnpb.ADRAckDelayExponent
	MinRetransmitTimeout time.Duration
	MaxRetransmitTimeout time.Duration

	TxOffset            []float32
	MaxADRDataRateIndex ttnpb.DataRateIndex

	TxParamSetupReqSupport bool

	DefaultMaxEIRP float32

	Rx1Channel  map[uint8]uint8
	Rx1DataRate map[string]ttnpb.DataRateIndex

	DefaultRx2Parameters band.Rx2Parameters

	BootDwellTime band.DwellTime
}

func makeRx1Channel(f func(uint8) (uint8, error)) map[uint8]uint8 {
	m := make(map[uint8]uint8)
	for i := 0; i <= 255; i++ {
		idx := uint8(i)

		ch, err := f(idx)
		if err != nil {
			continue
		}
		m[idx] = ch
	}
	return m
}

func makeRx1DataRate(f func(ttnpb.DataRateIndex, ttnpb.DataRateOffset, bool) (ttnpb.DataRateIndex, error)) map[string]ttnpb.DataRateIndex {
	m := make(map[string]ttnpb.DataRateIndex)
	for _, drIdxInt32 := range ttnpb.DataRateIndex_value {
		for _, drOffInt32 := range ttnpb.DataRateOffset_value {
			for _, dwellTime := range []bool{false, true} {
				name := fmt.Sprintf("%v_%v_%v", drIdxInt32, drOffInt32, dwellTime)

				dr, err := f(ttnpb.DataRateIndex(drIdxInt32), ttnpb.DataRateOffset(drOffInt32), dwellTime)
				if err != nil {
					continue
				}

				m[name] = dr
			}
		}
	}
	return m
}

func makeBand(b band.Band) serializableBand {
	return serializableBand{
		ID: b.ID,

		Beacon: serializableBeacon{
			DataRateIndex: b.Beacon.DataRateIndex,
			CodingRate:    b.Beacon.CodingRate,
			Frequencies:   b.Beacon.Frequencies,
		},
		PingSlotFrequencies: b.PingSlotFrequencies,

		MaxUplinkChannels: b.MaxUplinkChannels,
		UplinkChannels:    b.UplinkChannels,

		MaxDownlinkChannels: b.MaxDownlinkChannels,
		DownlinkChannels:    b.DownlinkChannels,

		SubBands: b.SubBands,

		DataRates:        makeDataRates(b.DataRates),
		StrictCodingRate: b.StrictCodingRate,

		FreqMultiplier:   b.FreqMultiplier,
		ImplementsCFList: b.ImplementsCFList,
		CFListType:       b.CFListType,

		ReceiveDelay1: b.ReceiveDelay1,
		ReceiveDelay2: b.ReceiveDelay2,

		JoinAcceptDelay1: b.JoinAcceptDelay1,
		JoinAcceptDelay2: b.JoinAcceptDelay2,
		MaxFCntGap:       b.MaxFCntGap,

		SupportsDynamicADR:   b.SupportsDynamicADR,
		ADRAckLimit:          b.ADRAckLimit,
		ADRAckDelay:          b.ADRAckDelay,
		MinRetransmitTimeout: b.MinRetransmitTimeout,
		MaxRetransmitTimeout: b.MaxRetransmitTimeout,

		TxOffset:            b.TxOffset,
		MaxADRDataRateIndex: b.MaxADRDataRateIndex,

		TxParamSetupReqSupport: b.TxParamSetupReqSupport,

		DefaultMaxEIRP: b.DefaultMaxEIRP,

		Rx1Channel:  makeRx1Channel(b.Rx1Channel),
		Rx1DataRate: makeRx1DataRate(b.Rx1DataRate),

		// Missing: GenerateChMasks
		// Missing: ParseChMask

		DefaultRx2Parameters: b.DefaultRx2Parameters,

		BootDwellTime: b.BootDwellTime,
	}
}

func testBand(t *testing.T, band serializableBand, version ttnpb.PHYVersion) {
	reference := path.Join("testdata", fmt.Sprintf("%v_%v.json", band.ID, version))
	if os.Getenv("TEST_WRITE_GOLDEN") == "1" {
		b, err := json.MarshalIndent(band, "", "  ")
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(reference, b, 0o600); err != nil {
			t.Fatal(err)
		}
	} else {
		b, err := os.ReadFile(reference)
		if err != nil {
			t.Fatal(err)
		}

		referenceBand := serializableBand{}
		if err := json.Unmarshal(b, &referenceBand); err != nil {
			t.Fatal(err)
		}

		assertions.New(t).So(band, should.Resemble, referenceBand)
	}
}

func TestBandDefinitions(t *testing.T) {
	t.Parallel()
	for name, versions := range band.All {
		for version, band := range versions {
			version, band := version, band
			t.Run(fmt.Sprintf("%v/%v", name, version), func(t *testing.T) {
				t.Parallel()
				testBand(t, makeBand(band), version)
			})
		}
	}
}

func TestLatest(t *testing.T) {
	for name := range band.All {
		t.Run(name, func(t *testing.T) {
			if _, ok := band.LatestVersion[name]; !ok {
				t.Fatal("Band not found in the latest version map")
			}
		})
	}

	for name := range band.LatestVersion {
		t.Run(name, func(t *testing.T) {
			if _, ok := band.All[name]; !ok {
				t.Fatal("Band not found in the all map")
			}
		})
	}
}
