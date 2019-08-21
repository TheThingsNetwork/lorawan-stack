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

/* eslint-disable no-invalid-this */

import traverse from 'traverse'

import { trackerProxy, trackObject, removeDecorations } from '../util/obj-tracking'

/**
 * Entity class serves as an abstraction on data returned from the API. It will
 * keep track of changes and generate update masks that the API requires.
 */
class Entity {
  constructor(data, isNew = true) {
    this._isNew = isNew
    this._changed = []

    // We do not want to manipulate the original data so we take a copy instead
    const clonedData = traverse(data).clone()
    const trackedObject = trackObject(clonedData)
    this.applyValues(trackedObject)
    this._rawData = data

    // Proxy the returned instance to keep track of modifications
    return new Proxy(this, trackerProxy(clonedData))
  }

  applyValues(values) {
    for (const key in values) {
      this[key] = values[key]
    }
  }

  clearValues() {
    removeDecorations(this)
  }

  toObject() {
    const output = {}
    for (const key in this._rawData) {
      const leaf = this[key]
      if (typeof leaf === 'object') {
        // Clone and remove _changed property on objects
        output[key] = removeDecorations(leaf, true)
      } else {
        output[key] = leaf
      }
    }

    return output
  }

  getUpdateMask() {
    let res = []
    let trails = []
    const traverseMask = function(tree, trail) {
      for (const key in tree) {
        if (key === '_changed') {
          continue
        }

        const leaf = tree[key]
        if (typeof leaf !== 'object') {
          if (tree._changed && tree._changed.indexOf(key) !== -1) {
            trails.push([...trail, key])
          }
        } else {
          traverseMask(leaf, [...trail, key])
        }
      }
    }

    for (const key in this._rawData) {
      const trail = [key]
      trails = []
      if (typeof this[key] === 'object') {
        traverseMask(this[key], trail)
        if (trails.length !== 0) {
          res = [...res, ...trails]
        }
      } else if (this._changed && this._changed.indexOf(key) !== -1) {
        res = [...res, trail]
      }
    }

    return res.map(e => e.join('.'))
  }

  mask(mask = this.getUpdateMask()) {
    const paths = 'paths' in mask ? mask.paths : mask
    const res = this.toObject()

    traverse(res).forEach(function(item) {
      // "this" now contains information about the node
      // see https://github.com/substack/js-traverse#context
      const path = this.path.join('.')
      if (this.notRoot && paths.indexOf(path) === -1) {
        this.remove()
      }
    })

    return res
  }

  save(data) {
    this._isNew = false
    this.applyValues(data)
  }
}

// In order to strip the console output of the object from all decorated props,
// we modify the inspect method to return the plain object representation
Entity.prototype.inspect = function() {
  return this.toObject()
}

export default Entity
