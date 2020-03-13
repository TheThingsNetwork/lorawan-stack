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

import sharedMessages from '../../../../lib/shared-messages'

import m from '../../../components/device-data-form/messages'

import { parseLorawanMacVersion, ACTIVATION_MODES } from '../utils'

const validationSchema = Yup.object()
  .shape({
    _external_js: Yup.boolean(),
    _activation_mode: Yup.mixed().oneOf([
      ACTIVATION_MODES.ABP,
      ACTIVATION_MODES.OTAA,
      ACTIVATION_MODES.MULTICAST,
    ]),
    _may_edit_keys: Yup.boolean().default(false),
    _may_read_keys: Yup.boolean().default(false),
    _joined: Yup.boolean().default(false),
    lorawan_version: Yup.string().required(sharedMessages.validateRequired),
    lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
    frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
    supports_class_c: Yup.boolean().default(false),
    session: Yup.object().when(
      ['_activation_mode', 'lorawan_version', '_joined', '_may_edit_keys', '_may_read_keys'],
      (mode, version, isJoined, mayEditKeys, mayReadKeys, schema) => {
        if (mode === ACTIVATION_MODES.ABP || mode === ACTIVATION_MODES.MULTICAST || isJoined) {
          const isNewVersion = parseLorawanMacVersion(version) >= 110
          return schema.shape({
            dev_addr: Yup.lazy(() => {
              const schema = Yup.string().length(4 * 2, m.validate8) // 4 Byte hex

              if (mayReadKeys && mayEditKeys) {
                // Force the field to be required only if the user can see and edit the `dev_addr`,
                // otherwise the user is not able to edit any other fields in the NS form without
                // resetting the `dev_addr`.
                return schema.required(sharedMessages.validateRequired)
              }

              return schema
            }),
            keys: Yup.object().shape({
              f_nwk_s_int_key: Yup.lazy(value =>
                Boolean(value) && Boolean(value.key)
                  ? Yup.object().shape({
                      key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
                    })
                  : Yup.object().strip(),
              ),
              s_nwk_s_int_key: Yup.lazy(value =>
                isNewVersion && Boolean(value) && Boolean(value.key)
                  ? Yup.object().shape({
                      key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
                    })
                  : Yup.object().strip(),
              ),
              nwk_s_enc_key: Yup.lazy(value =>
                isNewVersion && Boolean(value) && Boolean(value.key)
                  ? Yup.object().shape({
                      key: Yup.string().length(16 * 2, m.validate32), // 16 Byte hex
                    })
                  : Yup.object().strip(),
              ),
            }),
          })
        }
        return schema.strip()
      },
    ),
    mac_settings: Yup.object().when(['_activation_mode'], (mode, schema) => {
      if (mode === ACTIVATION_MODES.ABP) {
        return schema.shape({
          resets_f_cnt: Yup.boolean(),
        })
      }

      return schema.strip()
    }),
  })
  .noUnknown()

export default validationSchema
