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

import PropTypes from '../../../lib/prop-types'
import EventsSubscription from '../../containers/events-subscription'
import { startOrganizationEventsStream } from '../../store/actions/organizations'
@connect(
  null,
  (dispatch, ownProps) => ({
    onRestart: () => dispatch(startOrganizationEventsStream(ownProps.orgId)),
  }),
)
export default class OrganizationEvents extends React.Component {
  static propTypes = {
    errorSelector: PropTypes.func.isRequired,
    eventsSelector: PropTypes.func.isRequired,
    onClear: PropTypes.func.isRequired,
    onRestart: PropTypes.func.isRequired,
    orgId: PropTypes.string.isRequired,
    statusSelector: PropTypes.func.isRequired,
    widget: PropTypes.bool,
  }

  static defaultProps = {
    widget: false,
  }

  render() {
    const {
      orgId,
      widget,
      onClear,
      eventsSelector,
      errorSelector,
      statusSelector,
      onRestart,
    } = this.props

    return (
      <EventsSubscription
        id={orgId}
        widget={widget}
        eventsSelector={eventsSelector}
        statusSelector={statusSelector}
        errorSelector={errorSelector}
        onClear={onClear}
        onRestart={onRestart}
        toAllUrl={`/organizations/${orgId}/data`}
      />
    )
  }
}
