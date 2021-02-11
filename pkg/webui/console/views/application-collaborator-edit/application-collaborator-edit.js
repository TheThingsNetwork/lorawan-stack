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

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import toast from '@ttn-lw/components/toast'

import withRequest from '@ttn-lw/lib/components/with-request'

import CollaboratorForm from '@console/components/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const isUser = collaborator => collaborator.ids && 'user_ids' in collaborator.ids

@withRequest(
  ({ getCollaborator }) => getCollaborator(),
  ({ fetching, collaborator }) => fetching || !Boolean(collaborator),
)
@withBreadcrumb('apps.single.collaborators.edit', function (props) {
  const { appId, collaboratorId, collaboratorType } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/collaborators/${collaboratorType}/${collaboratorId}`}
      content={sharedMessages.edit}
    />
  )
})
export default class ApplicationCollaboratorEdit extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    collaborator: PropTypes.collaborator.isRequired,
    collaboratorId: PropTypes.string.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    redirectToList: PropTypes.func.isRequired,
    removeCollaborator: PropTypes.func.isRequired,
    rights: PropTypes.rights.isRequired,
    updateCollaborator: PropTypes.func.isRequired,
  }

  state = {
    error: undefined,
  }

  @bind
  async handleSubmit(updatedCollaborator) {
    const { updateCollaborator } = this.props

    await updateCollaborator(updatedCollaborator)
  }

  handleSubmitSuccess() {
    toast({
      message: sharedMessages.collaboratorUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDelete() {
    const { collaborator, collaboratorId, redirectToList, appId, removeCollaborator } = this.props
    const collaborator_type = isUser(collaborator) ? 'user' : 'organization'

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaboratorId,
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
      await this.setState({ error })
    }
  }

  render() {
    const { collaborator, collaboratorId, rights, pseudoRights, redirectToList } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.collaboratorEdit} values={{ collaboratorId }} />
        <Row>
          <Col lg={8} md={12}>
            <CollaboratorForm
              error={this.state.error}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
              onDelete={this.handleDelete}
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
}
