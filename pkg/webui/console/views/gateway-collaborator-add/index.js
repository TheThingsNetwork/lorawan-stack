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
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import api from '@console/api'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import CollaboratorForm from '@console/components/collaborator-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getGatewaysRightsList } from '@console/store/actions/gateways'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
} from '@console/store/selectors/gateways'
import { selectSelectedCollaborator } from '@console/store/selectors/collaborators'

@connect(
  state => ({
    gtwId: selectSelectedGatewayId(state),
    collaborators: selectSelectedCollaborator(state),
    fetching: selectGatewayRightsFetching(state),
    error: selectGatewayRightsError(state),
    rights: selectGatewayRights(state),
    pseudoRights: selectGatewayPseudoRights(state),
  }),
  (dispatch, ownProps) => ({
    getGatewaysRightsList: gtwId => dispatch(getGatewaysRightsList(gtwId)),
    redirectToList: gtwId => dispatch(push(`/gateways/${gtwId}/collaborators`)),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    getGatewaysRightsList: () => dispatchProps.getGatewaysRightsList(stateProps.gtwId),
    redirectToList: () => dispatchProps.redirectToList(stateProps.gtwId),
  }),
)
@withBreadcrumb('gtws.single.collaborators.add', function(props) {
  const gtwId = props.gtwId
  return <Breadcrumb path={`/gateways/${gtwId}/collaborators/add`} content={sharedMessages.add} />
})
@bind
export default class GatewayCollaboratorAdd extends React.Component {
  static propTypes = {
    gtwId: PropTypes.string.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    redirectToList: PropTypes.func.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  state = {
    error: '',
  }

  handleSubmit(collaborator) {
    const { gtwId } = this.props

    return api.gateway.collaborators.add(gtwId, collaborator)
  }

  render() {
    const { rights, redirectToList, pseudoRights } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.addCollaborator} />
        <Row>
          <Col lg={8} md={12}>
            <CollaboratorForm
              error={this.state.error}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={redirectToList}
              pseudoRights={pseudoRights}
              rights={rights}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
