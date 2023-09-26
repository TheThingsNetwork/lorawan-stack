// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { id as gatewayIdRegexp } from '@ttn-lw/lib/regexp'
import { selectGsConfig } from '@ttn-lw/lib/selectors/env'

const gsEnabled = selectGsConfig().enabled

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    gateway_id: Yup.string()
      .min(3, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .matches(gatewayIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
      .required(sharedMessages.validateRequired),
    eui: Yup.string()
      .test(
        'has 16 or 12 characters',
        Yup.passValues(sharedMessages.validateLength)({ length: 16 }),
        value => value && (value.length === 12 || value.length === 16),
      )
      .test(
        "doesn't have 12 characters",
        Yup.passValues(sharedMessages.validateMacAddressEntered),
        value => value && value.length !== 12,
      ),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  require_authenticated_connection: Yup.boolean(),
  location_public: Yup.boolean(),
  status_public: Yup.boolean(),
  frequency_plan_id: gsEnabled
    ? Yup.string()
        .max(64, Yup.passValues(sharedMessages.validateTooLong))
        .required(sharedMessages.validateRequired)
    : Yup.string(),
})

export default validationSchema
