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

import Events from '@console/components/events'

import PropTypes from '@ttn-lw/lib/prop-types'

const OrganizationEvents = props => {
  const { orgId, events, widget, onClear, truncated } = props

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

  return <Events events={events} onClear={onClear} entityId={orgId} truncated={truncated} />
}

OrganizationEvents.propTypes = {
  events: PropTypes.events,
  onClear: PropTypes.func.isRequired,
  orgId: PropTypes.string.isRequired,
  truncated: PropTypes.bool.isRequired,
  widget: PropTypes.bool,
}

OrganizationEvents.defaultProps = {
  widget: false,
  events: [],
}

export default OrganizationEvents
