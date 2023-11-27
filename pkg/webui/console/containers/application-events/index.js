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

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import {
  clearApplicationEventsStream,
  pauseApplicationEventsStream,
  resumeApplicationEventsStream,
  setApplicationEventsFilter,
} from '@console/store/actions/applications'

import {
  selectApplicationEvents,
  selectApplicationEventsPaused,
  selectApplicationEventsTruncated,
  selectApplicationEventsFilter,
} from '@console/store/selectors/applications'

const ApplicationEvents = props => {
  const {
    appId,
    events,
    widget,
    paused,
    onClear,
    onPauseToggle,
    truncated,
    onFilterChange,
    filter,
  } = props

  const content = useMemo(() => {
    if (widget) {
      return (
        <Events.Widget entityId={appId} events={events} toAllUrl={`/applications/${appId}/data`} />
      )
    }

    return (
      <Events
        entityId={appId}
        events={events}
        paused={paused}
        onClear={onClear}
        truncated={truncated}
        filter={filter}
        onPauseToggle={onPauseToggle}
        onFilterChange={onFilterChange}
      />
    )
  }, [appId, events, filter, onClear, onFilterChange, onPauseToggle, paused, truncated, widget])

  return <Require featureCheck={mayViewApplicationEvents}>{content}</Require>
}

ApplicationEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  events: PropTypes.events,
  filter: PropTypes.eventFilter,
  onClear: PropTypes.func.isRequired,
  onFilterChange: PropTypes.func.isRequired,
  onPauseToggle: PropTypes.func.isRequired,
  paused: PropTypes.bool.isRequired,
  truncated: PropTypes.bool.isRequired,
  widget: PropTypes.bool,
}

ApplicationEvents.defaultProps = {
  widget: false,
  events: [],
  filter: undefined,
}

export default connect(
  (state, props) => {
    const { appId } = props

    return {
      events: selectApplicationEvents(state, appId),
      paused: selectApplicationEventsPaused(state, appId),
      truncated: selectApplicationEventsTruncated(state, appId),
      filter: selectApplicationEventsFilter(state, appId),
    }
  },
  (dispatch, ownProps) => ({
    onClear: () => dispatch(clearApplicationEventsStream(ownProps.appId)),
    onPauseToggle: paused =>
      paused
        ? dispatch(resumeApplicationEventsStream(ownProps.appId))
        : dispatch(pauseApplicationEventsStream(ownProps.appId)),
    onFilterChange: filterId => dispatch(setApplicationEventsFilter(ownProps.appId, filterId)),
  }),
)(ApplicationEvents)
