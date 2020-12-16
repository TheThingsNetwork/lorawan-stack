// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { hasSpecial, hasUpper, hasDigit, hasMinLength, hasMaxLength } from '@ttn-lw/lib/password'

export default requirements => {
  const passwordValidation = Yup.string()
    .default('')
    .required(sharedMessages.validateRequired)
    .test(
      'min-length',
      { message: sharedMessages.validateTooShort, values: { min: requirements.min_length } },
      password => hasMinLength(password, requirements.min_length),
    )
    .test(
      'max-length',
      { message: sharedMessages.validateTooLong, values: { max: requirements.max_length } },
      password => hasMaxLength(password, requirements.max_length),
    )
    .test(
      'min-special',
      { message: sharedMessages.validateSpecial, values: { special: requirements.min_special } },
      password => hasSpecial(password, requirements.min_special),
    )
    .test(
      'min-upper',
      { message: sharedMessages.validateUppercase, values: { upper: requirements.min_uppercase } },
      password => hasUpper(password, requirements.min_uppercase),
    )
    .test(
      'min-digit',
      { message: sharedMessages.validateDigit, values: { digit: requirements.min_digits } },
      password => hasDigit(password, requirements.min_digits),
    )

  return Yup.object().shape({
    password: passwordValidation,
    confirmPassword: Yup.string()
      .default('')
      .required(sharedMessages.validateRequired)
      .oneOf([Yup.ref('password'), null], sharedMessages.validatePasswordMatch),
  })
}
