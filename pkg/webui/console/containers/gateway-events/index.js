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

import React, { useMemo } from 'react'
import { connect } from 'react-redux'

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
  const {
    gtwId,
    events,
    widget,
    paused,
    onPauseToggle,
    onClear,
    onFilterChange,
    truncated,
    filter,
  } = props

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
  events: PropTypes.events,
  filter: PropTypes.eventFilter,
  gtwId: PropTypes.string.isRequired,
  onClear: PropTypes.func.isRequired,
  onFilterChange: PropTypes.func.isRequired,
  onPauseToggle: PropTypes.func.isRequired,
  paused: PropTypes.bool.isRequired,
  truncated: PropTypes.bool.isRequired,
  widget: PropTypes.bool,
}

GatewayEvents.defaultProps = {
  widget: false,
  events: [],
  filter: undefined,
}

export default connect(
  (state, props) => {
    const { gtwId } = props

    return {
      events: selectGatewayEvents(state, gtwId),
      paused: selectGatewayEventsPaused(state, gtwId),
      truncated: selectGatewayEventsTruncated(state, gtwId),
      filter: selectGatewayEventsFilter(state, gtwId),
    }
  },
  (dispatch, ownProps) => ({
    onClear: () => dispatch(clearGatewayEventsStream(ownProps.gtwId)),
    onPauseToggle: paused =>
      paused
        ? dispatch(resumeGatewayEventsStream(ownProps.gtwId))
        : dispatch(pauseGatewayEventsStream(ownProps.gtwId)),
    onFilterChange: filterId => dispatch(setGatewayEventsFilter(ownProps.gtwId, filterId)),
  }),
)(GatewayEvents)
