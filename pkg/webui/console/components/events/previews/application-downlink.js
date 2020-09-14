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

import { base64ToHex } from '@console/lib/bytes'

import messages from '../messages'

import DescriptionList from './shared/description-list'

const ApplicationDownlinkPreview = React.memo(({ event }) => {
  const { data, identifiers } = event
  const deviceIds = identifiers[0].device_ids
  const hex = base64ToHex(data.frm_payload)

  return (
    <DescriptionList>
      <DescriptionList.Byte title={messages.devAddr} data={deviceIds.dev_addr} />
      <DescriptionList.Item title={messages.fPort}>{data.f_port}</DescriptionList.Item>
      <DescriptionList.Byte title={messages.frmPayload} data={hex} />
    </DescriptionList>
  )
})

ApplicationDownlinkPreview.propTypes = {
  event: PropTypes.event.isRequired,
}

export default ApplicationDownlinkPreview
