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

import ApplicationEventsList from '@console/components/events-list/application'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import {
  clearApplicationEventsStream,
  startApplicationEventsStream,
} from '@console/store/actions/applications'

import {
  selectApplicationEvents,
  selectApplicationEventsError,
} from '@console/store/selectors/applications'

const ApplicationEvents = props => {
  const { appId, events, error, onRestart, widget, onClear } = props

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
      <ApplicationEventsList.Widget
        events={events}
        toAllUrl={`/applications/${appId}/data`}
        appId={appId}
      />
    )
  }

  return <ApplicationEventsList events={events} onClear={onClear} appId={appId} />
}

ApplicationEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  error: PropTypes.error,
  events: PropTypes.events,
  onClear: PropTypes.func.isRequired,
  onRestart: PropTypes.func.isRequired,
  widget: PropTypes.bool,
}

ApplicationEvents.defaultProps = {
  widget: false,
  events: [],
  error: undefined,
}

export default withFeatureRequirement(mayViewApplicationEvents)(
  connect(
    (state, props) => {
      const { appId } = props

      return {
        events: selectApplicationEvents(state, appId),
        error: selectApplicationEventsError(state, appId),
      }
    },
    (dispatch, ownProps) => ({
      onClear: () => dispatch(clearApplicationEventsStream(ownProps.appId)),
      onRestart: () => dispatch(startApplicationEventsStream(ownProps.appId)),
    }),
  )(ApplicationEvents),
)
