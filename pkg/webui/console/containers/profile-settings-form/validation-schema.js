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

import m from './messages'

export default Yup.object().shape({
  _profile_picture_source: Yup.string()
    .oneOf(['gravatar', 'upload'])
    .when('$initialProfilePictureSource', ([ppSource], schema) => schema.default(ppSource)),
  profile_picture: Yup.object()
    .nullable()
    .when(
      ['_profile_picture_source', '$useGravatarConfig', '$disableUploadConfig'],
      ([ppSource, useGravatarConfig, uploadDisabled], schema) => {
        if (!useGravatarConfig && uploadDisabled) {
          return schema.strip()
        }
        if (ppSource === 'upload' && useGravatarConfig) {
          return schema.required(m.imageRequired)
        }

        if (ppSource === 'gravatar') {
          // To use gravatar, the profile picture value has to be `null`.
          return schema.transform(() => null)
        }

        return schema
      },
    ),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  primary_email_address: Yup.string()
    .email(sharedMessages.validateEmail)
    .required(sharedMessages.validateRequired),
})
