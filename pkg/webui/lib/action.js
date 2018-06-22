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

import { ulid } from 'ulid'
import check from './check-types'

const validator = {
  get (obj, type) {
    if (obj[type]) {
      return obj[type]
    }

    throw new TypeError(`${type} is not a valid action type, did you misspell it?`)
  },
}

/**
 * Creates an object of action creators based on the definition.
 *
 * @param {string} namespace - The namespace of the actions.
 * @param {object} def - The action creator definitions.
 * @param {object} def.key - A definition of an action creator, named by key
 * @param {object} def.key.types - A proptype object describing the type of payload
 * @param {function} def.key.transform - A transformer for the arguments, run before the type check
 *
 * @returns {object} - An object containing the action creators.
 */
export default function (namespace, def) {
  if (!namespace) {
    throw new TypeError('actions requires a namespace to be passed')
  }

  if (!def) {
    throw new TypeError('actions requires a definition to be passed')
  }

  const res =
    Object.keys(def)
      .reduce(function (acc, key) {
        const val = def[key]
        return {
          ...acc,
          [key]: action(namespace, key, val),
        }
      }, {})

  return new Proxy(res, validator)
}

// default transformer
const t = function (...args) {
  if (args.length === 0) {
    return {}
  }

  if (args[0] instanceof Error) {
    return {
      error: args[0],
    }
  }

  return args[0]
}

const normalize = function (obj = {}) {
  let types = obj
  let transform = t

  if (obj.types || (obj.transform && !obj.transform.isRequired)) {
    types = obj.types || {}
    transform = obj.transform || t
  }

  return {
    types,
    transform,
  }
}

const action = function (ns, key, def = {}) {
  const type = `${ns}/${key}`

  const fn = function (...args) {
    const d = normalize(def)

    // transform the arguments
    const payload = d.transform(...args)

    // check if the type is ok
    check(d.types, payload, 'argument', `action ${type}`)

    return {
      type,
      payload,
      meta: {
        created: new Date(),
        id: ulid(),
      },
    }
  }

  fn.type = type
  fn.toString = () => type

  return fn
}
