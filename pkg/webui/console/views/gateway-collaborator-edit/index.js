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
  getGatewayCollaboratorsList,
  getGatewaysRightsList,
} from '../../store/actions/gateways'
import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayUniversalRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
} from '../../store/selectors/gateway'

import api from '../../api'

@connect(function (state, props) {
  const gtwId = selectSelectedGatewayId(state, props)
  const { collaboratorId } = props.match.params
  const collaboratorsFetching = state.collaborators.gateways.fetching
  const collaboratorsError = state.collaborators.gateways.error

  const gtwCollaborators = state.collaborators.gateways[gtwId]
  const collaborator = gtwCollaborators ? gtwCollaborators.collaborators
    .find(c => c.id === collaboratorId) : undefined

  const fetching = selectGatewayRightsFetching(state) || collaboratorsFetching
  const error = selectGatewayRightsError(state) || collaboratorsError

  return {
    collaboratorId,
    collaborator,
    gtwId,
    rights: selectGatewayRights(state),
    universalRights: selectGatewayUniversalRights(state),
    fetching,
    error,
  }

}, (dispatch, ownProps) => ({
  async loadData (gtwId) {
    dispatch(getGatewaysRightsList(gtwId))
    dispatch(getGatewayCollaboratorsList(gtwId))
  },
  redirectToList (gtwId) {
    dispatch(replace(`/console/gateways/${gtwId}/collaborators`))
  },
}), (stateProps, dispatchProps, ownProps) => ({
  ...stateProps, ...dispatchProps, ...ownProps,
  loadData: () => dispatchProps.loadData(stateProps.gtwId),
  redirectToList: () => dispatchProps.redirectToList(stateProps.gtwId),
}))
@withBreadcrumb('gtws.single.collaborators.edit', function (props) {
  const { gtwId, collaboratorId } = props

  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}/collaborators/${collaboratorId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class GatewayCollaboratorEdit extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { loadData } = this.props

    loadData()
  }

  handleSubmit (updatedCollaborator) {
    const { gtwId } = this.props

    return api.gateway.collaborators.update(gtwId, updatedCollaborator)
  }

  handleSubmitSuccess () {
    toast({
      message: sharedMessages.collaboratorUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  async handleDelete (updatedCollaborator) {
    const { gtwId } = this.props

    return api.gateway.collaborators.remove(gtwId, updatedCollaborator)
  }

  render () {
    const { collaborator, rights, fetching, error, redirectToList, universalRights } = this.props

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
