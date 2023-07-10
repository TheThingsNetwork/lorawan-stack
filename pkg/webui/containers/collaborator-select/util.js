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

export const composeOption = value => ({
  value:
    'user_ids' in value?.ids
      ? value.ids?.user_ids.user_id
      : value.ids?.organization_ids.organization_id,
  label:
    'user_ids' in value?.ids
      ? value.ids?.user_ids.user_id
      : value.ids?.organization_ids.organization_id,
  icon: 'user_ids' in value?.ids ? 'user' : 'organization',
})

export const composeContactOption = value => ({
  value: 'user_ids' in value ? value.user_ids.user_id : value.organization_ids.organization_id,
  label: 'user_ids' in value ? value.user_ids.user_id : value.organization_ids.organization_id,
  icon: 'user_ids' in value ? 'user' : 'organization',
})

export const encodeContact = value =>
  value
    ? {
        [`${value.icon}_ids`]: {
          [`${value.icon}_id`]: value.value,
        },
      }
    : null

export const decodeContact = value => (value ? composeContactOption(value) : null)

export const organizationSchema = Yup.object().shape({
  organization_id: Yup.string().matches(contactIdRegex, sharedMessages.validateAlphanum),
})

export const userSchema = Yup.object().shape({
  user_id: Yup.string().matches(contactIdRegex, sharedMessages.validateAlphanum),
})
