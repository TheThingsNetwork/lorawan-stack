// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'

import api from '@console/api'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import toast from '@ttn-lw/components/toast'

import PubsubForm from '@console/components/pubsub-form'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  editPubsub: 'Edit Pub/Sub',
  updateSuccess: 'Pub/Sub updated',
  deleteSuccess: 'Pub/Sub deleted',
})

@withBreadcrumb('apps.single.integrations.edit', function (props) {
  const {
    appId,
    match: {
      params: { pubsubId },
    },
  } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/integrations/${pubsubId}`}
      content={sharedMessages.edit}
    />
  )
})
export default class ApplicationPubsubEdit extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    match: PropTypes.match.isRequired,
    navigateToList: PropTypes.func.isRequired,
    pubsub: PropTypes.pubsub.isRequired,
    updatePubsub: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(pubsub) {
    const { updatePubsub } = this.props

    await updatePubsub(pubsub)
  }

  @bind
  handleSubmitSuccess() {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDelete() {
    const {
      appId,
      match: {
        params: { pubsubId },
      },
    } = this.props

    await api.application.pubsubs.delete(appId, pubsubId)
  }

  @bind
  async handleDeleteSuccess() {
    const { navigateToList } = this.props

    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })

    navigateToList()
  }

  render() {
    const { pubsub, appId } = this.props

    return (
      <Container>
        <PageTitle title={m.editPubsub} />
        <Row>
          <Col lg={8} md={12}>
            <PubsubForm
              update
              appId={appId}
              initialPubsubValue={pubsub}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
              onDelete={this.handleDelete}
              onDeleteSuccess={this.handleDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
