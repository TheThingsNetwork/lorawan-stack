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

package scheduling

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
)

// ConcentratorTime is the time relative to the concentrator start time (nanoseconds).
type ConcentratorTime int64

// NewEmission returns a new Emission with the given values.
func NewEmission(starts ConcentratorTime, duration time.Duration) Emission {
	return Emission{t: starts, d: duration}
}

// Emission contains the scheduled time and duration of an emission.
type Emission struct {
	t ConcentratorTime
	d time.Duration
}

// Starts returns the time when the emission starts.
func (em Emission) Starts() ConcentratorTime { return em.t }

// Ends returns the time when the emission ends.
func (em Emission) Ends() ConcentratorTime { return em.t + ConcentratorTime(em.d) }

// Duration returns the duration of the emission.
func (em Emission) Duration() time.Duration { return em.d }

// OffAir returns the time-off-air of the emission.
func (em Emission) OffAir(toa frequencyplans.TimeOffAir) time.Duration {
	d := time.Duration(float32(em.d) * toa.Fraction)
	if d < toa.Duration {
		d = toa.Duration
	}
	return d
}

// Within returns the duration of the emission that happens within the given window.
func (em Emission) Within(from, to ConcentratorTime) time.Duration {
	if em.Ends() < from || em.t > to {
		return 0
	}
	if em.t < from {
		return time.Duration(em.Ends() - from)
	}
	return em.d
}

// EndsWithOffAir returns the time when the emission ends plus the time-off-air.
func (em Emission) EndsWithOffAir(toa frequencyplans.TimeOffAir) ConcentratorTime {
	return em.Ends() + ConcentratorTime(em.OffAir(toa))
}

// BeforeWithOffAir returns the time between the end of this emission to the start of the given other emission, considering time-off-air.
func (em Emission) BeforeWithOffAir(other Emission, toa frequencyplans.TimeOffAir) time.Duration {
	return time.Duration(other.Starts() - em.EndsWithOffAir(toa))
}

// AfterWithOffAir returns the time between the end of the given other emission to the start of this emission, considering time-off-air.
func (em Emission) AfterWithOffAir(other Emission, toa frequencyplans.TimeOffAir) time.Duration {
	return time.Duration(em.Starts() - other.EndsWithOffAir(toa))
}

// Emissions is an list of emissions.
type Emissions []Emission

// Insert inserts the given emission to the emissions by preserving order.
func (ems Emissions) Insert(em Emission) Emissions {
	for i := range ems {
		if ems[i].t > em.t {
			return append(ems[:i], append([]Emission{em}, ems[i:]...)...)
		}
	}
	return append(ems, em)
}
