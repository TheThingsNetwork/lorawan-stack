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

export const getTechnicalContact = values =>
  values.technical_contact !== undefined && values.technical_contact !== null
    ? {
        _technical_contact_id: values.technical_contact.user_ids
          ? values.technical_contact.user_ids.user_id
          : values.technical_contact.organization_ids.organization_id,
        _technical_contact_type: values.technical_contact.user_ids ? 'user' : 'organization',
      }
    : {
        _technical_contact_id: '',
        _technical_contact_type: '',
      }

export const getAdministrativeContact = values =>
  values.administrative_contact !== undefined && values.administrative_contact !== null
    ? {
        _administrative_contact_id: values.administrative_contact.user_ids
          ? values.administrative_contact.user_ids.user_id
          : values.administrative_contact.organization_ids.organization_id,
        _administrative_contact_type: values.administrative_contact.user_ids
          ? 'user'
          : 'organization',
      }
    : {
        _administrative_contact_id: '',
        _administrative_contact_type: '',
      }

export const composeContact = (type, id) => ({
  [`${type}_ids`]: {
    [`${id}_id`]: id,
  },
})
