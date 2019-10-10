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
import List from '../../list'
import ErrorNotification from '../../error-notification'
import getEventComponentByName from '../../event/types'

import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'

import style from './widget.styl'

const m = defineMessages({
  latestEvents: 'Latest Events',
  seeAllActivity: 'See all activity',
  unknown: 'Unknown',
})

@bind
class EventsWidget extends React.PureComponent {
  renderEvent(event) {
    const { component: Component, type } = getEventComponentByName(event.name)

    return (
      <List.Item>
        <Component event={event} type={type} widget />
      </List.Item>
    )
  }

  render() {
    const { className, events, toAllUrl, emitterId, limit, error, onRestart } = this.props

    let truncatedEvents = events
    if (events.length > limit) {
      truncatedEvents = events.slice(0, limit)
    }

    return (
      <aside className={className}>
        <div className={style.header}>
          <Message className={style.headerTitle} content={m.latestEvents} />
          {!error && (
            <Link to={toAllUrl}>
              <Message className={style.seeAllMessage} content={m.seeAllActivity} />→
            </Link>
          )}
        </div>
        {error ? (
          <ErrorNotification
            small
            title={sharedMessages.eventsCannotShow}
            error={error}
            action={onRestart}
            actionMessage={sharedMessages.restartStream}
            buttonIcon="refresh"
          />
        ) : (
          <List
            bordered
            listClassName={style.list}
            size="small"
            items={truncatedEvents}
            renderItem={this.renderEvent}
            emptyMessage={sharedMessages.noEvents}
            emptyMessageValues={{ entityId: emitterId }}
          />
        )}
      </aside>
    )
  }
}

EventsWidget.propTypes = {
  className: PropTypes.string,
  /** An entity identifer. */
  emitterId: PropTypes.node.isRequired,
  error: PropTypes.error,
  /**
   * A list of events to be displayed in the widget. Events are expected
   * to be sorted in the descending order by their time.
   */
  events: PropTypes.events,
  /** The number of events to displayed in the widget. */
  limit: PropTypes.number,
  onRestart: PropTypes.func.isRequired,
  /** A url to the page with full version of the events component. */
  toAllUrl: PropTypes.string.isRequired,
}

EventsWidget.defaultProps = {
  className: undefined,
  events: [],
  limit: 5,
  error: undefined,
}

export default EventsWidget
