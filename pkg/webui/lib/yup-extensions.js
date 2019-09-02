// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

const StringSchema = Yup.string

/**
 * `NullableStringSchemaType` is an extension for the default `yup.string` schema type.
 * It transforms the value to `null` if it is empty and skips validation.
 */
class NullableStringSchemaType extends StringSchema {
  constructor() {
    super()

    const self = this

    self.withMutation(function() {
      self
        .transform(function(value) {
          if (self.isType(value) && Boolean(value)) {
            return value
          }

          return null
        })
        .nullable(true)
    })
  }
}

Yup.nullableString = () => new NullableStringSchemaType()

Yup.addMethod(Yup.string, 'emptyOrLength', function(exactLength, message) {
  // eslint-disable-next-line no-invalid-this
  return this.test(
    'empty-or-length',
    message,
    value => !Boolean(value) || value.length === exactLength,
  )
})

export default Yup
