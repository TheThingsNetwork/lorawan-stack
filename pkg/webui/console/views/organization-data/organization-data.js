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
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import OrganizationEvents from '../../containers/organization-events'

import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

import style from './organization-data.styl'

const m = defineMessages({
  orgData: 'Organization Data',
})

@withBreadcrumb('orgs.single.data', function(props) {
  return (
    <Breadcrumb
      path={`/organizations/${props.orgId}/data`}
      icon="data"
      content={sharedMessages.data}
    />
  )
})
export default class Data extends React.Component {
  static propTypes = {
    orgId: PropTypes.string.isRequired,
  }

  render() {
    const { orgId } = this.props

    return (
      <Container>
        <Row>
          <Col sm={12}>
            <IntlHelmet title={m.orgData} />
            <Message component="h2" content={m.orgData} />
          </Col>
        </Row>
        <Row>
          <Col sm={12} className={style.wrapper}>
            <OrganizationEvents orgId={orgId} />
          </Col>
        </Row>
      </Container>
    )
  }
}
