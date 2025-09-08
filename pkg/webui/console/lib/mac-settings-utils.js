// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

// 0...7
export const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={sharedMessages.secondInterval} values={{ count: value }} />,
  }
})
// 0...15
export const adrAckLimitOptions = Array.from({ length: 16 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `ADR_ACK_LIMIT_${value}`,
    label: <Message content={sharedMessages.adrAckValue} values={{ count: value }} />,
  }
})
// 0...15
export const adrAckDelayOptions = Array.from({ length: 16 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `ADR_ACK_DELAY_${value}`,
    label: <Message content={sharedMessages.adrAckValue} values={{ count: value }} />,
  }
})
export const maxDutyCycleOptions = [
  { value: 'DUTY_CYCLE_1', label: '100%' },
  { value: 'DUTY_CYCLE_16', label: '6.25%' },
  { value: 'DUTY_CYCLE_128', label: '0.781%' },
  { value: 'DUTY_CYCLE_1024', label: '0.098%' },
  { value: 'DUTY_CYCLE_16384', label: '0.006%' },
]
