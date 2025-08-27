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

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const dynamicFrequencyTest = value => value === 0 || value >= 100000 || value === null

export const initialValues = {
  ids: {
    profile_id: '',
  },
  mac_settings: {
    beacon_frequency: null,
    desired_beacon_frequency: null,
    class_b_timeout: null,
    class_c_timeout: null,
    rx1_delay: 1,
    desired_rx1_delay: 1,
    rx1_data_rate_offset: 0,
    desired_rx1_data_rate_offset: 0,
    resets_f_cnt: false,
    ping_slot_data_rate_index: 0,
    desired_ping_slot_data_rate_index: 0,
    rx2_data_rate_index: 0,
    desired_rx2_data_rate_index: 0,
    rx2_frequency: null,
    ping_slot_periodicity: null,
    desired_rx2_frequency: null,
    ping_slot_frequency: null,
    desired_ping_slot_frequency: null,
    factory_preset_frequencies: [],
    supports_32_bit_f_cnt: true,
    max_duty_cycle: null,
    desired_max_duty_cycle: '',
    adr: {
      dynamic: {
        _use_default_nb_trans: true,
        margin: undefined,
        _override_nb_trans_defaults: false,
      },
    },
    desired_adr_ack_limit_exponent: null,
    desired_adr_ack_delay_exponent: null,
    status_time_periodicity: '',
    status_count_periodicity: null,
  },
}

const factoryPresetFreqNumericTest = frequencies =>
  frequencies.every(freq => {
    if (typeof freq !== 'undefined') {
      return !isNaN(parseInt(freq))
    }

    return true
  })

const factoryPresetFreqRequiredTest = frequencies =>
  frequencies.every(freq => typeof freq !== 'undefined' && freq !== '')

