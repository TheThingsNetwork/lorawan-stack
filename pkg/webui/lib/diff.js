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

import { observableDiff, applyChange } from 'deep-diff'
import { get, has, set } from 'lodash'

import { warn } from './log'

/**
 * Computes the structural differences (deep) between `original` and `updated`
 * objects by applying any differences (add/remove/update).
 *
 * @param {object} original - The original object.
 * @param {object} updated - The updated version of the `original` object.
 * @param {object} options - Options object.
 * @param {Array} options.exclude - A list of field names that should not be included in
 * the final diff.
 * @param {boolean} options.patchArraysItems - Whether to include diffs on arrays
 * on per element basis. If disabled changed arrays will be added in full.
 * @param {Array} options.patchInFull - Paths to only patch as a whole without traversal.
 * @returns {object} - A new object representing the structural differences
 * between `original` and `updated`.
 */
export default (
  original,
  updated,
  { exclude = [], patchArraysItems = true, patchInFull = [] } = {},
) => {
  const result = {}

  observableDiff(original, updated, d => {
    const { kind: diffKind, rhs: diffValue, path: diffPath } = d
    const diffEntry = diffPath[diffPath.length - 1]

    // Do not add new entries that are of type `undefined`.
    if (diffKind === 'N' && typeof diffValue === 'undefined') {
      return
    }

    // Do not diff array items if requested.
    if (diffKind === 'A') {
      if (!patchArraysItems) {
        set(result, diffPath, get(updated, diffPath))
        return
      }
      warn(
        'diff() has included a diff within an array. Make sure that this patched array is not sent to the backend, since it will not apply patches of arrays, but only arrays as a whole. Please refactor this by using { includeArrays: false }.',
      )
    }

    // Apply full patches to paths as requested.
    const path = patchInFull.find(p => diffPath.join('.').startsWith(`${p}.`))
    if (path) {
      if (!has(result, path)) {
        set(result, path, get(updated, path))
      }
      return
    }

    if (!exclude.includes(diffEntry)) {
      applyChange(result, undefined, d)
    }
  })

  return result
}
