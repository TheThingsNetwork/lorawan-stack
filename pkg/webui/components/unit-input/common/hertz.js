// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'

import FactoredUnitInput from '../factored'

const units = [
  { label: 'Hz', value: 'hz', factor: 1 },
  { label: 'kHz', value: 'khz', factor: 1000 },
  { label: 'MHz', value: 'mhz', factor: 1000000 },
]

const HertzInput = props => (
  <FactoredUnitInput
    inputWidth="xs"
    units={units}
    baseUnit={units[0].value}
    defaultUnit="mhz"
    {...props}
  />
)

export default HertzInput
