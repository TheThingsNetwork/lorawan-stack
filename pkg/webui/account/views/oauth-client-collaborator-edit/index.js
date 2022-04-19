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

import React, { useState, useCallback } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { replace } from 'connected-react-router'
import { connect } from 'react-redux'

import tts from '@account/api/tts'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import toast from '@ttn-lw/components/toast'

import CollaboratorForm from '@account/components/collaborator-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import withRequest from '@ttn-lw/lib/components/with-request'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getCollaborator } from '@account/store/actions/collaborators'
import {
  selectClientPseudoRights,
  selectClientRegularRights,
  selectClientRightsError,
  selectClientRightsFetching,
  selectSelectedClientId,
} from '@account/store/selectors/clients'
import {
  selectCollaboratorError,
  selectCollaboratorFetching,
  selectOrganizationCollaborator,
  selectUserCollaborator,
} from '@account/store/selectors/collaborators'

const showSuccessToast = () => {
  toast({
    message: sharedMessages.collaboratorUpdateSuccess,
    type: toast.types.SUCCESS,
  })
}

const OAuthClientCollaboratorEdit = props => {
  const {
    clientId,
    removeCollaborator,
    updateCollaborator,
    collaborator,
    collaboratorId,
    rights,
    pseudoRights,
    redirectToList,
    collaboratorType,
  } = props

  const [error, setError] = useState(undefined)

  const handleSubmit = useCallback(
    updatedCollaborator => updateCollaborator(updatedCollaborator),
    [updateCollaborator],
  )

  const handleDelete = useCallback(async () => {
    const collaborator_ids = {
      [`${collaboratorType}_ids`]: {
        [`${collaboratorType}_id`]: collaboratorId,
      },
    }
    const updatedCollaborator = {
      ids: collaborator_ids,
    }

    try {
      await removeCollaborator(updatedCollaborator)
      toast({
        message: sharedMessages.collaboratorDeleteSuccess,
        type: toast.types.SUCCESS,
      })
      redirectToList(clientId)
    } catch (error) {
      setError(error)
    }
  }, [clientId, collaboratorId, collaboratorType, redirectToList, removeCollaborator])

  return (
    <Container>
      <PageTitle title={sharedMessages.collaboratorEdit} values={{ collaboratorId }} />
      <Row>
        <Col lg={8} md={12}>
          <CollaboratorForm
            error={error}
            onSubmit={handleSubmit}
            onSubmitSuccess={showSuccessToast}
            onDelete={handleDelete}
            onDeleteSuccess={redirectToList}
            collaborator={collaborator}
            pseudoRights={pseudoRights}
            rights={rights}
            update
          />
        </Col>
      </Row>
    </Container>
  )
}

OAuthClientCollaboratorEdit.propTypes = {
  clientId: PropTypes.string.isRequired,
  collaborator: PropTypes.collaborator.isRequired,
  collaboratorId: PropTypes.string.isRequired,
  collaboratorType: PropTypes.oneOf(['collaborator', 'user']).isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  redirectToList: PropTypes.func.isRequired,
  removeCollaborator: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
  updateCollaborator: PropTypes.func.isRequired,
}

export default connect(
  (state, props) => {
    const clientId = selectSelectedClientId(state)

    const { collaboratorId, collaboratorType } = props.match.params

    const collaborator =
      collaboratorType === 'user'
        ? selectUserCollaborator(state)
        : selectOrganizationCollaborator(state)
    const fetching = selectClientRightsFetching(state) || selectCollaboratorFetching(state)
    const error = selectClientRightsError(state) || selectCollaboratorError(state)

    return {
      collaboratorId,
      collaboratorType,
      collaborator,
      clientId,
      rights: selectClientRegularRights(state),
      pseudoRights: selectClientPseudoRights(state),
      fetching,
      error,
    }
  },
  dispatch => ({
    getCollaborator: (clientId, collaboratorId, isUser) => {
      dispatch(getCollaborator('client', clientId, collaboratorId, isUser))
    },
    redirectToList: clientId => {
      dispatch(replace(`/oauth-clients/${clientId}/collaborators`))
    },
    updateCollaborator: (clientId, patch) => tts.Clients.Collaborators.update(clientId, patch),
    removeCollaborator: (clientId, patch) => tts.Clients.Collaborators.update(clientId, patch),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    getCollaborator: () =>
      dispatchProps.getCollaborator(
        stateProps.clientId,
        stateProps.collaboratorId,
        stateProps.collaboratorType === 'user',
      ),
    redirectToList: () => dispatchProps.redirectToList(stateProps.clientId),
    updateCollaborator: patch => tts.Applications.Collaborators.update(stateProps.clientId, patch),
    removeCollaborator: patch => tts.Applications.Collaborators.update(stateProps.clientId, patch),
  }),
)(
  withRequest(
    ({ getCollaborator }) => getCollaborator(),
    ({ fetching, collaborator }) => fetching || !Boolean(collaborator),
  )(OAuthClientCollaboratorEdit),
)
