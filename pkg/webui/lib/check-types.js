// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// based on https://github.com/facebook/prop-types/blob/master/checkPropTypes.js

// sigh...
import secret from "prop-types/lib/ReactPropTypesSecret"

/**
 * Assert that the values match with the type specs.
 * Throws if a type is wrong.
 *
 * @param {object} specs - Map of name to a PropType
 * @param {object} values - Runtime values that need to be type-checked
 * @param {string} location - e.g. "prop", "context", "child context"
 * @param {string} component - Name of the component for error messages.
 */
export default function (specs, values, location, component) {
  Object.keys(specs).forEach(function (key) {
    try {
      // This is intentionally an invariant that gets caught. It's the same
      // behavior as without this statement except with a better message.
      if (typeof specs[key] !== "function") {
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
