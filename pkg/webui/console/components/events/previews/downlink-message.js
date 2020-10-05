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

import PropTypes from '@ttn-lw/lib/prop-types'
import getByPath from '@ttn-lw/lib/get-by-path'

import messages from '../messages'

import DescriptionList from './shared/description-list'

const DownLinkMessagePreview = React.memo(({ event }) => {
  const { data } = event
  if ('scheduled' in data) {
    const rawPayload = getByPath(data, 'raw_payload')
    const txPower = getByPath(data, 'scheduled.downlink.tx_power')
    const bandwidth = getByPath(data, 'scheduled.data_rate.lora.bandwidth')

    return (
      <DescriptionList>
        <DescriptionList.Byte title={messages.rawPayload} data={rawPayload} convertToHex />
        <DescriptionList.Item title={messages.txPower} data={txPower} />
        <DescriptionList.Item title={messages.bandwidth} data={bandwidth} />
      </DescriptionList>
    )
  }

  if ('request' in data) {
    const devAddr = event.identifiers[0].device_ids.device_addr
    const frmPayload = getByPath(data, 'event.payload.mac_payload.frm_payload')
    const rx1Delay = getByPath(data, 'request.rx1_delay')

    return (
      <DescriptionList>
        <DescriptionList.Item title={messages.devAddr} data={devAddr} />
        <DescriptionList.Byte title={messages.frmPayload} data={frmPayload} convertToHex />
        <DescriptionList.Item title={messages.fPort} data={rx1Delay} />
      </DescriptionList>
    )
  }

  return null
})

DownLinkMessagePreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default DownLinkMessagePreview
