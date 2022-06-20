// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import sharedMessages from '@ttn-lw/lib/shared-messages'
import Yup from '@ttn-lw/lib/yup'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'
import { id as deviceIdRegexp } from '@ttn-lw/lib/regexp'

import { address as addressRegexp } from '@console/lib/regexp'
import { ACTIVATION_MODES, parseLorawanMacVersion } from '@console/lib/device-utils'

import { REGISTRATION_TYPES } from '../../utils'

import { DEVICE_CLASS_MAP } from './constants'

const factoryPresetFreqNumericTest = frequencies =>
  frequencies.every(freq => {
    if (typeof freq !== 'undefined') {
      return !isNaN(parseInt(freq))
    }

    return true
  })

const factoryPresetFreqRequiredTest = frequencies =>
  frequencies.every(freq => typeof freq !== 'undefined' && freq !== '')

const deviceIdSchema = Yup.string()
  .matches(deviceIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
  .min(2, Yup.passValues(sharedMessages.validateTooShort))
  .max(36, Yup.passValues(sharedMessages.validateTooLong))
  .required(sharedMessages.validateRequired)
const joinEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength))
const devEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength))

const idsSchema = Yup.object({
  ids: Yup.object().when(
    ['_activation_mode', 'lorawan_version'],
    (activationMode, version, schema) => {
      if (activationMode === ACTIVATION_MODES.OTAA) {
        return schema.shape({
          device_id: deviceIdSchema,
          join_eui: joinEUISchema.required(sharedMessages.validateRequired),
          dev_eui: devEUISchema.required(sharedMessages.validateRequired),
        })
      }

      if (activationMode === ACTIVATION_MODES.ABP) {
        if (parseLorawanMacVersion(version) === 104) {
          return schema.shape({
            device_id: deviceIdSchema,
            dev_eui: devEUISchema.required(sharedMessages.validateRequired),
            join_eui: Yup.string().strip(),
          })
        }

        return schema.shape({
          device_id: deviceIdSchema,
          dev_eui: Yup.lazy(value =>
            !value
              ? Yup.string().strip()
              : Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength)),
          ),
          join_eui: Yup.string().strip(),
        })
      }

      return schema.shape({
        device_id: deviceIdSchema,
        join_eui: Yup.string().strip(),
      })
    },
  ),
})

