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

import messages from '../messages'

import DescriptionList from './shared/description-list'

const DownLinkMessagePreview = React.memo(({ event }) => {
  const { data } = event
  let txPower, bandwidth, frmPayload

  if ('scheduled' in data) {
    if ('downlink' in data.scheduled) {
      txPower = data.scheduled.downlink.txPower
    }
    if ('data_rate' in data.scheduled && 'lora' in data.scheduled.data_rate) {
      bandwidth = data.scheduled.data_rate.lora.bandwidth
    }

    return (
      <DescriptionList>
        <DescriptionList.Item title={messages.txPower} data={txPower} />
        <DescriptionList.Item title={messages.bandwidth} data={bandwidth} />
      </DescriptionList>
    )
  }

  if ('request' in data) {
    const devAddr = event.identifiers[0].device_ids.device_addr
    if ('payload' in data && 'mac_payload' in data.payload) {
      frmPayload = data.payload.mac_payload.frm_payload
    }
    const rx1Delay = data.request.rx1_delay

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
