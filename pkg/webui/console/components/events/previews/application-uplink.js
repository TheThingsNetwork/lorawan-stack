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
import sharedMessages from '@ttn-lw/lib/shared-messages'
import getByPath from '@ttn-lw/lib/get-by-path'

import messages from '../messages'

import DescriptionList from './shared/description-list'
import JSONPayload from './shared/json-payload'

const ApplicationUplinkPreview = React.memo(({ event }) => {
  const { data, identifiers } = event
  const deviceIds = identifiers[0].device_ids
  let snr

  if ('rx_metadata' in data) {
    snr = data.rx_metadata[0].snr
  }

  const bandwidth = getByPath(data, 'settings.data_rate.lora.bandwidth')

  return (
    <DescriptionList>
      <DescriptionList.Byte title={messages.devAddr} data={deviceIds.dev_addr} />
      {'decoded_payload' in data ? (
        <DescriptionList.Item title={sharedMessages.payload}>
          <JSONPayload data={data.decoded_payload} />
          {data.frm_payload && (
            <DescriptionList.Byte key="frm_payload" data={data.frm_payload} convertToHex />
          )}
        </DescriptionList.Item>
      ) : (
        <DescriptionList.Byte title={messages.frmPayload} data={data.frm_payload} convertToHex />
      )}
      <DescriptionList.Item title={messages.fPort} data={data.f_port} />
      <DescriptionList.Item title={messages.snr} data={snr} />
      <DescriptionList.Item title={messages.bandwidth} data={bandwidth} />
    </DescriptionList>
  )
})

ApplicationUplinkPreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default ApplicationUplinkPreview
