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

/**
 * Computes the structural differences (deep) between `original` and `updated` objects
 * by applying any differences (add/remove/update).
 * @param {Object} original - The original object.
 * @param {Object} updated - The updated version of the `original` object.
 * @param {Array} exclude - A list of field names that should not be included in the
 * final diff.
 * @returns {Object} - A new object representing the structural differences between
 * `original` and `updated`.
 */
export default function(original, updated, exclude = []) {
  const result = {}

  observableDiff(original, updated, function(d) {
    const entry = d.path[d.path.length - 1]
    if (!exclude.includes(entry)) {
      applyChange(result, undefined, d)
    }
  })

  return result
}
