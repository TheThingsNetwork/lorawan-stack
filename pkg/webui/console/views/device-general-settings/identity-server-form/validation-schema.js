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
import { id as deviceIdRegexp, address as addressRegexp } from '../../../lib/regexp'
import m from '../../../components/device-data-form/messages'
import { selectJsConfig } from '../../../../lib/selectors/env'

import { parseLorawanMacVersion } from '../utils'

const jsConfig = selectJsConfig()

const random16BytesString = () => randomByteString(32)
const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object()
  .shape({
    ids: Yup.object().shape({
      device_id: Yup.string()
        .matches(deviceIdRegexp, sharedMessages.validateAlphanum)
        .min(2, sharedMessages.validateTooShort)
        .max(36, sharedMessages.validateTooLong)
        .required(sharedMessages.validateRequired),
    }),
    name: Yup.string()
      .min(2, sharedMessages.validateTooShort)
      .max(50, sharedMessages.validateTooLong),
    description: Yup.string().max(2000, sharedMessages.validateTooLong),
    network_server_address: Yup.string().matches(
      addressRegexp,
      sharedMessages.validateAddressFormat,
    ),
    application_server_address: Yup.string().matches(
      addressRegexp,
      sharedMessages.validateAddressFormat,
    ),
    _external_js: Yup.boolean(),
    _supports_join: Yup.boolean(),
    _lorawan_version: Yup.string(),
    join_server_address: Yup.string().when(
      ['_supports_join', ' _external_js'],
      (supportsJoin, externalJs, schema) => {
        if (!supportsJoin) {
          return schema.strip()
        }

        if (externalJs) {
          return schema.transform(() => '')
        }

        return schema
          .matches(addressRegexp, sharedMessages.validateAddressFormat)
          .transform(toUndefined)
          .default(new URL(jsConfig.base_url).hostname)
      },
    ),
    resets_join_nonces: Yup.bool().when(
      ['_supports_join', '_lorawan_version', '_external_js'],
      (supportsJoin, lorawanVersion, externalJs, schema) => {
        if (!supportsJoin || parseLorawanMacVersion(lorawanVersion) < 110) {
          return schema.strip()
        }

        if (externalJs) {
          return schema.transform(() => false)
        }

        return schema
      },
    ),
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
  })
  .noUnknown()

export default validationSchema
