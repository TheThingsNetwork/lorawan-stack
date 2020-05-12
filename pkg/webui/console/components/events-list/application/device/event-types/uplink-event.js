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
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { base64ToHex } from '@console/lib/bytes'

import { hasUplinkMessageData, hasMacData } from '../../../shared/utils/types'
import DescriptionList from '../../../shared/components/description-list'
import Entry from '../../../shared/components/entries'
import JSONPayload from '../../../shared/components/json-payload'
import messages from '../../../messages'

import style from './event-types.styl'

const { Overview, Details } = Event

const DefaultDataEntry = React.memo(({ event }) => {
  const { data = {}, identifiers } = event
  const deviceIds = identifiers[0].device_ids

  if (hasUplinkMessageData(event)) {
    const { uplink_message } = data

    const entries = []
    entries.push(
      <DescriptionList.Byte key="dev_addr" title={messages.devAddr} data={deviceIds.dev_addr} />,
    )

    if (uplink_message.f_port) {
      entries.push(
        <DescriptionList.Item key="f_port" title={messages.fPort}>
          {uplink_message.f_port}
        </DescriptionList.Item>,
      )
    }

    if (uplink_message.decoded_payload) {
      entries.push(
        <DescriptionList.Item key="decoded_payload" title={sharedMessages.payload}>
          <JSONPayload data={uplink_message.decoded_payload} />
          {uplink_message.frm_payload && (
            <DescriptionList.Byte
              key="frm_payload"
              data={base64ToHex(uplink_message.frm_payload)}
            />
          )}
        </DescriptionList.Item>,
      )
    } else if (uplink_message.frm_payload) {
      entries.push(
        <DescriptionList.Byte
          key="frm_payload"
          title={messages.frmPayload}
          data={base64ToHex(uplink_message.frm_payload)}
        />,
      )
    }

    return <Entry.Data>{entries}</Entry.Data>
  } else if (hasMacData(event)) {
    const { payload } = data

    const entries = [
      <DescriptionList.Byte key="dev_addr" title={messages.devAddr} data={deviceIds.dev_addr} />,
    ]

    if (payload.mac_payload.frm_payload) {
      const hex = base64ToHex(payload.mac_payload.frm_payload)

      entries.push(
        <DescriptionList.Byte key="frm_payload" title={messages.frmPayload} data={hex} />,
      )
    }

    return <Entry.Data>{entries}</Entry.Data>
  }

  return (
    <Entry.Data>
      <DescriptionList>
        <DescriptionList.Byte title={messages.devAddr} data={deviceIds.dev_addr} />
      </DescriptionList>
    </Entry.Data>
  )
})

DefaultDataEntry.propTypes = {
  event: PropTypes.event.isRequired,
}

const DeviceUplinkEvent = props => {
  const { event, deviceId, widget } = props
  const { name, time } = event

  const showData = 'data' in event && !widget

  return (
    <Event event={event} expandable={showData}>
      <Overview>
        <Entry.Icon className={style.icon} iconName="uplink" />
        <Entry.Time time={time} />
        {Boolean(deviceId) && <Entry.ID entityId={deviceId} />}
        <Entry.Type eventName={name} />
        {!widget && <DefaultDataEntry event={event} />}
      </Overview>
      {showData && <Details />}
    </Event>
  )
}

DeviceUplinkEvent.propTypes = {
  deviceId: PropTypes.string,
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

DeviceUplinkEvent.defaultProps = {
  deviceId: undefined,
  widget: false,
}

export default DeviceUplinkEvent
