// Copyright © 2019 The Things Industries B.V.

package toa

import (
	"testing"
	"time"

	assertions "github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestInvalidLoRa2400(t *testing.T) {
	a := assertions.New(t)

	// Invalid coding rate.
	{
		downlink, err := buildLoRaDownlinkFromParameters(10, 2422000000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 10,
					Bandwidth:       812000,
				},
			},
		}, "1/9")
		_, err = Compute(len(downlink.RawPayload), *downlink.GetScheduled())
		a.So(err, should.NotBeNil)
	}

	// Invalid spreading factor.
	{
		downlink, err := buildLoRaDownlinkFromParameters(10, 2422000000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 0,
					Bandwidth:       812000,
				},
			},
		}, "4/5")
		_, err = Compute(len(downlink.RawPayload), *downlink.GetScheduled())
		a.So(err, should.NotBeNil)
	}

	// Invalid bandwidth.
	{
		downlink, err := buildLoRaDownlinkFromParameters(10, 2422000000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 7,
					Bandwidth:       0,
				},
			},
		}, "4/5")
		_, err = Compute(len(downlink.RawPayload), *downlink.GetScheduled())
		a.So(err, should.NotBeNil)
	}
}

func TestDifferentLoRa2400SFs(t *testing.T) {
	a := assertions.New(t)
	sfTests := map[ttnpb.DataRate]time.Duration{
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 5, Bandwidth: 812000}}}:  1665000,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 6, Bandwidth: 812000}}}:  3093600,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 812000}}}:  5556700,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 8, Bandwidth: 812000}}}:  10482800,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 9, Bandwidth: 812000}}}:  19073900,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 10, Bandwidth: 812000}}}: 36886700,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 11, Bandwidth: 812000}}}: 73773400,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 12, Bandwidth: 812000}}}: 142502500,
	}
	for dr, ns := range sfTests {
		dl, err := buildLoRaDownlinkFromParameters(10, 2422000000, dr, "4/5")
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(ns), 50)
	}
}

func TestDifferentLoRa2400BWs(t *testing.T) {
	a := assertions.New(t)
	bwTests := map[ttnpb.DataRate]time.Duration{
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 203000}}}:  22226600,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 406000}}}:  11113300,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 812000}}}:  5556700,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 1625000}}}: 2776600,
	}
	for dr, ns := range bwTests {
		dl, err := buildLoRaDownlinkFromParameters(10, 2422000000, dr, "4/5")
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, ns, 50)
	}
}

func TestDifferentLoRa2400CRs(t *testing.T) {
	a := assertions.New(t)
	crTests := map[string]time.Duration{
		"4/5": 5556700,
		"4/6": 6029600,
		"4/8": 6817700,
	}
	for cr, ns := range crTests {
		dl, err := buildLoRaDownlinkFromParameters(10, 2422000000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 7,
					Bandwidth:       812000,
				},
			},
		}, cr)
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, ns, 50)
	}
}

func TestDifferentLoRa2400PayloadSizes(t *testing.T) {
	a := assertions.New(t)
	plTests := map[int]time.Duration{
		1:   102147800,
		10:  142502500,
		20:  192945800,
		50:  344275900,
		100: 596492600,
		230: 1252256200,
	}
	for size, ns := range plTests {
		dl, err := buildLoRaDownlinkFromParameters(size, 2422000000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 12,
					Bandwidth:       812000,
				},
			},
		}, "4/5")
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, ns, 50)
	}
}

func TestDifferentLoRa2400CRCs(t *testing.T) {
	a := assertions.New(t)
	crcTests := map[bool]time.Duration{
		true:  6029600,
		false: 5556700,
	}
	for crc, ns := range crcTests {
		dl, err := buildLoRaDownlinkFromParameters(10, 2422000000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 7,
					Bandwidth:       812000,
				},
			},
		}, "4/5")
		dl.GetScheduled().EnableCRC = crc
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, ns, 50)
	}
}
