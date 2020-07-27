// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Event from '@ttn-lw/components/events-list/event'

import PropTypes from '@ttn-lw/lib/prop-types'

import Entry from '../../shared/components/entries'
import { isGatewayConnectEvent, isGatewayDisconnectEvent } from '../../shared/utils/types'

const { Overview } = Event

const getIcon = event => {
  if (isGatewayConnectEvent(event)) {
    return 'gateway_connect'
  }

  if (isGatewayDisconnectEvent(event)) {
    return 'gateway_disconnect'
  }

  return 'event'
}

const GatewayConnectionEvent = props => {
  const { event } = props
  const { name, time } = event

  const iconName = getIcon(event)

  return (
    <Event event={event}>
      <Overview>
        <Entry.Icon iconName={iconName} />
        <Entry.Time time={time} />
        <Entry.Type eventName={name} />
        <Entry.Data />
      </Overview>
    </Event>
  )
}

GatewayConnectionEvent.propTypes = {
  event: PropTypes.event.isRequired,
}

export default GatewayConnectionEvent
