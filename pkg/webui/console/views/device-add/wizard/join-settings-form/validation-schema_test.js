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

describe('<JoinSettingsForm /> validation schema', () => {
  let schema

  beforeEach(() => {
    schema = {
      root_keys: {},
    }
  })

  describe('cannot edit keys', () => {
    it('should strip `root_keys`', () => {
      const schema = {
        root_keys: { app_key: { key: '1'.repeat(32) } },
      }

      schema.root_keys.app_key = { key: '1'.repeat(32) }

      let validatedValue = validationSchema.validateSync(schema, {
        context: {
          lorawanVersion: '1.0.0',
          meyEditKeys: false,
        },
      })

      expect(validatedValue).toBeDefined()
      expect(validatedValue.root_keys).toBeUndefined()

      validatedValue = validationSchema.validateSync(schema, {
        context: {
          lorawanVersion: '1.1.0',
          meyEditKeys: false,
        },
      })

      expect(validatedValue).toBeDefined()
      expect(validatedValue.root_keys).toBeUndefined()
    })
  })

  describe('can edit keys', () => {
    describe('is `lorawan_version` 1.0.0', () => {
      const validate = schema =>
        validationSchema.validateSync(schema, {
          context: {
            mayEditKeys: true,
            lorawanVersion: '1.0.0',
          },
        })

      it('should handle `app_key`', () => {
        const appKey = '1'.repeat(32)

        schema.root_keys.app_key = { key: appKey }

        const validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.root_keys).toBeDefined()

        const { root_keys } = validatedValue
        expect(root_keys.app_key).toBeDefined()
        expect(root_keys.app_key.key).toBe(appKey)
      })

      it('should strip `nwk_key`', () => {
        const appKey = '1'.repeat(32)
        const nwkKey = '2'.repeat(32)

        schema.root_keys.app_key = { key: appKey }
        schema.root_keys.nwk_key = { key: nwkKey }

        const validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.root_keys).toBeDefined()

        const { root_keys } = validatedValue
        expect(root_keys.app_key).toBeDefined()
        expect(root_keys.app_key.key).toBe(appKey)
        expect(root_keys.nwk_key).toBeUndefined()
      })
    })

    describe('is `lorawan_version` 1.1.0', () => {
      const validate = schema =>
        validationSchema.validateSync(schema, {
          context: {
            mayEditKeys: true,
            lorawanVersion: '1.1.0',
          },
        })

      it('should handle `app_key`', () => {
        const appKey = '1'.repeat(32)

        schema.root_keys.app_key = { key: appKey }

        const validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.root_keys).toBeDefined()

        const { root_keys } = validatedValue
        expect(root_keys.app_key).toBeDefined()
        expect(root_keys.app_key.key).toBe(appKey)
      })

      it('should handle `nwk_key`', () => {
        const appKey = '1'.repeat(32)
        const nwkKey = '2'.repeat(32)

        schema.root_keys.app_key = { key: appKey }
        schema.root_keys.nwk_key = { key: nwkKey }

        const validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.root_keys).toBeDefined()

        const { root_keys } = validatedValue
        expect(root_keys.app_key).toBeDefined()
        expect(root_keys.app_key.key).toBe(appKey)
        expect(root_keys.nwk_key).toBeDefined()
        expect(root_keys.nwk_key.key).toBe(nwkKey)
      })
    })
  })
})
