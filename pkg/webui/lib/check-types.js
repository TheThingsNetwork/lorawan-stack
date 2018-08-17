// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// based on https://github.com/facebook/prop-types/blob/master/checkPropTypes.js

// sigh...
import secret from 'prop-types/lib/ReactPropTypesSecret'

/**
 * Assert that the values match with the type specs.
 * Throws if a type is wrong.
 *
 * @param {Object} specs - Map of name to a PropType
 * @param {Object} values - Runtime values that need to be type-checked
 * @param {string} location - e.g. "prop", "context", "child context"
 * @param {string} component - Name of the component for error messages.
 */
export default function (specs, values, location, component) {
  Object.keys(specs).forEach(function (key) {
    try {
      // This is intentionally an invariant that gets caught. It's the same
      // behavior as without this statement except with a better message.
      if (typeof specs[key] !== 'function') {
        throw new TypeError(`${component}: ${location} type \`${key}\` is invalid; it must be a function, usually from the prop-types package, but received \`${typeof specs[key]}\`.'`)
      }

      const error = specs[key](values, key, component, location, null, secret)
      if (error instanceof Error) {
        throw error
      }

      if (error) {
        throw new TypeError(`${component}: type specification of ${location} \`${key}\` is invalid; the type checker function must return \`null\` or an \`Error\` but returned a ${typeof error}. You may have forgotten to pass an argument to the type checker creator (arrayOf, instanceOf, objectOf, oneOf, oneOfType, and shape all require an argument).`)
      }
    } catch (error) {
      throw new Error(`Failed ${location} type: ${error.message}`)
    }
  })
}
