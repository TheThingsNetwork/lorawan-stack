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

/**
 * These hooks re-implement the now removed useBlocker and usePrompt hooks in 'react-router-dom'.
 * Thanks for the idea @piecyk https://github.com/remix-run/react-router/issues/8139#issuecomment-953816315
 * Source: https://github.com/remix-run/react-router/commit/256cad70d3fd4500b1abcfea66f3ee622fb90874#diff-b60f1a2d4276b2a605c05e19816634111de2e8a4186fe9dd7de8e344b65ed4d3L344-L381.
 */
import { useContext, useEffect, useCallback } from 'react'
import { UNSAFE_NavigationContext as NavigationContext } from 'react-router-dom'
/**
 * Blocks all navigation attempts. This is useful for preventing the page from
 * changing until some condition is met, like saving form data.
 *
 * @param {Function} blocker - Blocking function.
 * @param {boolean} when - Whether to block or not.
 * @see https://reactrouter.com/api/useBlocker
 */
export const useBlocker = (blocker, when = true) => {
  const { navigator } = useContext(NavigationContext)

  useEffect(() => {
    if (!when) return

    const unblock = navigator.block(tx => {
      const autoUnblockingTx = {
        ...tx,
        retry: () => {
          // Automatically unblock the transition so it can play all the way
          // through before retrying it. TODO: Figure out how to re-enable
          // this block if the transition is cancelled for some reason.
          unblock()
          tx.retry()
        },
      }

      blocker(autoUnblockingTx)
    })

    return unblock
  }, [navigator, blocker, when])
}
/**
 * Prompts the user with an Alert before they leave the current screen.
 *
 * @param {Function} message - Message to display.
 * @param {boolean} when - Whether to prompt or not.
 */
export const usePrompt = (message, when = true) => {
  const blocker = useCallback(
    tx => {
      // eslint-disable-next-line no-alert
      if (window.confirm(message)) tx.retry()
    },
    [message],
  )

  useBlocker(blocker, when)
}
