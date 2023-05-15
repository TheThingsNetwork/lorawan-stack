// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { isEqual } from 'lodash'

import toast from '@ttn-lw/components/toast'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Collapse from '@ttn-lw/components/collapse'

import withRequest from '@ttn-lw/lib/components/with-request'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import diff from '@ttn-lw/lib/diff'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import {
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
  selectCollaboratorError,
} from '@ttn-lw/lib/store/selectors/collaborators'

import {
  checkFromState,
  mayEditBasicGatewayInformation,
  mayDeleteGateway,
  mayEditGatewaySecrets,
  mayPurgeEntities,
  mayViewOrEditGatewayApiKeys,
  mayViewOrEditGatewayCollaborators,
} from '@console/lib/feature-checks'
import { mapFormValueToAttributes } from '@console/lib/attributes'

import { updateGateway, deleteGateway } from '@console/store/actions/gateways'
import { getApiKeysList } from '@console/store/actions/api-keys'

import {
  selectApiKeysTotalCount,
  selectApiKeysFetching,
  selectApiKeysError,
} from '@console/store/selectors/api-keys'
import { selectSelectedGateway, selectSelectedGatewayId } from '@console/store/selectors/gateways'

import LorawanSettingsForm from './lorawan-settings-form'
import BasicSettingsForm from './basic-settings-form'
import m from './messages'

@connect(
  state => {
    const mayViewApiKeys = checkFromState(mayViewOrEditGatewayApiKeys, state)
    const mayViewCollaborators = checkFromState(mayViewOrEditGatewayCollaborators, state)
    const apiKeysCount = selectApiKeysTotalCount(state)
    const collaboratorsCount = selectCollaboratorsTotalCount(state)
    const mayEditSecrets = checkFromState(mayEditGatewaySecrets, state)
    const mayPurgeGtw = checkFromState(mayPurgeEntities, state)
    const mayDeleteGtw = checkFromState(mayDeleteGateway, state)

    const entitiesFetching = selectApiKeysFetching(state) || selectCollaboratorsFetching(state)
    const error = selectApiKeysError(state) || selectCollaboratorError(state)

    const fetching =
      entitiesFetching ||
      (mayViewApiKeys && apiKeysCount === undefined) ||
      (mayViewCollaborators && collaboratorsCount === undefined)
    const hasApiKeys = apiKeysCount > 0
    // Note: there is always at least one collaborator.
    const hasAddedCollaborators = collaboratorsCount > 1
    const isPristine = !hasAddedCollaborators && !hasApiKeys

    return {
      gateway: selectSelectedGateway(state),
      gtwId: selectSelectedGatewayId(state),
      mayEditSecrets,
      mayViewApiKeys,
      mayViewCollaborators,
      fetching,
      error,
      mayPurge: mayPurgeGtw,
      shouldConfirmDelete:
        !isPristine || !mayViewCollaborators || !mayViewApiKeys || Boolean(error),
      mayDeleteGateway: mayDeleteGtw,
    }
  },
  dispatch => ({
    ...bindActionCreators(
      {
        updateGateway: attachPromise(updateGateway),
        deleteGateway: attachPromise(deleteGateway),
        getApiKeysList,
        getCollaboratorsList,
      },
      dispatch,
    ),
    onDeleteSuccess: () => dispatch(replace('/gateways')),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    deleteGateway: (id, purge = false) => dispatchProps.deleteGateway(id, { purge }),
    loadData: () => {
      if (stateProps.mayDeleteGateway) {
        if (stateProps.mayViewApiKeys) {
          dispatchProps.getApiKeysList('gateway', stateProps.gtwId)
        }

        if (stateProps.mayViewCollaborators) {
          dispatchProps.getCollaboratorsList('gateway', stateProps.gtwId)
        }
      }
    },
  }),
)
@withFeatureRequirement(mayEditBasicGatewayInformation, {
  redirect: ({ gtwId }) => `/gateways/${gtwId}`,
})
@withRequest(({ loadData }) => loadData())
@withBreadcrumb('gateways.single.general-settings', props => {
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
    mayEditSecrets: PropTypes.bool.isRequired,
    mayPurge: PropTypes.bool.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    shouldConfirmDelete: PropTypes.bool.isRequired,
    updateGateway: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(values) {
    const { gtwId, updateGateway, gateway } = this.props
    const formValues = { ...values }

    const attributes = mapFormValueToAttributes(formValues.attributes)
    if (isEqual(gateway.attributes || {}, attributes)) {
      delete formValues.attributes
    }

    const changed = diff(gateway, formValues)

    const update = 'attributes' in changed ? { ...changed, attributes } : changed
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
  async handleDelete(shouldPurge) {
    const { gtwId, deleteGateway } = this.props

    return deleteGateway(gtwId, shouldPurge)
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
    const { gtwId, gateway, shouldConfirmDelete, mayPurge, mayEditSecrets } = this.props

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
                gtwId={gtwId}
                gateway={gateway}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                onDelete={this.handleDelete}
                onDeleteSuccess={this.handleDeleteSuccess}
                onDeleteFailure={this.handleDeleteFailure}
                mayDeleteGateway={mayDeleteGateway}
                mayEditSecrets={mayEditSecrets}
                shouldConfirmDelete={shouldConfirmDelete}
                mayPurge={mayPurge}
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
