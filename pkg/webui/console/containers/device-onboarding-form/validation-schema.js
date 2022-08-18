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

import { REGISTRATION_TYPES } from './utils'
import claimValidationSchema from './device-provisioning-form-section/device-claiming-form-section/validation-schema'
import registrationValidationSchema from './device-provisioning-form-section/device-registration-form-section/validation-schema'
import repositoryValidationSchema from './device-type-form-section/device-type-repository-form-section/validation-schema'
import manualValidationSchema from './device-type-form-section/device-type-manual-form-section/validation-schema'

const validationSchema = Yup.object({
  _registration: Yup.mixed().oneOf([REGISTRATION_TYPES.SINGLE, REGISTRATION_TYPES.MULTIPLE]),
})
  .when('._claim', {
    is: true,
    then: schema => schema.concat(claimValidationSchema),
    otherwise: schema => schema.concat(registrationValidationSchema),
  })
  .when('._inputMethod', {
    is: 'device-repository',
    then: schema => schema.concat(repositoryValidationSchema),
    otherwise: schema => schema.concat(manualValidationSchema),
  })

export default validationSchema
