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

import Spinner from '../../components/spinner'
import ErrorMessage from './error-message'

import '../../styles/main.styl'

// React grid configuration
// Keep these in line with styles/variables.less
setConfiguration({
  breakpoints: [480, 768, 1080, 1280],
  containerWidths: [768, 1000, 1140, 1140],
  gutterWidth: 28,
})

@connect(state => ({
  initialized: state.init.initialized,
  error: state.init.error,
}))
export default class Init extends React.PureComponent {
  componentDidMount() {
    this.props.dispatch({ type: 'INITIALIZE_REQUEST' })
  }

  render() {
    const { initialized, error } = this.props

    if (error) {
      return <ErrorMessage content={error} />
    }

    if (!initialized) {
      return <Spinner center>Please wait…</Spinner>
    }

    return this.props.children
  }
}
