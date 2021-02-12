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

import { attributeValidCheck, attributeTooShortCheck } from '@console/lib/attributes'
import { id as deviceIdRegexp, address as addressRegexp } from '@console/lib/regexp'
import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'

const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object()
  .shape({
    ids: Yup.object().shape({
      device_id: Yup.string()
        .matches(deviceIdRegexp, sharedMessages.validateAlphanum)
        .min(2, Yup.passValues(sharedMessages.validateTooShort))
        .max(36, Yup.passValues(sharedMessages.validateTooLong))
        .required(sharedMessages.validateRequired),
    }),
    name: Yup.string()
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(50, Yup.passValues(sharedMessages.validateTooLong)),
    description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
    network_server_address: Yup.string().matches(
      addressRegexp,
      Yup.passValues(sharedMessages.validateAddressFormat),
    ),
    application_server_address: Yup.string().matches(
      addressRegexp,
      Yup.passValues(sharedMessages.validateAddressFormat),
    ),
    _external_js: Yup.boolean(),
    join_server_address: Yup.string().when(['$supportsJoin'], (supportsJoin, schema) => {
      if (!supportsJoin) {
        return schema.strip()
      }

      return schema
        .matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
        .default('')
    }),
    resets_join_nonces: Yup.bool().when(
      ['$supportsJoin', '$lorawanVersion', '_external_js'],
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
      ['_external_js', '$lorawanVersion', '$supportsJoin'],
      (externalJs, version, supportsJoin, schema) => {
        if (!supportsJoin) {
          return schema.strip()
        }

        const keySchema = Yup.lazy(() =>
          !externalJs
            ? Yup.object().shape({
                key: Yup.string()
                  .emptyOrLength(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                  .transform(toUndefined)
                  .default(generate16BytesKey),
              })
            : Yup.object().strip(),
        )

        if (externalJs) {
          return schema.shape({
            nwk_key: Yup.object().strip(),
            app_key: Yup.object().strip(),
          })
        }

        if (parseLorawanMacVersion(version) < 110) {
          return schema.shape({
            nwk_key: Yup.object().strip(),
            app_key: keySchema,
          })
        }

        return schema.shape({
          nwk_key: keySchema,
          app_key: keySchema,
        })
      },
    ),
    attributes: Yup.array()
      .max(10, Yup.passValues(sharedMessages.attributesValidateTooMany))
      .test(
        'has no empty string values',
        sharedMessages.attributesValidateRequired,
        attributeValidCheck,
      )
      .test(
        'has key length longer than 2',
        sharedMessages.attributeKeyValidateTooShort,
        attributeTooShortCheck,
      ),
  })
  .noUnknown()

export default validationSchema
