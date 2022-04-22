// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { useIntl } from 'react-intl'
import { Col, Row, Container } from 'react-grid-system'

import applicationIcon from '@assets/misc/application.svg'

import DataSheet from '@ttn-lw/components/data-sheet'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import DateTime from '@ttn-lw/lib/components/date-time'
import withRequest from '@ttn-lw/lib/components/with-request'

import EntityTitleSection from '@console/components/entity-title-section'

import { mayPerformAdminActions } from '@account/lib/feature-checks'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  selectSelectedClient,
  selectSelectedClientId,
  selectClientFetching,
} from '@account/store/selectors/clients'
import { selectCollaboratorsTotalCount } from '@account/store/selectors/collaborators'
import { getCollaboratorsList } from '@account/store/actions/collaborators'

import style from './oauth-client-overview.styl'

const { Content } = EntityTitleSection

const capitalize = str => str.charAt(0).toUpperCase() + str.slice(1)

const OAuthClientOverview = props => {
  const {
    oauthClientId,
    oauthClient: { created_at, updated_at, state, state_description, secret },
    fetching,
    collaboratorsTotalCount,
  } = props
  const { formatMessage } = useIntl()

  const sheetData = [
    {
      header: sharedMessages.generalInformation,
      items: [
        { key: sharedMessages.oauthClientId, value: oauthClientId, type: 'code', sensitive: false },
        { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
        { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
        { key: sharedMessages.state, value: capitalize(formatMessage({ id: `enum:${state}` })) },
        {
          key: sharedMessages.stateDescription,
          value: state_description,
        },
      ],
    },
  ]

  // Add secret, if it is available.
  if (secret) {
    sheetData[0].items.push({
      key: sharedMessages.secret,
      value: secret,
      type: 'byte',
      sensitive: true,
      enableUint32: true,
    })
  }

  const bottomBarLeft = (
    <>
      {mayPerformAdminActions && (
        <Content.EntityCount
          icon="collaborators"
          value={collaboratorsTotalCount}
          keyMessage={sharedMessages.collaboratorCounted}
          errored={false}
          toAllUrl={`/oauth-clients/${oauthClientId}/collaborators`}
        />
      )}
    </>
  )

  return (
    <>
      <div className={style.titleSection}>
        <Container>
          <IntlHelmet title={sharedMessages.overview} />
          <Row>
            <Col sm={12}>
              <EntityTitleSection
                id={oauthClientId}
                icon={applicationIcon}
                iconAlt={sharedMessages.overview}
              >
                <Content fetching={fetching} bottomBarLeft={bottomBarLeft} />
              </EntityTitleSection>
            </Col>
          </Row>
        </Container>
      </div>
      <Container>
        <Row>
          <Col sm={12} lg={6}>
            <DataSheet data={sheetData} />
          </Col>
        </Row>
      </Container>
    </>
  )
}

OAuthClientOverview.propTypes = {
  collaboratorsTotalCount: PropTypes.number,
  fetching: PropTypes.bool.isRequired,
  oauthClient: PropTypes.shape({
    created_at: PropTypes.string,
    updated_at: PropTypes.string,
    state: PropTypes.string,
    state_description: PropTypes.string,
    secret: PropTypes.string,
  }).isRequired,
  oauthClientId: PropTypes.string.isRequired,
}

OAuthClientOverview.defaultProps = {
  collaboratorsTotalCount: undefined,
}

export default connect(
  state => {
    const oauthClientId = selectSelectedClientId(state)

    return {
      oauthClientId,
      oauthClient: selectSelectedClient(state),
      fetching: selectClientFetching(state),
      collaboratorsTotalCount: selectCollaboratorsTotalCount(state),
    }
  },
  dispatch => ({
    loadData: oauthClientId => {
      dispatch(getCollaboratorsList('client', oauthClientId))
    },
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    loadData: id => dispatchProps.loadData(id),
  }),
)(
  withRequest(
    ({ oauthClientId, loadData }) => loadData(oauthClientId),
    () => false,
  )(OAuthClientOverview),
)
