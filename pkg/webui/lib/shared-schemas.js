// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { userId as contactIdRegex } from '@ttn-lw/lib/regexp'

const organizationSchema = Yup.object().shape({
  organization_id: Yup.string().matches(contactIdRegex, sharedMessages.validateAlphanum),
})

const userSchema = Yup.object().shape({
  user_id: Yup.string().matches(contactIdRegex, sharedMessages.validateAlphanum),
})

export const contactSchema = Yup.object().shape({
  administrative_contact: Yup.object()
    .when(['organization_ids'], {
      is: organizationIds => Boolean(organizationIds),
      then: schema => schema.concat(organizationSchema),
      otherwise: schema => schema.concat(userSchema),
    })
    .nullable()
    .required(sharedMessages.validateRequired),
  technical_contact: Yup.object()
    .when(['organization_ids'], {
      is: organizationIds => Boolean(organizationIds),
      then: schema => schema.concat(organizationSchema),
      otherwise: schema => schema.concat(userSchema),
    })
    .nullable()
    .required(sharedMessages.validateRequired),
})

export default contactSchema
