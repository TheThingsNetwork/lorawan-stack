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

import m from '../../../components/device-data-form/messages'

import { parseLorawanMacVersion, ACTIVATION_MODES } from '../utils'

const random16BytesString = () => randomByteString(32)
const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object().shape({
  lorawan_version: Yup.string().required(sharedMessages.validateRequired),
  lorawan_phy_version: Yup.string().required(sharedMessages.validateRequired),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  _activation_mode: Yup.string(),
  session: Yup.object().when(['_activation_mode', 'lorawan_version'], (mode, version, schema) => {
    if (mode === ACTIVATION_MODES.ABP || mode === ACTIVATION_MODES.MULTICAST) {
      const isNewVersion = parseLorawanMacVersion(version) >= 110
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
        }),
      })
    }
    return schema.strip()
  }),
  mac_settings: Yup.object().when(['_activation_mode'], (mode, schema) => {
    if (mode === 'abp') {
      return schema.shape({
        resets_f_cnt: Yup.boolean(),
      })
    }

    return schema.strip()
  }),
  root_keys: Yup.object().when(
    ['_external_js', 'lorawan_version', '_activation_mode'],
    (externalJs, version, mode, schema) => {
      if (mode === ACTIVATION_MODES.OTAA) {
        const strippedSchema = Yup.object().strip()
        const keySchema = Yup.lazy(() => {
          return !externalJs
            ? Yup.object().shape({
                key: Yup.string()
                  .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
                  .transform(toUndefined)
                  .default(random16BytesString),
              })
            : strippedSchema
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
      }

      return schema.strip()
    },
  ),
})

export default validationSchema
