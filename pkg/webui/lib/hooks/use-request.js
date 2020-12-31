// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { useState, useEffect, useRef } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import { attachPromise } from '@console/store/actions/lib'

import { selectIsOnlineStatus } from '@console/store/selectors/status'

const useRequest = requestAction => {
  const dispatch = useDispatch()
  const [fetching, setFetching] = useState(true)
  const [error, setError] = useState('')
  const [result, setResult] = useState()
  const isOnline = useSelector(selectIsOnlineStatus)
  const prevIsOnlineRef = useRef()
  useEffect(() => {
    prevIsOnlineRef.current = isOnline
  })
  const prevIsOnline = prevIsOnlineRef.current

  useEffect(() => {
    if (prevIsOnline === undefined || (prevIsOnline === false && isOnline)) {
      // Make the request initially and additionally when the online state
      // has changed to `online`.
      const promise = dispatch(attachPromise(requestAction))
        .then(() => {
          setResult(result)
          setFetching(false)
        })
        .catch(error => {
          setError(error)
          setFetching(false)
        })

      return () => {
        // Cancel the promise on unmount (if still pending).
        promise.cancel()
      }
    }

    // We use the `isOnline` prop as dependency here since we want the effect
    // to trigger only initially and when the online state changed.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOnline])

  return [fetching, error, result]
}

export default useRequest
