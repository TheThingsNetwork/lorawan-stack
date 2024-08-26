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

import React, { useState, useCallback, useEffect, useRef } from 'react'
import classnames from 'classnames'
import { useSearchParams } from 'react-router-dom'

import EVENT_STORE_LIMIT from '@console/constants/event-store-limit'

import Icon, {
  IconInfoCircle,
  IconFileDownload,
  IconTrash,
  IconPlayerPlay,
  IconPlayerPause,
  IconX,
  IconExternalLink,
  IconClipboard,
  IconClipboardCheck,
} from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import Routes from '@ttn-lw/components/switch'

import Message from '@ttn-lw/lib/components/message'

import EventSplitFrameContext from '@console/containers/event-split-frame/context'

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
    entityId,
    truncated,
    darkTheme,
    framed,
    titleMessage,
    entityName,
    disableFiltering,
    filter,
    onFilterChange,
  }) => {
    const [focus, setFocus] = useState({ eventId: undefined, visible: false })
    const [xOffset, setXOffset] = useState(0)
    const [copyIcon, setCopyIcon] = useState(IconClipboard)
    const { setIsOpen } = React.useContext(EventSplitFrameContext)
    const onPause = useCallback(() => onPauseToggle(paused), [onPauseToggle, paused])
    const onExport = useCallback(() => {
      const eventLogData = composeDataUri(JSON.stringify(events, undefined, 2))
      downloadDataUriAsFile(eventLogData, `${entityId}_live_data_${Date.now()}.json`)
    }, [entityId, events])
    const [searchParams, setSearchParams] = useSearchParams()
    const firstRender = useRef(true)
    const selectedEvent = events.find(event => getEventId(event) === focus.eventId)

    useEffect(() => {
      if (copyIcon === IconClipboardCheck) {
        const timeout = setTimeout(() => {
          setCopyIcon(IconClipboard)
        }, 2000)
        return () => clearTimeout(timeout)
      }
    }, [copyIcon])

    useEffect(
      () => () => {
        if (firstRender.current && paused) {
          onPause()
        }
      },
      [onPause, paused],
    )
    useEffect(() => {
      const eventId = searchParams.get('eventId')
      if (!eventId) {
        setFocus({ eventId: undefined, visible: false })
        if (firstRender.current && paused) {
          onPause() // Resume stream
        }
        firstRender.current = false
        return
      }
      setFocus({ eventId, visible: eventId })
      if (firstRender.current && !paused) {
        // Make sure that the element is present in the DOM
        setTimeout(() => {
          const element = document.getElementById(eventId)
          element?.scrollIntoView({ behavior: 'smooth' })
        }, 200)

        onPause() // Pause stream
      }
    }, [onPause, paused, searchParams])

    const handleRowClick = useCallback(
      eventId => {
        if (eventId !== focus.eventId) {
          setSearchParams(
            {
              eventId,
            },
            { replace: true },
          )
        } else {
          setSearchParams({}, { replace: true })
        }
      },
      [focus.eventId, setSearchParams],
    )

    const handleEventInfoCloseClick = useCallback(() => {
      setSearchParams({}, { replace: true })
    }, [setSearchParams])

    const handleScroll = useCallback(event => {
      const xOffset = event.target.scrollLeft
      setXOffset(xOffset)
    }, [])

    const handleCopyClick = useCallback(() => {
      if (selectedEvent) {
        navigator.clipboard.writeText(JSON.stringify(selectedEvent, undefined, 2))
        setCopyIcon(IconClipboardCheck)
      }
    }, [selectedEvent])

    const handleVerboseFilterChange = useCallback(() => {
      onFilterChange(Boolean(filter) ? undefined : 'default')
    }, [onFilterChange, filter])

    return (
      <div
        className={classnames(style.container, {
          [style.themeDark]: darkTheme,
          [style.framed]: framed,
        })}
      >
        <section className={style.header}>
          <div className={style.headerCells} style={{ left: xOffset * -1 }}>
            <Message content={sharedMessages.time} className={style.cellTime} component="div" />
            {!scoped && (
              <Message content={sharedMessages.entityId} className={style.cellId} component="div" />
            )}
            <Message content={sharedMessages.type} className={style.cellType} component="div" />
            <Message content={m.dataPreview} className={style.cellData} component="div" />
          </div>
          {!framed && (
            <div className={style.stickyContainer}>
              <div className={style.actions}>
                <>
                  {!disableFiltering && (
                    <label className={style.toggleContainer}>
                      <Message content={m.verboseStream} className={style.toggleLabel} />
                      <Routes onChange={handleVerboseFilterChange} checked={!Boolean(filter)} />
                    </label>
                  )}
                  <Button
                    onClick={onExport}
                    message={sharedMessages.exportJson}
                    secondary
                    icon={IconFileDownload}
                    small
                  />
                  <Button
                    onClick={onPause}
                    message={paused ? sharedMessages.resume : sharedMessages.pause}
                    className={style.pauseButton}
                    secondary
                    warning={paused}
                    icon={paused ? IconPlayerPlay : IconPlayerPause}
                    small
                  />
                  <Button
                    onClick={onClear}
                    message={sharedMessages.clear}
                    secondary
                    icon={IconTrash}
                    small
                  />
                </>
              </div>
            </div>
          )}
          {framed && (
            <div className={style.actions}>
              <div className="d-inline-flex al-center mr-cs-s c-text-neutral-light">
                <Message
                  component="span"
                  className="md:d-none fs-s"
                  content={titleMessage}
                  values={{
                    entityName,
                    strong: msg => <strong>{msg}</strong>,
                  }}
                />
              </div>
              <div className={style.buttonBar}>
                <Button.Link
                  className={style.framedButton}
                  icon={IconExternalLink}
                  to="data"
                  tooltip={m.goToLiveData}
                />
                <Button
                  className={style.framedButton}
                  onClick={onPause}
                  warning={paused}
                  icon={paused ? IconPlayerPlay : IconPlayerPause}
                  tooltip={paused ? sharedMessages.resume : sharedMessages.pause}
                />
                <Button
                  className={style.framedButton}
                  onClick={() => setIsOpen(false)}
                  icon={IconX}
                />
              </div>
            </div>
          )}
        </section>
        <section className={style.body}>
          <EventsList
            events={events}
            paused={paused}
            scoped={scoped}
            entityId={entityId}
            onRowClick={handleRowClick}
            activeId={focus.eventId}
            onScroll={handleScroll}
          />
        </section>
        {truncated && (
          <div className={style.truncated}>
            <Icon icon={IconInfoCircle} />
            <Message content={m.eventsTruncated} values={{ limit: EVENT_STORE_LIMIT }} />
          </div>
        )}
        <section
          className={classnames(style.sidebarContainer, { [style.expanded]: focus.visible })}
        >
          <div className={style.sidebarHeader}>
            <Message content={m.eventDetails} className={style.sidebarTitle} />
            <div className="j-end d-flex mr-cs-xs">
              {Boolean(selectedEvent) && !selectedEvent.isSynthetic && (
                <Button
                  icon={copyIcon}
                  className={style.sidebarButton}
                  tabIndex={focus.visible ? '0' : '-1'}
                  onClick={handleCopyClick}
                />
              )}
              <Button
                icon={IconX}
                className={style.sidebarButton}
                tabIndex={focus.visible ? '0' : '-1'}
                onClick={handleEventInfoCloseClick}
              />
            </div>
          </div>
          <div className={style.sidebarContent}>
            {Boolean(focus.eventId) && (
              <EventDetails
                event={events.find(event => getEventId(event) === focus.eventId)}
                darkTheme={darkTheme}
              />
            )}
          </div>
        </section>
      </div>
    )
  },
)

Events.propTypes = {
  darkTheme: PropTypes.bool,
  disableFiltering: PropTypes.bool,
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
  events: PropTypes.events.isRequired,
  filter: PropTypes.string,
  framed: PropTypes.bool,
  onClear: PropTypes.func,
  onFilterChange: PropTypes.func,
  onPauseToggle: PropTypes.func,
  paused: PropTypes.bool.isRequired,
  scoped: PropTypes.bool,
  titleMessage: PropTypes.message,
  truncated: PropTypes.bool.isRequired,
}

Events.defaultProps = {
  darkTheme: false,
  entityName: undefined,
  framed: false,
  scoped: false,
  titleMessage: undefined,
  onClear: () => null,
  onPauseToggle: () => null,
  onFilterChange: () => null,
  filter: undefined,
  disableFiltering: false,
}

Events.Widget = Widget

export default Events
