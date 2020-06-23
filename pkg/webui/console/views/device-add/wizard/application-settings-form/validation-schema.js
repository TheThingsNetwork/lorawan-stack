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

import * as Yup from 'yup'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const validationSchema = Yup.object()
  .shape({
    skip_payload_crypto: Yup.boolean().default(false),
    session: Yup.object().when(
      ['skip_payload_crypto', '$mayEditKeys'],
      (skipPayloadCrypto, mayEditKeys, schema) => {
        if (skipPayloadCrypto || !mayEditKeys) {
          return schema.strip()
        }

        return schema.shape({
          keys: Yup.object().shape({
            app_s_key: Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // A 16 Byte hex.
                .required(sharedMessages.validateRequired),
            }),
          }),
        })
      },
    ),
  })
  .noUnknown()

export default validationSchema