export const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    profile_id: Yup.string().required(sharedMessages.validateRequired),
  }),
  mac_settings: Yup.object().shape({
    beacon_frequency: Yup.number()
      .nullable()
      .test('is-valid-freq', sharedMessages.validateFreqDynamic, dynamicFrequencyTest),
    desired_beacon_frequency: Yup.number()
      .nullable()
      .test('is-valid-freq', sharedMessages.validateFreqDynamic, dynamicFrequencyTest),
    class_b_timeout: Yup.lazy(value => {
      if (!Boolean(value)) {
        return Yup.string().strip()
      }

      return Yup.string()
    }),
    class_c_timeout: Yup.lazy(value => {
      if (!Boolean(value)) {
        return Yup.string().strip()
      }

      return Yup.string()
    }),
    rx1_delay: Yup.lazy(delay => {
      if (delay === undefined || delay === '') {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(1, Yup.passValues(sharedMessages.validateNumberGte))
        .max(15, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    desired_rx1_delay: Yup.lazy(delay => {
      if (delay === undefined || delay === '') {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(1, Yup.passValues(sharedMessages.validateNumberGte))
        .max(15, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    rx1_data_rate_offset: Yup.lazy(value => {
      if (value === undefined || value === '') {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(7, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    desired_rx1_data_rate_offset: Yup.lazy(value => {
      if (value === undefined || value === '') {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(7, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    resets_f_cnt: Yup.boolean().default(false),
    ping_slot_data_rate_index: Yup.lazy(dataRate => {
      if (dataRate === '' || dataRate === undefined) {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(15, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    desired_ping_slot_data_rate_index: Yup.lazy(dataRate => {
      if (dataRate === '' || dataRate === undefined) {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(15, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    rx2_data_rate_index: Yup.lazy(dataRate => {
      if (dataRate === '' || dataRate === undefined) {
        return Yup.number().strip()
      }

      return Yup.number()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(15, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    desired_rx2_data_rate_index: Yup.lazy(dataRate => {
      if (dataRate === '' || dataRate === undefined) {
        return Yup.number().strip()
      }

      return Yup.number()
        .nullable()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(15, Yup.passValues(sharedMessages.validateNumberLte))
    }),
    rx2_frequency: Yup.lazy(frequency => {
      if (frequency === undefined || frequency === '') {
        return Yup.number().strip()
      }
      return Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte)).nullable()
    }),
    ping_slot_periodicity: Yup.string().nullable(),
    desired_rx2_frequency: Yup.lazy(frequency => {
      if (frequency === undefined || frequency === '') {
        return Yup.number().strip()
      }
      return Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte)).nullable()
    }),
    ping_slot_frequency: Yup.number()
      .test('is-valid-freq', sharedMessages.validateFreqDynamic, dynamicFrequencyTest)
      .nullable(),
    desired_ping_slot_frequency: Yup.lazy(frequency => {
      if (!Boolean(frequency)) {
        return Yup.number().strip()
      }

      return Yup.number().test(
        'is-valid-freq',
        sharedMessages.validateFreqDynamic,
        dynamicFrequencyTest,
      )
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
    max_duty_cycle: Yup.string().nullable(),
    desired_max_duty_cycle: Yup.lazy(value => {
      if (!Boolean(value)) {
        return Yup.string().strip()
      }

      return Yup.string()
    }),
    adr: Yup.lazy(value => {
      if (value && 'static' in value) {
        return Yup.object().shape({
          static: Yup.object().shape({
            data_rate_index: Yup.number().nullable(),
            tx_power_index: Yup.number()
              .nullable()
              .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
            nb_trans: Yup.number()
              .nullable()
              .min(1, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte)),
          }),
        })
      } else if (value && 'disabled' in value) {
        return Yup.object().shape({
          disabled: Yup.object().shape({}).noUnknown(),
        })
      }

      return Yup.object().shape({
        dynamic: Yup.lazy(value =>
          Yup.object().shape({
            margin: Yup.number().nullable(),
            min_nb_trans: Yup.number()
              .min(1, Yup.passValues(sharedMessages.validateNumberGte))
              .max(value?.max_nb_trans || 3, Yup.passValues(sharedMessages.validateNumberLte))
              .nullable(),
            max_nb_trans: Yup.number()
              .min(value?.min_nb_trans || 1, Yup.passValues(sharedMessages.validateNumberGte))
              .max(3, Yup.passValues(sharedMessages.validateNumberLte))
              .nullable(),
            overrides: Yup.lazy(value => {
              if (!Boolean(value)) {
                return Yup.object().nullable()
              }

              return Yup.object().shape(
                Object.keys(value).reduce((acc, key) => {
                  acc[key] = Yup.object().shape({
                    _data_rate_index: Yup.string().default(key.startsWith('_') ? null : key),
                    min_nb_trans: Yup.lazy(value => {
                      if (value === undefined) {
                        return Yup.number().strip()
                      }

                      return Yup.number()
                        .min(1, Yup.passValues(sharedMessages.validateNumberGte))
                        .max(
                          value[key]?.max_nb_trans || 3,
                          Yup.passValues(sharedMessages.validateNumberLte),
                        )
                    }),
                    max_nb_trans: Yup.lazy(value => {
                      if (value === undefined) {
                        return Yup.number().strip()
                      }

                      return Yup.number()
                        .min(
                          value[key]?.min_nb_trans || 1,
                          Yup.passValues(sharedMessages.validateNumberGte),
                        )
                        .max(3, Yup.passValues(sharedMessages.validateNumberLte))
                    }),
                  })

                  return acc
                }, {}),
              )
            }),
          }),
        ),
      })
    }),
    desired_adr_ack_limit_exponent: Yup.string().when(['adr'], ([adr], schema) => {
      if (!('dynamic' in adr)) {
        return schema.strip()
      }

      return schema.nullable()
    }),
    desired_adr_ack_delay_exponent: Yup.string().when(['adr'], ([adr], schema) => {
      if (!('dynamic' in adr)) {
        return schema.strip()
      }

      return schema.nullable()
    }),
    status_time_periodicity: Yup.lazy(value => {
      if (!Boolean(value)) {
        return Yup.string().strip()
      }

      return Yup.string()
    }),
    status_count_periodicity: Yup.lazy(value => {
      if (value === undefined || value === '') {
        return Yup.number().strip()
      }

      return Yup.number().nullable()
    }),
  }),
})
