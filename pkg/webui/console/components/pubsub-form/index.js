// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import FileInput from '@ttn-lw/components/file-input'
import Radio from '@ttn-lw/components/radio-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Select from '@ttn-lw/components/select'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PubsubFormatSelector from '@console/containers/pubsub-formats-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import m from './messages'
import {
  mapPubsubToFormValues,
  mapFormValuesToPubsub,
  blankValues,
  mapNatsFormValues,
} from './mapping'
import { qosOptions } from './qos-options'
import providers from './providers'
import validationSchema from './validation-schema'

const pathPlaceholder = 'sub-topic'

const PubsubForm = props => {
  const form = useRef(null)
  const modalResolve = useRef(() => {})
  const modalReject = useRef(() => {})

  const {
    initialPubsubValue,
    update,
    mqttDisabled,
    natsDisabled,
    appId,
    onSubmit,
    existCheck,
    onDelete,
  } = props
  const [error, setError] = useState(undefined)
  const [provider, setProvider] = useState(blankValues._provider)
  const [mqttSecure, setMqttSecure] = useState(true)
  const [mqttUseCredentials, setMqttUseCredentials] = useState(true)
  const [natsUseCredentials, setNatsUseCredentials] = useState(true)
  const [displayOverwriteModal, setDisplayOverwriteModal] = useState(false)
  const [existingId, setExistingId] = useState(undefined)

  const initialValues = useMemo(() => {
    if (update && initialPubsubValue) {
      return mapPubsubToFormValues(initialPubsubValue)
    }
    return {
      ...blankValues,
      _provider: provider,
    }
  }, [initialPubsubValue, provider, update])

  useEffect(() => {
    if (natsDisabled && mqttDisabled) {
      setProvider(blankValues._provider)
    } else {
      setProvider(natsDisabled ? providers.MQTT : providers.NATS)
    }

    if (update && 'nats' in initialPubsubValue) {
      const { password, username } = mapNatsFormValues(initialPubsubValue.nats)
      setProvider(providers.NATS)
      setNatsUseCredentials(Boolean(password || username))
    } else if (update && 'mqtt' in initialPubsubValue) {
      setProvider(providers.MQTT)
      setMqttSecure(initialPubsubValue.mqtt.use_tls)
      setMqttUseCredentials(
        Boolean(initialPubsubValue.mqtt.username || initialPubsubValue.mqtt.password),
      )
    }
  }, [initialPubsubValue, mqttDisabled, natsDisabled, update])

  const handleSubmit = useCallback(
    async (values, { resetForm }) => {
      const castedValues = validationSchema.cast(values)
      const pubsub = mapFormValuesToPubsub(castedValues, appId)
      setError('')

      try {
        if (!update) {
          const pubsubId = pubsub.ids.pub_sub_id
          const exists = await existCheck(pubsubId)
          if (exists) {
            setDisplayOverwriteModal(true)
            setExistingId(pubsubId)
            await new Promise((resolve, reject) => {
              modalResolve.current = resolve
              modalReject.current = reject
            })
          }
        }
        await onSubmit(pubsub)

        resetForm({ values })
      } catch (error) {
        resetForm({ values })
        setError(error)
      }
    },
    [appId, existCheck, onSubmit, update],
  )

  const handleDelete = useCallback(async () => {
    try {
      await onDelete()
      form.current.resetForm()
    } catch (error) {
      setError(error)
    }
  }, [onDelete])

  const handleProviderSelect = useCallback(event => {
    setProvider(event.target.value)
  }, [])

  const handleUseCredentialsChangeNats = useCallback(event => {
    setNatsUseCredentials(event.target.checked)
  }, [])

  const handleMqttUseTlsChange = useCallback(event => {
    setMqttSecure(event.target.checked)
  }, [])

  const handleUseCredentialsChangeMqtt = useCallback(event => {
    setMqttUseCredentials(event.target.checked)
  }, [])

  const handleReplaceModalDecision = useCallback(mayReplace => {
    if (mayReplace) {
      modalResolve.current()
    } else {
      modalReject.current()
    }
    setDisplayOverwriteModal(false)
  }, [])

  const natsSection = useMemo(
    () => (
      <>
        <Form.SubTitle title={m.natsConfig} />
        <Form.Field name="nats.secure" label={m.useSecureConnection} component={Checkbox} />
        <Form.Field
          name="nats._use_credentials"
          label={m.useCredentials}
          component={Checkbox}
          onChange={handleUseCredentialsChangeNats}
        />
        <Form.Field
          name="nats.username"
          title={sharedMessages.username}
          placeholder={m.usernamePlaceholder}
          component={Input}
          required={natsUseCredentials}
          disabled={!natsUseCredentials}
        />
        <Form.Field
          name="nats.password"
          title={sharedMessages.password}
          placeholder={m.passwordPlaceholder}
          component={Input}
          required={natsUseCredentials}
          disabled={!natsUseCredentials}
          sensitive
        />
        <Form.Field
          name="nats.address"
          title={sharedMessages.address}
          placeholder={m.natsAddressPlaceholder}
          component={Input}
          autoComplete="on"
          required
        />
        <Form.Field
          name="nats.port"
          title={sharedMessages.port}
          placeholder={m.natsPortPlaceholder}
          component={Input}
          autoComplete="on"
          required
        />
      </>
    ),
    [handleUseCredentialsChangeNats, natsUseCredentials],
  )

  const mqttSection = useMemo(
    () => (
      <>
        <Form.SubTitle title={m.mqttConfig} />
        <Form.Field
          name="mqtt.use_tls"
          label={m.useSecureConnection}
          component={Checkbox}
          onChange={handleMqttUseTlsChange}
        />
        {mqttSecure && (
          <>
            <Form.Field
              name="mqtt.tls_ca"
              title={m.tlsCa}
              component={FileInput}
              message={m.selectPemFile}
              providedMessage={m.pemFileProvided}
              accept=".pem"
              required
            />
            <Form.Field
              name="mqtt.tls_client_cert"
              title={m.tlsClientCert}
              component={FileInput}
              message={m.selectPemFile}
              providedMessage={m.pemFileProvided}
              accept=".pem"
              required
            />
            <Form.Field
              name="mqtt.tls_client_key"
              title={m.tlsClientKey}
              component={FileInput}
              message={m.selectPemFile}
              providedMessage={m.pemFileProvided}
              accept=".pem"
              required
            />
          </>
        )}
        <Form.Field
          name="mqtt.server_url"
          title={sharedMessages.serverUrl}
          placeholder={m.mqttServerUrlPlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="mqtt.client_id"
          title={sharedMessages.clientId}
          placeholder={m.mqttClientIdPlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="mqtt._use_credentials"
          label={m.useCredentials}
          component={Checkbox}
          onChange={handleUseCredentialsChangeMqtt}
        />
        <Form.Field
          name="mqtt.username"
          title={sharedMessages.username}
          placeholder={m.usernamePlaceholder}
          component={Input}
          required={mqttUseCredentials}
          disabled={!mqttUseCredentials}
        />
        <Form.Field
          name="mqtt.password"
          title={sharedMessages.password}
          placeholder={m.passwordPlaceholder}
          component={Input}
          disabled={!mqttUseCredentials}
          sensitive
        />
        <Form.Field
          title={m.subscribeQos}
          name="mqtt.subscribe_qos"
          component={Select}
          required
          options={qosOptions}
        />
        <Form.Field
          title={m.publishQos}
          name="mqtt.publish_qos"
          component={Select}
          required
          options={qosOptions}
        />
      </>
    ),
    [handleMqttUseTlsChange, handleUseCredentialsChangeMqtt, mqttSecure, mqttUseCredentials],
  )

  const messageTypesSection = useMemo(
    () => (
      <>
        <Form.SubTitle title={sharedMessages.eventEnabledTypes} className="mb-0" />
        <Message component="p" content={m.messageInfo} className="mt-0 mb-ls-xxs" />
        <PubsubFormatSelector name="format" required />
        <Form.Field
          name="base_topic"
          title={sharedMessages.pubsubBaseTopic}
          placeholder="base-topic"
          component={Input}
        />
        <Form.Field
          name="uplink_message"
          type="toggled-input"
          enabledMessage={sharedMessages.uplinkMessage}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventUplinkMessageDesc}
        />
        <Form.Field
          name="uplink_normalized"
          type="toggled-input"
          enabledMessage={sharedMessages.uplinkNormalized}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventUplinkNormalizedDesc}
        />
        <Form.Field
          name="join_accept"
          type="toggled-input"
          enabledMessage={sharedMessages.joinAccept}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventJoinAcceptDesc}
        />
        <Form.Field
          name="downlink_ack"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkAck}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkAckDesc}
        />
        <Form.Field
          name="downlink_nack"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkNack}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkNackDesc}
        />
        <Form.Field
          name="downlink_sent"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkSent}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkSentDesc}
        />
        <Form.Field
          name="downlink_failed"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkFailed}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkFailedDesc}
        />
        <Form.Field
          name="downlink_queued"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkQueued}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkQueuedDesc}
        />
        <Form.Field
          name="downlink_queue_invalidated"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkQueueInvalidated}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          tooltipId={tooltipIds.DOWNLINK_QUEUE_INVALIDATED}
          description={sharedMessages.eventDownlinkQueueInvalidatedDesc}
        />
        <Form.Field
          name="location_solved"
          type="toggled-input"
          enabledMessage={sharedMessages.locationSolved}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventLocationSolvedDesc}
        />
        <Form.Field
          name="service_data"
          type="toggled-input"
          enabledMessage={sharedMessages.serviceData}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventServiceDataDesc}
        />
        <Form.Field
          name="downlink_push"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkPush}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkPushDesc}
        />
        <Form.Field
          name="downlink_replace"
          type="toggled-input"
          enabledMessage={sharedMessages.downlinkReplace}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
          description={sharedMessages.eventDownlinkReplaceDesc}
        />
      </>
    ),
    [],
  )

  return (
    <>
      <Message
        content={m.pubsubsDescription}
        values={{
          Link: val => (
            <Link.DocLink path="/integrations/pubsub" secondary>
              {val}
            </Link.DocLink>
          ),
        }}
        component="p"
      />
      <hr className="mb-ls-m" />
      <Form
        onSubmit={handleSubmit}
        validationSchema={validationSchema}
        initialValues={initialValues}
        error={error}
        formikRef={form}
      >
        <PortalledModal
          title={sharedMessages.idAlreadyExists}
          message={{ ...m.alreadyExistsModalMessage, values: { id: existingId } }}
          buttonMessage={m.replacePubsub}
          onComplete={handleReplaceModalDecision}
          approval
          visible={displayOverwriteModal}
        />
        <Form.SubTitle title={sharedMessages.generalSettings} />
        <Form.Field
          name="pub_sub_id"
          title={sharedMessages.pubsubId}
          placeholder={m.idPlaceholder}
          component={Input}
          required
          autoFocus
          disabled={update}
        />
        <Form.Field
          horizontal
          title={sharedMessages.provider}
          name="_provider"
          component={Radio.Group}
          description={natsDisabled || mqttDisabled ? m.providerDescription : undefined}
          disabled={natsDisabled || mqttDisabled}
        >
          <Radio label="NATS" value={providers.NATS} onChange={handleProviderSelect} />
          <Radio label="MQTT" value={providers.MQTT} onChange={handleProviderSelect} />
        </Form.Field>
        {provider === providers.NATS && natsSection}
        {provider === providers.MQTT && mqttSection}
        {messageTypesSection}
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={update ? sharedMessages.saveChanges : sharedMessages.addPubsub}
          />
          {update && (
            <ModalButton
              type="button"
              icon="delete"
              danger
              naked
              message={m.deletePubsub}
              modalData={{
                message: {
                  values: { pubsubId: initialPubsubValue.ids.pub_sub_id },
                  ...m.modalWarning,
                },
              }}
              onApprove={handleDelete}
            />
          )}
        </SubmitBar>
      </Form>
    </>
  )
}
PubsubForm.propTypes = {
  appId: PropTypes.string.isRequired,
  existCheck: PropTypes.func,
  initialPubsubValue: PropTypes.pubsub,
  mqttDisabled: PropTypes.bool.isRequired,
  natsDisabled: PropTypes.bool.isRequired,
  onDelete: PropTypes.func,
  onSubmit: PropTypes.func.isRequired,
  update: PropTypes.bool.isRequired,
}

PubsubForm.defaultProps = {
  initialPubsubValue: undefined,
  existCheck: () => false,
  onDelete: () => null,
}

export default PubsubForm
