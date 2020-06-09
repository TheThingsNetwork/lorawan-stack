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

/**
 * @example
 * const object = {
 *  a: 'b',
 *  b: {
 *    a: 'c',
 *    c: 'd',
 *  },
 * }
 *
 * const resultingObject = omitDeep(object, 'a') // result is { b: { c: 'd' } }
 * @param {object|Array} value - Multinested object or array.
 * @param {string[]} key - Array of properties / keys to exclude from `value`.
 * @param {string[]} regexp - Regular expression for excluding properties from `value`.
 * @returns {object|Array} The new object/array without specified properties/keys.
 */
export default function omitDeep(value, key, regexp = /(?!)/) {
  if (Array.isArray(value)) {
    return value.map(i => omitDeep(i, key))
  } else if (typeof value === 'object' && value !== null) {
    return Object.keys(value).reduce((newObject, k) => {
      if (key.includes(k) || regexp.test(k)) return newObject
      return Object.assign({ [k]: omitDeep(value[k], key) }, newObject)
    }, {})
  }
  return value
}
