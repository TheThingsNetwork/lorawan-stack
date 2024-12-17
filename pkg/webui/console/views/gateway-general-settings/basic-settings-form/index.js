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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'

import { GATEWAY } from '@console/constants/entities'

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Button from '@ttn-lw/components/button'
import { IconTrash } from '@ttn-lw/components/icon'

import CollaboratorSelect from '@ttn-lw/containers/collaborator-select'
import { decodeContact, encodeContact } from '@ttn-lw/containers/collaborator-select/util'

import Message from '@ttn-lw/lib/components/message'

import DeleteEntityHeaderModal from '@console/containers/delete-entity-header-modal'

import Require from '@console/lib/components/require'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { encodeAttributes, decodeAttributes } from '@console/lib/attributes'
import { checkFromState, mayDeleteGateway } from '@console/lib/feature-checks'

import { selectIsConfiguration } from '@console/store/selectors/identity-server'
import { selectUserId } from '@console/store/selectors/user'

import m from '../messages'

import validationSchema from './validation-schema'

const decodeSecret = value => {
  if (Boolean(value)) {
    return atob(value)
  }

  return ''
}

const encodeSecret = value => {
  if (Boolean(value)) {
    return btoa(value)
  }

  return ''
}

const BasicSettingsForm = React.memo(props => {
  const { gateway, gtwId, onSubmit, mayEditSecrets, supportsClaiming } = props

  const userId = useSelector(selectUserId)
  const isConfig = useSelector(selectIsConfiguration)
  const mayDeleteGtw = useSelector(state => checkFromState(mayDeleteGateway, state))
  const isResctrictedUser =
    isConfig && isConfig.collaborator_rights?.set_others_as_contacts === false
  const [error, setError] = React.useState(undefined)
  const [deleteGtwVisible, setDeleteGtwVisible] = React.useState(false)

  const handleOpenDeleteGatewayModal = useCallback(() => {
    setDeleteGtwVisible(true)
  }, [])

  const initialValues = React.useMemo(() => validationSchema.cast(gateway), [gateway])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      if (castedValues?.lbs_lns_secret?.value === '') {
        castedValues.lbs_lns_secret = null
      }

      const update = castedValues
      setError(undefined)
      try {
        await onSubmit(update)
        resetForm({ values: castedValues })
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [onSubmit],
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
        decode={decodeSecret}
        encode={encodeSecret}
        component={Input}
        disabled={!mayEditSecrets}
        sensitive
      />
      <Form.Field
        title={sharedMessages.gatewayStatus}
        name="status_public"
        component={Checkbox}
        label={sharedMessages.gatewayStatusPublic}
        description={sharedMessages.statusDescription}
        tooltipId={tooltipIds.GATEWAY_STATUS_PUBLIC}
      />
      <Form.Field
        title={sharedMessages.gatewayLocation}
        name="location_public"
        component={Checkbox}
        label={sharedMessages.gatewayLocationPublic}
        description={sharedMessages.locationDescription}
        tooltipId={tooltipIds.GATEWAY_LOCATION_PUBLIC}
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
        encode={encodeAttributes}
        decode={decodeAttributes}
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
      <Form.SubTitle title={sharedMessages.contactInformation} className="mb-cs-s" />
      <CollaboratorSelect
        name="administrative_contact"
        title={sharedMessages.adminContact}
        placeholder={sharedMessages.contactFieldPlaceholder}
        entity={'gateway'}
        entityId={gtwId}
        encode={encodeContact}
        decode={decodeContact}
        required
        isResctrictedUser={isResctrictedUser}
        userId={userId}
      />
      <Message
        content={m.adminContactDescription}
        component="p"
        className="mt-cs-xs c-text-neutral-light"
      />
      <CollaboratorSelect
        name="technical_contact"
        title={sharedMessages.technicalContact}
        placeholder={sharedMessages.contactFieldPlaceholder}
        entity={'gateway'}
        entityId={gtwId}
        encode={encodeContact}
        decode={decodeContact}
        required
        isResctrictedUser={isResctrictedUser}
        userId={userId}
      />
      <Message
        content={m.techContactDescription}
        component="p"
        className="mt-cs-xs c-text-neutral-light"
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
        <Require condition={mayDeleteGtw}>
          <Button
            onClick={handleOpenDeleteGatewayModal}
            message={
              supportsClaiming
                ? sharedMessages.unclaimAndDeleteGateway
                : sharedMessages.deleteGateway
            }
            type="button"
            icon={IconTrash}
            naked
            danger
          />
          <DeleteEntityHeaderModal
            entity={GATEWAY}
            entityId={gtwId}
            entityName={gateway.name}
            setVisible={setDeleteGtwVisible}
            visible={deleteGtwVisible}
            supportsClaiming={supportsClaiming}
          />
        </Require>
      </SubmitBar>
    </Form>
  )
})

BasicSettingsForm.propTypes = {
  gateway: PropTypes.gateway.isRequired,
  gtwId: PropTypes.string.isRequired,
  mayEditSecrets: PropTypes.bool.isRequired,
  onSubmit: PropTypes.func.isRequired,
  supportsClaiming: PropTypes.bool.isRequired,
}

export default BasicSettingsForm
