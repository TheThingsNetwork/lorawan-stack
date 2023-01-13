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

package band

var (
	ParseChMask   = parseChMask
	ParseChMask16 = parseChMask16
	ParseChMask72 = parseChMask72
	ParseChMask96 = parseChMask96

	GenerateChMask16     = generateChMask16
	GenerateChMask96     = generateChMask96
	MakeGenerateChMask72 = makeGenerateChMask72

	ErrUnsupportedChMaskCntl = errUnsupportedChMaskCntl

	CompareDataRates = compareDataRates

	BoolPtr = boolPtr
)
