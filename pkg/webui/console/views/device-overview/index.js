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
import DeviceEvents from '../../containers/device-events'
import DeviceMap from '../../components/device-map'

import style from './device-overview.styl'

const m = defineMessages({
  activationInfo: 'Activation Information',
  rootKeyId: 'Root Key ID',
  sessionInfo: 'Session Information',
  latestData: 'Latest Data',
  rootKeys: 'Root Keys',
  keysNotExposed: 'Keys are not exposed',
})

@connect(function({ device }, props) {
  return {
    device: device.device,
  }
})
@bind
class DeviceOverview extends React.Component {
  get deviceInfo() {
    const {
      ids,
      description,
      version_ids = {},
      root_keys = {},
      session = {},
      created_at,
    } = this.props.device

    // Get session keys
    const { keys: sessionKeys = {} } = session

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
          {
            key: sharedMessages.description,
            value: description || <Message content={sharedMessages.noDesc} />,
          },
          { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
        ],
      },
    ]

    // Add version info, if it is available
    if (Object.keys(version_ids).length > 0) {
      sheetData.push({
        header: sharedMessages.hardware,
        items: [
          { key: sharedMessages.brand, value: version_ids.brand_id },
          { key: sharedMessages.model, value: version_ids.model_id },
          { key: sharedMessages.hardwareVersion, value: version_ids.hardware_version },
          { key: sharedMessages.firmwareVersion, value: version_ids.firmware_version },
        ],
      })
    }

    // Add activation info, if available
    const activationInfoData = {
      header: m.activationInfo,
      items: [],
    }

    if (ids.join_eui || ids.dev_eui) {
      activationInfoData.items.push(
        { key: sharedMessages.joinEUI, value: ids.join_eui, type: 'byte', sensitive: false },
        { key: sharedMessages.devEUI, value: ids.dev_eui, type: 'byte', sensitive: false },
      )

      // Add root keys, if available
      if (Object.keys(root_keys).length > 0) {
        const infoEntry = {
          key: m.rootKeyId,
          value: root_keys.root_key_id,
          type: 'code',
          sensitive: false,
        }
        if (!Boolean(root_keys.app_key) && !Boolean(root_keys.nwk_key)) {
          infoEntry.subItems = [
            {
              key: m.rootKeys,
              value: <Message content={m.keysNotExposed} />,
            },
          ]
        } else {
          infoEntry.subItems = [
            {
              key: sharedMessages.appKey,
              value: root_keys.app_key.key,
              type: 'byte',
              sensitive: true,
            },
            ...(root_keys.nwk_key
              ? {
                  key: sharedMessages.nwkKey,
                  value: root_keys.nwk_key.key,
                  type: 'byte',
                  sensitive: true,
                }
              : { key: sharedMessages.nwkKey, value: undefined }),
          ]
        }
        activationInfoData.items.push(infoEntry)
      } else {
        activationInfoData.items.push({
          key: m.rootKeys,
          value: <Message content={sharedMessages.provisionedOnExternalJoinServer} />,
        })
      }
    }

    sheetData.push(activationInfoData)

    // Add session info, if available

    const sessionInfoData = {
      header: m.sessionInfo,
      items: [],
    }

    if (Object.keys(sessionKeys).length > 0) {
      sessionInfoData.items.push(
        { key: sharedMessages.devAddr, value: ids.dev_addr, type: 'byte', sensitive: false },
        {
          key: sharedMessages.fwdNtwkKey,
          value: f_nwk_s_int_key.key,
          type: 'code',
          sensitive: true,
        },
        {
          key: sharedMessages.sNtwkSIKey,
          value: s_nwk_s_int_key.key,
          type: 'code',
          sensitive: true,
        },
        {
          key: sharedMessages.ntwkSEncKey,
          value: nwk_s_enc_key.key,
          type: 'code',
          sensitive: true,
        },
        { key: sharedMessages.appSKey, value: app_s_key.key, type: 'code', sensitive: true },
      )
    }

    sheetData.push(sessionInfoData)

    return (
      <div className={style.overviewInfo}>
        <div>
          <DataSheet data={sheetData} />
        </div>
      </div>
    )
  }

  render() {
    const { device } = this.props
    const devIds = device && device.ids
    return (
      <Container>
        <IntlHelmet title={sharedMessages.overview} />
        <Row className={style.head}>
          <Col md={12} lg={6}>
            {this.deviceInfo}
          </Col>
          <Col md={12} lg={6} className={style.latestEvents}>
            <DeviceEvents devIds={devIds} widget />
            <DeviceMap device={device} />
          </Col>
        </Row>
      </Container>
    )
  }
}

export default DeviceOverview
