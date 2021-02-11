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

import api from '@console/api'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import PubsubForm from '@console/components/pubsub-form'

import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
  }),
  dispatch => ({
    navigateToList: appId => dispatch(push(`/applications/${appId}/integrations/pubsubs`)),
  }),
)
@withBreadcrumb('apps.single.integrations.add', function (props) {
  const { appId } = props
  return (
    <Breadcrumb path={`/applications/${appId}/integrations/add`} content={sharedMessages.add} />
  )
})
export default class ApplicationPubsubAdd extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    navigateToList: PropTypes.func.isRequired,
  }

  @bind
  async existCheck(pubsubId) {
    const { appId } = this.props

    try {
      await api.application.pubsubs.get(appId, pubsubId, [])
      return true
    } catch (error) {
      if (isNotFoundError(error)) {
        return false
      }

      throw error
    }
  }

  @bind
  async handleSubmit(pubsub) {
    const { appId } = this.props

    await api.application.pubsubs.create(appId, pubsub)
  }

  @bind
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
              existCheck={this.existCheck}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
