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
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Message from '../../../lib/components/message'
import ApplicationEvents from '../../containers/application-events'

import { selectSelectedApplicationId } from '../../store/selectors/applications'

import style from './application-data.styl'

const m = defineMessages({
  appData: 'Application Data',
})

@connect(state => ({ appId: selectSelectedApplicationId(state) }))
@withBreadcrumb('apps.single.data', function(props) {
  return (
    <Breadcrumb
      path={`/applications/${props.appId}/data`}
      icon="data"
      content={sharedMessages.data}
    />
  )
})
export default class Data extends React.Component {
  render() {
    const { appId } = this.props

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={m.appData} />
            <Message component="h2" content={m.appData} />
          </Col>
        </Row>
        <Row>
          <Col sm={12} className={style.wrapper}>
            <ApplicationEvents appId={appId} />
          </Col>
        </Row>
      </Container>
    )
  }
}
