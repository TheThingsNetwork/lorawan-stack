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

import { ACTIVATION_MODES } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

describe('<BasicSettingsForm /> validation schema', () => {
  const deviceId = 'test-device-id'
  const deviceName = 'test-device-name'
  const deviceDescription = 'test-device-description'
  const deviceJoinEUI = '1'.repeat(16)
  const deviceDevEUI = '2'.repeat(16)

  const createValidation = context => schema => validationSchema.validateSync(schema, { context })

  let schema

  beforeEach(() => {
    schema = {
      ids: {
        device_id: deviceId,
      },
      name: deviceName,
      description: deviceDescription,
    }
  })

  describe('is `OTAA` mode', () => {
    const validate = schema =>
      createValidation({
        activationMode: ACTIVATION_MODES.OTAA,
        lorawanVersion: '1.0.0',
      })(schema)

    it('should require `join_eui`', done => {
      schema.ids.dev_eui = deviceDevEUI

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('ids.join_eui')
        done()
      }
    })

    it('should require `dev_eui`', done => {
      schema.ids.join_eui = deviceJoinEUI

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('ids.dev_eui')
        done()
      }
    })

    it('should process valid schema', () => {
      schema.ids.join_eui = deviceJoinEUI
      schema.ids.dev_eui = deviceDevEUI

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.ids).toBeDefined()

      const { ids } = validatedValue
      expect(ids.join_eui).toBe(deviceJoinEUI)
      expect(ids.dev_eui).toBe(deviceDevEUI)
    })
  })

  describe('is `ABP` activation mode', () => {
    describe('is `lorawan_version` 1.0.4', () => {
      const validate = schema =>
        createValidation({
          activationMode: ACTIVATION_MODES.ABP,
          lorawanVersion: '1.0.4',
        })(schema)

      it('should require `dev_eui`', done => {
        try {
          validate(schema)
          done.fail('should fail')
        } catch (error) {
          expect(error).toBeDefined()
          expect(error.name).toBe('ValidationError')
          expect(error.path).toBe('ids.dev_eui')
        }

        schema.ids.dev_eui = deviceDevEUI

        const validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.ids).toBeDefined()
        expect(validatedValue.ids.dev_eui).toBe(deviceDevEUI)
        done()
      })
    })

    describe('is `lorawan_version` 1.0.0', () => {
      const validate = schema =>
        createValidation({
          activationMode: ACTIVATION_MODES.ABP,
          lorawanVersion: '1.0.0',
        })(schema)

      it('should process valid schema w/ or w/o `dev_eui`', () => {
        let validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.ids).toBeDefined()
        expect(validatedValue.ids.device_id).toBe(deviceId)
        expect(validatedValue.name).toBe(deviceName)
        expect(validatedValue.description).toBe(deviceDescription)

        schema.ids.dev_eui = deviceDevEUI

        validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.ids).toBeDefined()

        const { ids } = validatedValue
        expect(ids.device_id).toBe(deviceId)
        expect(ids.dev_eui).toBe(deviceDevEUI)
        expect(validatedValue.name).toBe(deviceName)
        expect(validatedValue.description).toBe(deviceDescription)
      })
    })
  })

  describe('is `multicast` activation mode', () => {
    describe('is `lorawan_version` 1.0.4', () => {
      const validate = schema =>
        createValidation({
          activationMode: ACTIVATION_MODES.MULTICAST,
          lorawanVersion: '1.0.4',
        })(schema)

      it('should require `dev_eui`', done => {
        try {
          validate(schema)
          done.fail('should fail')
        } catch (error) {
          expect(error).toBeDefined()
          expect(error.name).toBe('ValidationError')
          expect(error.path).toBe('ids.dev_eui')
          done()
        }
      })
    })

    describe('is `lorawan_version` 1.0.0', () => {
      const validate = schema =>
        createValidation({
          activationMode: ACTIVATION_MODES.MULTICAST,
          lorawanVersion: '1.0.0',
        })(schema)

      it('should process valid schema w/ or w/o `dev_eui`', () => {
        let validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.ids).toBeDefined()
        expect(validatedValue.ids.device_id).toBe(deviceId)
        expect(validatedValue.name).toBe(deviceName)
        expect(validatedValue.description).toBe(deviceDescription)

        schema.ids.dev_eui = deviceDevEUI

        validatedValue = validate(schema)

        expect(validatedValue).toBeDefined()
        expect(validatedValue.ids).toBeDefined()

        const { ids } = validatedValue
        expect(ids.device_id).toBe(deviceId)
        expect(ids.dev_eui).toBe(deviceDevEUI)
        expect(validatedValue.name).toBe(deviceName)
        expect(validatedValue.description).toBe(deviceDescription)
      })
    })
  })

  describe('is `none` activation mode', () => {
    const validate = schema =>
      createValidation({
        activationMode: ACTIVATION_MODES.MULTICAST,
        lorawanVersion: '1.0.0',
      })(schema)

    it('should process valid schema', () => {
      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.ids).toBeDefined()
      expect(validatedValue.ids.device_id).toBe(deviceId)
      expect(validatedValue.name).toBe(deviceName)
      expect(validatedValue.description).toBe(deviceDescription)
    })
  })
})
