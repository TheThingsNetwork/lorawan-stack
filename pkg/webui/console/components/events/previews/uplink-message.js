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

import PropTypes from '@ttn-lw/lib/prop-types'
import getByPath from '@ttn-lw/lib/get-by-path'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import messages from '../messages'

import DescriptionList from './shared/description-list'

const UplinkMessagePreview = React.memo(({ event }) => {
  const { data } = event
  let fPort, snr, devAddr, fCnt, joinEui, devEui, rssi, isConfirmed, dataRate

  if ('payload' in data) {
    if ('mac_payload' in data.payload) {
      devAddr = getByPath(data, 'payload.mac_payload.f_hdr.dev_addr')
      fPort = getByPath(data, 'payload.mac_payload.f_port')
      fCnt = getByPath(data, 'payload.mac_payload.f_hdr.f_cnt')
    }

    if ('join_request_payload' in data.payload) {
      joinEui = getByPath(data, 'payload.join_request_payload.join_eui')
      devEui = getByPath(data, 'payload.join_request_payload.dev_eui')
    }

    isConfirmed = getByPath(data, 'payload.m_hdr.m_type') === 'CONFIRMED_UP'
  }

  if ('rx_metadata' in data) {
    snr = data.rx_metadata[0].snr
    rssi = data.rx_metadata[0].rssi
  }

  if ('settings' in data && 'data_rate' in data.settings) {
    const bandwidth = getByPath(data, 'settings.data_rate.lora.bandwidth')
    const spreadingFactor = getByPath(data, 'settings.data_rate.lora.spreading_factor')
    dataRate = `SF${spreadingFactor}BW${bandwidth / 1000}`
  }

  return (
    <DescriptionList>
      <DescriptionList.Byte title={messages.devAddr} data={devAddr} />
      <DescriptionList.Item title={messages.fCnt} data={fCnt} highlight />
      <DescriptionList.Item title={messages.fPort} data={fPort} />
      {isConfirmed && (
        <DescriptionList.Item>
          <Message content={messages.confirmedUplink} />
        </DescriptionList.Item>
      )}
      <DescriptionList.Byte title={sharedMessages.joinEUI} data={joinEui} />
      <DescriptionList.Byte title={sharedMessages.devEUI} data={devEui} />
      <DescriptionList.Item title={messages.dataRate} data={dataRate} />
      <DescriptionList.Item title={messages.snr} data={snr} />
      <DescriptionList.Item title={messages.rssi} data={rssi} />
    </DescriptionList>
  )
})

UplinkMessagePreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default UplinkMessagePreview
