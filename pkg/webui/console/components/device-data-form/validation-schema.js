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

import sharedMessages from '../../../lib/shared-messages'
import { id as deviceIdRegexp, address as addressRegexp } from '../../lib/regexp'
import randomByteString from '../../lib/random-bytes'
import m from './messages'

const random16BytesString = () => randomByteString(32)
const isABP = mode => mode === 'abp'
const isOTAA = mode => mode === 'otaa'
const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object({
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string().max(2000, sharedMessages.validateTooLong),
  lorawan_version: Yup.string().required(sharedMessages.validateRequired),
  lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  supports_class_c: Yup.boolean(),
  network_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
  application_server_address: Yup.string().matches(
    addressRegexp,
    sharedMessages.validateAddressFormat,
  ),
  join_server_address: Yup.string().matches(addressRegexp, sharedMessages.validateAddressFormat),
  activation_mode: Yup.string().required(),
  supports_join_nonces: Yup.boolean(),
}) // OTAA related entries
  .shape({
    ids: Yup.object()
      .shape({
        device_id: Yup.string()
          .matches(deviceIdRegexp, sharedMessages.validateAlphanum)
          .min(2, sharedMessages.validateTooShort)
          .max(36, sharedMessages.validateTooLong)
          .required(sharedMessages.validateRequired),
      })
      .when('activation_mode', {
        is: isOTAA,
        then: schema =>
          schema.shape({
            join_eui: Yup.string()
              .length(8 * 2, m.validate16) // 8 Byte hex
              .required(sharedMessages.validateRequired),
            dev_eui: Yup.string()
              .length(8 * 2, m.validate16) // 8 Byte hex
              .required(sharedMessages.validateRequired),
          }),
        otherwise: schema =>
          schema.shape({
            join_eui: Yup.string().strip(),
            dev_eui: Yup.string().strip(),
          }),
      }),
    mac_settings: Yup.object().when('activation_mode', {
      is: isOTAA,
      then: schema =>
        schema.shape({
          resets_f_cnt: Yup.boolean(),
        }),
      otherwise: schema => schema.strip(),
    }),
    root_keys: Yup.object().when('activation_mode', {
      is: isOTAA,
      then: schema =>
        schema.shape({
          nwk_key: Yup.object().shape({
            key: Yup.string()
              .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
              .transform(toUndefined)
              .default(random16BytesString),
          }),
          app_key: Yup.object().shape({
            key: Yup.string()
              .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
              .transform(toUndefined)
              .default(random16BytesString),
          }),
        }),
      otherwise: schema => schema.strip(),
    }),
  }) // ABP related entries
  .shape({
    resets_join_nonces: Yup.boolean().when('activation_mode', {
      is: isOTAA,
      then: schema => schema,
      otherwise: schema => schema.strip(),
    }),
    session: Yup.object().when(['activation_mode', 'lorawan_version'], (mode, version, schema) => {
      if (isABP(mode)) {
        // Check if the version is 1.1.x or higher
        const isNewVersion =
          Boolean(version) && parseInt(version.replace(/\D/g, '').padEnd(3, 0)) >= 110

        return schema.shape({
          dev_addr: Yup.string()
            .length(4 * 2, m.validate8) // 4 Byte hex
            .required(sharedMessages.validateRequired),
          keys: Yup.object().shape({
            f_nwk_s_int_key: Yup.object().shape({
              key: Yup.string()
                .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
                .transform(toUndefined)
                .default(random16BytesString),
            }),
            s_nwk_s_int_key: Yup.lazy(() =>
              isNewVersion
                ? Yup.object().shape({
                    key: Yup.string()
                      .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
                      .transform(toUndefined)
                      .default(random16BytesString),
                  })
                : Yup.object().strip(),
            ),
            nwk_s_enc_key: Yup.lazy(() =>
              isNewVersion
                ? Yup.object().shape({
                    key: Yup.string()
                      .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
                      .transform(toUndefined)
                      .default(random16BytesString),
                  })
                : Yup.object().strip(),
            ),
            app_s_key: Yup.object().shape({
              key: Yup.string()
                .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
                .transform(toUndefined)
                .default(random16BytesString),
            }),
          }),
        })
      }

      return schema.strip()
    }),
  })

export default validationSchema
