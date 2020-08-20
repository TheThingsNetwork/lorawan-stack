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
import { isEqual } from 'lodash'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import List from '../../list'
import eventsContext from '../context'

import EventErrorBoundary from './error-boundary'

import style from './list.styl'

const getEventKey = event => {
  return `${event.time}-${event.name}`
}

const ListContainer = React.memo(
  props => {
    const { className, widget, renderEvent, entityId, items } = props

    const cls = classnames(className, {
      [style.listContainer]: !widget,
      [style.listContainerWidget]: widget,
    })

    const listCls = classnames({
      [style.list]: !widget,
    })

    return (
      <List
        className={cls}
        listClassName={listCls}
        bordered={widget}
        size={widget ? 'small' : 'none'}
        items={items}
        rowKey={getEventKey}
        renderItem={renderEvent}
        emptyMessage={sharedMessages.noEvents}
        emptyMessageValues={{
          entityId,
          pre: msg => (
            <pre className={style.entityId} key="entity-id">
              {msg}
            </pre>
          ),
        }}
      />
    )
  },
  (prevProps, nextProps) => {
    const { paused: newPaused, items: newItems, ...newProps } = nextProps
    const { paused: oldPaused, items: oldItems, ...oldProps } = prevProps

    // Update if empty list and in the `paused` state.
    if (newItems.length === 0 && newPaused) {
      return false
    }

    // Do not update if in the `paused` state.
    if (!oldPaused && newPaused) {
      return true
    }

    // Always update on new events if not in the `paused` state.
    if (oldItems !== newItems) {
      return false
    }

    return isEqual(newProps, oldProps)
  },
)

ListContainer.propTypes = {
  className: PropTypes.string,
  entityId: PropTypes.string.isRequired,
  items: PropTypes.events.isRequired,
  paused: PropTypes.bool.isRequired,
  renderEvent: PropTypes.func.isRequired,
  widget: PropTypes.bool.isRequired,
}

ListContainer.defaultProps = {
  className: undefined,
}

const EventsList = props => {
  const { className } = props
  const { events, renderEvent, widget, entityId, paused } = React.useContext(eventsContext)

  const renderItem = React.useCallback(
    event => {
      return (
        <List.Item>
          <EventErrorBoundary event={event} widget={widget}>
            {renderEvent(event)}
          </EventErrorBoundary>
        </List.Item>
      )
    },
    [renderEvent, widget],
  )

  return (
    <ListContainer
      className={className}
      paused={paused}
      widget={widget}
      entityId={entityId}
      items={events}
      renderEvent={renderItem}
    />
  )
}

EventsList.propTypes = {
  className: PropTypes.string,
}

EventsList.defaultProps = {
  className: undefined,
}

export default EventsList
