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

const icu = require('messageformat-parser')

const stringify = function (token) {
  if (typeof token === 'string') {
    return token.replace(/[A-Z]/g, 'X').replace(/[^X,.~`?:\-_=+!@#$%*(){}[\]"';/ \s]/g, 'x')
  }

  if (token.type === 'argument') {
    return `{${token.arg}}`
  }

  if (token.type === 'octothorpe') {
    return '#'
  }

  let res = `{${token.arg}, ${token.type},`

  for (const c of token.cases) {
    const k = c.key === 'other' ? c.key : `=${c.key}`
    const s = c.tokens.map(stringify).join('')
    res += ` ${k} {${s}}`
  }

  res += '}'

  return res
}

/**
 * Replace all the non-ICU text in a format string with x'es.
 *
 * For example
 *   "The {name} should contain at least {min, plural, =1 {one character} other {# characters}}"
 *   "Xxx {name} xxxxxx xxxxxxx xx xxxxx {min, plural, =1 {xxx xxxxxxxxx} other {# xxxxxxxxxx}}"
 *
 * @param {string} format - The format string.
 * @returns {string} - The updated format.
 */
module.exports = function (format) {
  return icu.parse(format).map(stringify).join('')
}
