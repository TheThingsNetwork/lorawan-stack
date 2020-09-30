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

import React, { useState, useCallback } from 'react'
import classnames from 'classnames'

import hamburgerMenuClose from '@assets/misc/hamburger-menu-close.svg'

import Button from '@ttn-lw/components/button'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import EventsList from './list'
import EventDetails from './details'
import Widget from './widget'
import m from './messages'
import { getEventId } from './utils'

import style from './events.styl'

const Events = React.memo(({ events, scoped, onClear, entityId, truncated }) => {
  const [paused, setPause] = useState(false)
  const [focus, setFocus] = useState({ eventId: undefined, visible: false })
  const onPause = useCallback(() => setPause(paused => !paused), [])
  const handleRowClick = useCallback(
    eventId => {
      if (eventId !== focus.eventId) {
        setFocus({ eventId, visible: eventId })
      } else {
        setFocus({ eventId: undefined, visible: false })
      }
    },
    [focus],
  )

  const handleEventInfoCloseClick = useCallback(() => {
    setFocus({ entityId: undefined, visible: false })
  }, [])

  return (
    <div className={style.container}>
      <section className={style.header}>
        <div className={style.headerCells}>
          <Message content={sharedMessages.time} className={style.cellTime} component="div" />
          {!scoped && (
            <Message content={sharedMessages.entityId} className={style.cellId} component="div" />
          )}
          <Message content={sharedMessages.type} className={style.cellType} component="div" />
          <Message content={m.dataPreview} className={style.cellData} component="div" />
          <div className={style.stickyContainer}>
            <div className={style.actions}>
              <Button
                onClick={onPause}
                message={paused ? sharedMessages.resume : sharedMessages.pause}
                naked
                secondary
                icon={paused ? 'play_arrow' : 'pause'}
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
        </div>
      </section>
      <section className={style.body}>
        <EventsList
          events={events}
          paused={paused}
          scoped={scoped}
          entityId={entityId}
          onRowClick={handleRowClick}
          activeId={focus.eventId}
        />
      </section>
      {truncated && (
        <div className={style.truncated}>
          <Icon icon="info" />
          <Message content={m.eventsTruncated} />
        </div>
      )}
      <section className={classnames(style.sidebarContainer, { [style.expanded]: focus.visible })}>
        <div className={style.sidebarHeader}>
          <Message content={m.eventDetails} className={style.sidebarTitle} />
          <button className={style.sidebarCloseButton} onClick={handleEventInfoCloseClick}>
            <img src={hamburgerMenuClose} alt="Close event info" />
          </button>
        </div>
        <div className={style.sidebarContent}>
          {Boolean(focus.eventId) && (
            <EventDetails event={events.find(event => getEventId(event) === focus.eventId)} />
          )}
        </div>
      </section>
    </div>
  )
})

Events.propTypes = {
  entityId: PropTypes.string.isRequired,
  events: PropTypes.events.isRequired,
  onClear: PropTypes.func,
  scoped: PropTypes.bool,
  truncated: PropTypes.bool.isRequired,
}

Events.defaultProps = {
  scoped: false,
  onClear: () => null,
}

Events.Widget = Widget

export default Events
