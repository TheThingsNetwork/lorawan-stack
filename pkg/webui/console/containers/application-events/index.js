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
  clearApplicationEventsStream,
} from '../../store/actions/application'

import {
  selectApplicationEvents,
  selectApplicationEventsStatus,
} from '../../store/selectors/applications'

@connect(
  null,
  (dispatch, ownProps) => ({
    onClear: () => dispatch(clearApplicationEventsStream(ownProps.appId)),
  }))
@bind
class ApplicationEvents extends React.Component {
  render () {
    const {
      appId,
      widget,
      onClear,
    } = this.props

    return (
      <EventsSubscription
        id={appId}
        widget={widget}
        eventsSelector={selectApplicationEvents}
        statusSelector={selectApplicationEventsStatus}
        onClear={onClear}
        toAllUrl={`/console/applications/${appId}/data`}
      />
    )
  }
}

ApplicationEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  widget: PropTypes.bool,
}

ApplicationEvents.defaultProps = {
  widget: false,
}

export default ApplicationEvents
