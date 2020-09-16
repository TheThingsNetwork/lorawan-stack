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
import { Col, Row, Container } from 'react-grid-system'

import DataSheet from '@ttn-lw/components/data-sheet'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import DateTime from '@ttn-lw/lib/components/date-time'

import OrganizationTitleSection from '@console/containers/organization-title-section'
import OrganizationEvents from '@console/containers/organization-events'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewOrganizationInformation } from '@console/lib/feature-checks'

import style from './organization-overview.styl'

@withFeatureRequirement(mayViewOrganizationInformation, {
  redirect: '/',
})
class Overview extends React.Component {
  static propTypes = {
    orgId: PropTypes.string.isRequired,
    organization: PropTypes.organization.isRequired,
  }

  render() {
    const {
      orgId,
      organization: { ids, created_at, updated_at },
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
      <>
        <div className={style.titleSection}>
          <Container>
            <Row>
              <Col sm={12}>
                <OrganizationTitleSection orgId={ids.organization_id} />
              </Col>
            </Row>
          </Container>
        </div>
        <Container>
          <IntlHelmet title={sharedMessages.overview} />
          <Row>
            <Col sm={12} lg={6}>
              <DataSheet data={sheetData} />
            </Col>
            <Col sm={12} lg={6}>
              <OrganizationEvents orgId={orgId} widget />
            </Col>
          </Row>
        </Container>
      </>
    )
  }
}

export default Overview
