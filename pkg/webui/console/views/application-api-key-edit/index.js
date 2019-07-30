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
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { ApiKeyEditForm } from '../../components/api-key-form'
import withRequest from '../../../lib/components/with-request'

import {
  getApplicationApiKey,
  getApplicationsRightsList,
} from '../../store/actions/applications'
import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationUniversalRights,
  selectApplicationRightsError,
  selectApplicationRightsFetching,
  selectApplicationApiKey,
  selectApplicationApiKeyError,
  selectApplicationApiKeyFetching,
} from '../../store/selectors/applications'

import api from '../../api'

@connect(function (state, props) {
  const { apiKeyId } = props.match.params

  const keyFetching = selectApplicationApiKeyFetching(state)
  const rightsFetching = selectApplicationRightsFetching(state)
  const keyError = selectApplicationApiKeyError(state)
  const rightsError = selectApplicationRightsError(state)

  return {
    keyId: apiKeyId,
    appId: selectSelectedApplicationId(state),
    apiKey: selectApplicationApiKey(state),
    rights: selectApplicationRights(state),
    universalRights: selectApplicationUniversalRights(state),
    fetching: keyFetching || rightsFetching,
    error: keyError || rightsError,
  }
},
dispatch => ({
  loadData (appId, apiKeyId) {
    dispatch(getApplicationsRightsList(appId))
    dispatch(getApplicationApiKey(appId, apiKeyId))
  },
  deleteSuccess: appId => dispatch(replace(`/applications/${appId}/api-keys`)),
}))
@withRequest(
  ({ loadData, appId, keyId }) => loadData(appId, keyId),
  ({ fetching, apiKey }) => fetching || !Boolean(apiKey)
)
@withBreadcrumb('apps.single.api-keys.edit', function (props) {
  const { appId, keyId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/api-keys/${keyId}`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class ApplicationApiKeyEdit extends React.Component {

  constructor (props) {
    super(props)

    this.deleteApplicationKey = id => api.application.apiKeys.delete(props.appId, id)
    this.editApplicationKey = key => api.application.apiKeys.update(
      props.appId,
      props.keyId,
      key
    )
  }

  onDeleteSuccess () {
    const { appId, deleteSuccess } = this.props

    deleteSuccess(appId)
  }

  render () {
    const { apiKey, rights, universalRights } = this.props

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.keyEdit} />
            <Message component="h2" content={sharedMessages.keyEdit} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyEditForm
              rights={rights}
              universalRights={universalRights}
              apiKey={apiKey}
              onEdit={this.editApplicationKey}
              onDelete={this.deleteApplicationKey}
              onDeleteSuccess={this.onDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
