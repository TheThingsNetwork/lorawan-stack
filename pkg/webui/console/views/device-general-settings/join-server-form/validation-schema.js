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

import randomByteString from '../../../lib/random-bytes'
import sharedMessages from '../../../../lib/shared-messages'
import { address as addressRegexp } from '../../../lib/regexp'
import m from '../../../components/device-data-form/messages'

import { parseLorawanMacVersion } from '../utils'

const random16BytesString = () => randomByteString(32)
const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object().shape({
  join_server_address: Yup.string().when('_external_js', {
    is: false,
    then: schema => schema.matches(addressRegexp, sharedMessages.validateAddressFormat),
    otherwise: schema => schema.default(''),
  }),
  root_keys: Yup.object().when(
    ['_external_js', '_lorawan_version'],
    (externalJs, version, schema) => {
      const strippedSchema = Yup.object().strip()
      const keySchema = Yup.lazy(() => {
        return !externalJs
          ? Yup.object().shape({
              key: Yup.string()
                .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
                .transform(toUndefined)
                .default(random16BytesString),
            })
          : Yup.object().strip()
      })

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
  _external_js: Yup.boolean(),
  _lorawan_version: Yup.string(),
})

export default validationSchema
