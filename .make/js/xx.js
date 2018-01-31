// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

import icu from "messageformat-parser"

const stringify = function (token) {
  if (typeof token === "string") {
    return token.replace(/[A-Z]/g, "X").replace(/[^X,.~`?:\-_=+!@#$%*(){}[\]"';/ \s]/g, "x")
  }

  if (token.type === "argument") {
    return `{${token.arg}}`
  }

  if (token.type === "octothorpe") {
    return "#"
  }

  let res = `{${token.arg}, ${token.type},`

  for (const c of token.cases) {
    const k = c.key === "other" ? c.key : `=${c.key}`
    const s = c.tokens.map(stringify).join("")
    res += ` ${k} {${s}}`
  }

  res += "}"

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
export default function (format) {
  return icu.parse(format).map(stringify).join("")
}
