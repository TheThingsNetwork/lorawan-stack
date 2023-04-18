// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { getDataRate, getSignalInformation } from '@console/components/events/utils'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import messages from '../messages'

import DescriptionList from './shared/description-list'
import JSONPayload from './shared/json-payload'

const ApplicationUplinkNormalizedPreview = React.memo(({ event }) => {
  const { data, identifiers } = event
  const deviceIds = identifiers[0].device_ids
  const { snr, rssi } = getSignalInformation(data)
  const dataRate = getDataRate(data)

  return (
    <DescriptionList>
      <DescriptionList.Byte title={messages.devAddr} data={deviceIds.dev_addr} />
      {data.normalized_payload.soil && (
        <DescriptionList.Item title={sharedMessages.normalizedPayloadSoil}>
          <JSONPayload data={data.normalized_payload.soil} />
        </DescriptionList.Item>
      )}
      {data.normalized_payload.air && (
        <DescriptionList.Item title={sharedMessages.normalizedPayloadAir}>
          <JSONPayload data={data.normalized_payload.air} />
        </DescriptionList.Item>
      )}
      {data.normalized_payload.wind && (
        <DescriptionList.Item title={sharedMessages.normalizedPayloadWind}>
          <JSONPayload data={data.normalized_payload.wind} />
        </DescriptionList.Item>
      )}
      <DescriptionList.Item title={messages.fPort} data={data.f_port} />
      <DescriptionList.Item title={messages.dataRate} data={dataRate} />
      <DescriptionList.Item title={messages.snr} data={snr} />
      <DescriptionList.Item title={messages.rssi} data={rssi} />
    </DescriptionList>
  )
})

ApplicationUplinkNormalizedPreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default ApplicationUplinkNormalizedPreview
