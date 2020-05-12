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

import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Events from '../events'

import style from './widget.styl'

const m = defineMessages({
  latestEvents: 'Latest events',
  seeAllActivity: 'See all activity',
})

const EventsWidget = props => {
  const { className, events, renderEvent, toAllUrl, limit, entityId } = props

  let truncatedEvents = events
  if (events.length > limit) {
    truncatedEvents = events.slice(0, limit)
  }

  return (
    <aside className={className}>
      <Events events={truncatedEvents} renderEvent={renderEvent} entityId={entityId} widget>
        <Events.Header className={style.header}>
          <Message content={m.latestEvents} component="b" />
          <Link className={style.seeAllLink} secondary to={toAllUrl}>
            <Message content={m.seeAllActivity} /> →
          </Link>
        </Events.Header>
        <Events.List />
      </Events>
    </aside>
  )
}

EventsWidget.propTypes = {
  className: PropTypes.string,
  entityId: PropTypes.string.isRequired,
  events: PropTypes.events,
  limit: PropTypes.number,
  renderEvent: PropTypes.func.isRequired,
  toAllUrl: PropTypes.string.isRequired,
}

EventsWidget.defaultProps = {
  className: undefined,
  events: [],
  limit: 5,
}

export default EventsWidget
