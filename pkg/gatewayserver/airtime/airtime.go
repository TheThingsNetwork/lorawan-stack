
package airtime

import (
	"errors"
	"math"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// ComputeLoRa computes the time-on-air given a PHY payload size in bytes, a datr
// identifier and LoRa coding rate identifier. Note that this function operates
// on the PHY payload size and does not add the LoRaWAN header.
//
// See http://www.semtech.com/images/datasheet/LoraDesignGuide_STD.pdf, page 7
func ComputeLoRa(payloadSize uint, datr string, codr string) (time.Duration, error) {
	// Determine CR
	var cr float64
	switch codr {
	case "4/5":
		cr = 1
	case "4/6":
		cr = 2
	case "4/7":
		cr = 3
	case "4/8":
		cr = 4
	default:
		return 0, errors.New("Invalid Codr")
	}
	// Determine DR
	var datarate types.DataRate
	err := 
	dr, err := types.ParseDataRate(datr)
	if err != nil {
		return 0, err
	}
	// Determine DE
	var de float64
	if dr.Bandwidth == 125 && (dr.SpreadingFactor == 11 || dr.SpreadingFactor == 12) {
		de = 1.0
	}
	pl := float64(payloadSize)
	sf := float64(dr.SpreadingFactor)
	bw := float64(dr.Bandwidth)
	h := 0.0 // 0 means header is enabled

	tSym := math.Pow(2, float64(dr.SpreadingFactor)) / bw

	payloadNb := 8.0 + math.Max(0.0, math.Ceil((8.0*pl-4.0*sf+28.0+16.0-20.0*h)/(4.0*(sf-2.0*de)))*(cr+4.0))
	timeOnAir := (payloadNb + 12.25) * tSym * 1000000 // in nanoseconds

	return time.Duration(timeOnAir), nil
}

// ComputeFSK computes the time-on-air given a PHY payload size in bytes and a
// bitrate, Note that this function operates on the PHY payload size and does
// not add the LoRaWAN header.
func ComputeFSK(payloadSize uint, bitrate int) (time.Duration, error) {
	tPkt := int64(time.Second) * (int64(payloadSize) + 5 + 3 + 1 + 2) * 8 / int64(bitrate)
	return time.Duration(tPkt), nil
}
