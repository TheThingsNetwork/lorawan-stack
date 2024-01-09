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

import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import Events from '@console/components/events'

import { getApplicationId, getDeviceId, combineDeviceIds } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  clearDeviceEventsStream,
  pauseDeviceEventsStream,
  resumeDeviceEventsStream,
  setDeviceEventsFilter,
} from '@console/store/actions/devices'

import {
  selectDeviceEvents,
  selectDeviceEventsPaused,
  selectDeviceEventsTruncated,
  selectDeviceEventsFilter,
} from '@console/store/selectors/devices'

const DeviceEvents = props => {
  const { devIds, widget } = props

  const appId = getApplicationId(devIds)
  const devId = getDeviceId(devIds)
  const combinedId = combineDeviceIds(appId, devId)

  const events = useSelector(state => selectDeviceEvents(state, combinedId))
  const paused = useSelector(state => selectDeviceEventsPaused(state, combinedId))
  const truncated = useSelector(state => selectDeviceEventsTruncated(state, combinedId))
  const filter = useSelector(state => selectDeviceEventsFilter(state, combinedId))

  const dispatch = useDispatch()

  const onClear = useCallback(() => {
    dispatch(clearDeviceEventsStream(devIds))
  }, [devIds, dispatch])

  const onPauseToggle = useCallback(
    paused => {
      if (paused) {
        dispatch(resumeDeviceEventsStream(devIds))
        return
      }
      dispatch(pauseDeviceEventsStream(devIds))
    },
    [devIds, dispatch],
  )

  const onFilterChange = useCallback(
    filterId => {
      dispatch(setDeviceEventsFilter(devIds, filterId))
    },
    [devIds, dispatch],
  )

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
      paused={paused}
      filter={filter}
      onClear={onClear}
      onPauseToggle={onPauseToggle}
      onFilterChange={onFilterChange}
      truncated={truncated}
      scoped
      widget
    />
  )
}

DeviceEvents.propTypes = {
  devIds: PropTypes.shape({
    device_id: PropTypes.string,
    application_ids: PropTypes.shape({
      application_id: PropTypes.string,
    }),
  }).isRequired,
  widget: PropTypes.bool,
}

DeviceEvents.defaultProps = {
  widget: false,
}

export default DeviceEvents
