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

import { defineMessages } from 'react-intl'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  validateCode: 'Claim authentication code must consist only of numbers and letters',
})

const validationSchema = Yup.object({
  ids: Yup.object().shape({
    dev_eui: Yup.string()
      .length(8 * 2, Yup.passValues(sharedMessages.validateLength))
      .required(sharedMessages.validateRequired),
  }),
  authentication_code: Yup.string().when(['_claim'], {
    is: 'true',
    then: schema =>
      schema
        .matches(/^[A-Z0-9]{1,32}$/, Yup.passValues(m.validateCode))
        .required(sharedMessages.validateRequired),
    otherwise: schema => schema.strip(),
  }),
})

export default validationSchema
