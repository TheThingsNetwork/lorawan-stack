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

import Spinner from '../../../components/spinner'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { ApiKeyCreateForm } from '../../../components/api-key-form'

import { getApplicationsRightsList } from '../../store/actions/applications'
import {
  applicationRightsSelector,
  applicationRightsErrorSelector,
  applicationRightsFetchingSelector,
} from '../../store/selectors/application'
import api from '../../api'

@connect(function (state, props) {
  const appId = props.match.params.appId

  return {
    appId,
    fetching: applicationRightsFetchingSelector(state, props),
    error: applicationRightsErrorSelector(state, props),
    rights: applicationRightsSelector(state, props),
  }
})
@withBreadcrumb('apps.single.api-keys.add', function (props) {
  const appId = props.appId
  return (
    <Breadcrumb
      path={`/console/applications/${appId}/api-keys/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class ApplicationApiKeyAdd extends React.Component {

  constructor (props) {
    super(props)

    this.createApplicationKey = key => api.application.apiKeys.create(props.appId, key)
  }

  componentDidMount () {
    const { dispatch, appId } = this.props

    dispatch(getApplicationsRightsList(appId))
  }

  handleApprove () {
    const { dispatch, appId } = this.props

    dispatch(replace(`/console/applications/${appId}/api-keys`))
  }

  render () {
    const { rights, fetching, error } = this.props

    if (error) {
      return 'ERROR'
    }

    if (fetching || !rights.length) {
      return <Spinner center />
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.addApiKey} />
            <Message component="h2" content={sharedMessages.addApiKey} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <ApiKeyCreateForm
              rights={rights}
              onCreate={this.createApplicationKey}
              onCreateSuccess={this.handleApprove}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
