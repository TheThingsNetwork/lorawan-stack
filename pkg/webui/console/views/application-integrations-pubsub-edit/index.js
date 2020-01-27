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
import { defineMessages } from 'react-intl'
import { replace } from 'connected-react-router'

import PropTypes from '../../../lib/prop-types'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import PageTitle from '../../../components/page-title'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import PubsubForm from '../../components/pubsub-form'
import toast from '../../../components/toast'
import diff from '../../../lib/diff'
import sharedMessages from '../../../lib/shared-messages'
import withRequest from '../../../lib/components/with-request'

import {
  selectSelectedPubsub,
  selectPubsubFetching,
  selectPubsubError,
} from '../../store/selectors/pubsubs'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { getPubsub } from '../../store/actions/pubsubs'

import api from '../../api'

const m = defineMessages({
  editPubsub: 'Edit PubSub',
  updateSuccess: 'Successfully updated PubSub',
  deleteSuccess: 'Successfully deleted PubSub',
})

const pubsubEntitySelector = [
  'base_topic',
  'format',
  'provider.nats',
  'provider.mqtt',
  'downlink_ack',
  'downlink_failed',
  'downlink_nack',
  'downlink_push',
  'downlink_queued',
  'downlink_replace',
  'downlink_sent',
  'join_accept',
  'location_solved',
  'uplink_message',
]

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    pubsub: selectSelectedPubsub(state),
    fetching: selectPubsubFetching(state),
    error: selectPubsubError(state),
  }),
  function(dispatch, { match }) {
    const { appId, pubsubId } = match.params
    return {
      getPubsub: () => dispatch(getPubsub(appId, pubsubId, pubsubEntitySelector)),
      navigateToList: () => dispatch(replace(`/applications/${appId}/integrations/pubsubs`)),
    }
  },
)
@withRequest(({ getPubsub }) => getPubsub(), ({ fetching, pubsub }) => fetching || !Boolean(pubsub))
@withBreadcrumb('apps.single.integrations.edit', function(props) {
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
@bind
export default class ApplicationPubsubEdit extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    match: PropTypes.match.isRequired,
    navigateToList: PropTypes.func.isRequired,
    pubsub: PropTypes.pubsub.isRequired,
  }

  async handleSubmit(pubsub) {
    const {
      appId,
      match: {
        params: { pubsubId },
      },
      pubsub: originalPubsub,
    } = this.props
    const patch = diff(originalPubsub, pubsub, ['ids'])

    await api.application.pubsubs.update(appId, pubsubId, patch)
  }

  handleSubmitSuccess() {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  async handleDelete() {
    const {
      appId,
      match: {
        params: { pubsubId },
      },
    } = this.props

    await api.application.pubsubs.delete(appId, pubsubId)
  }

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
