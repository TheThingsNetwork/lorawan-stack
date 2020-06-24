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

import PAGE_SIZES from '@console/constants/page-sizes'

import DataSheet from '@ttn-lw/components/data-sheet'
import Status from '@ttn-lw/components/status'
import Spinner from '@ttn-lw/components/spinner'

import DateTime from '@ttn-lw/lib/components/date-time'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'
import withRequest from '@ttn-lw/lib/components/with-request'

import KeyValueTag from '@console/components/key-value-tag'
import EntityTitleSection from '@console/components/entity-title-section'

import DevicesTable from '@console/containers/devices-table'
import ApplicationEvents from '@console/containers/application-events'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewApplicationInfo } from '@console/lib/feature-checks'

import style from './application-overview.styl'

@withRequest(
  ({ appId, loadData }) => loadData(appId),
  () => false,
)
@withFeatureRequirement(mayViewApplicationInfo, {
  redirect: '/',
})
class ApplicationOverview extends React.Component {
  static propTypes = {
    apiKeysTotalCount: PropTypes.number,
    appId: PropTypes.string.isRequired,
    application: PropTypes.application.isRequired,
    applicationLastSeen: PropTypes.string,
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
    apiKeysTotalCount: undefined,
    applicationLastSeen: undefined,
    collaboratorsTotalCount: undefined,
    devicesTotalCount: undefined,
    link: undefined,
  }

  render() {
    const {
      apiKeysTotalCount,
      appId,
      application: { name, description, created_at, updated_at },
      applicationLastSeen,
      collaboratorsTotalCount,
      devicesTotalCount,
      link,
      mayViewApplicationApiKeys,
      mayViewApplicationCollaborators,
      mayViewApplicationLink,
      mayViewDevices,
      statusBarFetching,
    } = this.props

    const linkStatus = typeof link === 'boolean' ? (link ? 'good' : 'bad') : 'mediocre'
    let linkLabel = sharedMessages.fetching
    let linkElement
    if (typeof link === 'boolean') {
      if (link) {
        if (applicationLastSeen) {
          linkElement = (
            <Status className={style.status} status={linkStatus} flipped>
              <Message content={sharedMessages.lastSeen} />{' '}
              <DateTime.Relative value={applicationLastSeen} />
            </Status>
          )
        } else {
          linkLabel = sharedMessages.linked
        }
      } else {
        linkLabel = sharedMessages.notLinked
      }
    }

    if (!linkElement) {
      linkElement = (
        <Status className={style.status} label={linkLabel} status={linkStatus} flipped />
      )
    }

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
              {mayViewApplicationLink && linkElement}
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
              <DataSheet data={sheetData} className={style.generalInformation} />
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
