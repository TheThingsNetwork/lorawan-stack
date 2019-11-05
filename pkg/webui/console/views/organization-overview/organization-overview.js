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
import { Col, Row, Container } from 'react-grid-system'

import sharedMessages from '../../../lib/shared-messages'
import DateTime from '../../../lib/components/date-time'
import IntlHelmet from '../../../lib/components/intl-helmet'
import PropTypes from '../../../lib/prop-types'
import { getOrganizationId } from '../../../lib/selectors/id'
import OrganizationEvents from '../../containers/organization-events'
import DataSheet from '../../../components/data-sheet'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayViewOrganizationInformation } from '../../lib/feature-checks'

import style from './organization-overview.styl'

@withFeatureRequirement(mayViewOrganizationInformation, {
  redirect: '/',
})
class Overview extends React.Component {
  static propTypes = {
    organization: PropTypes.organization.isRequired,
  }

  get organizationInfo() {
    const {
      organization: { ids, name, description, created_at, updated_at },
    } = this.props

    const sheetData = [
      {
        header: sharedMessages.generalInformation,
        items: [
          {
            key: sharedMessages.organizationId,
            value: ids.organization_id,
            type: 'code',
            sensitive: false,
          },
          { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
          { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
        ],
      },
    ]

    return (
      <div>
        <div className={style.title}>
          <h2>{name || ids.organization_id}</h2>
          {description && <span className={style.description}>{description}</span>}
        </div>
        <DataSheet data={sheetData} />
      </div>
    )
  }

  render() {
    const { organization } = this.props
    const orgId = getOrganizationId(organization)

    return (
      <Container>
        <IntlHelmet title={sharedMessages.overview} />
        <Row>
          <Col sm={12} lg={6}>
            {this.organizationInfo}
          </Col>
          <Col sm={12} lg={6}>
            <div className={style.latestEvents}>
              <OrganizationEvents orgId={orgId} widget />
            </div>
          </Col>
        </Row>
      </Container>
    )
  }
}

export default Overview
