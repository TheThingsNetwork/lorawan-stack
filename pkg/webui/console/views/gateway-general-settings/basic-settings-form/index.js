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

import DeleteModalButton from '@ttn-lw/console/components/delete-modal-button'

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import KeyValueMap from '@ttn-lw/components/key-value-map'

import Require from '@console/lib/components/require'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { mapAttributesToFormValue } from '@console/lib/attributes'

import m from '../messages'

import validationSchema from './validation-schema'

const BasicSettingsForm = React.memo(props => {
  const {
    gateway,
    gtwId,
    onSubmit,
    onSubmitSuccess,
    onDelete,
    onDeleteFailure,
    onDeleteSuccess,
    mayDeleteGateway,
    mayEditSecrets,
    shouldConfirmDelete,
    mayPurge,
  } = props

  const [error, setError] = React.useState(undefined)

  const onGatewayDelete = React.useCallback(
    async shouldPurge => {
      try {
        await onDelete(shouldPurge)
        onDeleteSuccess()
      } catch (error) {
        onDeleteFailure()
      }
    },
    [onDelete, onDeleteFailure, onDeleteSuccess],
  )

  const initialValues = React.useMemo(() => {
    const initialValues = {
      ...gateway,
      attributes: mapAttributesToFormValue(gateway.attributes),
    }

    if (Boolean(initialValues.lbs_lns_secret) && initialValues.lbs_lns_secret.value !== undefined) {
      try {
        initialValues.lbs_lns_secret.value = atob(initialValues.lbs_lns_secret.value)
      } catch (e) {
        initialValues.lbs_lns_secret.value = ''
      }
    }
    return validationSchema.cast(initialValues)
  }, [gateway])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      if (Boolean(castedValues.lbs_lns_secret) && castedValues.lbs_lns_secret.value !== undefined) {
        castedValues.lbs_lns_secret.value = btoa(castedValues.lbs_lns_secret.value)
      }
      setError(undefined)
      try {
        await onSubmit(castedValues)
        resetForm({ values: castedValues })
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [onSubmit, onSubmitSuccess],
  )

  return (
    <Form
      error={error}
      onSubmit={onFormSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
      enableReinitialize
    >
      <Form.Field
        title={sharedMessages.gatewayID}
        name="ids.gateway_id"
        placeholder={sharedMessages.gatewayIdPlaceholder}
        required
        disabled
        component={Input}
        tooltipId={tooltipIds.GATEWAY_ID}
      />
      <Form.Field
        title={sharedMessages.gatewayEUI}
        name="ids.eui"
        type="byte"
        min={8}
        max={8}
        placeholder={sharedMessages.gatewayEUI}
        component={Input}
        tooltipId={tooltipIds.GATEWAY_EUI}
      />
      <Form.Field
        title={sharedMessages.gatewayName}
        placeholder={sharedMessages.gatewayNamePlaceholder}
        name="name"
        component={Input}
        tooltipId={tooltipIds.GATEWAY_NAME}
      />
      <Form.Field
        title={sharedMessages.gatewayDescription}
        description={sharedMessages.gatewayDescDescription}
        placeholder={sharedMessages.gatewayDescPlaceholder}
        name="description"
        type="textarea"
        component={Input}
        tooltipId={tooltipIds.GATEWAY_DESCRIPTION}
      />
      <Form.Field
        title={sharedMessages.gatewayServerAddress}
        description={sharedMessages.gsServerAddressDescription}
        placeholder={sharedMessages.addressPlaceholder}
        name="gateway_server_address"
        component={Input}
      />
      <Form.Field
        title={sharedMessages.requireAuthenticatedConnection}
        name="require_authenticated_connection"
        component={Checkbox}
        label={sharedMessages.enabled}
        description={sharedMessages.requireAuthenticatedConnectionDescription}
        tooltipId={tooltipIds.REQUIRE_AUTHENTICATED_CONNECTION}
      />
      <Form.Field
        title={sharedMessages.lbsLNSSecret}
        description={sharedMessages.lbsLNSSecretDescription}
        name="lbs_lns_secret.value"
        component={Input}
        disabled={!mayEditSecrets}
      />
      <Form.Field
        title={sharedMessages.gatewayStatus}
        name="status_public"
        component={Checkbox}
        label={sharedMessages.public}
        description={sharedMessages.statusDescription}
        tooltipId={tooltipIds.GATEWAY_STATUS}
      />
      <Form.Field
        title={sharedMessages.gatewayLocation}
        name="location_public"
        component={Checkbox}
        label={sharedMessages.public}
        description={sharedMessages.locationDescription}
        tooltipId={tooltipIds.GATEWAY_LOCATION}
      />
      <Form.Field
        name="attributes"
        title={sharedMessages.attributes}
        keyPlaceholder={sharedMessages.key}
        valuePlaceholder={sharedMessages.value}
        addMessage={sharedMessages.addAttributes}
        component={KeyValueMap}
        description={sharedMessages.attributeDescription}
        tooltipId={tooltipIds.GATEWAY_ATTRIBUTES}
      />
      <Form.Field
        title={sharedMessages.automaticUpdates}
        name="auto_update"
        component={Checkbox}
        description={sharedMessages.autoUpdateDescription}
      />
      <Form.Field
        title={sharedMessages.channel}
        description={sharedMessages.updateChannelDescription}
        placeholder={sharedMessages.stable}
        name="update_channel"
        component={Input}
        autoComplete="on"
      />
      <Form.Field
        title={sharedMessages.packetBroker}
        label={sharedMessages.disabled}
        name="disable_packet_broker_forwarding"
        component={Checkbox}
        description={m.disablePacketBrokerForwarding}
        tooltipId={tooltipIds.DISABLE_PACKET_BROKER_FORWARDING}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
        <Require featureCheck={mayDeleteGateway}>
          <DeleteModalButton
            entityId={gtwId}
            entityName={gateway.name}
            message={m.deleteGateway}
            onApprove={onGatewayDelete}
            shouldConfirm={shouldConfirmDelete}
            mayPurge={mayPurge}
          />
        </Require>
      </SubmitBar>
    </Form>
  )
})

BasicSettingsForm.propTypes = {
  gateway: PropTypes.gateway.isRequired,
  gtwId: PropTypes.string.isRequired,
  mayDeleteGateway: PropTypes.shape({}).isRequired,
  mayEditSecrets: PropTypes.bool.isRequired,
  mayPurge: PropTypes.bool.isRequired,
  onDelete: PropTypes.func.isRequired,
  onDeleteFailure: PropTypes.func.isRequired,
  onDeleteSuccess: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
  shouldConfirmDelete: PropTypes.bool.isRequired,
}

export default BasicSettingsForm
