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

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { ACTIVATION_MODES, parseLorawanMacVersion, DEVICE_CLASSES } from '@console/lib/device-utils'

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

const validationSchema = Yup.object({
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  lorawan_version: Yup.string().required(sharedMessages.validateRequired),
  lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
  supports_class_b: Yup.boolean().default(false),
  supports_class_c: Yup.boolean().default(false),
  supports_join: Yup.boolean().default(false),
  multicast: Yup.boolean().default(false),
  mac_settings: Yup.object({
    rx1_delay: Yup.lazy(delay => {
      if (!Boolean(delay) || typeof delay.value === 'undefined') {
        return Yup.object().strip()
      }

      return Yup.object().when('$activationMode', {
        is: ACTIVATION_MODES.ABP,
        then: schema =>
          schema.shape({
            value: Yup.number()
              .min(1, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
          }),
        otherwise: schema => schema.strip(),
      })
    }),
    rx1_data_rate_offset: Yup.number().when('$activationMode', {
      is: ACTIVATION_MODES.ABP,
      then: schema =>
        schema
          .min(0, Yup.passValues(sharedMessages.validateNumberGte))
          .max(7, Yup.passValues(sharedMessages.validateNumberLte)),
      otherwise: schema => schema.strip(),
    }),
    resets_f_cnt: Yup.boolean().when('$activationMode', {
      is: ACTIVATION_MODES.ABP,
      then: schema => schema.default(false),
      otherwise: schema => schema.strip(),
    }),
    rx2_data_rate_index: Yup.lazy(dataRate => {
      if (!Boolean(dataRate) || typeof dataRate.value === 'undefined') {
        return Yup.object().strip()
      }

      return Yup.object({
        value: Yup.number()
          .min(0, Yup.passValues(sharedMessages.validateNumberGte))
          .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
      })
    }),
    rx2_frequency: Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte)),
    ping_slot_periodicity: Yup.lazy(periodicity => {
      if (!Boolean(periodicity) || typeof periodicity.value === 'undefined') {
        return Yup.object().strip()
      }

      return Yup.object().when('$isClassB', {
        is: true,
        then: schema =>
          schema
            .shape({
              value: Yup.string(),
            })
            .required(sharedMessages.validateRequired),
        otherwise: schema => schema.strip(),
      })
    }),
    ping_slot_frequency: Yup.number().when('$isClassB', {
      is: true,
      then: schema => schema.min(100000, Yup.passValues(sharedMessages.validateNumberGte)),
      otherwise: schema => schema.strip(),
    }),
    factory_preset_frequencies: Yup.lazy(frequencies => {
      if (!Boolean(frequencies)) {
        return Yup.array().strip()
      }

      return Yup.array()
        .default([])
        .test(
          'is-valid-frequency',
          sharedMessages.validateFreqNumberic,
          factoryPresetFreqNumericTest,
        )
        .test(
          'is-empty-frequency',
          sharedMessages.validateFreqRequired,
          factoryPresetFreqRequiredTest,
        )
    }),
    supports_32_bit_f_cnt: Yup.boolean().default(true),
  }),
  session: Yup.object().when(
    ['lorawan_version', '$activationMode'],
    (version, activationMode, schema) => {
      if (activationMode === ACTIVATION_MODES.OTAA || activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      const lwVersion = parseLorawanMacVersion(version)

      return schema.shape({
        dev_addr: Yup.string()
          .length(4 * 2, Yup.passValues(sharedMessages.validateLength)) // 4 Byte hex.
          .required(sharedMessages.validateRequired),
        keys: Yup.object().shape({
          f_nwk_s_int_key: Yup.object({
            key: Yup.string()
              .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
              .required(sharedMessages.validateRequired),
          }),
          s_nwk_s_int_key: Yup.lazy(() =>
            lwVersion >= 110
              ? Yup.object().shape({
                  key: Yup.string()
                    .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                    .required(sharedMessages.validateRequired),
                })
              : Yup.object().strip(),
          ),
          nwk_s_enc_key: Yup.lazy(() =>
            lwVersion >= 110
              ? Yup.object().shape({
                  key: Yup.string()
                    .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                    .required(sharedMessages.validateRequired),
                })
              : Yup.object().strip(),
          ),
        }),
      })
    },
  ),
  _device_class: Yup.mixed().when(['$activationMode'], (mode, schema) => {
    const isMulticast = mode === ACTIVATION_MODES.MULTICAST

    if (isMulticast) {
      return schema
        .oneOf([DEVICE_CLASSES.CLASS_B, DEVICE_CLASSES.CLASS_C])
        .default(DEVICE_CLASSES.CLASS_B)
        .required(sharedMessages.validateRequired)
    }

    return schema
      .oneOf(Object.values(DEVICE_CLASSES))
      .default(DEVICE_CLASSES.CLASS_A)
      .required(sharedMessages.validateRequired)
  }),
}).noUnknown()

export default validationSchema
