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
import { Col, Row } from 'react-grid-system'

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import DevicesTable from '../../../containers/devices-table'

import style from './application-overview.styl'

const DEVICES_TABLE_SIZE = 5

@connect(function ({ application }, props) {
  return {
    application: application.application,
  }
})
class ApplicationOverview extends React.Component {

  get applicationInfo () {
    const {
      application_id,
      description,
      created_at,
      updated_at,
    } = this.props.application

    return (
      <div>
        <h2 className={style.id}>
          {application_id}
        </h2>
        <p>{description}</p>
        <ul className={style.attributes}>
          <li className={style.attributesEntry}>
            <strong className={style.key}>
              <Message content={sharedMessages.createdAt} />
            </strong>
            <span className={style.value}>{created_at.toLocaleDateString()}</span>
          </li>
          <li className={style.attributesEntry}>
            <strong className={style.key}>
              <Message content={sharedMessages.updatedAt} />
            </strong>
            <span className={style.value}>{updated_at.toLocaleDateString()}</span>
          </li>
        </ul>
      </div>
    )
  }

  render () {
    return (
      <React.Fragment>
        <Row className={style.head}>
          <Col sm={12} lg={6}>
            {this.applicationInfo}
          </Col>
          <Col sm={12} lg={6}>
            <div className={style.latestEvents}>Latest Events Placeholder</div>
          </Col>
        </Row>
        <Row>
          <Col sm={12} className={style.table}>
            <DevicesTable pageSize={DEVICES_TABLE_SIZE} />
          </Col>
        </Row>
      </React.Fragment>
    )
  }
}

export default ApplicationOverview
