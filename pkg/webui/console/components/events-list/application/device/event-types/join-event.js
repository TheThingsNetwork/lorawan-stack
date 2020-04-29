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

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { hasJoinAcceptData, hasJoinRequestData } from '../../../shared/utils/types'
import DescriptionList from '../../../shared/components/description-list'
import Entry from '../../../shared/components/entries'
import messages from '../../../messages'

import style from './event-types.styl'

const { Overview, Details } = Event

const DefaultDataEntry = React.memo(({ event }) => {
  const { identifiers } = event
  const ids = identifiers[0].device_ids

  if (hasJoinAcceptData(event) || hasJoinRequestData(event)) {
    return (
      <Entry.Data>
        <DescriptionList>
          <DescriptionList.Byte title={messages.devAddr} data={ids.dev_addr} />
          <DescriptionList.Byte title={sharedMessages.joinEUI} data={ids.join_eui} />
          <DescriptionList.Byte title={sharedMessages.devEUI} data={ids.dev_eui} />
        </DescriptionList>
      </Entry.Data>
    )
  }

  return (
    <Entry.Data>
      <DescriptionList>
        <DescriptionList.Byte title={messages.devAddr} data={ids.dev_addr} />
      </DescriptionList>
    </Entry.Data>
  )
})

DefaultDataEntry.propTypes = {
  event: PropTypes.event.isRequired,
}

const DeviceJoinEvent = props => {
  const { event, deviceId, widget } = props
  const { name, time } = event

  const showData = 'data' in event && !widget

  return (
    <Event event={event} expandable={showData}>
      <Overview>
        <Entry.Icon className={style.icon} iconName="join" />
        <Entry.Time time={time} />
        {Boolean(deviceId) && <Entry.ID entityId={deviceId} />}
        <Entry.Type eventName={name} />
        {!widget && <DefaultDataEntry event={event} />}
      </Overview>
      {showData && <Details />}
    </Event>
  )
}

DeviceJoinEvent.propTypes = {
  deviceId: PropTypes.string,
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

DeviceJoinEvent.defaultProps = {
  deviceId: undefined,
  widget: false,
}

export default DeviceJoinEvent
