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

import '@ttn-lw/lib/yup'

import { ACTIVATION_MODES } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

describe('<ConfigurationForm /> validation schema', () => {
  const testHost = 'test-host'

  const createValidation = context => schema =>
    validationSchema.validateSync(schema, {
      context,
    })

  describe('has NS, JS and AS', () => {
    let schema

    beforeEach(() => {
      schema = {
        application_server_address: undefined,
        network_server_address: undefined,
        join_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: true,
        jsEnabled: true,
        asEnabled: true,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
        mayEditKeys: true,
      })(schema)

    it('should process `OTAA` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.OTAA)
      expect(validatedValue.supports_join).toBe(true)
      expect(validatedValue.multicast).toBe(false)
    })

    it('should process `ABP` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.ABP

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.ABP)
      expect(validatedValue.supports_join).toBe(false)
      expect(validatedValue.multicast).toBe(false)
    })

    it('should process `multicast` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.MULTICAST)
      expect(validatedValue.supports_join).toBe(false)
      expect(validatedValue.multicast).toBe(true)
    })

    it('should process `none` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.NONE)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should process `join_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBe(testHost)
    })

    it('should process `network_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)
    })

    it('should process `application_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBe(testHost)
    })

    it('should process `lorawan_version`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)
    })
  })

  describe('has JS and AS (no NS)', () => {
    let schema

    beforeEach(() => {
      schema = {
        application_server_address: undefined,
        network_server_address: undefined,
        join_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: false,
        jsEnabled: true,
        asEnabled: true,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
        mayEditKeys: true,
      })(schema)

    it('should fail on `ABP` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.ABP

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should fail on `multicast` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should process `OTAA` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.OTAA)
      expect(validatedValue.supports_join).toBe(true)
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should process `none` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.NONE)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should process `join_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBe(testHost)
    })

    it('should strip `network_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()
    })

    it('should process `application_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBe(testHost)
    })

    it('should process `lorawan_version`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)
    })
  })

  describe('has NS and AS (no JS)', () => {
    let schema

    beforeEach(() => {
      schema = {
        application_server_address: undefined,
        network_server_address: undefined,
        join_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: true,
        jsEnabled: false,
        asEnabled: true,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
        mayEditKeys: true,
      })(schema)

    it('should fail on `OTAA` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      try {
        const res = validate(schema)
        console.log(res)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should process `ABP` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.ABP

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.ABP)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBe(false)
    })

    it('should process `multicast` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.MULTICAST)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBe(true)
    })

    it('should process `none` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.NONE)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should strip `join_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()
    })

    it('should process `network_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)
    })

    it('should process `application_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBe(testHost)
    })

    it('should process `lorawan_version`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)
    })
  })

  describe('has AS (no JS and NS)', () => {
    let schema

    beforeEach(() => {
      schema = {
        application_server_address: undefined,
        network_server_address: undefined,
        join_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: false,
        jsEnabled: false,
        asEnabled: true,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
      })(schema)

    it('should fail on `ABP` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.ABP

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should fail on `multicast` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should fail on `OTAA` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should process `none` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.NONE)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should strip `join_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()
    })

    it('should strip `network_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()
    })

    it('should process `application_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()
    })

    it('should strip `lorawan_version`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBeUndefined()
    })
  })

  describe('has NS and JS (no AS)', () => {
    let schema

    beforeEach(() => {
      schema = {
        application_server_address: undefined,
        network_server_address: undefined,
        join_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: true,
        jsEnabled: true,
        asEnabled: false,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
        mayEditKeys: true,
      })(schema)

    it('should process `OTAA` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.OTAA)
      expect(validatedValue.supports_join).toBe(true)
      expect(validatedValue.multicast).toBe(false)
    })

    it('should process `ABP` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.ABP

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.ABP)
      expect(validatedValue.supports_join).toBe(false)
      expect(validatedValue.multicast).toBe(false)
    })

    it('should process `multicast` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.MULTICAST)
      expect(validatedValue.supports_join).toBe(false)
      expect(validatedValue.multicast).toBe(true)
    })

    it('should process `none` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.NONE)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should strip `join_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()
    })

    it('should process `network_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)
    })

    it('should process `application_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.application_server_address).toBeUndefined()
    })

    it('should process `lorawan_version`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)
      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)

      schema._activation_mode = ACTIVATION_MODES.ABP

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)

      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.lorawan_version).toBe(schema.lorawan_version)
    })
  })

  describe('is external JS', () => {
    let schema

    beforeEach(() => {
      schema = {
        join_server_addess: undefined,
        application_server_address: undefined,
        network_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: true,
        jsEnabled: true,
        asEnabled: false,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
        mayEditKeys: true,
      })(schema)

    it('should strip `join_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA
      schema._external_js = true

      const validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.join_server_address).toBeUndefined()
      expect(validatedValue.supports_join).toBe(true)
      expect(validatedValue.multicast).toBe(false)
    })
  })

  describe('cannot edit keys', () => {
    let schema

    beforeEach(() => {
      schema = {
        join_server_addess: undefined,
        application_server_address: undefined,
        network_server_address: undefined,
        lorawan_version: '1.0.0',
      }
    })

    const validate = schema =>
      createValidation({
        nsEnabled: true,
        jsEnabled: true,
        asEnabled: false,
        nsUrl: `http://${testHost}`,
        jsUrl: `http://${testHost}`,
        asUrl: `http://${testHost}`,
        mayEditKeys: false,
      })(schema)

    it('should fail on `ABP` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.ABP

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should fail on `multicast` activation mode', done => {
      schema._activation_mode = ACTIVATION_MODES.MULTICAST

      try {
        validate(schema)
        done.fail('should fail')
      } catch (error) {
        expect(error).toBeDefined()
        expect(error.name).toBe('ValidationError')
        expect(error.path).toBe('_activation_mode')
        done()
      }
    })

    it('should process `OTAA` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.OTAA

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.OTAA)
      expect(validatedValue.supports_join).toBe(true)
      expect(validatedValue.multicast).toBe(false)
    })

    it('should process `none` activation mode', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      const validatedValue = validate(schema)

      expect(validatedValue._activation_mode).toBe(ACTIVATION_MODES.NONE)
      expect(validatedValue.supports_join).toBeUndefined()
      expect(validatedValue.multicast).toBeUndefined()
    })

    it('should process `network_server_address`', () => {
      schema._activation_mode = ACTIVATION_MODES.NONE

      let validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBeUndefined()

      schema._activation_mode = ACTIVATION_MODES.OTAA

      validatedValue = validate(schema)

      expect(validatedValue).toBeDefined()
      expect(validatedValue.network_server_address).toBe(testHost)
    })
  })
})
