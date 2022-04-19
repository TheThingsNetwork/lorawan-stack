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
import { Col, Row, Container } from 'react-grid-system'
import PageTitle from '@ttn-lw/components/page-title'

import withRequest from '@ttn-lw/lib/components/with-request'

import OAuthClientForm from '@account/containers/oauth-client-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getUserRights } from '@account/store/actions/user'
import {
  selectUserIsAdmin,
  selectUserId,
  selectUserRights,
  selectUserRegularRights,
  selectUserPseudoRights,
  selectUserRightsFetching,
  selectUserRightsError,
} from '@account/store/selectors/user'
import { selectSelectedClient } from '@account/store/selectors/clients'

const OAuthClientGeneralSettings = props => {
  const { userId, pseudoRights, isAdmin, regularRights, oauthClient } = props

  return (
    <Container>
      <PageTitle title={sharedMessages.generalSettings} />
      <Row>
        <Col lg={8} md={12}>
          <OAuthClientForm
            initialValues={oauthClient}
            isAdmin={isAdmin}
            userId={userId}
            rights={regularRights}
            pseudoRights={pseudoRights}
            update
          />
        </Col>
      </Row>
    </Container>
  )
}

OAuthClientGeneralSettings.propTypes = {
  isAdmin: PropTypes.bool.isRequired,
  oauthClient: PropTypes.shape({}).isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  regularRights: PropTypes.rights.isRequired,
  userId: PropTypes.string.isRequired,
}

export default connect(
  state => ({
    userId: selectUserId(state),
    isAdmin: selectUserIsAdmin(state),
    fetching: selectUserRightsFetching(state),
    error: selectUserRightsError(state),
    rights: selectUserRights(state),
    regularRights: selectUserRegularRights(state),
    pseudoRights: selectUserPseudoRights(state),
    oauthClient: selectSelectedClient(state),
  }),
  dispatch => ({
    getUsersRightsList: userId => dispatch(attachPromise(getUserRights(userId))),
  }),
)(
  withRequest(
    ({ getUsersRightsList, userId }) => getUsersRightsList(userId),
    ({ fetching, rights }) => fetching || rights.length === 0,
  )(OAuthClientGeneralSettings),
)
