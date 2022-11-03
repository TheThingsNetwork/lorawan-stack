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

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { unit as unitRegexp } from '@console/lib/regexp'

import EncodedUnitInput from '../encoded'

const units = [
  { label: sharedMessages.milliseconds, value: 'ms' },
  { label: sharedMessages.seconds, value: 's' },
  { label: sharedMessages.minutes, value: 'm' },
  { label: sharedMessages.hours, value: 'h' },
]

const encoder = (value, unit) => {
  if (value === null) {
    return null
  }

  return value ? `${value}${unit}` : unit
}
const decoder = (rawValue = '') => {
  if (rawValue === null) {
    return { value: '', unit: null }
  }
  const value = rawValue.split(unitRegexp)[0]
  const unit = rawValue.split(value)[1] || null
  return {
    value: value ? Number(value) : undefined,
    unit,
  }
}

const DurationInput = props => (
  <EncodedUnitInput
    inputWidth="xxs"
    selectWidth="s"
    units={units}
    encode={encoder}
    decode={decoder}
    {...props}
  />
)

export default DurationInput
