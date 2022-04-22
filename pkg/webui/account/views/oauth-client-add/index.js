// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'

import withRequest from '@ttn-lw/lib/components/with-request'

import ClientAdd from '@account/containers/oauth-client-add'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

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

const m = defineMessages({
  addClient: 'Add OAuth Client',
})

const OAuthClientAdd = props => {
  const { userId, pseudoRights, isAdmin, regularRights } = props

  return (
    <Container>
      <PageTitle tall title={m.addClient} />
      <Row>
        <Col lg={8} md={12}>
          <ClientAdd
            isAdmin={isAdmin}
            userId={userId}
            rights={regularRights}
            pseudoRights={pseudoRights}
          />
        </Col>
      </Row>
    </Container>
  )
}

OAuthClientAdd.propTypes = {
  isAdmin: PropTypes.bool.isRequired,
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
  }),
  dispatch => ({
    getUsersRightsList: userId => dispatch(attachPromise(getUserRights(userId))),
  }),
)(
  withRequest(
    ({ getUsersRightsList, userId }) => getUsersRightsList(userId),
    ({ fetching, rights }) => fetching || rights.length === 0,
  )(OAuthClientAdd),
)
