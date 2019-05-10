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
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { ApiKeyEditForm } from '../../../components/api-key-form'

import { getApplicationApiKeyPageData } from '../../store/actions/application'
import {
  applicationRightsSelector,
  applicationRightsErrorSelector,
  applicationRightsFetchingSelector,
  applicationKeySelector,
  applicationKeysErrorSelector,
  applicationKeysFetchingSelector,
} from '../../store/selectors/application'

import api from '../../api'

@connect(function (state, props) {
  const { appId, apiKeyId } = props.match.params
  const ids = { id: appId, keyId: apiKeyId }

  const keysFetching = applicationKeysFetchingSelector(state, ids)
  const rightsFetching = applicationRightsFetchingSelector(state, props)
  const keysError = applicationKeysErrorSelector(state, ids)
  const rightsError = applicationRightsErrorSelector(state, props)
  const apiKey = applicationKeySelector(state, ids)
  const rights = applicationRightsSelector(state, props)

  return {
    keyId: apiKeyId,
    appId,
    apiKey,
    rights,
    fetching: keysFetching || rightsFetching,
    error: keysError || rightsError,
  }
})
@withBreadcrumb('apps.single.api-keys.edit', function (props) {
  const { appId, keyId } = props

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/api-keys/${keyId}/edit`}
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
      props.apiKey.id,
      key
    )
  }

  componentDidMount () {
    const { dispatch, appId } = this.props

    dispatch(getApplicationApiKeyPageData(appId))
  }

  onDeleteSuccess () {
    const { appId, dispatch } = this.props

    dispatch(replace(`/console/applications/${appId}/api-keys`))
  }

  render () {
    const { apiKey, rights, fetching, error } = this.props

    if (error) {
      return 'ERROR'
    }

    if (fetching || !apiKey) {
      return <Spinner center />
    }

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
