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
import { connect } from 'react-redux'

import CONNECTION_STATUS from '../../constants/connection-status'
import Events from '../../../components/events'
import PropTypes from '../../../lib/prop-types'

const { Widget } = Events

const mapConnectionStatusToWidget = function (status) {
  switch (status) {
  case CONNECTION_STATUS.CONNECTED:
    return Widget.CONNECTION_STATUS.GOOD
  case CONNECTION_STATUS.CONNECTING:
    return Widget.CONNECTION_STATUS.MEDIOCRE
  case CONNECTION_STATUS.DISCONNECTED:
    return Widget.CONNECTION_STATUS.BAD
  case CONNECTION_STATUS.UNKNOWN:
  default:
    return Widget.CONNECTION_STATUS.UNKNOWN
  }
}

@connect(function (state, props) {
  const {
    eventsSelector,
    statusSelector,
  } = props

  return {
    events: eventsSelector(state, props),
    connectionStatus: statusSelector(state, props),
  }
})
class EventsSubscription extends React.Component {
  render () {
    const {
      id,
      widget,
      events,
      connectionStatus,
      onClear,
      toAllUrl,
    } = this.props

    if (widget) {
      return (
        <Widget
          emitterId={id}
          events={events}
          connectionStatus={mapConnectionStatusToWidget(connectionStatus)}
          toAllUrl={toAllUrl}
        />
      )
    }

    return (
      <Events
        emitterId={id}
        events={events}
        onClear={onClear}
      />
    )
  }
}

EventsSubscription.propTypes = {
  id: PropTypes.string.isRequired,
  eventsSelector: PropTypes.func.isRequired,
  statusSelector: PropTypes.func,
  onClear: PropTypes.func,
  widget: PropTypes.bool,
  toAllUrl: PropTypes.string,
}

EventsSubscription.defaultProps = {
  widget: false,
  onClear: () => null,
  statusSelector: () => 'unknown',
  toAllUrl: null,
}

export default EventsSubscription
