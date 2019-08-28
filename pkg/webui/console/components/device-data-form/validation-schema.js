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
import m from './messages'

const baseSchemaShape = Yup.object({
  ids: Yup.object()
    .shape({
      device_id: Yup.string()
        .matches(deviceIdRegexp, sharedMessages.validateAlphanum)
        .min(2, sharedMessages.validateTooShort)
        .max(36, sharedMessages.validateTooLong)
        .required(sharedMessages.validateRequired),
    })
    .when('activation_mode', {
      is: 'otaa',
      then: Yup.object().shape({
        join_eui: Yup.string()
          .length(8 * 2, m.validate16)
          .required(sharedMessages.validateRequired), // 8 Byte hex
        dev_eui: Yup.string()
          .length(8 * 2, m.validate16)
          .required(sharedMessages.validateRequired), // 8 Byte hex
      }),
    }),
  mac_settings: Yup.object().shape({
    resets_f_cnt: Yup.boolean(),
  }),
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
})

export const updateFormValidationSchema = baseSchemaShape.shape({
  session: Yup.object().shape({
    dev_addr: Yup.string().length(4 * 2, m.validate8), // 4 Byte hex
    keys: Yup.object().shape({
      f_nwk_s_int_key: Yup.object().shape({
        key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
      }),
      s_nwk_s_int_key: Yup.object().shape({
        key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
      }),
      nwk_s_enc_key: Yup.object().shape({
        key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
      }),
      app_s_key: Yup.object().shape({
        key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
      }),
    }),
  }),
  root_keys: Yup.object().shape({
    nwk_key: Yup.object().shape({
      key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
    }),
    app_key: Yup.object().shape({
      key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
    }),
  }),
})

const stripIfEmpty = (originalSchema, testEmpty = () => false) =>
  function(value) {
    if (!Boolean(value) || testEmpty(value)) {
      return Yup.string().strip()
    }

    return originalSchema
  }

// See discussion in https://github.com/TheThingsNetwork/lorawan-stack/pull/1075 for reasons
// to strip the key values.
const stripKeyIfEmpty = originalSchema => stripIfEmpty(originalSchema, value => !Boolean(value.key))

export const createFormValidationSchema = baseSchemaShape.shape({
  session: Yup.object().shape({
    dev_addr: Yup.lazy(
      stripIfEmpty(
        Yup.string().length(4 * 2, m.validate8), // 4 Byte hex
      ),
    ),
    keys: Yup.object().shape({
      f_nwk_s_int_key: Yup.lazy(
        stripKeyIfEmpty(
          Yup.object().shape({
            key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
          }),
        ),
      ),
      s_nwk_s_int_key: Yup.lazy(
        stripKeyIfEmpty(
          Yup.object().shape({
            key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
          }),
        ),
      ),
      nwk_s_enc_key: Yup.lazy(
        stripKeyIfEmpty(
          Yup.object().shape({
            key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
          }),
        ),
      ),
      app_s_key: Yup.lazy(
        stripKeyIfEmpty(
          Yup.object().shape({
            key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
          }),
        ),
      ),
    }),
  }),
  root_keys: Yup.object().shape({
    nwk_key: Yup.lazy(
      stripKeyIfEmpty(
        Yup.object().shape({
          key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
        }),
      ),
    ),
    app_key: Yup.lazy(
      stripKeyIfEmpty(
        Yup.object().shape({
          key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
        }),
      ),
    ),
  }),
})
