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

import * as Yup from 'yup'
import { defineMessages } from 'react-intl'

import sharedMessages from '../../../lib/shared-messages'
import { ACTIVATION_MODES } from '../../lib/device-utils'

const m = defineMessages({
  validateFreqNumberic: 'All frequency values must be positive integers',
  validateFreqRequired: 'All frequency values are required. Please remove empty entries.',
})

const factoryPresetFreqNumericTest = frequencies => {
  return frequencies.every(freq => {
    if (typeof freq !== 'undefined') {
      return !isNaN(parseInt(freq))
    }

    return true
  })
}

const factoryPresetFreqRequiredTest = frequencies => {
  return frequencies.every(freq => typeof freq !== 'undefined' && freq !== '')
}

export default Yup.object({
  mac_settings: Yup.object({
    rx2_data_rate_index: Yup.lazy(dataRate => {
      if (!Boolean(dataRate) || typeof dataRate.value === 'undefined') {
        return Yup.object().strip()
      }

      return Yup.object({
        value: Yup.number(),
      })
    }),
    rx2_frequency: Yup.number().min(100000, {
      ...sharedMessages.validateNumberGte,
      values: { value: 100000 },
    }),
    resets_f_cnt: Yup.boolean().when('$activation_mode', {
      is: mode => mode === ACTIVATION_MODES.ABP,
      then: schema => schema.default(false),
      otherwise: schema => schema.strip(),
    }),
    rx1_delay: Yup.lazy(delay => {
      if (!Boolean(delay) || typeof delay.value === 'undefined') {
        return Yup.object().strip()
      }

      return Yup.object().when('$activation_mode', {
        is: mode => mode === ACTIVATION_MODES.ABP,
        then: schema =>
          schema.shape({
            value: Yup.number(),
          }),
        otherwise: schema => schema.strip(),
      })
    }),
    rx1_data_rate_offset: Yup.number().when('$activation_mode', {
      is: mode => mode === ACTIVATION_MODES.ABP || mode === ACTIVATION_MODES.OTAA,
      then: schema =>
        schema.min(0, { ...sharedMessages.validateNumberGte, values: { value: 0 } }).max(7, {
          ...sharedMessages.validateNumberLte,
          values: { value: 7 },
        }),
      otherwise: schema => schema.strip(),
    }),
    ping_slot_periodicity: Yup.lazy(periodicity => {
      if (!Boolean(periodicity) || typeof periodicity.value === 'undefined') {
        return Yup.object().strip()
      }

      return Yup.object().when('$class_b', {
        is: true,
        then: schema =>
          schema.shape({
            value: Yup.string(),
          }),
        otherwise: schema => schema.strip(),
      })
    }),
    ping_slot_frequency: Yup.number().when('$class_b', {
      is: true,
      then: schema =>
        schema.min(100000, {
          ...sharedMessages.validateNumberGte,
          values: { value: 100000 },
        }),
      otherwise: schema => schema.strip(),
    }),
    factory_preset_frequencies: Yup.lazy(frequencies => {
      if (!Boolean(frequencies)) {
        return Yup.array().strip()
      }

      return Yup.array()
        .default([])
        .test('is-valid-frequency', m.validateFreqNumberic, factoryPresetFreqNumericTest)
        .test('is-empty-frequency', m.validateFreqRequired, factoryPresetFreqRequiredTest)
        .when('$is_init', (isInit, schema) => {
          if (isInit) {
            return schema.transform(arr => arr.map((value, key) => ({ key, value })))
          }
          return schema.transform(arr => arr.map(({ value }) => value))
        })
    }),
  }),
}).noUnknown()
