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
import { replace } from 'connected-react-router'

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import { ApiKeyCreateForm } from '../../components/api-key-form'

import { getApplicationsRightsList } from '../../store/actions/applications'
import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationPseudoRights,
  selectApplicationRightsError,
  selectApplicationRightsFetching,
} from '../../store/selectors/applications'

import api from '../../api'
import PropTypes from '../../../lib/prop-types'

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    fetching: selectApplicationRightsFetching(state),
    error: selectApplicationRightsError(state),
    rights: selectApplicationRights(state),
    pseudoRights: selectApplicationPseudoRights(state),
  }),
  dispatch => ({
    getApplicationsRightsList: appId => dispatch(getApplicationsRightsList(appId)),
    navigateToList: appId => dispatch(replace(`/applications/${appId}/api-keys`)),
  }),
)
@withBreadcrumb('apps.single.api-keys.add', function(props) {
  const appId = props.appId
  return (
    <Breadcrumb
      path={`/applications/${appId}/api-keys/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class ApplicationApiKeyAdd extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    navigateToList: PropTypes.func.isRequired,
    pseudoRights: PropTypes.rights.isRequired,
    rights: PropTypes.rights.isRequired,
  }

  constructor(props) {
    super(props)

    this.createApplicationKey = key => api.application.apiKeys.create(props.appId, key)
  }

  handleApprove() {
    const { navigateToList, appId } = this.props

    navigateToList(appId)
  }

  render() {
    const { rights, pseudoRights } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.addApiKey} />
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyCreateForm
              rights={rights}
              pseudoRights={pseudoRights}
              onCreate={this.createApplicationKey}
              onCreateSuccess={this.handleApprove}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
