// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import { useState, useEffect } from 'react'

/**
 * Custom React hook for managing state with cookies.
 *
 * @param {string} cookieName - The name of the cookie to manage.
 * @param {*} defaultValue - The default value to use if the cookie is not set.
 * @returns {[*, Function]} A stateful value, and a function to update it.
 */
const useCookieState = (cookieName, defaultValue) => {
  const prefixedCookieName = `webui-${cookieName}`
  const [value, setValue] = useState(() => {
    const cookieValue = document.cookie
      .split('; ')
      .find(row => row.startsWith(`${prefixedCookieName}=`))
      ?.split('=')[1]

    return cookieValue !== undefined ? JSON.parse(decodeURIComponent(cookieValue)) : defaultValue
  })

  useEffect(() => {
    document.cookie = `${prefixedCookieName}=${encodeURIComponent(JSON.stringify(value))}`
  }, [prefixedCookieName, value])

  return [value, setValue]
}

export default useCookieState
