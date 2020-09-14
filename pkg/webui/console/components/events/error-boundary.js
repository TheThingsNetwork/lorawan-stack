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

import React from 'react'
import * as Sentry from '@sentry/browser'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import m from './messages'

import style from './events.styl'

class EventErrorBoundary extends React.Component {
  state = { hasErrored: false, error: undefined, expanded: false }

  static getDerivedStateFromError(error) {
    Sentry.captureException(error)
    return { hasErrored: true, error }
  }

  render() {
    const { hasErrored } = this.state
    const { children } = this.props

    if (hasErrored) {
      return (
        <div className={style.cellError}>
          <Icon icon="error" className={style.eventIcon} />
          <Message content={m.errorOverviewEntry} />
        </div>
      )
    }

    return children
  }
}

EventErrorBoundary.propTypes = {
  children: PropTypes.node.isRequired,
}

export default EventErrorBoundary
