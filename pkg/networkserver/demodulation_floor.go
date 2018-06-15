// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

// TODO: The values for BW250 and BW500 need to be verified
// (https://github.com/TheThingsIndustries/ttn/issues/876)

var demodulationFloor = map[uint32]map[uint32]float32{
	6: {
		125: -5,
		250: -2,
		500: 1,
	},
	7: {
		125: -7.5,
		250: -4.5,
		500: -1.5,
	},
	8: {
		125: -10,
		250: -7,
		500: -4,
	},
	9: {
		125: -12.5,
		250: -9.5,
		500: -6.5,
	},
	10: {
		125: -15,
		250: -12,
		500: -9,
	},
	11: {
		125: -17.5,
		250: -14.5,
		500: -11.5,
	},
	12: {
		125: -20,
		250: -17,
		500: -24,
	},
}
