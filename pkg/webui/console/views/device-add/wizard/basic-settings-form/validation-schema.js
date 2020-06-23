// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { id as deviceIdRegexp } from '@ttn-lw/lib/regexp'

import { ACTIVATION_MODES, parseLorawanMacVersion } from '@console/lib/device-utils'

const deviceIdSchema = Yup.string()
  .matches(deviceIdRegexp, sharedMessages.validateIdFormat)
  .min(2, Yup.passValues(sharedMessages.validateTooShort))
  .max(36, Yup.passValues(sharedMessages.validateTooLong))
  .required(sharedMessages.validateRequired)

const joinEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength)) // 8 Byte hex.
const devEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength)) // 8 Byte hex.

const validationSchema = Yup.object()
  .shape({
    ids: Yup.object().when(
      ['$activationMode', '$lorawanVersion'],
      (activationMode, version, schema) => {
        if (activationMode === ACTIVATION_MODES.OTAA) {
          return schema.shape({
            device_id: deviceIdSchema,
            join_eui: joinEUISchema.required(sharedMessages.validateRequired),
            dev_eui: devEUISchema.required(sharedMessages.validateRequired),
          })
        }

        if (
          activationMode === ACTIVATION_MODES.ABP ||
          activationMode === ACTIVATION_MODES.MULTICAST
        ) {
          if (parseLorawanMacVersion(version) === 104) {
            return schema.shape({
              device_id: deviceIdSchema,
              dev_eui: devEUISchema.required(sharedMessages.validateRequired),
            })
          }

          return schema.shape({
            device_id: deviceIdSchema,
            dev_eui: Yup.lazy(
              value =>
                !value
                  ? Yup.string().strip()
                  : Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength)), // 8 Byte hex.
            ),
          })
        }

        return schema.shape({
          device_id: deviceIdSchema,
        })
      },
    ),
    name: Yup.string()
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(50, Yup.passValues(sharedMessages.validateTooLong)),
    description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
  })
  .noUnknown()

export default validationSchema
