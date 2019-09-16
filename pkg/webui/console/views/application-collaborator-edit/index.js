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
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { Container, Col, Row } from 'react-grid-system'
import { replace } from 'connected-react-router'

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import CollaboratorForm from '../../components/collaborator-form'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import toast from '../../../components/toast'
import withRequest from '../../../lib/components/with-request'

import {
  getApplicationCollaborator,
  getApplicationsRightsList,
} from '../../store/actions/applications'
import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationUniversalRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
  selectApplicationUserCollaborator,
  selectApplicationOrganizationCollaborator,
  selectApplicationCollaboratorFetching,
  selectApplicationCollaboratorError,
} from '../../store/selectors/applications'

import api from '../../api'

@connect(
  function(state, props) {
    const appId = selectSelectedApplicationId(state)

    const { collaboratorId, collaboratorType } = props.match.params

    const collaborator =
      collaboratorType === 'user'
        ? selectApplicationUserCollaborator(state)
        : selectApplicationOrganizationCollaborator(state)
    const fetching =
      selectApplicationRightsFetching(state) || selectApplicationCollaboratorFetching(state)
    const error = selectApplicationRightsError(state) || selectApplicationCollaboratorError(state)

    return {
      collaboratorId,
      collaboratorType,
      collaborator,
      appId,
      rights: selectApplicationRights(state),
      universalRights: selectApplicationUniversalRights(state),
      fetching,
      error,
    }
  },
  (dispatch, ownProps) => ({
    loadData(appId, collaboratorId, isUser) {
      dispatch(getApplicationsRightsList(appId))
      dispatch(getApplicationCollaborator(appId, collaboratorId, isUser))
    },
    redirectToList(appId) {
      dispatch(replace(`/applications/${appId}/collaborators`))
    },
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    loadData: () =>
      dispatchProps.loadData(
        stateProps.appId,
        stateProps.collaboratorId,
        stateProps.collaboratorType === 'user',
      ),
    redirectToList: () => dispatchProps.redirectToList(stateProps.appId),
  }),
)
@withRequest(
  ({ loadData }) => loadData(),
  ({ fetching, collaborator }) => fetching || !Boolean(collaborator),
)
@withBreadcrumb('apps.single.collaborators.edit', function(props) {
  const { appId, collaboratorId, collaboratorType } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/collaborators/${collaboratorType}/${collaboratorId}`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class ApplicationCollaboratorEdit extends React.Component {
  state = {
    error: '',
  }

  async handleSubmit(updatedCollaborator) {
    const { appId } = this.props

    await api.application.collaborators.update(appId, updatedCollaborator)
  }

  handleSubmitSuccess() {
    toast({
      message: sharedMessages.collaboratorUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  async handleDelete() {
    const { collaborator, redirectToList, appId } = this.props
    const collaborator_type = collaborator.isUser ? 'user' : 'organization'

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator.id,
      },
    }
    const updatedCollaborator = {
      ids: collaborator_ids,
    }

    try {
      await api.application.collaborators.remove(appId, updatedCollaborator)
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
    const { collaborator, rights, universalRights, redirectToList } = this.props

    return (
      <Container>
        <Row>
          <Col>
            <IntlHelmet
              title={sharedMessages.collaboratorEdit}
              values={{ collaboratorId: collaborator.id }}
            />
            <Message
              component="h2"
              content={sharedMessages.collaboratorEditRights}
              values={{ collaboratorId: collaborator.id }}
            />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <CollaboratorForm
              error={this.state.error}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
              onDelete={this.handleDelete}
              onDeleteSuccess={redirectToList}
              collaborator={collaborator}
              universalRights={universalRights}
              rights={rights}
              update
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
