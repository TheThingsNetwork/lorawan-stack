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

import '@ttn-lw/lib/yup-extensions'

import validationSchema from './validation-schema'

describe('<ApplicationSettingsForm /> validation schema', () => {
  const validateWithKeys = schema =>
    validationSchema.validateSync(schema, { context: { mayEditKeys: true } })
  const validateWithoutKeys = schema =>
    validationSchema.validateSync(schema, { context: { mayEditKeys: false } })

  describe('`skip_payload_crypto` is not set', () => {
    const schema = { skip_payload_crypto: false }

    it('should require `app_s_key`', done => {
      try {
        validateWithKeys(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('session.keys.app_s_key.key')
      }

      const appSKey = '1'.repeat(32)
      schema.session = {
        keys: {
          app_s_key: {
            key: appSKey,
          },
        },
      }

      const validatedValue = validateWithKeys(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.session).toBeDefined()

      const { session } = validatedValue
      expect(session.keys).toBeDefined()

      const { keys } = session
      expect(keys.app_s_key).toBeDefined()
      expect(keys.app_s_key.key).toBe(appSKey)
      done()
    })
  })

  describe('`skip_payload_crypto` is set', () => {
    const schema = { skip_payload_crypto: true }

    it('should strip `app_s_key`', () => {
      const validatedValue = validateWithKeys(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.session).toBeUndefined()
    })
  })

  describe('cannot edit keys', () => {
    it('should strip `app_s_key`', () => {
      const schema = {
        keys: {
          app_s_key: {
            key: '1'.repeat(32),
          },
        },
      }

      const validatedValue = validateWithoutKeys(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.session).toBeUndefined()
    })
  })
})
