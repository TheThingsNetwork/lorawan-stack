// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'

import IntlHelmet from '../../../lib/components/intl-helmet'
import Message from '../../../lib/components/message'
import DataSheet from '../../../components/data-sheet'
import DateTime from '../../../lib/components/date-time'

import style from './device-overview.styl'

const m = defineMessages({
  hardware: 'Hardware',
  brand: 'Brand',
  model: 'Model',
  hardwareVersion: 'Hardware version',
  firmwareVersion: 'Firmware Version',
  activationInfo: 'Activation Info',
  rootKeyId: 'Root Key ID',
  sessionInfo: 'Session Info',
  latestData: 'Latest Data',
})

@connect(function ({ device }, props) {
  return {
    device: device.device,
  }
})
@bind
class DeviceOverview extends React.Component {

  get deviceInfo () {
    const {
      ids,
      description,
      version_ids = {},
      root_keys = {},
      session = {},
      created_at,
    } = this.props.device

    // Get session keys
    const { keys: sessionKeys = {}} = session

    const {
      f_nwk_s_int_key = { key: undefined },
      s_nwk_s_int_key = { key: undefined },
      nwk_s_enc_key = { key: undefined },
      app_s_key = { key: undefined },
    } = sessionKeys

    const sheetData = [
      {
        header: sharedMessages.generalInformation,
        items: [
          { key: sharedMessages.devID, value: ids.device_id, type: 'code', sensitive: false },
          { key: sharedMessages.description, value: description || <Message content={sharedMessages.noDesc} /> },
          { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
        ],
      },
    ]

    // Add version info, if it is available
    if (Object.keys(version_ids).length > 0) {
      sheetData.push({
        header: m.hardware,
        items: [
          { key: m.brand, value: version_ids.brand_id },
          { key: m.model, value: version_ids.model_id },
          { key: m.hardwareVersion, value: version_ids.hardware_version },
          { key: m.firmwareVersion, value: version_ids.firmware_version },
        ],
      })
    }

    // Add activation info
    const activationInfoData = {
      header: m.activationInfo,
      items: [
        { key: sharedMessages.devEUI, value: ids.dev_eui, type: 'byte', sensitive: false },
        { key: sharedMessages.joinEUI, value: ids.join_eui, type: 'byte', sensitive: false },
      ],
    }

    // Add root keys, if available
    if (Object.keys(root_keys).length > 0) {
      activationInfoData.items.push({
        key: m.rootKeyId,
        value: root_keys.root_key_id,
        type: 'code',
        sensitive: false,
        subItems: [
          { key: sharedMessages.appKey, value: root_keys.app_key.key, type: 'code', sensitive: true },
          { key: sharedMessages.networkKey, value: root_keys.nwk_key.key, type: 'code', sensitive: true },
        ],
      })
    }
    sheetData.push(activationInfoData)

    // Add session info
    sheetData.push({
      header: m.sessionInfo,
      items: [
        { key: sharedMessages.devAddr, value: ids.dev_addr, type: 'byte', sensitive: false },
        { key: sharedMessages.fwdNtwkKey, value: f_nwk_s_int_key.key, type: 'code', sensitive: true },
        { key: sharedMessages.sNtwkSIKey, value: s_nwk_s_int_key.key, type: 'code', sensitive: true },
        { key: sharedMessages.ntwkSEncKey, value: nwk_s_enc_key.key, type: 'code', sensitive: true },
        { key: sharedMessages.appSKey, value: app_s_key.key, type: 'code', sensitive: true },
      ],
    })

    return (
      <div className={style.overviewInfo}>
        <div>
          <DataSheet data={sheetData} />
        </div>
      </div>
    )
  }

  render () {
    return (
      <Container>
        <IntlHelmet
          title={sharedMessages.overview}
        />
        <Row className={style.head}>
          <Col md={12} lg={6}>
            {this.deviceInfo}
          </Col>
          <Col md={12} lg={6}>
            <div className={style.activityPlaceholder}>
              <h4><Message content={m.latestData} /></h4>
              <div>Activity Panel Placeholder</div>
            </div>
            <div className={style.locationPlaceholder}>
              <h4><Message content={sharedMessages.location} /></h4>
              <div>Location Map Placeholder</div>
            </div>
          </Col>
        </Row>
      </Container>
    )
  }
}

export default DeviceOverview
