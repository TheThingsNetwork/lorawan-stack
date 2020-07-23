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

import { hasJoinRequestData, hasMacData } from '../../shared/utils/types'
import DescriptionList from '../../shared/components/description-list'
import Entry from '../../shared/components/entries'
import messages from '../../messages'

import style from './event-types.styl'

const { Overview, Details } = Event

const DefaultDataEntry = React.memo(({ event }) => {
  const { data = {} } = event

  if (hasMacData(event)) {
    const { payload } = data

    const devAddr = payload.mac_payload.f_hdr.dev_addr
    const hex = base64ToHex(payload.mac_payload.frm_payload)
    const fPort = payload.mac_payload.f_port

    return (
      <Entry.Data>
        <DescriptionList>
          <DescriptionList.Byte title={messages.devAddr} data={devAddr} />
          <DescriptionList.Item title={messages.fPort}>{fPort}</DescriptionList.Item>
          <DescriptionList.Byte title={messages.frmPayload} data={hex} />
        </DescriptionList>
      </Entry.Data>
    )
  } else if (hasJoinRequestData(event)) {
    const { payload } = data

    return (
      <Entry.Data>
        <DescriptionList>
          <DescriptionList.Byte
            title={sharedMessages.joinEUI}
            data={payload.join_request_payload.join_eui}
          />
          <DescriptionList.Byte
            title={sharedMessages.devEUI}
            data={payload.join_request_payload.dev_eui}
          />
        </DescriptionList>
      </Entry.Data>
    )
  }

  return <Entry.Data />
})

DefaultDataEntry.propTypes = {
  event: PropTypes.event.isRequired,
}

const GatewayUplinkEvent = props => {
  const { event, widget } = props
  const { name, time } = event

  const showData = 'data' in event && !widget

  return (
    <Event event={event} expandable={showData}>
      <Overview>
        <Entry.Icon className={style.icon} iconName="uplink" />
        <Entry.Time time={time} />
        <Entry.Type eventName={name} />
        {!widget && <DefaultDataEntry event={event} />}
      </Overview>
      {showData && <Details />}
    </Event>
  )
}

GatewayUplinkEvent.propTypes = {
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

GatewayUplinkEvent.defaultProps = {
  widget: false,
}

export default GatewayUplinkEvent
