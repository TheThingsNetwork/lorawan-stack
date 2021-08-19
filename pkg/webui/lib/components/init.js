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
import { connect } from 'react-redux'
import 'focus-visible/dist/focus-visible'
import { setConfiguration } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import SourceSansRegular from '@assets/fonts/source-sans-pro-v13-latin_latin-ext-regular.woff2'
import SourceSans600 from '@assets/fonts/source-sans-pro-v13-latin_latin-ext-600.woff2'
import SourceSans700 from '@assets/fonts/source-sans-pro-v13-latin_latin-ext-700.woff2'
import IBMPlexMono from '@assets/fonts/ibm-plex-mono-regular.woff2'
import MaterialIcons from '@assets/fonts/materialicons.woff2'
import LAYOUT from '@ttn-lw/constants/layout'

import Spinner from '@ttn-lw/components/spinner'

import PropTypes from '@ttn-lw/lib/prop-types'

import ErrorMessage from './error-message'
import Message from './message'

import '@ttn-lw/styles/main.styl'

const m = defineMessages({
  initializing: 'Initializing…',
})

// Keep this list updated with fonts used in `/styles/fonts.styl`.
const fontsToPreload = [SourceSansRegular, SourceSans600, SourceSans700, IBMPlexMono, MaterialIcons]

setConfiguration({
  breakpoints: [
    LAYOUT.BREAKPOINTS.XXS,
    LAYOUT.BREAKPOINTS.S,
    LAYOUT.BREAKPOINTS.M,
    LAYOUT.BREAKPOINTS.L,
  ],
  containerWidths: [
    LAYOUT.CONTAINER_WIDTHS.XS,
    LAYOUT.CONTAINER_WIDTHS.S,
    LAYOUT.CONTAINER_WIDTHS.M,
    LAYOUT.CONTAINER_WIDTHS.L,
  ],
  gutterWidth: LAYOUT.GUTTER_WIDTH,
})

@connect(
  state => ({
    initialized: state.init.initialized,
    error: state.init.error,
  }),
  dispatch => ({
    initialize: () => dispatch({ type: 'INITIALIZE_REQUEST' }),
  }),
)
export default class Init extends React.PureComponent {
  static propTypes = {
    children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
    error: PropTypes.error,
    initialize: PropTypes.func.isRequired,
    initialized: PropTypes.bool,
  }

  static defaultProps = {
    initialized: false,
    error: undefined,
  }

  componentDidMount() {
    const { initialize } = this.props

    initialize()

    // Preload font files to avoid flashes of unstyled text.
    for (const fontUrl of fontsToPreload) {
      const linkElem = document.createElement('link')
      linkElem.setAttribute('rel', 'preload')
      linkElem.setAttribute('href', fontUrl)
      linkElem.setAttribute('as', 'font')
      linkElem.setAttribute('crossorigin', 'anonymous')
      document.getElementsByTagName('head')[0].appendChild(linkElem)
    }
  }

  render() {
    const { initialized, error } = this.props

    if (error) {
      return <ErrorMessage content={error} />
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

    return this.props.children
  }
}
