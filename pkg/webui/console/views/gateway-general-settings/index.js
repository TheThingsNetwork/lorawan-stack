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
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { Col, Row, Container } from 'react-grid-system'
import { bindActionCreators } from 'redux'
import { replace } from 'connected-react-router'

import toast from '../../../components/toast'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import GatewayDataForm from '../../components/gateway-data-form'
import ModalButton from '../../../components/button/modal-button'
import FormSubmit from '../../../components/form/submit'
import SubmitButton from '../../../components/submit-button'
import PageTitle from '../../../components/page-title'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'
import Require from '../../lib/components/require'
import diff from '../../../lib/diff'

import { updateGateway, deleteGateway } from '../../store/actions/gateways'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedGateway, selectSelectedGatewayId } from '../../store/selectors/gateways'
import { mayEditBasicGatewayInformation, mayDeleteGateway } from '../../lib/feature-checks'

const m = defineMessages({
  updateSuccess: 'Successfully updated gateway',
  deleteGateway: 'Delete Gateway',
  modalWarning:
    'Are you sure you want to delete "{gtwName}"? This action cannot be undone and it will not be possible to reuse the gateway ID and EUI!',
})

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
@withBreadcrumb('gateways.single.general-settings', function(props) {
  const { gtwId } = props

  return (
    <Breadcrumb
      path={`/gateways/${gtwId}/general-settings`}
      icon="general_settings"
      content={sharedMessages.generalSettings}
    />
  )
})
@bind
export default class GatewayGeneralSettings extends React.Component {
  static propTypes = {
    deleteGateway: PropTypes.func.isRequired,
    gateway: PropTypes.object.isRequired,
    gtwId: PropTypes.string.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    updateGateway: PropTypes.func.isRequired,
  }

  constructor(props) {
    super(props)

    this.formRef = React.createRef()
  }

  state = {
    error: '',
  }

  async handleSubmit(values) {
    const { gtwId, gateway, updateGateway } = this.props

    await this.setState({ error: '' })

    const { ids: valuesIds, ...valuesRest } = values
    const { ids: gatewayIds, ...gatewayRest } = gateway

    const idsDiff = diff(gatewayIds, valuesIds)
    const entityDiff = diff({ ...gatewayRest }, { ...valuesRest })

    let changed
    if (Object.keys(idsDiff).length) {
      changed = { ids: idsDiff, ...entityDiff }
    } else {
      changed = entityDiff
    }

    try {
      await updateGateway(gtwId, changed)
      this.formRef.current.resetForm()
      toast({
        title: gtwId,
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      this.formRef.current.resetForm(values)
      await this.setState({ error })
    }
  }

  async handleDelete() {
    const { gtwId, deleteGateway, onDeleteSuccess } = this.props

    await this.setState({ error: '' })

    try {
      await deleteGateway(gtwId)
      onDeleteSuccess()
    } catch (error) {
      this.formRef.current.setSubmitting(false)
      this.setState({ error })
    }
  }

  render() {
    const { gateway, gtwId } = this.props
    const { error } = this.state
    const {
      ids,
      gateway_server_address,
      frequency_plan_id,
      enforce_duty_cycle,
      name,
      description,
    } = gateway

    const initialValues = {
      ids: { ...ids },
      gateway_server_address,
      frequency_plan_id,
      enforce_duty_cycle,
      name,
      description,
    }

    return (
      <Container>
        <PageTitle title={sharedMessages.generalSettings} />
        <Row>
          <Col lg={8} md={12}>
            <GatewayDataForm
              error={error}
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              formRef={this.formRef}
              update
            >
              <FormSubmit component={SubmitButton} message={sharedMessages.saveChanges} />
              <Require featureCheck={mayDeleteGateway}>
                <ModalButton
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={m.deleteGateway}
                  modalData={{ message: { values: { gtwName: name || gtwId }, ...m.modalWarning } }}
                  onApprove={this.handleDelete}
                />
              </Require>
            </GatewayDataForm>
          </Col>
        </Row>
      </Container>
    )
  }
}
