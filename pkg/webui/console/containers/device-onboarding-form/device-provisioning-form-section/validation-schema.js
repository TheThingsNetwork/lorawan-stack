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

import claimValidationSchema from './device-claiming-form-section/validation-schema'
import registrationValidationSchema from './device-registration-form-section/validation-schema'

const joinEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength))
// Validation schema of the provisioning form section.
// Please observe the following rules to keep the validation schemas maintainable:
// 1. DO NOT USE ANY TYPE CONVERSIONS HERE. Use decocer/encoder on field level instead.
//    Consider all values as backend values. Exceptions may apply in consideration.
// 2. Comment each individual validation prop and use whitespace to structure visually.
// 3. Do not use ternary assignments but use plain if statements to ensure clarity.
const validationSchema = Yup.object({
  ids: Yup.object().when(['supports_join'], (isOTAA, schema) => {
    if (isOTAA) {
      return schema.shape({
        join_eui: joinEUISchema.required(sharedMessages.validateRequired),
      })
    }
  }),
  supports_join: Yup.bool().default(false),
}).when(['_claim'], {
  is: true,
  then: schema => schema.concat(claimValidationSchema),
  otherwise: schema => schema.concat(registrationValidationSchema),
})

export default validationSchema
