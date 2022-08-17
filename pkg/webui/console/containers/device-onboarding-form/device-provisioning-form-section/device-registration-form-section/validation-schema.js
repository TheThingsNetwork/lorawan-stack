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

const networkKeySchema = Yup.object({
  app_key: Yup.object({
    key: Yup.string()
      .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
})

const appKeySchema = Yup.object({
  app_key: Yup.object({
    key: Yup.string()
      .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
})

const devAddrSchema = Yup.object({
  dev_addr: Yup.string()
    .length(4 * 2, Yup.passValues(sharedMessages.validateLength))
    .required(sharedMessages.validateRequired),
})

const sessionKeysSchemaBase = Yup.object({
  app_s_key: Yup.object({
    key: Yup.string()
      .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
  f_nwk_s_int_key: Yup.object({
    key: Yup.string()
      .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
})

const sessionKeysVersion110Schema = Yup.object({
  s_nwk_s_int_key: Yup.object({
    key: Yup.string()
      .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
  nwk_s_enc_key: Yup.object({
    key: Yup.string()
      .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
})

const validationSchema = Yup.object({
  ids: Yup.object().shape({
    dev_eui: Yup.string().when('$supportsJoin', {
      is: true,
      then: schema =>
        schema
          .length(8 * 2, Yup.passValues(sharedMessages.validateLength))
          .required(sharedMessages.validateRequired),
    }),
    device_id: Yup.string(),
    join_eui: Yup.string()
      .length(8 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
}).when(
  ['supports_join', '_claim', 'lorawan_version', '$mayEditKeys'],
  (isOTAA, claim, version, mayEditKeys, schema) => {
    if (mayEditKeys || isOTAA || claim === false) {
      const rootKeysSchema = Yup.object().concat(appKeySchema)
      if (parseLorawanMacVersion(version) >= 110) {
        rootKeysSchema.concat(networkKeySchema)
      }
      schema.concat(
        Yup.object({
          root_keys: rootKeysSchema,
        }),
      )
    }

    if (!isOTAA) {
      const sessionKeysSchema = Yup.object().concat(sessionKeysSchemaBase)
      if (parseLorawanMacVersion(version) >= 110) {
        sessionKeysSchema.concat(sessionKeysVersion110Schema)
      }
      schema.concat(
        Yup.object({
          session: Yup.object({
            devAddrSchema,
            keys: sessionKeysSchema,
          }),
        }),
      )
    }
  },
)

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

export { validationSchema as default, initialValues }
