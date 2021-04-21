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

import DataSheet from '@ttn-lw/components/data-sheet'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import DeviceMap from '@console/components/device-map'

import DeviceEvents from '@console/containers/device-events'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { parseLorawanMacVersion } from '@console/lib/device-utils'

import { selectSelectedDevice } from '@console/store/selectors/devices'

import style from './device-overview.styl'

const m = defineMessages({
  activationInfo: 'Activation information',
  rootKeyId: 'Root key ID',
  sessionInfo: 'Session information',
  latestData: 'Latest data',
  rootKeys: 'Root keys',
  keysNotExposed: 'Keys are not exposed',
})

@connect((state, props) => ({
  device: selectSelectedDevice(state),
}))
class DeviceOverview extends React.Component {
  static propTypes = {
    device: PropTypes.device.isRequired,
  }

  get deviceInfo() {
    const {
      ids,
      description,
      version_ids = {},
      root_keys = {},
      session = {},
      created_at,
      lorawan_version,
      supports_join,
    } = this.props.device

    // Get session keys.
    const { keys: sessionKeys = {} } = session

    const lorawanVersion = parseLorawanMacVersion(lorawan_version)

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

    // Add version info, if it is available.
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

    // Add activation info, if available.
    const activationInfoData = {
      header: m.activationInfo,
      items: [],
    }

    if (ids.join_eui || ids.dev_eui) {
      const joinEUI =
        lorawanVersion < 100
          ? sharedMessages.appEUIJoinEUI
          : lorawanVersion >= 104
          ? sharedMessages.joinEUI
          : sharedMessages.appEUI

      activationInfoData.items.push(
        { key: joinEUI, value: ids.join_eui, type: 'byte', sensitive: false },
        { key: sharedMessages.devEUI, value: ids.dev_eui, type: 'byte', sensitive: false },
      )

      // Add root keys, if available.
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
            ...(root_keys.app_key
              ? {
                  key: sharedMessages.appKey,
                  value: root_keys.app_key.key,
                  type: 'byte',
                  sensitive: true,
                }
              : { key: sharedMessages.appKey, value: undefined }),
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
      } else if (supports_join) {
        activationInfoData.items.push({
          key: m.rootKeys,
          value: <Message content={sharedMessages.provisionedOnExternalJoinServer} />,
        })
      }
    }

    sheetData.push(activationInfoData)

    // Add session info, if available.

    const sessionInfoData = {
      header: m.sessionInfo,
      items: [],
    }

    if (Object.keys(sessionKeys).length > 0) {
      sessionInfoData.items.push(
        {
          key: sharedMessages.devAddr,
          value: ids.dev_addr,
          type: 'byte',
          sensitive: false,
          enableUint32: true,
        },
        {
          key: sharedMessages.nwkSKey,
          value: f_nwk_s_int_key.key,
          type: 'byte',
          sensitive: true,
        },
        {
          key: sharedMessages.sNwkSIKey,
          value: s_nwk_s_int_key.key,
          type: 'byte',
          sensitive: true,
        },
        {
          key: sharedMessages.nwkSEncKey,
          value: nwk_s_enc_key.key,
          type: 'byte',
          sensitive: true,
        },
        { key: sharedMessages.appSKey, value: app_s_key.key, type: 'byte', sensitive: true },
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
