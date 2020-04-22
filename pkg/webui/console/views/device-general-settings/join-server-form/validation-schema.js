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

import * as Yup from 'yup'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { parseLorawanMacVersion } from '../utils'

const validationSchema = Yup.object()
  .shape({
    net_id: Yup.nullableString()
      .emptyOrLength(3 * 2, Yup.passValues(sharedMessages.validateLength)) // 3 Byte hex.
      .default(''),
    root_keys: Yup.object().when(
      ['_external_js', '_lorawan_version', '_may_edit_keys', '_may_read_keys'],
      (externalJs, version, mayEditKeys, mayReadKeys, schema) => {
        const strippedSchema = Yup.object().strip()
        const keySchema = Yup.lazy(value => {
          return !externalJs && Boolean(value) && Boolean(value.key)
            ? Yup.object().shape({
                key: Yup.string().emptyOrLength(
                  16 * 2,
                  Yup.passValues(sharedMessages.validateLength),
                ), // 16 Byte hex.
              })
            : Yup.object().strip()
        })

        if (!mayEditKeys && !mayReadKeys) {
          return schema.strip()
        }

        if (externalJs) {
          return schema.shape({
            nwk_key: strippedSchema,
            app_key: strippedSchema,
          })
        }

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
    resets_join_nonces: Yup.boolean().when('_lorawan_version', {
      // Verify if lorawan version is 1.1.0 or higher.
      is: version => parseLorawanMacVersion(version) >= 110,
      then: schema => schema,
      otherwise: schema => schema.strip(),
    }),
    application_server_id: Yup.string()
      .max(100, Yup.passValues(sharedMessages.validateTooLong))
      .default(''),
    application_server_kek_label: Yup.string()
      .max(2048, Yup.passValues(sharedMessages.validateTooLong))
      .default(''),
    network_server_kek_label: Yup.string()
      .max(2048, Yup.passValues(sharedMessages.validateTooLong))
      .default(''),
    _external_js: Yup.boolean().default(false),
    _lorawan_version: Yup.string().default('1.1.0'),
    _may_edit_keys: Yup.boolean().default(false),
    _may_read_keys: Yup.boolean().default(false),
  })
  .noUnknown()

export default validationSchema
