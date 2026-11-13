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

package semtechws

import (
	"math"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
)

const (
	// XTime48BitLSBMask is the mask used to extract the lower 48 bits of the XTime.
	XTime48BitLSBMask = 0xFFFFFFFFFFFF
	// XTime32BitLSBMask is the mask used to extract the lower 32 bits of the XTime.
	XTime32BitLSBMask = 0xFFFFFFFF

	// The session ID consists of the radio unit (bits 8-14) and the session (bits 0-7).
	// The xTimeSessionMask is the mask used to extract the session from the session ID.
	xTimeSessionMask = 0xFF
)

// TimeFromUnixSeconds constructs a time.Time from the provided UNIX fractional timestamp.
func TimeFromUnixSeconds(tf float64) time.Time {
	sec, nsec := math.Modf(tf)
	return time.Unix(int64(sec), int64(nsec*1e9)).UTC()
}

// TimePtrFromUnixSeconds constructs a *time.Time from the provided UNIX fractional timestamp.
// If the timestamp is 0, this function returns nil.
func TimePtrFromUnixSeconds(tf float64) *time.Time {
	if tf == 0.0 {
		return nil
	}
	tm := TimeFromUnixSeconds(tf)
	return &tm
}

// TimeToUnixSeconds constructs a UNIX fractional timestamp from the provided time.Time.
func TimeToUnixSeconds(t time.Time) float64 {
	return float64(t.UnixNano()) / float64(1e9)
}

// TimeFromGPSTime constructs a time.Time from the provided GPS time in microseconds.
func TimeFromGPSTime(t int64) time.Time {
	return gpstime.Parse(time.Duration(t) * time.Microsecond)
}

// TimePtrFromGPSTime constructs a *time.Time from the provided GPS time in microseconds.
// If the timestamp is 0, this function returns nil.
func TimePtrFromGPSTime(t int64) *time.Time {
	if t == 0 {
		return nil
	}
	tm := TimeFromGPSTime(t)
	return &tm
}

// TimeToGPSTime contructs a GPS timestamp from the provided time.Time.
func TimeToGPSTime(t time.Time) int64 {
	return int64(gpstime.ToGPS(t) / time.Microsecond)
}

// TimePtrFromUpInfo contructs a *time.Time from the provided uplink metadata information.
// The GPS timestamp (microseconds) is used if present, then the RxTime. The function
// returns nil if both timestamps are unavailable.
func TimePtrFromUpInfo(gpsTime int64, rxTime float64) *time.Time {
	switch {
	case gpsTime != 0:
		return TimePtrFromGPSTime(gpsTime)
	case rxTime != 0.0:
		return TimePtrFromUnixSeconds(rxTime)
	default:
		return nil
	}
}

// TimePtrToUpInfoTime constructs the RxTime and GPSTime from the provided *time.Time.
// If the time is nil, this function returns (0.0, 0).
func TimePtrToUpInfoTime(t *time.Time) (float64, int64) {
	if t == nil {
		return 0.0, 0
	}
	return TimeToUnixSeconds(*t), TimeToGPSTime(*t)
}

// ConcentratorTimeToXTime contructs the XTime associated with the provided
// session ID and concentrator timestamp.
// Bytes 0-5 (6 bytes = 48 bits) are used for the timestamp.
// Bytes 6-8 are returned unmodified to the gateway from the xtime read on the latest uplink.
func ConcentratorTimeToXTime(id int32, t scheduling.ConcentratorTime) int64 {
	return int64(id)<<48 | (int64(t) / int64(time.Microsecond) & XTime48BitLSBMask)
}

// ConcentratorTimeToTimestampXTime constructs the XTime associated with the provided
// concentrator time for gateways that do not extend the concentrator timestamp.
// The XTime is the 32-bit unsigned concentrator timestamp, which rolls over in the same way
// as the concentrator timestamp of the gateway.
// See IsExtendedXTimeSession.
func ConcentratorTimeToTimestampXTime(t scheduling.ConcentratorTime) int64 {
	return int64(t) / int64(time.Microsecond) & XTime32BitLSBMask
}

// ConcentratorTimeFromXTime constructs the scheduling.ConcentratorTime associated
// with the provided XTime.
func ConcentratorTimeFromXTime(xTime int64) scheduling.ConcentratorTime {
	// The Basic Station epoch is the 48 LSB.
	return scheduling.ConcentratorTime(time.Duration(xTime&XTime48BitLSBMask) * time.Microsecond)
}

// TimestampFromXTime constructs the concentrator timestamp associated with the
// provided XTime.
func TimestampFromXTime(xTime int64) uint32 {
	// The concentrator timestamp is the 32 LSB.
	return uint32(xTime & XTime32BitLSBMask)
}

// SessionIDFromXTime constructs the session ID associated with the provided XTime.
func SessionIDFromXTime(xTime int64) int32 {
	return int32(xTime >> 48)
}

// IsExtendedXTimeSession returns whether the provided session ID belongs to a gateway
// that extends the 32-bit concentrator timestamp to 48 bits in the XTime.
// LoRa Basics Station never uses session 0 and accounts for rollovers in the lower 48 bits.
// Some implementations report the raw 32-bit concentrator timestamp as
// XTime, without a session and without accounting for rollovers.
func IsExtendedXTimeSession(id int32) bool {
	return id&xTimeSessionMask != 0
}
