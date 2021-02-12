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

import { parseLorawanMacVersion } from '@console/lib/device-utils'

const validationSchema = Yup.object()
  .shape({
    net_id: Yup.nullableString()
      .emptyOrLength(3 * 2, Yup.passValues(sharedMessages.validateLength)) // 3 Byte hex.
      .default(null),
    root_keys: Yup.object().when(
      ['$externalJs', '$lorawanVersion', '$mayEditKeys', '$mayEditkeys'],
      (externalJs, lorawanVersion, mayEditKeys, mayReadKeys, schema) => {
        const strippedSchema = Yup.object().strip()
        const keySchema = Yup.lazy(value =>
          !externalJs && Boolean(value) && Boolean(value.key)
            ? Yup.object().shape({
                key: Yup.string().emptyOrLength(
                  16 * 2,
                  Yup.passValues(sharedMessages.validateLength),
                ), // 16 Byte hex.
              })
            : Yup.object().strip(),
        )

        if (!mayEditKeys && !mayReadKeys) {
          return schema.strip()
        }

        if (externalJs) {
          return schema.shape({
            nwk_key: strippedSchema,
            app_key: strippedSchema,
          })
        }

        if (parseLorawanMacVersion(lorawanVersion) < 110) {
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
    resets_join_nonces: Yup.boolean().when('$lorawanVersion', {
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
  })
  .noUnknown()

export default validationSchema
