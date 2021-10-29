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

package lbslns

import (
	"math"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
)

// TimeFromUnixSeconds constructs a time.Time from the provided UNIX fractional timestamp.
func TimeFromUnixSeconds(tf float64) time.Time {
	sec, nsec := math.Modf(tf)
	return time.Unix(int64(sec), int64(nsec*1e9))
}

// TimeToUnixSeconds constructs a UNIX fractional timestamp from the provided time.Time.
func TimeToUnixSeconds(t time.Time) float64 {
	return float64(t.UnixNano()) / float64(1e9)
}

// TimeToGPSTime contructs a GPS timestamp from the provided time.Time.
func TimeToGPSTime(t time.Time) int64 {
	return int64(gpstime.ToGPS(t) / time.Microsecond)
}

// ConcentratorTimeToXTime contructs the XTime associated with the provided
// session ID and concentrator timestamp.
func ConcentratorTimeToXTime(id int32, t scheduling.ConcentratorTime) int64 {
	return int64(id)<<48 | (int64(t) / int64(time.Microsecond) & 0xFFFFFFFFFF)
}
