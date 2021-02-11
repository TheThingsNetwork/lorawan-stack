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
import { Col, Row, Container } from 'react-grid-system'
import { bindActionCreators } from 'redux'
import { replace } from 'connected-react-router'

import toast from '@ttn-lw/components/toast'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Collapse from '@ttn-lw/components/collapse'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import diff from '@ttn-lw/lib/diff'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayEditBasicGatewayInformation, mayDeleteGateway } from '@console/lib/feature-checks'
import { mapFormValueToAttributes } from '@console/lib/attributes'

import { updateGateway, deleteGateway } from '@console/store/actions/gateways'

import { selectSelectedGateway, selectSelectedGatewayId } from '@console/store/selectors/gateways'

import LorawanSettingsForm from './lorawan-settings-form'
import BasicSettingsForm from './basic-settings-form'
import m from './messages'

@connect(
  state => ({
    gateway: selectSelectedGateway(state),
    gtwId: selectSelectedGatewayId(state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        updateGateway: attachPromise(updateGateway),
        deleteGateway: attachPromise(deleteGateway),
      },
      dispatch,
    ),
    onDeleteSuccess: () => dispatch(replace('/gateways')),
  }),
)
@withFeatureRequirement(mayEditBasicGatewayInformation, {
  redirect: ({ gtwId }) => `/gateways/${gtwId}`,
})
@withBreadcrumb('gateways.single.general-settings', function (props) {
  const { gtwId } = props

  return (
    <Breadcrumb
      path={`/gateways/${gtwId}/general-settings`}
      content={sharedMessages.generalSettings}
    />
  )
})
export default class GatewayGeneralSettings extends React.Component {
  static propTypes = {
    deleteGateway: PropTypes.func.isRequired,
    gateway: PropTypes.gateway.isRequired,
    gtwId: PropTypes.string.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    updateGateway: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(values) {
    const { gtwId, updateGateway, gateway } = this.props

    const changed = diff(gateway, values)

    const update =
      'attributes' in changed
        ? { ...changed, attributes: mapFormValueToAttributes(values.attributes) }
        : changed
    return updateGateway(gtwId, update)
  }

  @bind
  async handleSubmitSuccess() {
    const { gateway } = this.props

    const {
      ids: { gateway_id: gatewayId },
    } = gateway

    toast({
      title: gatewayId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDelete() {
    const { gtwId, deleteGateway } = this.props

    return deleteGateway(gtwId)
  }

  @bind
  async handleDeleteSuccess() {
    const { gateway, onDeleteSuccess } = this.props
    const {
      ids: { gateway_id: gatewayId },
    } = gateway

    onDeleteSuccess()
    toast({
      title: gatewayId,
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDeleteFailure() {
    const { gateway } = this.props
    const {
      ids: { gateway_id: gatewayId },
    } = gateway

    toast({
      title: gatewayId,
      message: m.deleteFailure,
      type: toast.types.ERROR,
    })
  }

  render() {
    const { gateway } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.generalSettings} hideHeading />
        <Row>
          <Col lg={8} md={12}>
            <Collapse
              title={m.basicTitle}
              description={m.basicDescription}
              disabled={false}
              initialCollapsed={false}
            >
              <BasicSettingsForm
                gateway={gateway}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                onDelete={this.handleDelete}
                onDeleteSuccess={this.handleDeleteSuccess}
                onDeleteFailure={this.handleDeleteFailure}
                mayDeleteGateway={mayDeleteGateway}
              />
            </Collapse>
            <Collapse
              title={m.lorawanTitle}
              description={m.lorawanDescription}
              disabled={false}
              initialCollapsed
            >
              <LorawanSettingsForm
                gateway={gateway}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
              />
            </Collapse>
          </Col>
        </Row>
      </Container>
    )
  }
}
