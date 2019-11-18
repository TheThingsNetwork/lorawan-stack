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

import React, { Component } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import PubsubForm from '../../components/pubsub-form'

import sharedMessages from '../../../lib/shared-messages'

import { selectSelectedApplicationId } from '../../store/selectors/applications'

import api from '../../api'

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
  }),
  dispatch => ({
    navigateToList: appId => dispatch(push(`/applications/${appId}/integrations/pubsubs`)),
  }),
)
@withBreadcrumb('apps.single.integrations.add', function(props) {
  const { appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/integrations/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class ApplicationPubsubAdd extends Component {
  async handleSubmit(pubsub) {
    const { appId } = this.props

    await api.application.pubsubs.create(appId, pubsub)
  }

  handleSubmitSuccess() {
    const { navigateToList, appId } = this.props

    navigateToList(appId)
  }

  render() {
    const { appId } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.addPubsub} />
        <Row>
          <Col lg={8} md={12}>
            <PubsubForm
              appId={appId}
              update={false}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
