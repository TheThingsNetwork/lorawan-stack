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
import EventsSubscription from '../../containers/events-subscription'

import {
  clearGatewayEventsStream,
} from '../../store/actions/gateway'

import {
  selectGatewayEvents,
  selectGatewayEventsStatus,
} from '../../store/selectors/gateway'

@bind
class GatewayEvents extends React.Component {
  render () {
    const {
      gtwId,
      widget,
      onClear,
    } = this.props

    return (
      <EventsSubscription
        id={gtwId}
        widget={widget}
        eventsSelector={selectGatewayEvents}
        statusSelector={selectGatewayEventsStatus}
        onClear={onClear}
        toAllUrl={`/console/gateways/${gtwId}/data`}
      />
    )
  }
}

GatewayEvents.propTypes = {
  gtwId: PropTypes.string.isRequired,
  onClear: PropTypes.func.isRequired,
  widget: PropTypes.bool,
}

GatewayEvents.defaultProps = {
  widget: false,
}

export default connect(
  null,
  (dispatch, ownProps) => ({
    onClear: () => dispatch(clearGatewayEventsStream(ownProps.gtwId)),
  }))(GatewayEvents)
