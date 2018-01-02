// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package timeutil provides utilities to work with time.
package timeutil

import (
	"time"
)

// 1980-01-06T00:00:00+00:00
const gpsEpochSec = 315964800

// Leap seconds in GPS time
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

// IsGPSLeap reports whether the given GPS time, sec seconds since January 6, 1980 UTC, is a leap second in UTC.
func IsGPSLeap(sec int64) bool {
	i := int64(len(leaps)) - 1
	for ; i >= 0; i-- {
		if sec > leaps[i] {
			return false
		}
		if sec == leaps[i] {
			return true
		}
	}
	return false
}

// GPS returns the local Time corresponding to the given GPS time, sec seconds since January 6, 1980 UTC.
func GPS(sec int64) time.Time {
	i := int64(len(leaps))
	for ; i > 0; i-- {
		if sec > leaps[i-1] {
			break
		}
	}
	return time.Unix(sec+gpsEpochSec-i, 0)
}

// TimeToGPS returns t as a GPS time, the number of seconds elapsed since January 6, 1980 UTC.
func TimeToGPS(t time.Time) int64 {
	sec := t.Unix() - gpsEpochSec

	i := int64(len(leaps))
	for ; i > 0; i-- {
		if sec > leaps[i-1]-i {
			break
		}
	}
	return sec + i
}
