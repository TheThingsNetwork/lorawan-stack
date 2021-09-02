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

import "time"

// SubBandParameters contains the sub-band frequency range, duty cycle and Tx power.
type SubBandParameters struct {
	MinFrequency uint64
	MaxFrequency uint64
	DutyCycle    float32
	MaxEIRP      float32
}

// Comprises returns whether the duty cycle applies to the given frequency.
func (d SubBandParameters) Comprises(frequency uint64) bool {
	return frequency >= d.MinFrequency && frequency <= d.MaxFrequency
}

// MaxEmissionDuring the period passed as parameter, that is allowed by that duty cycle.
func (d SubBandParameters) MaxEmissionDuring(period time.Duration) time.Duration {
	return time.Duration(d.DutyCycle * float32(period))
}
