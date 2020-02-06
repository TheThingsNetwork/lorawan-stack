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

import { set, get, toPath } from 'lodash'

/**
 * @example
 *
 * const userDefaultsEmitter = createDefaultsEmitterFromFieldMask((fmKey, value) => {
 *  if (fmKey === 'state' && (!Boolean(value) || value === null)) {
 *    return 'STATE_REQUESTED'
 *  }
 * })
 *
 * const user = { ids: { user_id: 'user' }, isAdmin: false}
 * const fieldMask = ['isAdmin', state]
 * const userWithDefaults = userDefaultsEmitter(user, fieldMask)
 * // userWidthDefaults: { ids: { user_id: 'user' }, isAdmin: false, state: 'STATE_REQUESTED' }
 *
 * Creates a defaults emitter function. This could be helpful in cases when the http backend strips
 * nullable/falsy/default values while these values are required.
 * @param {Function} formatter - The customization function that defines how and which
 * fields should be updated. The function is invoked with two arguments: fieldMaskKey, value. Where
 * - `fieldMaskKey` is the field mask entry (string)
 * - `value` is the value mapped to the `fieldMaskKey` (any)
 * In order to update `formatter` should return a new value, if it returns `undefined` nothing is changed
 * in `sourceObject`.
 * @returns {Function} - The defaults emitter function. It accepts two arguments: sourceObject, fieldMask.
 * Where:
 * - `sourceObject` is the object to add default values to
 * - `fieldMask` the field mask array
 * The defaults emitter function returns the updated version copy of the `sourceObject`.
 */
const createDefaultsEmitterFromFieldMask = formatter => {
  return function(object, fieldMask) {
    const result = { ...object }

    for (const fmKey of fieldMask) {
      const path = toPath(fmKey)
      const value = get(object, path)

      const updated = formatter(fmKey, value)
      if (typeof updated !== 'undefined') {
        set(result, path, updated)
      }
    }

    return result
  }
}

// eslint-disable-next-line import/prefer-default-export
export { createDefaultsEmitterFromFieldMask }
