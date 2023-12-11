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

import React, { useCallback, useMemo } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import Events from '@console/components/events'

import Require from '@console/lib/components/require'

import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewGatewayEvents } from '@console/lib/feature-checks'

import {
  clearGatewayEventsStream,
  pauseGatewayEventsStream,
  resumeGatewayEventsStream,
  setGatewayEventsFilter,
} from '@console/store/actions/gateways'

import {
  selectGatewayEvents,
  selectGatewayEventsPaused,
  selectGatewayEventsTruncated,
  selectGatewayEventsFilter,
} from '@console/store/selectors/gateways'

const GatewayEvents = props => {
  const { gtwId, widget } = props

  const events = useSelector(state => selectGatewayEvents(state, gtwId))
  const paused = useSelector(state => selectGatewayEventsPaused(state, gtwId))
  const truncated = useSelector(state => selectGatewayEventsTruncated(state, gtwId))
  const filter = useSelector(state => selectGatewayEventsFilter(state, gtwId))

  const dispatch = useDispatch()

  const onClear = useCallback(() => {
    dispatch(clearGatewayEventsStream(gtwId))
  }, [dispatch, gtwId])

  const onPauseToggle = useCallback(
    paused => {
      if (paused) {
        dispatch(resumeGatewayEventsStream(gtwId))
        return
      }
      dispatch(pauseGatewayEventsStream(gtwId))
    },
    [dispatch, gtwId],
  )

  const onFilterChange = useCallback(
    filterId => {
      dispatch(setGatewayEventsFilter(gtwId, filterId))
    },
    [dispatch, gtwId],
  )

  const content = useMemo(() => {
    if (widget) {
      return (
        <Events.Widget
          events={events}
          entityId={gtwId}
          toAllUrl={`/gateways/${gtwId}/data`}
          scoped
        />
      )
    }

    return (
      <Events
        events={events}
        entityId={gtwId}
        paused={paused}
        onClear={onClear}
        onPauseToggle={onPauseToggle}
        onFilterChange={onFilterChange}
        truncated={truncated}
        filter={filter}
        scoped
      />
    )
  }, [events, filter, gtwId, onClear, onFilterChange, onPauseToggle, paused, truncated, widget])

  return <Require featureCheck={mayViewGatewayEvents}>{content}</Require>
}

GatewayEvents.propTypes = {
  gtwId: PropTypes.string.isRequired,
  widget: PropTypes.bool,
}

GatewayEvents.defaultProps = {
  widget: false,
}

export default GatewayEvents
