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

import bind from 'autobind-decorator'

import IntlHelmet from '../../../lib/components/intl-helmet'
import Icon from '../../../components/icon'
import DataSheet from '../../../components/data-sheet'
import Message from '../../../lib/components/message'

import style from './device-overview.styl'

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
      version_ids,
      root_keys,
      session,
    } = this.props.device

    const {
      f_nwk_s_int_key,
      s_nwk_s_int_key,
      nwk_s_enc_key,
      app_s_key,
    } = session.keys

    const sheetData = [
      {
        header: 'Hardware',
        items: [
          { key: 'Brand', value: version_ids.brand_id },
          { key: 'Model', value: version_ids.model_id },
          { key: 'Hardware Version', value: version_ids.hardware_version },
          { key: 'Firmware Version', value: version_ids.firmware_version },
        ],
      },
      {
        header: 'Activation Info',
        items: [
          { key: 'Device EUI', value: ids.dev_eui, type: 'byte', sensitive: false },
          { key: 'Join EUI', value: ids.join_eui, type: 'byte', sensitive: false },
          {
            key: 'Root Key ID',
            value: root_keys.root_key_id,
            type: 'code',
            sensitive: false,
            subItems: [
              { key: 'Application Key', value: root_keys.app_key.key, type: 'code', sensitive: true },
              { key: 'Network Key', value: root_keys.nwk_key.key, type: 'code', sensitive: true },
            ],
          },
        ],
      },
      {
        header: 'Session Info',
        items: [
          { key: 'Device Address', value: ids.dev_addr, type: 'byte', sensitive: false },
          { key: 'Forwarding Network Session Integrity Key', value: f_nwk_s_int_key.key, type: 'code', sensitive: true },
          { key: 'Serving Network Session Integrity Key', value: s_nwk_s_int_key.key, type: 'code', sensitive: true },
          { key: 'Network Session Encryption Key', value: nwk_s_enc_key.key, type: 'code', sensitive: true },
          { key: 'Application Session Key', value: app_s_key.key, type: 'code', sensitive: true },
        ],
      },
    ]

    return (
      <div className={style.overviewInfo}>
        <div className={style.overviewInfoGeneral}>
          <span className={style.devId}>{ids.device_id}</span>
          <span className={style.devDesc}>{description || <Message content={m.noDesc} />}</span>
          <div className={style.connectivity}>
            <span className={style.activityDot} />
            <span className={style.lastSeen}>Last seen 2 secs. ago</span>
            <span className={style.frameCountUp}><Icon icon="arrow_upward" className={style.frameCountIcon} />89.139</span>
            <span><Icon icon="arrow_downward" className={style.frameCountIcon} />0</span>
          </div>
        </div>
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
              <h4>Latest Data</h4>
              <div>Activity Panel Placeholder</div>
            </div>
            <div className={style.locationPlaceholder}>
              <h4>Location</h4>
              <div>Location Map Placeholder</div>
            </div>
          </Col>
        </Row>
      </Container>
    )
  }
}

export default DeviceOverview
