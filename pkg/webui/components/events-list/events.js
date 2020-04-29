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

import PropTypes from '@ttn-lw/lib/prop-types'

import eventsContext from './context'

import style from './events.styl'

const Events = props => {
  const { children, events, renderEvent, onClear, onPause, widget, entityId } = props

  const [paused, setPaused] = React.useState(false)
  const handlePauseChange = React.useCallback(() => {
    setPaused(!paused)
    onPause(!paused)
  }, [onPause, setPaused, paused])

  return (
    <eventsContext.Provider
      value={{
        events,
        renderEvent,
        widget,
        paused,
        onPause: handlePauseChange,
        onClear,
        entityId,
      }}
    >
      <div className={style.eventsContainer}>{children}</div>
    </eventsContext.Provider>
  )
}

Events.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  entityId: PropTypes.string.isRequired,
  events: PropTypes.arrayOf(PropTypes.event),
  onClear: PropTypes.func,
  onPause: PropTypes.func,
  renderEvent: PropTypes.func.isRequired,
  widget: PropTypes.bool,
}

Events.defaultProps = {
  events: [],
  widget: false,
  onClear: () => null,
  onPause: () => null,
}

export default Events
