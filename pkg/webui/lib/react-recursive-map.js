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

import React from 'react'

/**
 * Recursively maps react component children
 * @param {Object} children - asdnaj
 * @param {Function} fn - Transfomation function to be applied to each child
 * @returns {Object} Cloned children
 */
function recursiveMap (children, fn) {
  return React.Children.map(children, function (Child) {
    if (!React.isValidElement(Child)) {
      return Child
    }

    let child = Child
    if (child.props.children) {
      child = React.cloneElement(child, {
        children: recursiveMap(child.props.children, fn),
      })
    }

    return fn(child)
  })
}

export default recursiveMap
