// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

// Package gpstime provides utilities to work with GPS time.
package gpstime

import (
	"time"
)

var gpsEpoch = time.Date(1980, time.January, 6, 0, 0, 0, 0, time.UTC)

// Leap seconds represented as seconds since GPS epoch.
var leaps = [...]int64{
	46828800,
	78364801,
	109900802,
	173059203,
	252028804,
	315187205,
	346723206,
	393984007,
	425520008,
	457056009,
	504489610,
	551750411,
	599184012,
	820108813,
	914803214,
	1025136015,
	1119744016,
	1167264017,
	1341118800,
}

func seconds(d time.Duration) int64 {
	return int64(d / time.Second)
}

// IsLeap reports whether the given time.Duration elapsed since GPS epoch (January 6, 1980 UTC) is a leap second in UTC.
func IsLeapSecond(d time.Duration) bool {
	dSec := seconds(d)
	for i := len(leaps) - 1; i >= 0; i-- {
		if dSec > leaps[i] {
			return false
		}
		if dSec == leaps[i] {
			return true
		}
	}
	return false
}

// Parse returns the UTC Time corresponding to the given time.Duration elapsed since GPS epoch (January 6, 1980 UTC).
func Parse(d time.Duration) time.Time {
	dSec := seconds(d)
	i := int64(len(leaps))
	for ; i > 0; i-- {
		if dSec > leaps[i-1] {
			break
		}
	}
	return gpsEpoch.Add(d - time.Duration(i)*time.Second)
}

// ToGPS returns t as time.Duration elapsed since GPS epoch (January 6, 1980 UTC).
func ToGPS(t time.Time) time.Duration {
	d := t.Sub(gpsEpoch)
	dSec := seconds(d)
	i := int64(len(leaps))
	for ; i > 0; i-- {
		if dSec > leaps[i-1]-i {
			break
		}
	}
	return d + time.Duration(i)*time.Second
}
