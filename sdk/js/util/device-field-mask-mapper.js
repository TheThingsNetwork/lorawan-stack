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

/* eslint-disable no-console */
/* eslint-disable import/no-commonjs */
/* eslint-disable no-invalid-this */

const fs = require('fs')
const traverse = require('traverse')
const fieldMasks = require('../generated/device-field-masks.json')

const result = {}

for (const component in fieldMasks) {
  // Write get components
  for (const fieldMask of fieldMasks[component].get) {
    const path = [ ...fieldMask.split('.'), '_root' ]
    const val = traverse(result).get(path)
    if (val) {
      if (typeof val[0] === 'string') {
        traverse(result).set(path, [[ val[0], component ], val[1] ])
      } else {
        traverse(result).set(path, [[ ...val[0], component ], val[1] ])
      }
    } else {
      traverse(result).set(path, [ component, 'read_only' ])
    }
  }

  // Write set components
  for (const fieldMask of fieldMasks[component].set) {
    const path = [ ...fieldMask.split('.'), '_root' ]
    const val = traverse(result).get(path)
    if (val) {
      if (typeof val[1] === 'string') {
        traverse(result).set(path, [ val[0], val[1] !== 'read_only' ? [ val[1], component ] : component ])
      } else {
        traverse(result).set(path, [ val[0], [ ...val[1], component ]])
      }
    } else {
      traverse(result).set(path, [ component, component ])
    }
  }
}

// Rewrite single `_root` entries as plain array leaf
traverse(result).forEach(function () {
  if (Object.keys(this.node).length === 1 && this.node._root) {
    this.update(this.node._root)
  }
})

fs.writeFile(`${__dirname}/../generated/device-entity-map.json`, JSON.stringify(result, null, 2), function (err) {
  if (err) {
    return console.error(err)
  }
  console.log('File saved.')
})
