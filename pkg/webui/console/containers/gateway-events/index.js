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

import { mayViewGatewayEvents } from '@console/lib/feature-checks'

import { clearGatewayEventsStream } from '@console/store/actions/gateways'

import {
  selectGatewayEvents,
  selectGatewayEventsTruncated,
} from '@console/store/selectors/gateways'

const GatewayEvents = props => {
  const { gtwId, events, widget, onClear, truncated } = props

  if (widget) {
    return (
      <Events.Widget events={events} entityId={gtwId} toAllUrl={`/gateways/${gtwId}/data`} scoped />
    )
  }

  return <Events events={events} entityId={gtwId} onClear={onClear} truncated={truncated} scoped />
}

GatewayEvents.propTypes = {
  events: PropTypes.events,
  gtwId: PropTypes.string.isRequired,
  onClear: PropTypes.func.isRequired,
  truncated: PropTypes.bool.isRequired,
  widget: PropTypes.bool,
}

GatewayEvents.defaultProps = {
  widget: false,
  events: [],
}

export default withFeatureRequirement(mayViewGatewayEvents)(
  connect(
    (state, props) => {
      const { gtwId } = props

      return {
        events: selectGatewayEvents(state, gtwId),
        truncated: selectGatewayEventsTruncated(state, gtwId),
      }
    },
    (dispatch, ownProps) => ({
      onClear: () => dispatch(clearGatewayEventsStream(ownProps.gtwId)),
    }),
  )(GatewayEvents),
)
