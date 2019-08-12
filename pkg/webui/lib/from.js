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

/**
 * from returns an array that contains the values of the keys of a which have a trueish value in b.
 * If only is passed, it is used as a whitelist for keys that should be used.
 *
 * This can be used in conjunction with css modules and the classnames package to effectively created dynamic
 * classnames. Before:
 *
 *     import classnames from "classnames"
 *     import styles from "./some/styles.css"
 *
 *     const { foo, bar } = props
 *
 *     const classname = classnames(style.foo, {
 *       [style.bar]: bar,
 *       [style.baz]: baz,
 *     })
 *
 * After:
 *
 *     import classnames from "classnames"
 *     import styles from "./some/styles.css"
 *
 *     const { foo, bar } = props
 *
 *     const classname = classnames(style.foo, ...from(styles, { foo, bar }))
 *     // or
 *     const classname = classnames(style.foo, ...from(styles, props, [ "foo", "bar" ]))
 *
 *
 * @param {Object} a - The object to take the values from.
 * @param {Object} b - The object that controls which values will be taken.
 * @param {Array} only - Filter the keys by name.
 * @returns {Array} - An array of values from a for which the key in b had a trueish value.
 */
export default function(a = {}, b = {}, only) {
  const res = []
  for (const key in b) {
    if (!b[key] || !(key in a) || (only && !only.includes(key))) {
      continue
    }

    res.push(a[key])
  }

  return res
}
