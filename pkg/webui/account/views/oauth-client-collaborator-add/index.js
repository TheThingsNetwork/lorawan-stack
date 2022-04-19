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

import React, { useCallback } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { push } from 'connected-react-router'
import { connect } from 'react-redux'

import tts from '@account/api/tts'

import PageTitle from '@ttn-lw/components/page-title'

import CollaboratorForm from '@account/components/collaborator-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import withRequest from '@ttn-lw/lib/components/with-request'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  selectSelectedClientId,
  selectClientRegularRights,
  selectClientPseudoRights,
  selectClientRightsFetching,
  selectClientRightsError,
} from '@account/store/selectors/clients'
import { selectCollaborators } from '@account/store/selectors/collaborators'
import { getClientRights } from '@account/store/actions/clients'

const OAuthClientCollaboratorAdd = props => {
  const { rights, pseudoRights, redirectToList, addCollaborator, error } = props

  const handleSubmit = useCallback(collaborator => addCollaborator(collaborator), [addCollaborator])

  return (
    <Container>
      <PageTitle title={sharedMessages.addCollaborator} />
      <Row>
        <Col lg={8} md={12}>
          <CollaboratorForm
            error={error}
            onSubmit={handleSubmit}
            onSubmitSuccess={redirectToList}
            pseudoRights={pseudoRights}
            rights={rights}
          />
        </Col>
      </Row>
    </Container>
  )
}

OAuthClientCollaboratorAdd.propTypes = {
  addCollaborator: PropTypes.func.isRequired,
  error: PropTypes.error,
  pseudoRights: PropTypes.rights.isRequired,
  redirectToList: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
}

OAuthClientCollaboratorAdd.defaultProps = {
  error: undefined,
}

export default connect(
  state => ({
    clientId: selectSelectedClientId(state),
    collaborators: selectCollaborators(state),
    rights: selectClientRegularRights(state),
    pseudoRights: selectClientPseudoRights(state),
    fetching: selectClientRightsFetching(state),
    error: selectClientRightsError(state),
  }),
  dispatch => ({
    redirectToList: clientId => dispatch(push(`/oauth-clients/${clientId}/collaborators`)),
    addCollaborator: (clientId, collaborator) =>
      tts.Clients.Collaborators.add(clientId, collaborator),
    getClientRightsList: clientId => dispatch(attachPromise(getClientRights(clientId))),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    redirectToList: () => dispatchProps.redirectToList(stateProps.clientId),
    addCollaborator: collaborator =>
      dispatchProps.addCollaborator(stateProps.clientId, collaborator),
  }),
)(
  withRequest(({ getClientRightsList, clientId }) => getClientRightsList(clientId))(
    OAuthClientCollaboratorAdd,
  ),
)
