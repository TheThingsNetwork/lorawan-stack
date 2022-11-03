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

/* eslint-disable import/prefer-default-export */

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { unit as unitRegexp } from '@console/lib/regexp'

export const units = [
  { label: sharedMessages.milliseconds, value: 'ms' },
  { label: sharedMessages.seconds, value: 's' },
  { label: sharedMessages.minutes, value: 'm' },
  { label: sharedMessages.hours, value: 'h' },
]

export const durationDecoder = duration => {
  const value = duration.split(unitRegexp)[0]
  const unit = duration.split(value)[1]
  const textUnit = units.find(({ value }) => value === unit).label

  return `${value} ${textUnit}`
}
