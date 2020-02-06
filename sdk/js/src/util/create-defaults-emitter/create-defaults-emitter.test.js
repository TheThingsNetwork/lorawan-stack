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

import { createDefaultsEmitterFromFieldMask } from '.'

describe('create-defaults-emitter', () => {
  describe('createDefaultsEmitterFromFieldMask', () => {
    it('should pass correct values and field mask to emitter', () => {
      const obj = {
        field1: '',
        field2: null,
        field3: undefined,
        field4: false,
        field5: {},
        field6: [],
        field7: { field1: '', field2: null, field3: undefined, field4: false },
        field8: [42],
        field9: 'string',
        field10: 42,
        field11: true,
      }
      const fieldMask = [
        'field1',
        'field2',
        'field3',
        'field4',
        'field5',
        'field6',
        'field7',
        'field7.field1',
        'field7.field2',
        'field7.field3',
        'field7.field4',
        'field8',
        'field9',
        'field10',
        'field11',
      ]

      const defaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
        expect(fieldMask.includes(fmKey)).toBe(true)
        const keys = fmKey.split('.')

        let val
        for (const k of keys) {
          val = obj[k]
        }

        expect(val).toEqual(value)
      })
      defaultsEmitter(obj, fieldMask)
    })

    it('should set and update any falsy value except for `undefined`', () => {
      const values = {
        field1: '',
        field2: null,
        field3: false,
        field4: undefined,
      }

      const obj = {}
      const fieldMask = Object.keys(values)

      const defaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
        return values[fmKey]
      })
      let withDefaults = defaultsEmitter(obj, fieldMask)

      expect(Object.keys(withDefaults)).toHaveLength(3)
      expect(withDefaults.field1).toBe(values.field1)
      expect(withDefaults.field2).toBe(values.field2)
      expect(withDefaults.field3).toBe(values.field3)
      expect(withDefaults).not.toHaveProperty('field4')

      obj.field1 = 'value1'
      obj.field2 = 'value2'
      obj.field3 = 'value3'
      obj.field4 = 'value4'

      withDefaults = defaultsEmitter(obj, fieldMask)

      expect(Object.keys(withDefaults)).toHaveLength(4)
      expect(withDefaults.field1).toBe(values.field1)
      expect(withDefaults.field2).toBe(values.field2)
      expect(withDefaults.field3).toBe(values.field3)
      expect(withDefaults.field4).toBe(obj.field4)
    })

    it('should leave unrelated fields untouched', () => {
      const fieldName1 = 'fieldName1'
      const fieldName2 = 'fieldName2'
      const fieldName3 = 'fieldName3'
      const fieldValue1 = 'fieldValue1'
      const fieldValue2 = {}
      const fieldValue3 = 'fieldValue3'

      const obj = {
        [fieldName1]: fieldValue1,
        [fieldName2]: fieldValue2,
      }
      const fieldMask = [fieldName3]

      const defaultsEmitter = createDefaultsEmitterFromFieldMask(fmKey => {
        if (fmKey === fieldName3) {
          return fieldValue3
        }
      })
      const withDefaults = defaultsEmitter(obj, fieldMask)

      expect(withDefaults[fieldName1]).toBe(fieldValue1)
      expect(withDefaults[fieldName2]).toBe(fieldValue2)
      expect(withDefaults[fieldName2] === fieldValue2).toBe(true)
    })

    it('should set top-level fields', () => {
      const fieldName1 = 'fieldName1'
      const fieldName2 = 'fieldName2'
      const fieldValue1 = 'fieldValue1'
      const fieldValue2 = 'fieldValue2'

      const obj = {}
      const fieldMask = [fieldName1, fieldName2]

      const defaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
        if (fmKey === fieldName1) {
          return fieldValue1
        }

        if (fmKey === fieldName2) {
          return fieldValue2
        }
      })
      const withDefaults = defaultsEmitter(obj, fieldMask)

      expect(Object.keys(withDefaults)).toHaveLength(2)
      expect(withDefaults[fieldName1]).toBe(fieldValue1)
      expect(withDefaults[fieldName2]).toBe(fieldValue2)
    })

    it('should update top-level fields', () => {
      const fieldName1 = 'fieldName1'
      const fieldValue1 = 'updatedFieldValue1'

      const obj = { [fieldName1]: undefined }
      const fieldMask = [fieldName1]

      const defaultsEmitter = createDefaultsEmitterFromFieldMask(fmKey => {
        if (fmKey === fieldName1) {
          return fieldValue1
        }
      })
      const withDefaults = defaultsEmitter(obj, fieldMask)

      expect(Object.keys(withDefaults)).toHaveLength(1)
      expect(withDefaults[fieldName1]).toBe(fieldValue1)
    })

    it('should set nested fields', () => {
      const obj = {}
      const fieldName1 = 'fieldName1'
      const fieldName2 = 'fieldName2'
      const fieldName3 = 'fieldName3'
      const fieldValue = 'fieldValue'
      const combinedFieldMask = `${fieldName1}.${fieldName2}.${fieldName3}`
      const fieldMask = [combinedFieldMask]

      const defaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
        if (fmKey === combinedFieldMask) {
          return fieldValue
        }
      })
      const withDefaults = defaultsEmitter(obj, fieldMask)

      expect(Object.keys(withDefaults)).toHaveLength(1)
      expect(Object.keys(withDefaults[fieldName1])).toHaveLength(1)
      expect(Object.keys(withDefaults[fieldName1][fieldName2])).toHaveLength(1)
      expect(withDefaults[fieldName1][fieldName2][fieldName3]).toBe(fieldValue)
    })

    it('should update nested fields', () => {
      const fieldName1 = 'fieldName1'
      const fieldName2 = 'fieldName2'
      const fieldName3 = 'fieldName3'
      const fieldValue = 'fieldValue'
      const combinedFieldMask = `${fieldName1}.${fieldName2}`
      const newFieldValue = { newFieldName: 'newFieldValue' }

      const obj = {
        [fieldName1]: {
          [fieldName2]: {
            [fieldName3]: fieldValue,
          },
        },
      }
      const fieldMask = [combinedFieldMask]

      const defaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
        if (fmKey === combinedFieldMask) {
          return newFieldValue
        }
      })
      const withDefaults = defaultsEmitter(obj, fieldMask)

      expect(Object.keys(withDefaults)).toHaveLength(1)
      expect(Object.keys(withDefaults[fieldName1])).toHaveLength(1)
      expect(withDefaults[fieldName1][fieldName2]).toBe(newFieldValue)
    })
  })
})
