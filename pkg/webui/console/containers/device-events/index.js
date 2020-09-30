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

import Events from '@console/components/events'

import { getApplicationId, getDeviceId, combineDeviceIds } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'

import { clearDeviceEventsStream } from '@console/store/actions/devices'

import { selectDeviceEvents, selectDeviceEventsTruncated } from '@console/store/selectors/devices'

const DeviceEvents = props => {
  const { appId, devId, events, widget, onClear, truncated } = props

  if (widget) {
    return (
      <Events.Widget
        events={events}
        entityId={devId}
        toAllUrl={`/applications/${appId}/devices/${devId}/data`}
        scoped
      />
    )
  }

  return (
    <Events
      events={events}
      entityId={devId}
      onClear={onClear}
      truncated={truncated}
      scoped
      widget
    />
  )
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
  events: PropTypes.events,
  onClear: PropTypes.func.isRequired,
  truncated: PropTypes.bool.isRequired,
  widget: PropTypes.bool,
}

DeviceEvents.defaultProps = {
  widget: false,
  events: [],
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
      truncated: selectDeviceEventsTruncated(state, combinedId),
    }
  },
  (dispatch, ownProps) => {
    const { devIds } = ownProps

    return {
      onClear: () => dispatch(clearDeviceEventsStream(devIds)),
    }
  },
)(DeviceEvents)
