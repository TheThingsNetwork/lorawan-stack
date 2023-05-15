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

import EVENT_STORE_LIMIT from '@console/constants/event-store-limit'
import hamburgerMenuClose from '@assets/misc/hamburger-menu-close.svg'

import Button from '@ttn-lw/components/button'
import Routes from '@ttn-lw/components/switch'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { composeDataUri, downloadDataUriAsFile } from '@ttn-lw/lib/data-uri'

import EventsList from './list'
import EventDetails from './details'
import Widget from './widget'
import m from './messages'
import { getEventId } from './utils'

import style from './events.styl'

const Events = React.memo(
  ({
    events,
    scoped,
    paused,
    onClear,
    onPauseToggle,
    onFilterChange,
    entityId,
    truncated,
    filter,
    disableFiltering,
  }) => {
    const [focus, setFocus] = useState({ eventId: undefined, visible: false })
    const onPause = useCallback(() => onPauseToggle(paused), [onPauseToggle, paused])
    const onExport = useCallback(() => {
      const eventLogData = composeDataUri(JSON.stringify(events, undefined, 2))
      downloadDataUriAsFile(eventLogData, `${entityId}_live_data_${Date.now()}.json`)
    }, [entityId, events])
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

    const handleVerboseFilterChange = useCallback(() => {
      onFilterChange(Boolean(filter) ? undefined : 'default')
    }, [onFilterChange, filter])

    const handleEventInfoCloseClick = useCallback(() => {
      setFocus({ eventId: undefined, visible: false })
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
                {!disableFiltering && (
                  <label className={style.toggleContainer}>
                    <Message content={m.verboseStream} className={style.toggleLabel} />
                    <Routes onChange={handleVerboseFilterChange} checked={!Boolean(filter)} />
                  </label>
                )}
                <Button
                  onClick={onExport}
                  message={sharedMessages.exportJson}
                  naked
                  icon="file_download"
                />
                <Button
                  onClick={onPause}
                  message={paused ? sharedMessages.resume : sharedMessages.pause}
                  naked
                  warning={paused}
                  icon={paused ? 'play_arrow' : 'pause'}
                />
                <Button onClick={onClear} message={sharedMessages.clear} naked icon="delete" />
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
            <Message content={m.eventsTruncated} values={{ limit: EVENT_STORE_LIMIT }} />
          </div>
        )}
        <section
          className={classnames(style.sidebarContainer, { [style.expanded]: focus.visible })}
        >
          <div className={style.sidebarHeader}>
            <Message content={m.eventDetails} className={style.sidebarTitle} />
            <button
              className={style.sidebarCloseButton}
              onClick={handleEventInfoCloseClick}
              tabIndex={focus.visible ? '0' : '-1'}
            >
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
  },
)

Events.propTypes = {
  disableFiltering: PropTypes.bool,
  entityId: PropTypes.string.isRequired,
  events: PropTypes.events.isRequired,
  filter: PropTypes.eventFilter,
  onClear: PropTypes.func,
  onFilterChange: PropTypes.func,
  onPauseToggle: PropTypes.func,
  paused: PropTypes.bool.isRequired,
  scoped: PropTypes.bool,
  truncated: PropTypes.bool.isRequired,
}

Events.defaultProps = {
  disableFiltering: false,
  filter: undefined,
  scoped: false,
  onClear: () => null,
  onPauseToggle: () => null,
  onFilterChange: () => null,
}

Events.Widget = Widget

export default Events
