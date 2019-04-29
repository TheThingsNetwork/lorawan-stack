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

import Button from '../button'
import Message from '../../lib/components/message'
import List from '../list'
import getEventComponentByName from '../event/types'
import sharedMessages from '../../lib/shared-messages'
import PropTypes from '../../lib/prop-types'
import EventsWidget from './widget'

import style from './events.styl'

@bind
class Events extends React.PureComponent {

  renderEvent (event) {
    const { component: Component, type } = getEventComponentByName(event.name)

    return (
      <List.Item className={style.listItem}>
        <Component
          overviewClassName={style.event}
          expandedClassName={style.eventData}
          event={event}
          type={type}
        />
      </List.Item>
    )
  }

  onPause () {
    const { onPause } = this.props

    onPause()
  }

  onClear () {
    const { onClear } = this.props

    onClear()
  }

  getEventkey (event) {
    return `${event.time}-${event.name}`
  }

  render () {
    const {
      className,
      events,
      paused,
      onClear,
      onPause,
      emitterId,
    } = this.props

    const header = (
      <Header
        paused={paused}
        onPause={onPause}
        onClear={onClear}
      />
    )

    return (
      <List
        bordered
        size="none"
        className={className}
        listClassName={style.list}
        header={header}
        items={events}
        renderItem={this.renderEvent}
        rowKey={this.getEventkey}
        emptyMessage={sharedMessages.noEvents}
        emptyMessageValues={{ entityId: emitterId }}
      />
    )
  }
}

Events.propTypes = {
  events: PropTypes.arrayOf(PropTypes.event),
  paused: PropTypes.bool.isRequired,
  emitterId: PropTypes.string.isRequired,
  onPause: PropTypes.func.isRequired,
  onClear: PropTypes.func.isRequired,
}

Events.defaultProps = {
  events: [],
}

Events.Widget = EventsWidget

const Header = function (props) {
  const {
    paused,
    onPause,
    onClear,
  } = props

  const pauseMessage = paused ? sharedMessages.resume : sharedMessages.pause
  const pauseIcon = paused ? 'play_arrow' : 'pause'

  return (
    <div className={style.header}>
      <div className={style.headerColumns}>
        <Message className={style.headerColumnsTime} content={sharedMessages.time} />
        <Message className={style.headerColumnsId} content={sharedMessages.entityId} />
        <Message content={sharedMessages.data} />
      </div>
      <div className={style.headerActions}>
        <Button
          onClick={onPause}
          message={pauseMessage}
          naked
          secondary
          icon={pauseIcon}
        />
        <Button
          onClick={onClear}
          message={sharedMessages.clear}
          naked
          secondary
          icon="delete"
        />
      </div>
    </div>
  )
}

export default Events
