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

import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import Events from '@console/components/events'

import PropTypes from '@ttn-lw/lib/prop-types'

import {
  clearOrganizationEventsStream,
  pauseOrganizationEventsStream,
  resumeOrganizationEventsStream,
} from '@console/store/actions/organizations'

import {
  selectOrganizationEvents,
  selectOrganizationEventsPaused,
  selectOrganizationEventsTruncated,
} from '@console/store/selectors/organizations'

const OrganizationEvents = props => {
  const { orgId, widget } = props

  const events = useSelector(state => selectOrganizationEvents(state, orgId))
  const paused = useSelector(state => selectOrganizationEventsPaused(state, orgId))
  const truncated = useSelector(state => selectOrganizationEventsTruncated(state, orgId))

  const dispatch = useDispatch()

  const onPauseToggle = useCallback(
    paused => {
      if (paused) {
        dispatch(resumeOrganizationEventsStream(orgId))
        return
      }
      dispatch(pauseOrganizationEventsStream(orgId))
    },
    [dispatch, orgId],
  )

  const onClear = useCallback(() => {
    dispatch(clearOrganizationEventsStream(orgId))
  }, [dispatch, orgId])

  if (widget) {
    return (
      <Events.Widget
        events={events}
        toAllUrl={`/organizations/${orgId}/data`}
        entityId={orgId}
        scoped
      />
    )
  }

  return (
    <Events
      events={events}
      paused={paused}
      onClear={onClear}
      onPauseToggle={onPauseToggle}
      truncated={truncated}
      entityId={orgId}
      disableFiltering
    />
  )
}

OrganizationEvents.propTypes = {
  orgId: PropTypes.string.isRequired,
  widget: PropTypes.bool,
}

OrganizationEvents.defaultProps = {
  widget: false,
}

export default OrganizationEvents
