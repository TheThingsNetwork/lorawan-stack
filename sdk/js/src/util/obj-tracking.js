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
  let trackedObject = obj
  trackedObject = new Proxy(obj, trackerProxy(obj))

  for (const key in trackedObject) {
    const leaf = trackedObject[key]
    if (typeof leaf === 'object' && !(leaf instanceof Array)) {
      trackedObject[key] = trackObject(trackedObject[key])
    }
  }

  // Remove unwanted changed markings that have been added by the proxies
  trackedObject._changed = []

  return trackedObject
}

/**
 * Traverse through the object and remove all _changed decorations.
 * @param {Object} obj - The to be cleaned object.
 * @param {boolean} clone - Whether the obj should be cloned before cleaning.
 * @returns {Object} The cleaned object.
 */
function removeDecorations (obj, clone = false) {
  const subject = clone ? traverse(obj).clone() : obj

  traverse(subject).forEach(function (element) {
    if (this.key === '_changed') {
      this.remove()
    }
  })

  return subject
}

export { trackerProxy, trackObject, removeDecorations }
