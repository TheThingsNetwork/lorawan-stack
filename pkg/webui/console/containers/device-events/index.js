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
import bind from 'autobind-decorator'

import PropTypes from '../../../lib/prop-types'
import { getApplicationId, getDeviceId } from '../../../lib/selectors/id'
import EventsSubscription from '../../containers/events-subscription'

import { clearDeviceEventsStream, startDeviceEventsStream } from '../../store/actions/device'

import {
  selectDeviceEvents,
  selectDeviceEventsStatus,
  selectDeviceEventsError,
} from '../../store/selectors/device'

@connect(
  null,
  (dispatch, ownProps) => ({
    onClear: () => dispatch(clearDeviceEventsStream(ownProps.devIds)),
    onRestart: () => dispatch(startDeviceEventsStream(ownProps.devIds)),
  }),
)
@bind
class DeviceEvents extends React.Component {
  render() {
    const { devIds, widget, onClear, onRestart } = this.props

    const devId = getDeviceId(devIds)
    const appId = getApplicationId(devIds)

    return (
      <EventsSubscription
        id={devId}
        widget={widget}
        eventsSelector={selectDeviceEvents}
        statusSelector={selectDeviceEventsStatus}
        errorSelector={selectDeviceEventsError}
        onClear={onClear}
        onRestart={onRestart}
        toAllUrl={`/applications/${appId}/devices/${devId}/data`}
      />
    )
  }
}

DeviceEvents.propTypes = {
  devIds: PropTypes.object.isRequired,
  onClear: PropTypes.func.isRequired,
  onRestart: PropTypes.func.isRequired,
  widget: PropTypes.bool,
}

DeviceEvents.defaultProps = {
  widget: false,
}

export default DeviceEvents
