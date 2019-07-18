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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Link from '../../link'
import Message from '../../../lib/components/message'
import Status from '../../status'
import List from '../../list'
import Notification from '../../notification'
import PropTypes from '../../../lib/prop-types'
import getEventComponentByName from '../../event/types'

import DateTime from '../../../lib/components/date-time'
import sharedMessages from '../../../lib/shared-messages'
import style from './widget.styl'

const m = defineMessages({
  latestEvents: 'Latest Events',
  seeAllActivity: 'See all activity',
  unknown: 'Unknown',
})

@bind
class EventsWidget extends React.PureComponent {

  renderEvent (event) {
    const { component: Component, type } = getEventComponentByName(event.name)

    return (
      <List.Item>
        <Component
          event={event}
          type={type}
          widget
        />
      </List.Item>
    )
  }

  render () {
    const {
      className,
      events,
      toAllUrl,
      emitterId,
      connectionStatus,
      limit,
      error,
    } = this.props

    let latestActivityTime = null
    if (events.length) {
      const latestEvent = events[0]
      latestActivityTime = <DateTime.Relative value={latestEvent.time} />
    } else {
      latestActivityTime = <Message content={m.unknown} />
    }

    const statusMessage = (
      <span>
        <Message
          className={style.statusMessage}
          content={m.latestEvents}
        />
        {latestActivityTime}
      </span>
    )

    let truncatedEvents = events
    if (events.length > limit) {
      truncatedEvents = events.slice(0, limit)
    }

    return (
      <aside className={className}>
        <div className={style.header}>
          <Status
            label={statusMessage}
            status={connectionStatus}
          />
          {!error && (
            <Link to={toAllUrl}>
              <Message
                className={style.seeAllMessage}
                content={m.seeAllActivity}
              />
              →
            </Link>
          )}
        </div>
        {error
          ? <Notification small title={sharedMessages.eventsCannotShow} error={error} />
          : (
            <List
              bordered
              listClassName={style.list}
              size="small"
              items={truncatedEvents}
              renderItem={this.renderEvent}
              emptyMessage={sharedMessages.noEvents}
              emptyMessageValues={{ entityId: emitterId }}
            />
          )
        }
      </aside>
    )
  }
}

EventsWidget.propTypes = {
  /**
   * A list of events to be displayed in the widget. Events are expected
   * to be sorted in the descending order by their time.
   */
  events: PropTypes.array,
  /** A url to the page with full version of the events component. */
  toAllUrl: PropTypes.string.isRequired,
  /** An entity identifer. */
  emitterId: PropTypes.node.isRequired,
  /** A current status of the network. */
  connectionStatus: PropTypes.oneOf([ 'good', 'bad', 'mediocre', 'unknown' ]).isRequired,
  /** The number of events to displayed in the widget. */
  limit: PropTypes.number,
}

EventsWidget.defaultProps = {
  events: [],
  limit: 5,
}

const CONNECTION_STATUS = Object.freeze({
  GOOD: 'good',
  BAD: 'bad',
  MEDIOCRE: 'mediocre',
  UNKNOWN: 'unknown',
})

EventsWidget.CONNECTION_STATUS = CONNECTION_STATUS

export default EventsWidget
