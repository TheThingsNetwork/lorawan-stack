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

import { IconOrganization, IconUser } from '@ttn-lw/components/icon'

export const composeOption = value => {
  const data = value.ids || value
  return {
    value: 'user_ids' in data ? data.user_ids?.user_id : data.organization_ids?.organization_id,
    label: 'user_ids' in data ? data.user_ids?.user_id : data.organization_ids?.organization_id,
    icon: 'user_ids' in data ? IconUser : IconOrganization,
  }
}

export const encodeContact = value =>
  value
    ? {
        [`${value.icon}_ids`]: {
          [`${value.icon}_id`]: value.value,
        },
      }
    : null

export const decodeContact = value => (value ? composeOption(value) : null)
