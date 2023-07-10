// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { useSearchParams } from 'react-router-dom'

/**
 * `useQueryState` is a hook that allows to store a value in the URL query string.
 * It returns a stateful value, and a function to update it.
 * The updated value is stored in the URL query string.
 * @param {string} key - The key to store the value under in the URL query string.
 * @param {any} initialValue - The initial value to store in the URL query string.
 * @param {Function} parser - A function to parse the value from the URL query string.
 * @returns {[any, Function]} - A tuple containing the stateful value and a function to update it.
 */
const useQueryState = (key, initialValue, parser = v => v) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [state, setState] = useState(parser(searchParams.get(key)) || initialValue)

  useEffect(() => {
    if (state === undefined || state === '') {
      searchParams.delete(key)
    } else {
      searchParams.set(key, state)
    }
    setSearchParams(searchParams, { replace: true })
  }, [key, state, searchParams, setSearchParams])

  return [state, setState]
}

export default useQueryState
