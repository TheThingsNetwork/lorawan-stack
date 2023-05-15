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

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  dynamicFrequencyTest,
  parseLorawanMacVersion,
  ACTIVATION_MODES,
  isNonZeroSessionKey,
} from '@console/lib/device-utils'

import messages from '../messages'

const factoryPresetFreqNumericTest = frequencies =>
  frequencies.every(freq => {
    if (typeof freq !== 'undefined') {
      return !isNaN(parseInt(freq))
    }

    return true
  })

const factoryPresetFreqRequiredTest = frequencies =>
  frequencies.every(freq => typeof freq !== 'undefined' && freq !== '')

const validationSchema = Yup.object()
  .shape({
    _activation_mode: Yup.mixed()
      .oneOf([ACTIVATION_MODES.ABP, ACTIVATION_MODES.OTAA, ACTIVATION_MODES.MULTICAST])
      .required(sharedMessages.validateRequired),
    lorawan_version: Yup.string().required(sharedMessages.validateRequired),
    lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
    frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
    supports_class_b: Yup.boolean().when(['_device_classes'], ([deviceClasses = {}], schema) =>
      schema.transform(() => undefined).default(deviceClasses.class_b || false),
    ),
    supports_class_c: Yup.boolean().when(['_device_classes'], ([deviceClasses = {}], schema) =>
      schema.transform(() => undefined).default(deviceClasses.class_c || false),
    ),
    _device_classes: Yup.object({
      class_b: Yup.boolean(),
      class_c: Yup.boolean(),
    }).when(['_activation_mode'], ([mode], schema) => {
      if (mode === ACTIVATION_MODES.MULTICAST) {
        return schema.test(
          'has-class-checked',
          sharedMessages.validateRequired,
          classes =>
            !classes || Object.values(classes).some(supportsClass => Boolean(supportsClass)),
        )
      }

      return schema
    }),
    session: Yup.object().when(
      ['_activation_mode', 'lorawan_version', '$isJoined', '$mayEditKeys', '$mayReadKeys'],
      ([mode, version, isJoined, mayEditKeys, mayReadKeys], schema) => {
        if (mode === ACTIVATION_MODES.ABP || mode === ACTIVATION_MODES.MULTICAST || isJoined) {
          const isNewVersion = parseLorawanMacVersion(version) >= 110
          return schema.shape({
            dev_addr: Yup.lazy(() => {
              const schema = Yup.string().length(
                4 * 2,
                Yup.passValues(sharedMessages.validateLength),
              ) // 4 Byte hex.

              if (mayReadKeys && mayEditKeys) {
                // Force the field to be required only if the user can see and
                // edit the `dev_addr`, otherwise the user is not able to edit
                // any other fields in the NS form without resetting the
                // `dev_addr`.
                return schema.required(sharedMessages.validateRequired)
              }

              return schema
            }),
            keys: Yup.object().shape({
              f_nwk_s_int_key: Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                  .test('is-not-all-zero-key', messages.validateSessionKey, isNonZeroSessionKey)
                  .required(sharedMessages.validateRequired),
              }),
              s_nwk_s_int_key: Yup.lazy(() =>
                isNewVersion
                  ? Yup.object().shape({
                      key: Yup.string()
                        .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                        .test(
                          'is-not-all-zero-key',
                          messages.validateSessionKey,
                          isNonZeroSessionKey,
                        )
                        .required(sharedMessages.validateRequired),
                    })
                  : Yup.object().strip(),
              ),
              nwk_s_enc_key: Yup.lazy(() =>
                isNewVersion
                  ? Yup.object().shape({
                      key: Yup.string()
                        .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                        .test(
                          'is-not-all-zero-key',
                          messages.validateSessionKey,
                          isNonZeroSessionKey,
                        )
                        .required(sharedMessages.validateRequired),
                    })
                  : Yup.object().strip(),
              ),
            }),
          })
        }
        return schema.strip()
      },
    ),
    mac_settings: Yup.object().when(
      ['_activation_mode', 'supports_class_b', 'supports_class_c', 'lorawan_version'],
      ([mode, isClassB, isClassC, version], schema) => {
        const isNewVersion = parseLorawanMacVersion(version) >= 110

        return schema.shape({
          beacon_frequency: Yup.lazy(() => {
            if (!isClassB || mode === ACTIVATION_MODES.OTAA) {
              return Yup.number().strip()
            }

            return Yup.number().test(
              'is-valid-freq',
              sharedMessages.validateFreqDynamic,
              dynamicFrequencyTest,
            )
          }),
          desired_beacon_frequency: Yup.lazy(() => {
            if (!isClassB || mode === ACTIVATION_MODES.MULTICAST) {
              return Yup.number().strip()
            }

            return Yup.number().test(
              'is-valid-freq',
              sharedMessages.validateFreqDynamic,
              dynamicFrequencyTest,
            )
          }),
          class_b_timeout: Yup.lazy(value => {
            if (!isClassB || !Boolean(value)) {
              return Yup.string().strip()
            }

            return Yup.string()
          }),
          class_c_timeout: Yup.lazy(value => {
            if (!isClassC || !Boolean(value)) {
              return Yup.string().strip()
            }

            return Yup.string()
          }),
          rx1_delay: Yup.lazy(delay => {
            if (delay === undefined || delay === '' || mode !== ACTIVATION_MODES.ABP) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(1, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          desired_rx1_delay: Yup.lazy(delay => {
            if (delay === undefined || delay === '' || mode === ACTIVATION_MODES.MULTICAST) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(1, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          rx1_data_rate_offset: Yup.lazy(value => {
            if (value === undefined || value === '' || mode !== ACTIVATION_MODES.ABP) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(0, Yup.passValues(sharedMessages.validateNumberGte))
              .max(7, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          desired_rx1_data_rate_offset: Yup.lazy(value => {
            if (value === undefined || value === '' || mode === ACTIVATION_MODES.MULTICAST) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(0, Yup.passValues(sharedMessages.validateNumberGte))
              .max(7, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          resets_f_cnt: Yup.lazy(() => {
            if (mode !== ACTIVATION_MODES.ABP) {
              return Yup.boolean().strip()
            }

            return Yup.boolean().default(false)
          }),
          ping_slot_data_rate_index: Yup.lazy(dataRate => {
            if (
              !isClassB ||
              dataRate === '' ||
              dataRate === undefined ||
              mode === ACTIVATION_MODES.OTAA
            ) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(0, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          desired_ping_slot_data_rate_index: Yup.lazy(dataRate => {
            if (
              !isClassB ||
              dataRate === '' ||
              dataRate === undefined ||
              mode === ACTIVATION_MODES.MULTICAST
            ) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(0, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          rx2_data_rate_index: Yup.lazy(dataRate => {
            if (dataRate === '' || dataRate === undefined || mode === ACTIVATION_MODES.OTAA) {
              return Yup.number().strip()
            }

            return Yup.number()
              .min(0, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          desired_rx2_data_rate_index: Yup.lazy(dataRate => {
            if (dataRate === '' || dataRate === undefined || mode === ACTIVATION_MODES.MULTICAST) {
              return Yup.number().strip()
            }

            return Yup.number()
              .nullable()
              .min(0, Yup.passValues(sharedMessages.validateNumberGte))
              .max(15, Yup.passValues(sharedMessages.validateNumberLte))
          }),
          rx2_frequency: Yup.lazy(frequency => {
            if (frequency === undefined || frequency === '' || mode === ACTIVATION_MODES.OTAA) {
              return Yup.number().strip()
            }
            return Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte))
          }),
          ping_slot_periodicity: Yup.lazy(value => {
            if (isClassB) {
              if (mode === ACTIVATION_MODES.MULTICAST || mode === ACTIVATION_MODES.ABP) {
                return Yup.string().required(sharedMessages.validateRequired)
              }

              if (!value) {
                return Yup.string().strip()
              }

              return Yup.string()
            }

            return Yup.string().strip()
          }),
          desired_rx2_frequency: Yup.lazy(frequency => {
            if (
              frequency === undefined ||
              frequency === '' ||
              mode === ACTIVATION_MODES.MULTICAST
            ) {
              return Yup.number().strip()
            }
            return Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte))
          }),
          ping_slot_frequency: Yup.lazy(() => {
            if (!isClassB || mode === ACTIVATION_MODES.OTAA) {
              return Yup.number().strip()
            }

            return Yup.number().test(
              'is-valid-freq',
              sharedMessages.validateFreqDynamic,
              dynamicFrequencyTest,
            )
          }),
          desired_ping_slot_frequency: Yup.lazy(frequency => {
            if (!Boolean(frequency) || !isClassB || mode === ACTIVATION_MODES.MULTICAST) {
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
          max_duty_cycle: Yup.lazy(() => {
            if (mode !== ACTIVATION_MODES.ABP) {
              return Yup.string().strip()
            }

            return Yup.string()
          }),
          desired_max_duty_cycle: Yup.lazy(value => {
            if (!Boolean(value) || mode === ACTIVATION_MODES.MULTICAST) {
              return Yup.string().strip()
            }

            return Yup.string()
          }),
          use_adr: Yup.bool().nullable(),
          adr_margin: Yup.number().nullable(),
          adr: Yup.lazy(value => {
            if (value && 'static' in value) {
              return Yup.object().shape({
                static: Yup.object().shape({
                  data_rate_index: Yup.number().nullable(),
                  tx_power_index: Yup.number().nullable(),
                  nb_trans: Yup.number().nullable(),
                }),
              })
            } else if (value && 'disabled' in value) {
              return Yup.object().shape({
                disabled: Yup.object().shape({}).noUnknown(),
              })
            }

            return Yup.object().shape({
              dynamic: Yup.object().shape({
                margin: Yup.number().nullable(),
              }),
            })
          }),
          desired_adr_ack_limit_exponent: Yup.string().when(['adr'], ([adr], schema) => {
            if (!('dynamic' in adr) || !isNewVersion || mode === ACTIVATION_MODES.MULTICAST) {
              return schema.strip()
            }

            return schema
          }),
          desired_adr_ack_delay_exponent: Yup.string().when(['adr'], ([adr], schema) => {
            if (!('dynamic' in adr) || !isNewVersion || mode === ACTIVATION_MODES.MULTICAST) {
              return schema.strip()
            }

            return schema
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

            return Yup.number()
          }),
        })
      },
    ),
  })
  .noUnknown()

export default validationSchema
