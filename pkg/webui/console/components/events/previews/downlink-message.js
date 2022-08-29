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

import Message from '@ttn-lw/lib/components/message'

import { getDataRate } from '@console/components/events/utils'

import PropTypes from '@ttn-lw/lib/prop-types'
import getByPath from '@ttn-lw/lib/get-by-path'

import messages from '../messages'

import DescriptionList from './shared/description-list'

const DownLinkMessagePreview = React.memo(({ event }) => {
  const { data } = event

  if ('scheduled' in data) {
    const txPower = getByPath(data, 'scheduled.downlink.tx_power')
    const dataRate = getDataRate(data, 'scheduled')

    return (
      <DescriptionList>
        <DescriptionList.Item title={messages.txPower} data={txPower} />
        <DescriptionList.Item title={messages.dataRate} data={dataRate} />
      </DescriptionList>
    )
  }

  if ('request' in data) {
    const { name } = event
    if (name.startsWith('gs')) {
      const gatewayEUI = event.identifiers[0].gateway_ids.eui
      const lorawanClass = getByPath(data, 'request.class')
      const rx1Delay = getByPath(data, 'request.rx1_delay')
      const rx1DataRateIndex = getByPath(data, 'request.rx1_data_rate_index')
      const rx1Frequency = getByPath(data, 'request.rx1_frequency')
      const rx2Frequency = getByPath(data, 'request.rx2_frequency')
      const rx2DataRateIndex = getByPath(data, 'request.rx2_data_rate_index')
      return (
        <DescriptionList>
          <DescriptionList.Byte title={messages.gatewayEUI} data={gatewayEUI} />
          <DescriptionList.Item title={messages.class} data={lorawanClass} />
          <DescriptionList.Item title={messages.rx1Delay} data={rx1Delay} />
          <DescriptionList.Item title={messages.rx1DataRateIndex} data={rx1DataRateIndex} />
          <DescriptionList.Item title={messages.rx1Frequency} data={rx1Frequency} />
          <DescriptionList.Item title={messages.rx2Frequency} data={rx2Frequency} />
          <DescriptionList.Item title={messages.rx2DataRateIndex} data={rx2DataRateIndex} />
        </DescriptionList>
      )
    }
    const devAddr = event.identifiers[0].device_ids.dev_addr
    const frmPayload = getByPath(data, 'payload.mac_payload.frm_payload')
    const rx1Delay = getByPath(data, 'request.rx1_delay')
    const fPort = getByPath(data, 'payload.mac_payload.f_port')
    const isConfirmed = getByPath(data, 'payload.m_hdr.m_type') === 'CONFIRMED_DOWN'

    return (
      <DescriptionList>
        <DescriptionList.Byte title={messages.devAddr} data={devAddr} />
        <DescriptionList.Item title={messages.fPort} data={fPort} />
        {isConfirmed && (
          <DescriptionList.Item>
            <Message content={messages.confirmedDownlink} />
          </DescriptionList.Item>
        )}
        <DescriptionList.Byte title={messages.MACPayload} data={frmPayload} convertToHex />
        <DescriptionList.Item title={messages.rx1Delay} data={rx1Delay} />
      </DescriptionList>
    )
  }

  return null
})

DownLinkMessagePreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default DownLinkMessagePreview
