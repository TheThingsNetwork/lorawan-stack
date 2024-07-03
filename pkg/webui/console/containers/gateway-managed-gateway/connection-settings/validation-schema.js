// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import {
  ethernetValidationSchema,
  wifiValidationSchema,
} from '@console/containers/gateway-managed-gateway/shared/validation-schema'
import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/shared/utils'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

export const validationSchema = Yup.object().shape({
  settings: Yup.array().of(
    Yup.object()
      .shape({
        _connection_type: Yup.string()
          .oneOf(Object.values(CONNECTION_TYPES))
          .default(CONNECTION_TYPES.WIFI),
      })
      .when('._connection_type', {
        is: CONNECTION_TYPES.WIFI,
        then: schema =>
          schema
            .concat(
              Yup.object().shape({
                profile: Yup.string().required(sharedMessages.validateRequired),
              }),
            )
            .concat(wifiValidationSchema),
        otherwise: schema => schema.concat(ethernetValidationSchema),
      }),
  ),
})

export default validationSchema
