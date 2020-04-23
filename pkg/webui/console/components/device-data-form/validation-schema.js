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

import { id as deviceIdRegexp, address as addressRegexp } from '@console/lib/regexp'

const isABP = mode => mode === 'abp'
const isOTAA = mode => mode === 'otaa'

const validationSchema = Yup.object({
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
  lorawan_version: Yup.string().required(sharedMessages.validateRequired),
  lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  supports_class_c: Yup.boolean(),
  network_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
  application_server_address: Yup.string().matches(
    addressRegexp,
    sharedMessages.validateAddressFormat,
  ),
  join_server_address: Yup.string().when('_external_js', {
    is: false,
    then: schema => schema.matches(addressRegexp, sharedMessages.validateAddressFormat),
    otherwise: schema => schema.default(''),
  }),
  _activation_mode: Yup.string().required(),
  supports_join_nonces: Yup.boolean(),
  ids: Yup.object()
    .shape({
      device_id: Yup.string()
        .matches(deviceIdRegexp, sharedMessages.validateIdFormat)
        .min(2, Yup.passValues(sharedMessages.validateTooShort))
        .max(36, Yup.passValues(sharedMessages.validateTooLong))
        .required(sharedMessages.validateRequired),
    })
    .when(['_activation_mode', 'lorawan_version'], (mode, version, schema) => {
      const isLw104 = Boolean(version) && parseInt(version.replace(/\D/g, '').padEnd(3, 0)) === 104
      const isModeOTAA = isOTAA(mode)

      return schema.shape({
        join_eui: Yup.lazy(() =>
          isModeOTAA
            ? Yup.string()
                .length(8 * 2, Yup.passValues(sharedMessages.validateLength)) // 8 Byte hex.
                .required(sharedMessages.validateRequired)
            : Yup.string().strip(),
        ),
        dev_eui: Yup.lazy(
          () =>
            isModeOTAA || isLw104
              ? Yup.string()
                  .length(8 * 2, Yup.passValues(sharedMessages.validateLength)) // 8 Byte hex.
                  .required(sharedMessages.validateRequired)
              : Yup.nullableString().emptyOrLength(
                  8 * 2,
                  Yup.passValues(sharedMessages.validateLength),
                ), // 8 Byte hex.
        ),
      })
    }),
}) // OTAA related entries.
  .shape({
    mac_settings: Yup.object().when('_activation_mode', {
      is: isABP,
      then: schema =>
        schema.shape({
          resets_f_cnt: Yup.boolean(),
        }),
      otherwise: schema => schema.strip(),
    }),
    _external_js: Yup.boolean().default(true),
    _may_edit_keys: Yup.boolean().default(false),
    supports_join: Yup.boolean().when('_activation_mode', {
      is: 'otaa',
      then: schema => schema.default(true),
      otherwise: schema => schema.default(false),
    }),
    root_keys: Yup.object().when(
      ['_activation_mode', '_external_js', '_may_edit_keys'],
      (mode, externalJs, mayEditKeys, schema) => {
        if (isOTAA(mode) && !externalJs && mayEditKeys) {
          return schema.shape({
            nwk_key: Yup.lazy(value =>
              Boolean(value) && Boolean(value.key)
                ? Yup.object().shape({
                    key: Yup.string().emptyOrLength(
                      16 * 2,
                      Yup.passValues(sharedMessages.validateLength),
                    ), // 16 Byte hex.
                  })
                : Yup.object().strip(),
            ),
            app_key: Yup.lazy(value =>
              Boolean(value) && Boolean(value.key)
                ? Yup.object().shape({
                    key: Yup.string().emptyOrLength(
                      16 * 2,
                      Yup.passValues(sharedMessages.validateLength),
                    ), // 16 Byte hex.
                  })
                : Yup.object().strip(),
            ),
          })
        }

        return schema.strip()
      },
    ),
    net_id: Yup.nullableString().when('_external_js', {
      is: false,
      then: schema =>
        schema
          .emptyOrLength(3 * 2, Yup.passValues(sharedMessages.validateLength)) // 3 Byte hex.
          .default(''),
      otherwise: schema => schema.strip(),
    }),
    application_server_id: Yup.string().when('_external_js', {
      is: false,
      then: schema => schema.max(100, Yup.passValues(sharedMessages.validateTooLong)).default(''),
      otherwise: schema => schema.strip(),
    }),
    application_server_kek_label: Yup.string().when('_external_js', {
      is: false,
      then: schema => schema.max(2048, Yup.passValues(sharedMessages.validateTooLong)).default(''),
      otherwise: schema => schema.strip(),
    }),
    network_server_kek_label: Yup.string().when('_external_js', {
      is: false,
      then: schema => schema.max(2048, Yup.passValues(sharedMessages.validateTooLong)).default(''),
      otherwise: schema => schema.strip(),
    }),
  }) // ABP related entries.
  .shape({
    resets_join_nonces: Yup.boolean().when(['_activation_mode', '_external_js'], {
      is: (mode, externalJs) => isOTAA(mode) && !externalJs,
      then: schema => schema,
      otherwise: schema => schema.strip(),
    }),
    session: Yup.object().when(['_activation_mode', 'lorawan_version'], (mode, version, schema) => {
      if (isABP(mode)) {
        // Check if the version is 1.1.x or higher.
        const isNewVersion =
          Boolean(version) && parseInt(version.replace(/\D/g, '').padEnd(3, 0)) >= 110

        return schema.shape({
          dev_addr: Yup.string()
            .length(4 * 2, Yup.passValues(sharedMessages.validateLength)) // 4 Byte hex.
            .required(sharedMessages.validateRequired),
          keys: Yup.object().shape({
            f_nwk_s_int_key: Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                .required(sharedMessages.validateRequired),
            }),
            app_s_key: Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                .required(sharedMessages.validateRequired),
            }),
            s_nwk_s_int_key: Yup.lazy(() =>
              isNewVersion
                ? Yup.object().shape({
                    key: Yup.string()
                      .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                      .required(sharedMessages.validateRequired),
                  })
                : Yup.object().strip(),
            ),
            nwk_s_enc_key: Yup.lazy(() =>
              isNewVersion
                ? Yup.object().shape({
                    key: Yup.string()
                      .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // 16 Byte hex.
                      .required(sharedMessages.validateRequired),
                  })
                : Yup.object().strip(),
            ),
          }),
        })
      }

      return schema.strip()
    }),
  })

export default validationSchema
