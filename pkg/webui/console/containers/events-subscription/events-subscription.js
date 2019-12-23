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

import Events from '../../../components/events'
import PropTypes from '../../../lib/prop-types'

const { Widget } = Events

class EventsSubscription extends React.Component {
  static propTypes = {
    error: PropTypes.error,
    events: PropTypes.events,
    id: PropTypes.string.isRequired,
    onClear: PropTypes.func,
    onRestart: PropTypes.func.isRequired,
    toAllUrl: PropTypes.string,
    widget: PropTypes.bool,
  }

  static defaultProps = {
    widget: false,
    onClear: () => null,
    toAllUrl: undefined,
    events: [],
    error: undefined,
  }

  render() {
    const { id, widget, events, onClear, toAllUrl, onRestart, error } = this.props

    if (widget) {
      return (
        <Widget
          emitterId={id}
          events={events}
          toAllUrl={toAllUrl}
          onRestart={onRestart}
          error={error}
        />
      )
    }

    return (
      <Events
        emitterId={id}
        events={events}
        onClear={onClear}
        onRestart={onRestart}
        error={error}
      />
    )
  }
}

export default EventsSubscription
