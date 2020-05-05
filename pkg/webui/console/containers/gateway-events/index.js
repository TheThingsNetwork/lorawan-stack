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

import GatewayEventsList from '@console/components/events-list/gateway'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewGatewayEvents } from '@console/lib/feature-checks'

import { clearGatewayEventsStream, startGatewayEventsStream } from '@console/store/actions/gateways'

import { selectGatewayEvents, selectGatewayEventsError } from '@console/store/selectors/gateways'

const GatewayEvents = props => {
  const { gtwId, events, error, onRestart, widget, onClear } = props

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
      <GatewayEventsList.Widget
        events={events}
        toAllUrl={`/gateways/${gtwId}/data`}
        gtwId={gtwId}
      />
    )
  }

  return <GatewayEventsList events={events} onClear={onClear} gtwId={gtwId} />
}

GatewayEvents.propTypes = {
  error: PropTypes.error,
  events: PropTypes.events,
  gtwId: PropTypes.string.isRequired,
  onClear: PropTypes.func.isRequired,
  onRestart: PropTypes.func.isRequired,
  widget: PropTypes.bool,
}

GatewayEvents.defaultProps = {
  widget: false,
  events: [],
  error: undefined,
}

export default withFeatureRequirement(mayViewGatewayEvents)(
  connect(
    (state, props) => {
      const { gtwId } = props

      return {
        events: selectGatewayEvents(state, gtwId),
        error: selectGatewayEventsError(state, gtwId),
      }
    },
    (dispatch, ownProps) => ({
      onClear: () => dispatch(clearGatewayEventsStream(ownProps.gtwId)),
      onRestart: () => dispatch(startGatewayEventsStream(ownProps.gtwId)),
    }),
  )(GatewayEvents),
)
