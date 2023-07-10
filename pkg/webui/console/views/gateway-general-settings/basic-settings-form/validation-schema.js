// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { id as gatewayIdRegexp, userId as contactIdRegex } from '@ttn-lw/lib/regexp'

import {
  attributeValidCheck,
  attributeTooShortCheck,
  attributeKeyTooLongCheck,
  attributeValueTooLongCheck,
  attributesCountCheck,
} from '@console/lib/attributes'
import { addressWithOptionalScheme as addressWithOptionalSchemeRegexp } from '@console/lib/regexp'

const organizationSchema = Yup.object().shape({
  organization_id: Yup.string().matches(contactIdRegex, sharedMessages.validateAlphanum),
})

const userSchema = Yup.object().shape({
  user_id: Yup.string().matches(contactIdRegex, sharedMessages.validateAlphanum),
})

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    gateway_id: Yup.string()
      .matches(gatewayIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .required(sharedMessages.validateRequired),
    eui: Yup.nullableString().length(8 * 2, Yup.passValues(sharedMessages.validateLength)),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  update_channel: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
  gateway_server_address: Yup.string().matches(
    addressWithOptionalSchemeRegexp,
    Yup.passValues(sharedMessages.validateAddressFormat),
  ),
  require_authenticated_connection: Yup.boolean().default(false),
  // The API allows 2048 bytes. But since we convert to Base64 we need an additional 33% (at max) capacity. So 66% of 2048 = 1351,68 and hence this is set to 1350.
  lbs_lns_secret: Yup.lazy(secret => {
    if (!secret) {
      return Yup.object().strip()
    }

    return Yup.object({
      value: Yup.string().max(1350, Yup.passValues(sharedMessages.validateTooLong)),
    })
  }),
  location_public: Yup.boolean().default(false),
  status_public: Yup.boolean().default(false),
  update_location_from_status: Yup.boolean().default(false),
  auto_update: Yup.boolean().default(false),
  disable_packet_broker_forwarding: Yup.boolean().default(false),
  attributes: Yup.object()
    .nullable()
    .test(
      'has no more than 10 keys',
      sharedMessages.attributesValidateTooMany,
      attributesCountCheck,
    )
    .test('has no null values', sharedMessages.attributesValidateRequired, attributeValidCheck)
    .test(
      'has key length longer than 2',
      sharedMessages.attributeKeyValidateTooShort,
      attributeTooShortCheck,
    )
    .test(
      'has key length less than 36',
      sharedMessages.attributeKeyValidateTooLong,
      attributeKeyTooLongCheck,
    )
    .test(
      'has value length less than 200',
      sharedMessages.attributeValueValidateTooLong,
      attributeValueTooLongCheck,
    ),
  administrative_contact: Yup.object().when(['organization_ids'], {
    is: organizationIds => Boolean(organizationIds),
    then: schema => schema.concat(organizationSchema),
    otherwise: schema => schema.concat(userSchema),
  }),
  technical_contact: Yup.object().when(['organization_ids'], {
    is: organizationIds => Boolean(organizationIds),
    then: schema => schema.concat(organizationSchema),
    otherwise: schema => schema.concat(userSchema),
  }),
})

export default validationSchema
