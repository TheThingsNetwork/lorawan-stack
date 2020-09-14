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
import classnames from 'classnames'

import WidgetContainer from '@ttn-lw/components/widget-container'
import Status from '@ttn-lw/components/status'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import Event from './event'
import { EmptyMessage } from './list'
import { getEventId } from './utils'
import m from './messages'

import style from './events.styl'

const EventsWidget = ({ toAllUrl, className, events, scoped, entityId }) => {
  const title = (
    <Status flipped status="good">
      <Message content={sharedMessages.liveData} className={style.widgetHeaderTitle} />
    </Status>
  )
  return (
    <WidgetContainer
      title={title}
      toAllUrl={toAllUrl}
      linkMessage={m.seeAllActivity}
      className={className}
    >
      <div className={classnames(style.body, style.widgetContainer)}>
        {events.length === 0 ? (
          <EmptyMessage entityId={entityId} />
        ) : (
          <ol>
            {events.slice(0, 6).map(event => {
              const eventId = getEventId(event)
              return <Event event={event} eventId={eventId} key={eventId} scoped={scoped} widget />
            })}
          </ol>
        )}
      </div>
    </WidgetContainer>
  )
}

EventsWidget.propTypes = {
  className: PropTypes.string,
  entityId: PropTypes.string.isRequired,
  events: PropTypes.events.isRequired,
  scoped: PropTypes.bool,
  toAllUrl: PropTypes.string.isRequired,
}
EventsWidget.defaultProps = { className: undefined, scoped: false }

export default EventsWidget
