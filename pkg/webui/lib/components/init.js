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

import React, { useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import 'focus-visible/dist/focus-visible'
import { defineMessages } from 'react-intl'

import IBMPlexMono from '@assets/fonts/ibm-plex-mono-regular.woff2'
import TextSecurityDisc from '@assets/fonts/text-security-disc.woff2'

import Spinner from '@ttn-lw/components/spinner'

import PropTypes from '@ttn-lw/lib/prop-types'
import { initialize } from '@ttn-lw/lib/store/actions/init'
import {
  selectInitError,
  selectInitFetching,
  selectIsInitialized,
} from '@ttn-lw/lib/store/selectors/init'

import Message from './message'

import '@ttn-lw/styles/main.styl'
import '@ttn-lw/styles/utilities/general.styl'
import '@ttn-lw/styles/utilities/spacing.styl'
import '@ttn-lw/styles/utilities/tokens.styl'

const m = defineMessages({
  initializing: 'Initializing…',
})

// Keep this list updated with fonts used in `/styles/fonts.styl`.
const fontsToPreload = [IBMPlexMono, TextSecurityDisc]

const Init = ({ children }) => {
  const initialized = useSelector(state => !selectInitFetching(state) && selectIsInitialized(state))
  const error = useSelector(state => selectInitError(state))
  const dispatch = useDispatch()

  useEffect(() => {
    dispatch(initialize())

    // Preload font files to avoid flashes of unstyled text.
    for (const fontUrl of fontsToPreload) {
      const linkElem = document.createElement('link')
      linkElem.setAttribute('rel', 'preload')
      linkElem.setAttribute('href', fontUrl)
      linkElem.setAttribute('as', 'font')
      linkElem.setAttribute('crossorigin', 'anonymous')
      document.getElementsByTagName('head')[0].appendChild(linkElem)
    }
  }, [dispatch])

  if (error) {
    throw error
  }

  if (!initialized) {
    return (
      <div style={{ height: '100vh' }}>
        <Spinner center>
          <Message content={m.initializing} />
        </Spinner>
      </div>
    )
  }

  return children
}

Init.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
}

export default Init
