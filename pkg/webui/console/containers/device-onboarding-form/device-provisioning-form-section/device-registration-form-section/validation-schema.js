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

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import { parseLorawanMacVersion } from '@console/lib/device-utils'

// Validation schema of the device registration form section.
// Please observe the following rules to keep the validation schemas maintainable:
// 1. DO NOT USE ANY TYPE CONVERSIONS HERE. Use decocer/encoder on field level instead.
//    Consider all values as backend values. Exceptions may apply in consideration.
// 2. Comment each individual validation prop and use whitespace to structure visually.
// 3. Do not use ternary assignments but use plain if statements to ensure clarity.

const devEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength))

const validationSchema = Yup.object({
  ids: Yup.object().shape({
    dev_eui: devEUISchema.required(sharedMessages.validateRequired),
  }),

  root_keys: Yup.object().when(
    ['supports_join', 'lorawan_version', '$mayEditKeys', '$jsEnabled'],
    (isOTAA, version, mayEditKeys, jsEnabled, schema) => {
      if (!mayEditKeys || !isOTAA || !jsEnabled) {
        return schema.strip()
      }

      const strippedSchema = Yup.object().strip()
      const keySchema = Yup.lazy(() => {
        if (mayEditKeys) {
          return Yup.object().shape({
            key: Yup.string()
              .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
              .required(sharedMessages.validateRequired),
          })
        }

        return Yup.object().strip()
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
  session: Yup.object().when(
    ['lorawan_version', 'supports_join', '$nsEnabled', '$asEnabled'],
    (version, isOTAA, nsEnabled, asEnabled, schema) => {
      if (isOTAA || (!nsEnabled && !asEnabled)) {
        return schema.strip()
      }

      const lwVersion = parseLorawanMacVersion(version)

      return schema.shape({
        dev_addr: Yup.lazy(() => {
          if (nsEnabled) {
            return Yup.string()
              .length(4 * 2, Yup.passValues(sharedMessages.validateLength))
              .required(sharedMessages.validateRequired)
          }

          return Yup.string().strip()
        }),
        keys: Yup.object().shape({
          app_s_key: Yup.lazy(() => {
            if (asEnabled) {
              return Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                  .required(sharedMessages.validateRequired),
              })
            }

            return Yup.object().strip()
          }),
          f_nwk_s_int_key: Yup.lazy(() => {
            if (nsEnabled) {
              return Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                  .required(sharedMessages.validateRequired),
              })
            }

            return Yup.object().strip()
          }),
          s_nwk_s_int_key: Yup.lazy(() => {
            if (lwVersion >= 110 && nsEnabled) {
              return Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                  .required(sharedMessages.validateRequired),
              })
            }

            return Yup.object().strip()
          }),
          nwk_s_enc_key: Yup.lazy(() => {
            if (lwVersion >= 110 && nsEnabled) {
              return Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                  .required(sharedMessages.validateRequired),
              })
            }

            return Yup.object().strip()
          }),
        }),
      })
    },
  ),
  // Derived.
  application_server_address: Yup.string().when(
    ['$asUrl', '$asEnabled'],
    (asUrl, asEnabled, schema) => {
      if (!asEnabled) {
        return schema.strip()
      }

      return schema.default(getHostFromUrl(asUrl))
    },
  ),
  network_server_address: Yup.string().when(
    ['$nsUrl', '$nsEnabled', '$mayEditKeys'],
    (nsUrl, nsEnabled, mayEditKeys, schema) => {
      if (!nsEnabled || !mayEditKeys) {
        return schema.strip()
      }

      return schema.default(getHostFromUrl(nsUrl))
    },
  ),
  join_server_address: Yup.string().when(['$jsUrl', '$jsEnabled'], (jsUrl, jsEnabled, schema) => {
    if (!jsEnabled) {
      return schema.strip()
    }

    return schema.default(getHostFromUrl(jsUrl))
  }),
  lorawan_version: Yup.string(),
})

const initialValues = {
  ids: {
    device_id: undefined,
    dev_eui: undefined,
  },
  root_keys: {
    app_key: { key: undefined },
    nwk_key: { key: undefined },
  },
  session: {
    dev_addr: undefined,
    keys: {
      app_s_key: {
        key: undefined,
      },
      f_nwk_s_int_key: {
        key: undefined,
      },
      s_nwk_s_int_key: {
        key: undefined,
      },
      nwk_s_enc_key: {
        key: undefined,
      },
    },
  },
  join_server_address: undefined,
  application_server_address: undefined,
  network_server_address: undefined,
}

export { validationSchema as default, devEUISchema, initialValues }
