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

import React, { useRef, useState, useLayoutEffect, useCallback } from 'react'
import { FixedSizeList as List } from 'react-window'
import AutoSizer from 'react-virtualized-auto-sizer'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import Event from './event'
import { getEventId } from './utils'

import style from './events.styl'

const EmptyMessage = ({ entityId }) => (
  <div className={style.emptyMessageContainer}>
    <Message
      className={style.emptyMessageContent}
      content={sharedMessages.noEvents}
      values={{ pre: content => <pre key="entity-id">{content}</pre>, entityId }}
    />
  </div>
)

EmptyMessage.propTypes = {
  entityId: PropTypes.string.isRequired,
}

const EventsList = React.memo(({ events, scoped, entityId, onRowClick, activeId }) => {
  const outerListRef = useRef(undefined)
  const innerListRef = useRef(undefined)
  const [scrollOffset, setScrollOffset] = useState(0)
  const listHeight = 150
  const minHeight = 0.1
  const pageOffset = listHeight * 5
  const maxHeight =
    (innerListRef.current && innerListRef.current.style.height.replace('px', '')) || listHeight

  const keys = {
    PageUp: Math.max(minHeight, scrollOffset - pageOffset),
    PageDown: Math.min(scrollOffset + pageOffset, maxHeight),
    End: maxHeight,
    Home: minHeight,
  }

  const handleKeyDown = useCallback(
    ({ key }) => {
      if (keys[key]) {
        setScrollOffset(keys[key])
      }
    },
    [keys],
  )

  useLayoutEffect(() => {
    if (outerListRef.current) {
      outerListRef.current.scrollTo({
        left: 0,
        top: scrollOffset,
        behavior: 'auto',
      })
    }
  })

  if (!events.length) {
    return <EmptyMessage entityId={entityId} />
  }

  return (
    <AutoSizer>
      {({ height, width }) => (
        <ol onKeyDown={handleKeyDown} tabIndex="0">
          <List
            height={height}
            width={width}
            itemCount={events.length}
            itemSize={40}
            overscanCount={25}
            outerRef={outerListRef}
            innerRef={innerListRef}
          >
            {({ index, style }) => {
              const eventId = getEventId(events[index])
              return (
                <Event
                  event={events[index]}
                  eventId={eventId}
                  key={eventId}
                  scoped={scoped}
                  rowStyle={style}
                  onRowClick={onRowClick}
                  index={index}
                  active={eventId === activeId}
                />
              )
            }}
          </List>
        </ol>
      )}
    </AutoSizer>
  )
})

EventsList.propTypes = {
  activeId: PropTypes.string,
  entityId: PropTypes.string.isRequired,
  events: PropTypes.events.isRequired,
  onRowClick: PropTypes.func.isRequired,
  scoped: PropTypes.bool,
}

EventsList.defaultProps = {
  activeId: undefined,
  scoped: false,
}

export { EventsList as default, EmptyMessage }
