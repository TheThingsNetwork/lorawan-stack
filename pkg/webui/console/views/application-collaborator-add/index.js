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

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import CollaboratorForm from '../../components/collaborator-form'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import withRequest from '../../../lib/components/with-request'

import { getApplicationsRightsList } from '../../store/actions/applications'
import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationUniversalRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '../../store/selectors/applications'

import api from '../../api'

@connect(
  function(state) {
    return {
      appId: selectSelectedApplicationId(state),
      collaborators: state.collaborators.applications.collaborators,
      rights: selectApplicationRights(state),
      universalRights: selectApplicationUniversalRights(state),
      fetching: selectApplicationRightsFetching(state),
      error: selectApplicationRightsError(state),
    }
  },
  (dispatch, ownProps) => ({
    redirectToList: appId => dispatch(push(`/applications/${appId}/collaborators`)),
    getApplicationsRightsList: appId => dispatch(getApplicationsRightsList(appId)),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    redirectToList: () => dispatchProps.redirectToList(stateProps.appId),
    getApplicationsRightsList: () => dispatchProps.getApplicationsRightsList(stateProps.appId),
  }),
)
@withRequest(
  ({ getApplicationsRightsList }) => getApplicationsRightsList(),
  ({ fetching, rights }) => fetching || !Boolean(rights.length),
)
@withBreadcrumb('apps.single.collaborators.add', function(props) {
  const appId = props.appId
  return (
    <Breadcrumb
      path={`/applications/${appId}/collaborators/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class ApplicationCollaboratorAdd extends React.Component {
  state = {
    error: '',
  }

  async handleSubmit(collaborator) {
    const { appId } = this.props

    await api.application.collaborators.add(appId, collaborator)
  }

  render() {
    const { rights, universalRights, redirectToList } = this.props

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.addCollaborator} />
            <Message component="h2" content={sharedMessages.addCollaborator} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <CollaboratorForm
              error={this.state.error}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={redirectToList}
              universalRights={universalRights}
              rights={rights}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
