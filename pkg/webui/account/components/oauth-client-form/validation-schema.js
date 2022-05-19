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

import { approvalStates } from './utils'

const validationSchema = Yup.object().shape({
  owner_id: Yup.string().required(sharedMessages.validateRequired),
  ids: Yup.object().shape({
    client_id: Yup.string()
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .required(sharedMessages.validateRequired),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(2000, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string(),
  redirect_uris: Yup.array().max(10, Yup.passValues(sharedMessages.attributesValidateTooMany)),
  logout_redirect_uris: Yup.array().max(
    10,
    Yup.passValues(sharedMessages.attributesValidateTooMany),
  ),
  skip_authorization: Yup.bool(),
  endorsed: Yup.bool(),
  grants: Yup.array().max(3, Yup.passValues(sharedMessages.attributesValidateTooMany)),
  state: Yup.lazy(value => {
    if (value === '') {
      return Yup.string().strip()
    }

    return Yup.string().oneOf(approvalStates, sharedMessages.validateRequired)
  }),
  state_description: Yup.lazy(value => {
    if (value === '') {
      return Yup.string().strip()
    }

    return Yup.string()
  }),
  rights: Yup.array().min(1, sharedMessages.validateRights),
})

export default validationSchema
