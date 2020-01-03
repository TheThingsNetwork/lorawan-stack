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
import PropTypes from '../../../lib/prop-types'
import PAGE_SIZES from '../../constants/page-sizes'

import IntlHelmet from '../../../lib/components/intl-helmet'
import DateTime from '../../../lib/components/date-time'
import DevicesTable from '../../containers/devices-table'
import DataSheet from '../../../components/data-sheet'
import ApplicationEvents from '../../containers/application-events'
import EntityTitleSection from '../../components/entity-title-section'
import KeyValueTag from '../../components/key-value-tag'
import Status from '../../../components/status'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'
import withRequest from '../../../lib/components/with-request'

import { mayViewApplicationInfo } from '../../lib/feature-checks'

import style from './application-overview.styl'

@withRequest(({ appId, loadData }) => loadData(appId), () => false)
@withFeatureRequirement(mayViewApplicationInfo, {
  redirect: '/',
})
class ApplicationOverview extends React.Component {
  static propTypes = {
    apiKeysTotalCount: PropTypes.number,
    appId: PropTypes.string.isRequired,
    application: PropTypes.application.isRequired,
    collaboratorsTotalCount: PropTypes.number,
    devicesTotalCount: PropTypes.number,
    link: PropTypes.bool,
    mayViewApplicationApiKeys: PropTypes.bool.isRequired,
    mayViewApplicationCollaborators: PropTypes.bool.isRequired,
    mayViewApplicationLink: PropTypes.bool.isRequired,
    mayViewDevices: PropTypes.bool.isRequired,
    statusBarFetching: PropTypes.bool.isRequired,
  }

  static defaultProps = {
    collaboratorsTotalCount: undefined,
    apiKeysTotalCount: undefined,
    devicesTotalCount: undefined,
    link: undefined,
  }

  render() {
    const {
      appId,
      collaboratorsTotalCount,
      apiKeysTotalCount,
      devicesTotalCount,
      statusBarFetching,
      mayViewApplicationApiKeys,
      mayViewApplicationCollaborators,
      mayViewDevices,
      mayViewApplicationLink,
      link,
      application: { name, description, created_at, updated_at },
    } = this.props

    const linkStatus = typeof link === 'boolean' ? (link ? 'good' : 'bad') : 'mediocre'
    const linkLabel =
      typeof link === 'boolean'
        ? link
          ? sharedMessages.linked
          : sharedMessages.notLinked
        : sharedMessages.fetching

    const sheetData = [
      {
        header: sharedMessages.generalInformation,
        items: [
          { key: sharedMessages.appId, value: appId, type: 'code', sensitive: false },
          { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
          { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
        ],
      },
    ]

    return (
      <React.Fragment>
        <EntityTitleSection
          entityId={appId}
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
              {mayViewApplicationLink && (
                <Status className={style.status} label={linkLabel} status={linkStatus} flipped />
              )}
              {mayViewDevices && (
                <KeyValueTag
                  icon="devices"
                  value={devicesTotalCount}
                  keyMessage={sharedMessages.deviceCounted}
                />
              )}
              {mayViewApplicationCollaborators && (
                <KeyValueTag
                  icon="collaborators"
                  value={collaboratorsTotalCount}
                  keyMessage={sharedMessages.collaboratorCounted}
                />
              )}
              {mayViewApplicationApiKeys && (
                <KeyValueTag
                  icon="api_keys"
                  value={apiKeysTotalCount}
                  keyMessage={sharedMessages.apiKeyCounted}
                />
              )}
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
              <ApplicationEvents appId={appId} widget />
            </Col>
          </Row>
          <Row>
            <Col sm={12} className={style.table}>
              <DevicesTable pageSize={PAGE_SIZES.SMALL} devicePathPrefix="/devices" />
            </Col>
          </Row>
        </Container>
      </React.Fragment>
    )
  }
}

export default ApplicationOverview
