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

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import CollaboratorForm from '../../components/collaborator-form'
import toast from '../../../components/toast'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import withRequest from '../../../lib/components/with-request'

import { getGatewayCollaborator } from '../../store/actions/gateways'
import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
  selectGatewayUserCollaborator,
  selectGatewayOrganizationCollaborator,
  selectGatewayCollaboratorFetching,
  selectGatewayCollaboratorError,
} from '../../store/selectors/gateways'

import api from '../../api'

@connect(
  function(state, props) {
    const gtwId = selectSelectedGatewayId(state, props)

    const { collaboratorId, collaboratorType } = props.match.params

    const collaborator =
      collaboratorType === 'user'
        ? selectGatewayUserCollaborator(state)
        : selectGatewayOrganizationCollaborator(state)

    const fetching = selectGatewayRightsFetching(state) || selectGatewayCollaboratorFetching(state)
    const error = selectGatewayRightsError(state) || selectGatewayCollaboratorError(state)

    return {
      collaboratorId,
      collaboratorType,
      collaborator,
      gtwId,
      rights: selectGatewayRights(state),
      pseudoRights: selectGatewayPseudoRights(state),
      fetching,
      error,
    }
  },
  (dispatch, ownProps) => ({
    getGatewayCollaborator(gtwId, collaboratorId, isUser) {
      dispatch(getGatewayCollaborator(gtwId, collaboratorId, isUser))
    },
    redirectToList(gtwId) {
      dispatch(replace(`/gateways/${gtwId}/collaborators`))
    },
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    getGatewayCollaborator: () =>
      dispatchProps.getGatewayCollaborator(
        stateProps.gtwId,
        stateProps.collaboratorId,
        stateProps.collaboratorType === 'user',
      ),
    redirectToList: () => dispatchProps.redirectToList(stateProps.gtwId),
  }),
)
@withRequest(
  ({ getGatewayCollaborator }) => getGatewayCollaborator(),
  ({ fetching, collaborator }) => fetching || !Boolean(collaborator),
)
@withBreadcrumb('gtws.single.collaborators.edit', function(props) {
  const { gtwId, collaboratorId, collaboratorType } = props

  return (
    <Breadcrumb
      path={`/gateways/${gtwId}/collaborators/${collaboratorType}/${collaboratorId}`}
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

  handleSubmit(updatedCollaborator) {
    const { gtwId } = this.props

    return api.gateway.collaborators.update(gtwId, updatedCollaborator)
  }

  handleSubmitSuccess() {
    toast({
      message: sharedMessages.collaboratorUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  async handleDelete(updatedCollaborator) {
    const { gtwId } = this.props

    return api.gateway.collaborators.remove(gtwId, updatedCollaborator)
  }

  render() {
    const { collaborator, rights, redirectToList, pseudoRights } = this.props

    return (
      <Container>
        <PageTitle
          title={sharedMessages.collaboratorEdit}
          values={{ collaboratorId: collaborator.id }}
        />
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
