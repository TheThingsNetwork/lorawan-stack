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

import Events from '../../../components/events'
import PropTypes from '../../../lib/prop-types'

const { Widget } = Events

@connect(function(state, props) {
  const { id, eventsSelector, statusSelector, errorSelector } = props

  return {
    events: eventsSelector(state, id),
    connectionStatus: statusSelector(state, id),
    error: errorSelector(state, id),
  }
})
class EventsSubscription extends React.Component {
  render() {
    const { id, widget, events, onClear, toAllUrl, error } = this.props

    if (widget) {
      return <Widget emitterId={id} events={events} toAllUrl={toAllUrl} error={error} />
    }

    return <Events emitterId={id} events={events} onClear={onClear} error={error} />
  }
}

EventsSubscription.propTypes = {
  errorSelector: PropTypes.func,
  eventsSelector: PropTypes.func.isRequired,
  id: PropTypes.string.isRequired,
  onClear: PropTypes.func,
  toAllUrl: PropTypes.string,
  widget: PropTypes.bool,
}

EventsSubscription.defaultProps = {
  widget: false,
  onClear: () => null,
  errorSelector: () => undefined,
  toAllUrl: null,
}

export default EventsSubscription
