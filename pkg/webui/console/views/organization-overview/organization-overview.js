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
import OrganizationEvents from '../../containers/organization-events'
import DataSheet from '../../../components/data-sheet'
import EntityTitleSection from '../../components/entity-title-section'
import KeyValueTag from '../../components/key-value-tag'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import withRequest from '../../../lib/components/with-request'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayViewOrganizationInformation } from '../../lib/feature-checks'

@withFeatureRequirement(mayViewOrganizationInformation, {
  redirect: '/',
})
@withRequest(({ orgId, loadData }) => loadData(orgId), () => false)
class Overview extends React.Component {
  static propTypes = {
    apiKeysTotalCount: PropTypes.number,
    collaboratorsTotalCount: PropTypes.number,
    orgId: PropTypes.string.isRequired,
    organization: PropTypes.organization.isRequired,
    statusBarFetching: PropTypes.bool.isRequired,
  }

  static defaultProps = {
    collaboratorsTotalCount: undefined,
    apiKeysTotalCount: undefined,
  }

  render() {
    const {
      orgId,
      organization: { ids, name, description, created_at, updated_at },
      collaboratorsTotalCount,
      apiKeysTotalCount,
      statusBarFetching,
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
      <React.Fragment>
        <EntityTitleSection
          entityId={orgId}
          entityName={name}
          description={description}
          creationDate={created_at}
        >
          {statusBarFetching ? (
            <Spinner after={0} faded micro inline>
              <Message content={sharedMessages.fetching} />
            </Spinner>
          ) : (
            <React.Fragment>
              <KeyValueTag
                icon="collaborators"
                value={collaboratorsTotalCount}
                keyMessage={sharedMessages.collaboratorCounted}
              />
              <KeyValueTag
                icon="api_keys"
                value={apiKeysTotalCount}
                keyMessage={sharedMessages.apiKeyCounted}
              />
            </React.Fragment>
          )}
        </EntityTitleSection>
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
      </React.Fragment>
    )
  }
}

export default Overview
