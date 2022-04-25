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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import toast from '@ttn-lw/components/toast'
import CollaboratorForm from '@ttn-lw/components/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const showSuccessToast = () => {
  toast({
    message: sharedMessages.collaboratorUpdateSuccess,
    type: toast.types.SUCCESS,
  })
}

const ApplicationCollaboratorEdit = props => {
  const {
    appId,
    removeCollaborator,
    updateCollaborator,
    collaborator,
    collaboratorId,
    rights,
    pseudoRights,
    redirectToList,
    collaboratorType,
  } = props

  useBreadcrumbs(
    'apps.single.collaborators.edit',
    <Breadcrumb
      path={`/applications/${appId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />,
  )

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
      redirectToList(appId)
    } catch (error) {
      setError(error)
    }
  }, [appId, collaboratorId, collaboratorType, redirectToList, removeCollaborator])

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

ApplicationCollaboratorEdit.propTypes = {
  appId: PropTypes.string.isRequired,
  collaborator: PropTypes.collaborator.isRequired,
  collaboratorId: PropTypes.string.isRequired,
  collaboratorType: PropTypes.oneOf(['collaborator', 'user']).isRequired,
  pseudoRights: PropTypes.rights.isRequired,
  redirectToList: PropTypes.func.isRequired,
  removeCollaborator: PropTypes.func.isRequired,
  rights: PropTypes.rights.isRequired,
  updateCollaborator: PropTypes.func.isRequired,
}

export default ApplicationCollaboratorEdit
