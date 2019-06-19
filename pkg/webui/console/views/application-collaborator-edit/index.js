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
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import toast from '../../../components/toast'

import {
  getApplicationCollaboratorsList,
  getApplicationsRightsList,
} from '../../store/actions/applications'
import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationUniversalRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '../../store/selectors/applications'

import api from '../../api'

@connect(function (state, props) {
  const appId = selectSelectedApplicationId(state)
  const { collaboratorId } = props.match.params
  const collaboratorsFetching = state.collaborators.applications.fetching
  const collaboratorsError = state.collaborators.applications.error

  const appCollaborators = state.collaborators.applications[appId]
  const collaborator = appCollaborators ? appCollaborators.collaborators
    .find(c => c.id === collaboratorId) : undefined

  const fetching = selectApplicationRightsFetching(state) || collaboratorsFetching
  const error = selectApplicationRightsError(state) || collaboratorsError

  return {
    collaboratorId,
    collaborator,
    appId,
    rights: selectApplicationRights(state),
    universalRights: selectApplicationUniversalRights(state),
    fetching,
    error,
  }
}, function (dispatch, ownProps) {
  const appId = ownProps.match.params.appId
  return {
    async loadData () {
      await dispatch(getApplicationsRightsList(appId))
      dispatch(getApplicationCollaboratorsList(appId))
    },
    redirectToList () {
      dispatch(replace(`/console/applications/${appId}/collaborators`))
    },
  }
})
@withBreadcrumb('apps.single.collaborators.edit', function (props) {
  const { appId, collaboratorId } = props

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/collaborators/${collaboratorId}/edit`}
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

  componentDidMount () {
    const { loadData, appId } = this.props

    loadData(appId)
  }

  async handleSubmit (updatedCollaborator) {
    const { appId } = this.props

    await api.application.collaborators.update(appId, updatedCollaborator)
  }

  handleSubmitSuccess () {
    toast({
      message: sharedMessages.collaboratorUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  async handleDelete () {
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

  render () {
    const {
      collaborator,
      rights,
      fetching,
      error,
      universalRights,
      redirectToList,
    } = this.props

    if (error) {
      throw error
    }

    if (fetching || !collaborator) {
      return <Spinner center />
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
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
              universalRightLiterals={universalRights}
              rights={rights}
              update
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
