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
import { defineMessages } from 'react-intl'
import ErrorStackParser from 'error-stack-parser'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Event from '../../event'

import style from './error-boundary.styl'

const m = defineMessages({
  title: 'Error cause',
  overviewEntry: 'Something went wrong when displaying this event',
})

const formatError = error => {
  if (error instanceof Error) {
    const cause = ErrorStackParser.parse(error)[0] || {}

    return {
      name: error.name,
      message: error.message,
      cause: cause.functionName,
    }
  }

  return error
}

class EventErrorBoundary extends React.Component {
  state = { hasErrored: false, error: undefined }

  static getDerivedStateFromError(error) {
    return { hasErrored: true, error }
  }

  render() {
    const { hasErrored, error } = this.state
    const { children, event, widget } = this.props

    if (hasErrored) {
      const content =
        typeof error === 'string' ? error : JSON.stringify(formatError(error), null, 2)

      return (
        <Event event={event} expandable widget={widget}>
          <Event.Overview>
            <Event.Overview.Entry className={style.overviewEntry}>
              <Message content={m.overviewEntry} />
            </Event.Overview.Entry>
          </Event.Overview>
          <Event.Details>
            <Message className={style.errorTitle} content={m.title} component="h4" />
            <pre>{content}</pre>
          </Event.Details>
        </Event>
      )
    }

    return children
  }
}

EventErrorBoundary.propTypes = {
  children: PropTypes.node.isRequired,
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool.isRequired,
}

export default EventErrorBoundary