const rootKeysSchema = Yup.object({
  root_keys: Yup.object().when(
    [
      'lorawan_version',
      '$mayEditKeys',
      '_activation_mode',
      '$jsEnabled',
      '$jsUrl',
      'join_server_address',
    ],
    (version, mayEditKeys, mode, jsEnabled, jsUrl, jsHost, schema) => {
      if (
        !jsEnabled ||
        !mayEditKeys ||
        mode !== ACTIVATION_MODES.OTAA ||
        getHostFromUrl(jsUrl) !== jsHost
      ) {
        return schema.strip()
      }

      const strippedSchema = Yup.object().strip()
      const keySchema = Yup.object().shape({
        key: Yup.string()
          .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
          .required(sharedMessages.validateRequired),
      })

      if (parseLorawanMacVersion(version) < 110) {
        return schema.shape({
          nwk_key: strippedSchema,
          app_key: keySchema,
        })
      }

      return schema.shape({
        nwk_key: keySchema,
        app_key: keySchema,
      })
    },
  ),
})
const sessionSchema = Yup.object({
  session: Yup.object().when(
    ['lorawan_version', '_activation_mode', '$asEnabled', '$mayEditKeys'],
    (version, activationMode, asEnabled, mayEditKeys, schema) => {
      if (activationMode === ACTIVATION_MODES.OTAA || activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      const lwVersion = parseLorawanMacVersion(version)

      return schema.shape({
        dev_addr: Yup.string()
          .length(4 * 2, Yup.passValues(sharedMessages.validateLength))
          .required(sharedMessages.validateRequired),
        keys: Yup.object().shape({
          app_s_key: Yup.lazy(() =>
            asEnabled && mayEditKeys
              ? Yup.object().shape({
                  key: Yup.string()
                    .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                    .required(sharedMessages.validateRequired),
                })
              : Yup.object().strip(),
          ),
          f_nwk_s_int_key: Yup.object().shape({
            key: Yup.string()
              .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
              .required(sharedMessages.validateRequired),
          }),
          s_nwk_s_int_key: Yup.lazy(() =>
            lwVersion >= 110
              ? Yup.object().shape({
                  key: Yup.string()
                    .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                    .required(sharedMessages.validateRequired),
                })
              : Yup.object().strip(),
          ),
          nwk_s_enc_key: Yup.lazy(() =>
            lwVersion >= 110
              ? Yup.object().shape({
                  key: Yup.string()
                    .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                    .required(sharedMessages.validateRequired),
                })
              : Yup.object().strip(),
          ),
        }),
      })
    },
  ),
})

const macSettingsSchema = Yup.object({
  mac_settings: Yup.object().when(
    [
      '$nsEnabled',
      '_activation_mode',
      'supports_class_b',
      'supports_class_c',
      '$hasRxDataRateOffset',
      '$hasRxDelay',
      '$hasRxDataRateIndex',
      '$hasPingSlotDataRateIndex',
      '$hasBeaconFrequency',
      '$hasClassBTimeout',
      '$hasClassCTimeout',
      '$hasPingSlotFrequency',
      '$hasRxFrequency',
    ],
    (
      nsEnabled,
      mode,
      isClassB,
      isClassC,
      hasRxDelay,
      hasRxDataRateOffset,
      hasRxDataRateIndex,
      hasPingSlotDataRateIndex,
      hasBeaconFrequency,
      hasClassBTimeout,
      hasClassCTimeout,
      hasPingSlotFrequency,
      hasRxFrequency,
      schema,
    ) => {
      if (!nsEnabled) {
        return schema.strip()
      }

      return schema.shape({
        resets_f_cnt: Yup.lazy(() => {
          if (mode !== ACTIVATION_MODES.ABP) {
            return Yup.boolean().strip()
          }

          return Yup.boolean()
        }),
        rx1_data_rate_offset: Yup.lazy(value => {
          if (mode !== ACTIVATION_MODES.ABP || (value === undefined && !hasRxDataRateOffset)) {
            return Yup.number().strip()
          }

          const schema = Yup.number()
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(7, Yup.passValues(sharedMessages.validateNumberLte))

          if (hasRxDataRateOffset) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
        }),
        rx1_delay: Yup.lazy(delay => {
          if (
            mode !== ACTIVATION_MODES.ABP ||
            ((delay === undefined || delay === '') && !hasRxDelay)
          ) {
            return Yup.number().strip()
          }

          const schema = Yup.number()
            .min(1, Yup.passValues(sharedMessages.validateNumberGte))
            .max(15, Yup.passValues(sharedMessages.validateNumberLte))

          if (hasRxDelay) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
        }),
        factory_preset_frequencies: Yup.lazy(frequencies => {
          if (!Boolean(frequencies)) {
            return Yup.array().strip()
          }

          return Yup.array()
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
            )
        }),
        rx2_frequency: Yup.lazy(frequency => {
          if ((frequency === undefined || frequency === '') && !hasRxFrequency) {
            return Yup.number().strip()
          }

          const schema = Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte))

          if (hasRxFrequency) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
        }),
        beacon_frequency: Yup.lazy(frequency => {
          if (!isClassB || ((frequency === undefined || frequency === '') && !hasBeaconFrequency)) {
            return Yup.number().strip()
          }

          const schema = Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte))

          if (hasBeaconFrequency) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
        }),
        ping_slot_frequency: Yup.lazy(frequency => {
          if (
            !isClassB ||
            ((frequency === undefined || frequency === '') && !hasPingSlotFrequency)
          ) {
            return Yup.number().strip()
          }

          const schema = Yup.number().min(100000, Yup.passValues(sharedMessages.validateNumberGte))

          if (hasPingSlotFrequency) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
        }),
        rx2_data_rate_index: Yup.lazy(dataRate => {
          if ((dataRate === '' || dataRate === undefined) && !hasRxDataRateIndex) {
            return Yup.number().strip()
          }

          const schema = Yup.number()
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(15, Yup.passValues(sharedMessages.validateNumberLte))

          if (hasRxDataRateIndex) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
        }),
        ping_slot_data_rate_index: Yup.lazy(dataRate => {
          if (
            !isClassB ||
            ((dataRate === '' || dataRate === undefined) && !hasPingSlotDataRateIndex)
          ) {
            return Yup.number().strip()
          }

          const schema = Yup.number()
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(15, Yup.passValues(sharedMessages.validateNumberLte))

          if (hasPingSlotDataRateIndex) {
            return schema.required(sharedMessages.validateRequired)
          }

          return schema
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
        class_b_timeout: Yup.lazy(value => {
          if (!isClassB || (!Boolean(value) && !hasClassBTimeout)) {
            return Yup.string().strip()
          }

          if (hasClassBTimeout) {
            return Yup.string().required(sharedMessages.validateRequired)
          }

          return Yup.string()
        }),
        class_c_timeout: Yup.lazy(value => {
          if (!isClassC || (!Boolean(value) && !hasClassCTimeout)) {
            return Yup.string().strip()
          }

          if (hasClassBTimeout) {
            return Yup.string().required(sharedMessages.validateRequired)
          }

          return Yup.string()
        }),
      })
    },
  ),
})

