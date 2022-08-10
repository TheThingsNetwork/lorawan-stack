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
    dev_eui: devEUISchema.when('supports_join', supports_joins => {
      if (supports_joins) {
        return devEUISchema.required(sharedMessages.validateRequired)
      }

      return devEUISchema.default(null).nullable()
    }),
    device_id: Yup.string().when('_claim', claim => {
      if (!claim) {
        return Yup.string().required(sharedMessages.validateRequired)
      }

      return Yup.string().default(null).nullable()
    }),
  }),

  root_keys: Yup.object().when(
    ['supports_join', 'lorawan_version', '_inputMethod', '$mayEditKeys'],
    (isOTAA, version, inputMethod, mayEditKeys, schema) => {
      const notRequiredKeySchema = Yup.object().shape({
        key: Yup.string().default('').nullable(),
      })
      const requiredKeySchema = Yup.lazy(() => {
        if (mayEditKeys) {
          return Yup.object().shape({
            key: Yup.string()
              .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
              .required(sharedMessages.validateRequired),
          })
        }

        return notRequiredKeySchema
      })

      if (!mayEditKeys || !isOTAA || inputMethod === 'manual') {
        return schema.shape({
          nwk_key: notRequiredKeySchema,
          app_key: notRequiredKeySchema,
        })
      }

      if (parseLorawanMacVersion(version) < 110) {
        return schema.shape({
          nwk_key: notRequiredKeySchema,
          app_key: requiredKeySchema,
        })
      }

      return schema.shape({
        nwk_key: requiredKeySchema,
        app_key: requiredKeySchema,
      })
    },
  ),
  session: Yup.object().when(['lorawan_version', 'supports_join'], (version, isOTAA, schema) => {
    const lwVersion = parseLorawanMacVersion(version)
    const notRequiredKeySchema = Yup.object().shape({
      key: Yup.string().default('').nullable(),
    })

    return schema.shape({
      dev_addr: Yup.lazy(() => {
        if (!isOTAA) {
          return Yup.string()
            .length(4 * 2, Yup.passValues(sharedMessages.validateLength))
            .required(sharedMessages.validateRequired)
        }

        return notRequiredKeySchema
      }),
      keys: Yup.object().shape({
        app_s_key: Yup.lazy(() => {
          if (!isOTAA) {
            return Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                .required(sharedMessages.validateRequired),
            })
          }

          return notRequiredKeySchema
        }),
        f_nwk_s_int_key: Yup.lazy(() => {
          if (!isOTAA) {
            return Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                .required(sharedMessages.validateRequired),
            })
          }

          return notRequiredKeySchema
        }),
        s_nwk_s_int_key: Yup.lazy(() => {
          if (!isOTAA && lwVersion >= 110) {
            return Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                .required(sharedMessages.validateRequired),
            })
          }

          return notRequiredKeySchema
        }),
        nwk_s_enc_key: Yup.lazy(() => {
          if (!isOTAA && lwVersion >= 110) {
            return Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                .required(sharedMessages.validateRequired),
            })
          }

          return notRequiredKeySchema
        }),
      }),
    })
  }),
  lorawan_version: Yup.string().default(null).nullable(),
})

const initialValues = {
  ids: {
    device_id: '',
    dev_eui: '',
  },
  root_keys: {
    app_key: { key: '' },
    nwk_key: { key: '' },
  },
  session: {
    dev_addr: '',
    keys: {
      app_s_key: {
        key: '',
      },
      f_nwk_s_int_key: {
        key: '',
      },
      s_nwk_s_int_key: {
        key: '',
      },
      nwk_s_enc_key: {
        key: '',
      },
    },
  },
}

export { validationSchema as default, devEUISchema, initialValues }
