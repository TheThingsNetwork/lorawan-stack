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

import { isUndefined } from 'lodash'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import Yup from '@ttn-lw/lib/yup'

// Validation schemas of the device type manual form section.
// Please observe the following rules to keep the validation schemas maintainable:
// 1. DO NOT USE ANY TYPE CONVERSIONS HERE. Use decocer/encoder on field level instead.
//    Consider all values as backend values. Exceptions may apply in consideration.
// 2. Comment each individual validation prop and use whitespace to structure visually.
// 3. Do not use ternary assignments but use plain if statements to ensure clarity.

const factoryPresetFreqRequiredTest = frequencies =>
  frequencies.every(freq => typeof freq !== 'undefined' && freq !== '')

const factoryPresetFreqNumericTest = frequencies =>
  frequencies.every(freq => {
    if (typeof freq !== 'undefined') {
      return !isNaN(parseInt(freq))
    }

    return true
  })

const advancedSettingsSchema = Yup.object({
  supports_class_b: Yup.boolean().required(sharedMessages.validateRequired),
  supports_class_c: Yup.boolean().required(sharedMessages.validateRequired),
  supports_join: Yup.boolean().when('$mayEditKeys', {
    is: false,
    then: schema => schema.oneOf([true]),
  }),
  multicast: Yup.boolean().when('$mayEditKeys', {
    is: false,
    then: schema => schema.oneOf([false]),
  }),
  _default_ns_settings: Yup.bool(),
  _skip_js_registration: Yup.bool(),
})

const macSettingsSchema = Yup.object({
  mac_settings: Yup.lazy(macSettings =>
    Yup.object().when(
      ['multicast', 'supports_join', 'supports_class_b', '$defaultMacSettings'],
      (multicast, supports_join, supports_class_b, defaultMacSettings, schema) => {
        if (!defaultMacSettings || !macSettings) {
          return schema
        }

        const shape = {
          resets_f_cnt: Yup.boolean(),
          rx1_data_rate_offset: Yup.number()
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(7, Yup.passValues(sharedMessages.validateNumberLte)),
          rx1_delay: Yup.number()
            .min(1, Yup.passValues(sharedMessages.validateNumberGte))
            .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
          factory_preset_frequencies: Yup.array()
            .default([])
            .test(
              'is-valid-frequency',
              sharedMessages.validateFreqNumeric,
              factoryPresetFreqNumericTest,
            )
            .test(
              'is-empty-frequency',
              sharedMessages.validateFreqRequired,
              factoryPresetFreqRequiredTest,
            ),
          rx2_frequency: Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte)),
          beacon_frequency: Yup.number().min(
            100000,
            Yup.passValues(sharedMessages.validateNumberGte),
          ),
          ping_slot_frequency: Yup.number().min(
            100000,
            Yup.passValues(sharedMessages.validateNumberGte),
          ),
          rx2_data_rate_index: Yup.number()
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
          ping_slot_data_rate_index: Yup.number()
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
          ping_slot_periodicity: Yup.lazy(() => {
            // Ping slot periodicity does not have a default value, so it has to be
            // set explicitly when not using OTAA.
            if ((multicast || !supports_join) && supports_class_b) {
              return Yup.string().default(null).typeError(sharedMessages.validateRequired)
            }

            return Yup.string().default(null).nullable()
          }),
          class_b_timeout: Yup.string(),
          class_c_timeout: Yup.string(),
        }

        // Each MAC setting that does have a corresponding default setting is required.
        // We can use the default MAC settings object to cycle through the list of
        // values and make them required.
        for (const key of Object.keys(shape)) {
          if (!(key in macSettings)) {
            delete shape[key]
          } else if (!isUndefined(defaultMacSettings[key]) && shape[key].type !== 'lazy') {
            // Compose a new schema that makes the field mandatory unless it was
            // already stripped. Due to the way Yup works, we need to convert empty strings
            // to `null` as they would otherwise be converted to `undefined` and automatically
            // stripped by formik.
            // See https://github.com/jaredpalmer/formik/issues/805
            const oldSchema = shape[key].clone()
            shape[key] = oldSchema
              .default(null)
              .typeError(sharedMessages.validateRequired)
              .required(sharedMessages.validateRequired)
          }
        }

        return schema.shape(shape)
      },
    ),
  ),
})

const validationSchema = Yup.object({
  lorawan_version: Yup.string().required(sharedMessages.validateRequired),
  lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
})
  .concat(advancedSettingsSchema)
  .concat(macSettingsSchema)

export default validationSchema
