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

import ErrorNotification from '@ttn-lw/components/error-notification'

import DeviceEventsList from '@console/components/events-list/application/device'

import { getApplicationId, getDeviceId, combineDeviceIds } from '@ttn-lw/lib/selectors/id'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { clearDeviceEventsStream, startDeviceEventsStream } from '@console/store/actions/devices'

import { selectDeviceEvents, selectDeviceEventsError } from '@console/store/selectors/devices'

const DeviceEvents = props => {
  const { appId, devId, events, error, onRestart, widget, onClear } = props

  if (error) {
    return (
      <ErrorNotification
        small
        title={sharedMessages.eventsCannotShow}
        content={error}
        action={onRestart}
        actionMessage={sharedMessages.restartStream}
        buttonIcon="refresh"
      />
    )
  }

  if (widget) {
    return (
      <DeviceEventsList.Widget
        events={events}
        toAllUrl={`/applications/${appId}/devices/${devId}/data`}
        deviceId={devId}
      />
    )
  }

  return <DeviceEventsList events={events} onClear={onClear} deviceId={devId} />
}

DeviceEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  devId: PropTypes.string.isRequired,
  devIds: PropTypes.shape({
    device_id: PropTypes.string,
    application_ids: PropTypes.shape({
      application_id: PropTypes.string,
    }),
  }).isRequired,
  error: PropTypes.error,
  events: PropTypes.events,
  onClear: PropTypes.func.isRequired,
  onRestart: PropTypes.func.isRequired,
  widget: PropTypes.bool,
}

DeviceEvents.defaultProps = {
  widget: false,
  events: [],
  error: undefined,
}

export default connect(
  (state, props) => {
    const { devIds } = props

    const appId = getApplicationId(devIds)
    const devId = getDeviceId(devIds)
    const combinedId = combineDeviceIds(appId, devId)

    return {
      devId,
      appId,
      events: selectDeviceEvents(state, combinedId),
      error: selectDeviceEventsError(state, combinedId),
    }
  },
  (dispatch, ownProps) => {
    const { devIds } = ownProps

    return {
      onClear: () => dispatch(clearDeviceEventsStream(devIds)),
      onRestart: () => dispatch(startDeviceEventsStream(devIds)),
    }
  },
)(DeviceEvents)
