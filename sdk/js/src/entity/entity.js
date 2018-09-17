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

/* eslint-disable no-invalid-this */

import deleteKey from 'object-delete-key'
import traverse from 'traverse'

// The tracker proxy which will keep track of changes via setters and stores
// changed properties in the _changed array
const trackerProxy = trackedData => ({
  set (obj, prop, value) {
    if (prop in trackedData) {
      if (obj._changed === undefined) {
        obj._changed = []
      }
      if (obj._changed.indexOf(prop) === -1) {
        obj._changed.push(prop)
      }
    }
    return Reflect.set(obj, prop, value)
  },
})

/**
 * Traverse through the object and apply a tracker proxy to all child objects.
 * @param {Object} obj - The to be tracked object.
 * @returns {Object} The tracked object.
 */
function trackObject (obj) {
  for (const key in obj) {
    const leaf = obj[key]
    if (typeof leaf === 'object' && !(leaf instanceof Array)) {
      obj[key] = new Proxy(leaf, trackerProxy(leaf))
      trackObject(obj[key])
      obj[key]._changed = []
    }
  }

  return obj
}

/**
 * Entity class serves as an abstraction on data returned from the API. It will
 * keep track of changes and generate update masks that the API requires.
 */
class Entity {
  constructor (data, isNew = true) {
    this._isNew = isNew
    this._changed = []

    // We do not want to manipulate the original data so we take a copy instead
    const clonedData = JSON.parse(JSON.stringify(data))
    const trackedObject = trackObject(clonedData)
    this.applyValues(trackedObject)
    this._rawData = data

    // Proxy the returned instance to keep track of modifications
    return new Proxy(this, trackerProxy(clonedData))
  }

  applyValues (values) {
    for (const key in values) {
      this[key] = values[key]
    }
  }

  clearValues () {
    for (const key in this) {
      if (!(key.startsWith('_'))) {
        delete this[key]
      }
    }
  }

  toObject () {
    const output = {}
    for (const key in this._rawData) {
      const leaf = this[key]
      if (typeof leaf === 'object') {
        // Clone and remove _changed property on objects
        output[key] = deleteKey(Object.assign(leaf instanceof Array ? [] : {}, leaf), { key: '_changed' })
      } else {
        output[key] = leaf
      }
    }

    return output
  }

  getUpdateMask () {
    let res = []
    let trails = []
    const traverseMask = function (tree, trail) {
      for (const key in tree) {
        if (key === '_changed') {
          continue
        }

        const leaf = tree[key]
        if (typeof leaf !== 'object') {
          if (tree._changed && tree._changed.indexOf(key) !== -1) {
            trails.push([ ...trail, key ])
          }
        } else {
          traverseMask(leaf, [ ...trail, key ])
        }
      }
    }

    for (const key in this._rawData) {
      const trail = [ key ]
      trails = []
      if (typeof this[key] === 'object') {
        traverseMask(this[key], trail)
        if (trails.length !== 0) {
          res = [ ...res, ...trails ]
        }
      } else if (this._changed.indexOf(key) !== -1) {
        res = [ ...res, trail ]
      }
    }

    return res.map(e => e.join('.'))
  }

  mask (mask = this.getUpdateMask()) {
    const paths = ('paths' in mask) ? mask.paths : mask
    const res = this.toObject()

    traverse(res).forEach(function (item) {
      const path = this.path.join('.')
      if (this.notRoot && paths.indexOf(path) === -1) {
        this.remove()
      }
    })

    return res
  }

  save (data) {
    this._isNew = false
    this.applyValues(data)
  }

}

// In order to strip the console output of the object from all decorated props,
// we modify the inspect method to return the plain object representation
Entity.prototype.inspect = function () {
  return this.toObject()
}

export default Entity
