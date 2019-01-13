// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import Tabs from '../../../components/tabs'
import Icon from '../../../components/icon'

import style from './device-overview.styl'

const tabs = [
  { title: 'Overview', name: 'overview' },
  { title: 'Data', name: 'data' },
  { title: 'Location', name: 'location' },
  { title: 'Payload Formatter', name: 'develop' },
  { title: 'General Settings', name: 'general-settings' },
]

@connect(function ({ device }, props) {
  return {
    device: device.device,
  }
})
class DeviceOverview extends React.Component {

  handleTabChange () {

  }

  get deviceInfo () {
    const {
      ids,
      description,
    } = this.props.device

    return (
      <div className={style.overviewInfo}>
        <div className={style.overviewInfoGeneral}>
          <span className={style.devId}>{ids.device_id}</span>
          <span className={style.devDesc}>{description}</span>
          <div className={style.connectivity}>
            <span className={style.activityDot} />
            <span className={style.lastSeen}>Last seen 2 secs. ago</span>
            <span className={style.frameCountUp}><Icon icon="arrow_upward" className={style.frameCountIcon} />89.139</span>
            <span><Icon icon="arrow_downward" className={style.frameCountIcon} />0</span>
          </div>
        </div>
      </div>
    )
  }

  render () {
    const { device_id } = this.props.device
    return (
      <Container>
        <Row className={style.head}>
          <Col lg={12}>
            <div className={style.title}>

              <h2 className={style.id}>
                {device_id}
              </h2>
            </div>
            <Tabs
              narrow
              active="overview"
              tabs={tabs}
              onTabChange={this.handleTabChange}
              className={style.tabs}
            />
          </Col>
          <Col sm={12} lg={6}>
            {this.deviceInfo}
          </Col>
        </Row>
      </Container>
    )
  }
}

export default DeviceOverview
