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
import sharedMessages from '@ttn-lw/lib/shared-messages'

import messages from '../messages'

import DescriptionList from './shared/description-list'

const UplinkMessagePreview = React.memo(({ event }) => {
  const { data } = event
  let frmPayload, fPort, snr, devAddr, fCnt, joinEui, devEui

  if ('payload' in data && 'mac_payload' in data.payload) {
    devAddr = getByPath(data, 'payload.mac_payload.f_hdr.dev_addr')
    frmPayload = getByPath(data, 'payload.mac_payload.frm_payload')
    fPort = getByPath(data, 'payload.mac_payload.f_port')
    fCnt = getByPath(data, 'payload.mac_payload.f_hdr.f_cnt')
  }

  if ('payload' in data && 'join_request_payload' in data.payload) {
    joinEui = getByPath(data, 'payload.join_request_payload.join_eui')
    devEui = getByPath(data, 'payload.join_request_payload.dev_eui')
  }

  if ('rx_metadata' in data) {
    snr = data.rx_metadata[0].snr
  }

  const rawPayload = getByPath(data, 'raw_payload')
  const bandwidth = getByPath(data, 'settings.data_rate.lora.bandwidth')

  return (
    <DescriptionList>
      <DescriptionList.Byte title={messages.devAddr} data={devAddr} />
      <DescriptionList.Item title={messages.fPort} data={fPort} />
      <DescriptionList.Item title={messages.fCnt} data={fCnt} />
      <DescriptionList.Byte title={sharedMessages.joinEUI} data={joinEui} />
      <DescriptionList.Byte title={sharedMessages.devEUI} data={devEui} />
      <DescriptionList.Byte title={messages.frmPayload} data={frmPayload} convertToHex />
      <DescriptionList.Item title={messages.bandwidth} data={bandwidth} />
      <DescriptionList.Item title={messages.snr} data={snr} />
      <DescriptionList.Byte title={messages.rawPayload} data={rawPayload} convertToHex />
    </DescriptionList>
  )
})

UplinkMessagePreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default UplinkMessagePreview
