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

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import { clearApplicationEventsStream } from '@console/store/actions/applications'

import {
  selectApplicationEvents,
  selectApplicationEventsTruncated,
} from '@console/store/selectors/applications'

const ApplicationEvents = props => {
  const { appId, events, widget, onClear, truncated } = props

  if (widget) {
    return (
      <Events.Widget entityId={appId} events={events} toAllUrl={`/applications/${appId}/data`} />
    )
  }

  return <Events entityId={appId} events={events} onClear={onClear} truncated={truncated} />
}

ApplicationEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  events: PropTypes.events,
  onClear: PropTypes.func.isRequired,
  truncated: PropTypes.bool.isRequired,
  widget: PropTypes.bool,
}

ApplicationEvents.defaultProps = {
  widget: false,
  events: [],
}

export default withFeatureRequirement(mayViewApplicationEvents)(
  connect(
    (state, props) => {
      const { appId } = props

      return {
        events: selectApplicationEvents(state, appId),
        truncated: selectApplicationEventsTruncated(state, appId),
      }
    },
    (dispatch, ownProps) => ({
      onClear: () => dispatch(clearApplicationEventsStream(ownProps.appId)),
    }),
  )(ApplicationEvents),
)
