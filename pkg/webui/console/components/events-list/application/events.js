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

import Events from '@ttn-lw/components/events-list'

import PropTypes from '@ttn-lw/lib/prop-types'

import Column from '../shared/components/columns'

import renderEvent from './render-event'

const ApplicationEvents = props => {
  const { events, onClear, onPause, appId } = props

  return (
    <Events
      events={events}
      renderEvent={renderEvent}
      onClear={onClear}
      onPause={onPause}
      entityId={appId}
    >
      <Events.Header>
        <Column.Time />
        <Column.ID />
        <Column.Type />
        <Column.Data />
      </Events.Header>
      <Events.List />
    </Events>
  )
}

ApplicationEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  events: PropTypes.arrayOf(PropTypes.event).isRequired,
  onClear: PropTypes.func.isRequired,
  onPause: PropTypes.func,
}

ApplicationEvents.defaultProps = {
  onPause: () => null,
}

export default ApplicationEvents
