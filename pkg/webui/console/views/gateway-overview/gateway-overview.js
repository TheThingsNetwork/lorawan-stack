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
import bind from 'autobind-decorator'
import { Container, Col, Row } from 'react-grid-system'

import DataSheet from '../../../components/data-sheet'
import GatewayConnection from '../../containers/gateway-connection'
import GatewayEvents from '../../containers/gateway-events'
import Tag from '../../../components/tag'

import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'
import DateTime from '../../../lib/components/date-time'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import GatewayMap from '../../components/gateway-map'
import EntityTitleSection from '../../components/entity-title-section'
import KeyValueTag from '../../components/key-value-tag'
import Spinner from '../../../components/spinner'
import withRequest from '../../../lib/components/with-request'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayEditBasicGatewayInformation } from '../../lib/feature-checks'

@withRequest(({ gtwId, loadData }) => loadData(gtwId), () => false)
@withFeatureRequirement(mayEditBasicGatewayInformation, {
  redirect: '/',
})
@bind
export default class GatewayOverview extends React.Component {
  static propTypes = {
    apiKeysTotalCount: PropTypes.number,
    collaboratorsTotalCount: PropTypes.number,
    gateway: PropTypes.gateway.isRequired,
    gtwId: PropTypes.string.isRequired,
    statusBarFetching: PropTypes.bool.isRequired,
  }

  static defaultProps = {
    apiKeysTotalCount: undefined,
    collaboratorsTotalCount: undefined,
  }

  render() {
    const {
      gtwId,
      collaboratorsTotalCount,
      apiKeysTotalCount,
      statusBarFetching,
      gateway: {
        ids,
        name,
        description,
        created_at,
        updated_at,
        frequency_plan_id,
        gateway_server_address,
      },
    } = this.props

    const sheetData = [
      {
        header: sharedMessages.generalInformation,
        items: [
          {
            key: sharedMessages.gatewayID,
            value: gtwId,
            type: 'code',
            sensitive: false,
          },
          {
            key: sharedMessages.gatewayEUI,
            value: ids.eui,
            type: 'code',
            sensitive: false,
          },
          {
            key: sharedMessages.gatewayDescription,
            value: description || <Message content={sharedMessages.none} />,
          },
          {
            key: sharedMessages.createdAt,
            value: <DateTime value={created_at} />,
          },
          {
            key: sharedMessages.updatedAt,
            value: <DateTime value={updated_at} />,
          },
          {
            key: sharedMessages.gatewayServerAddress,
            value: gateway_server_address,
            type: 'code',
            sensitive: false,
          },
        ],
      },
      {
        header: sharedMessages.lorawanInformation,
        items: [
          {
            key: sharedMessages.frequencyPlan,
            value: <Tag content={frequency_plan_id} />,
          },
        ],
      },
    ]

    return (
      <React.Fragment>
        <EntityTitleSection
          entityId={gtwId}
          entityName={name}
          description={description}
          creationDate={created_at}
        >
          <GatewayConnection gtwId={gtwId} />
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
              <GatewayEvents gtwId={gtwId} widget />
              <GatewayMap gtwId={gtwId} gateway={this.props.gateway} />
            </Col>
          </Row>
        </Container>
      </React.Fragment>
    )
  }
}
