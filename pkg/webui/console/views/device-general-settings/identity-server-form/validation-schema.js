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

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectJsConfig, selectNsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'

import { attributeValidCheck, attributeTooShortCheck } from '@console/lib/attributes'
import { id as deviceIdRegexp, address as addressRegexp } from '@console/lib/regexp'
import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'

const jsConfig = selectJsConfig()
const asConfig = selectAsConfig()
const nsConfig = selectNsConfig()

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
    network_server_address: Yup.string()
      .matches(addressRegexp, sharedMessages.validateAddressFormat)
      .when(['_default_addresses'], {
        is: true,
        then: schema =>
          schema
            .transform(value => (nsConfig.enabled ? undefined : value))
            .default(getHostnameFromUrl(nsConfig.base_url)),
      }),
    application_server_address: Yup.string()
      .matches(addressRegexp, sharedMessages.validateAddressFormat)
      .when(['_default_addresses'], {
        is: true,
        then: schema =>
          schema
            .transform(value => (asConfig.enabled ? undefined : value))
            .default(getHostnameFromUrl(asConfig.base_url)),
      }),
    _external_js: Yup.boolean(),
    _supports_join: Yup.boolean(),
    _lorawan_version: Yup.string(),
    _default_addresses: Yup.boolean(),
    join_server_address: Yup.string().when(
      ['_supports_join', ' _external_js', '_default_addresses'],
      (supportsJoin, externalJs, useDefaultAddresses, schema) => {
        if (!supportsJoin) {
          return schema.strip()
        }

        if (externalJs) {
          return schema.transform(() => '')
        }

        return schema
          .matches(addressRegexp, sharedMessages.validateAddressFormat)
          .transform(value => (useDefaultAddresses && jsConfig.enabled ? undefined : value))
          .default(getHostnameFromUrl(jsConfig.base_url))
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
                  .emptyOrLength(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                  .transform(toUndefined)
                  .default(generate16BytesKey),
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
    attributes: Yup.array()
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