const validationSchema = Yup.object({
  supports_class_b: Yup.boolean().when(
    ['_device_class', '$nsEnabled'],
    (deviceClass, nsEnabled, schema) => {
      if (!nsEnabled) {
        return schema.strip()
      }

      return schema
        .transform(() => undefined)
        .default(
          deviceClass === DEVICE_CLASS_MAP.CLASS_B || deviceClass === DEVICE_CLASS_MAP.CLASS_B_C,
        )
    },
  ),
  supports_class_c: Yup.boolean().when(['_device_class'], (deviceClass, schema) =>
    schema
      .transform(() => undefined)
      .default(
        deviceClass === DEVICE_CLASS_MAP.CLASS_C || deviceClass === DEVICE_CLASS_MAP.CLASS_B_C,
      ),
  ),
  supports_join: Yup.boolean().when(
    ['$jsEnabled', '_activation_mode'],
    (jsEnabled, activationMode, schema) => {
      if (!jsEnabled || activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      if (
        activationMode === ACTIVATION_MODES.ABP ||
        activationMode === ACTIVATION_MODES.MULTICAST
      ) {
        return schema.transform(() => undefined).default(false)
      }

      if (activationMode === ACTIVATION_MODES.OTAA) {
        return schema.transform(() => undefined).default(true)
      }

      return schema
    },
  ),
  multicast: Yup.boolean()
    .transform(() => undefined)
    .when(['$nsEnabled', '_activation_mode'], (nsEnabled, activationMode, schema) => {
      if (!nsEnabled || activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      if (activationMode === ACTIVATION_MODES.OTAA || activationMode === ACTIVATION_MODES.ABP) {
        return schema.transform(() => undefined).default(false)
      }

      if (activationMode === ACTIVATION_MODES.MULTICAST) {
        return schema.transform(() => undefined).default(true)
      }

      return schema
    }),
  _device_class: Yup.string().when(['_activation_mode'], (mode, schema) => {
    if (mode === ACTIVATION_MODES.MULTICAST) {
      return schema.required(sharedMessages.validateRequired)
    }

    return schema.oneOf(Object.values(DEVICE_CLASS_MAP))
  }),
  _default_ns_settings: Yup.bool(),
  _activation_mode: Yup.mixed().when(
    ['$nsEnabled', '$jsEnabled', '$mayEditKeys'],
    (nsEnabled, jsEnabled, mayEditKeys, schema) => {
      const canCreateNs = nsEnabled && mayEditKeys
      const canCreateJs = jsEnabled

      if (!canCreateJs && !canCreateNs) {
        return schema.oneOf([ACTIVATION_MODES.NONE]).required(sharedMessages.validateRequired)
      }

      if (!canCreateNs) {
        return schema
          .oneOf([ACTIVATION_MODES.OTAA, ACTIVATION_MODES.NONE])
          .required(sharedMessages.validateRequired)
      }

      if (!canCreateJs) {
        return schema
          .oneOf([ACTIVATION_MODES.ABP, ACTIVATION_MODES.MULTICAST, ACTIVATION_MODES.NONE])
          .required(sharedMessages.validateRequired)
      }

      return schema.oneOf(Object.values(ACTIVATION_MODES)).required(sharedMessages.validateRequired)
    },
  ),
  lorawan_version: Yup.string().when(['_activation_mode'], (activationMode, schema) => {
    if (activationMode === ACTIVATION_MODES.NONE) {
      return schema.strip()
    }

    return schema.required(sharedMessages.validateRequired)
  }),
  lorawan_phy_version: Yup.string().when('_activation_mode', {
    is: ACTIVATION_MODES.NONE,
    then: schema => schema.strip(),
    otherwise: schema => schema.required(sharedMessages.validateRequired),
  }),
  frequency_plan_id: Yup.string().when(
    ['_activation_mode', '$nsEnabled'],
    (mode, nsEnabled, schema) => {
      if (mode === ACTIVATION_MODES.NONE || !nsEnabled) {
        return schema.strip()
      }

      return schema.required(sharedMessages.validateRequired)
    },
  ),
  _registration: Yup.mixed()
    .oneOf([REGISTRATION_TYPES.SINGLE, REGISTRATION_TYPES.MULTIPLE])
    .default(REGISTRATION_TYPES.SINGLE),
  _external_servers: Yup.bool().when(
    ['_activation_mode', '$jsEnabled', '$nsEnabled', '$asEnabled'],
    (activationMode, jsEnabled, nsEnabled, asEnabled, schema) => {
      if (activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      if (activationMode === ACTIVATION_MODES.OTAA) {
        return jsEnabled && nsEnabled ? schema.default(false) : schema.default(true)
      }

      return nsEnabled && asEnabled ? schema.default(false) : schema.default(true)
    },
  ),
  join_server_address: Yup.string().when(
    ['_activation_mode', '$jsEnabled'],
    (activationMode, jsEnabled, schema) => {
      if (activationMode !== ACTIVATION_MODES.OTAA || !jsEnabled) {
        return schema.strip()
      }

      return schema.matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
    },
  ),
  application_server_address: Yup.string().when(
    ['$asUrl', '$asEnabled', '_activation_mode'],
    (asUrl, asEnabled, activationMode, schema) => {
      if (activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      if (!asEnabled) {
        return schema.matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
      }

      return schema
        .matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
        .default(getHostFromUrl(asUrl))
    },
  ),
  network_server_address: Yup.string().when(
    ['$nsUrl', '$nsEnabled', '$mayEditKeys', '_activation_mode'],
    (nsUrl, nsEnabled, mayEditKeys, activationMode, schema) => {
      if (activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      if (!nsEnabled) {
        return schema.matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
      }

      if (!mayEditKeys) {
        if (activationMode === ACTIVATION_MODES.OTAA) {
          return schema
            .matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
            .default(getHostFromUrl(nsUrl))
        }

        return schema.matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
      }

      return schema
        .matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
        .default(getHostFromUrl(nsUrl))
    },
  ),
})
  .concat(idsSchema)
  .concat(rootKeysSchema)
  .concat(sessionSchema)
  .concat(macSettingsSchema)
  .noUnknown()

export {
  validationSchema as default,
  rootKeysSchema,
  sessionSchema,
  idsSchema,
  macSettingsSchema,
  devEUISchema,
}
