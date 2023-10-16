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

import React, { Component } from 'react'
import bind from 'autobind-decorator'

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

export default class PubsubForm extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    existCheck: PropTypes.func,
    initialPubsubValue: PropTypes.pubsub,
    mqttDisabled: PropTypes.bool.isRequired,
    natsDisabled: PropTypes.bool.isRequired,
    onDelete: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    update: PropTypes.bool.isRequired,
  }

  static defaultProps = {
    initialPubsubValue: undefined,
    existCheck: () => false,
    onDelete: () => null,
  }

  constructor(props) {
    super(props)

    this.form = React.createRef()
    this.modalResolve = () => null
    this.modalReject = () => null

    const { initialPubsubValue, update, mqttDisabled, natsDisabled } = this.props

    this.state = {
      error: undefined,
      mqttDisabled,
      provider: blankValues._provider,
      mqttUseCredentials: true,
      natsUseCredentials: true,
      natsDisabled,
      displayOverwriteModal: false,
      existingId: undefined,
    }

    if (natsDisabled && mqttDisabled) {
      this.state.provider = blankValues._provider
    } else {
      this.state.provider = natsDisabled ? providers.MQTT : providers.NATS
    }

    if (update && 'nats' in initialPubsubValue) {
      const { password, username } = mapNatsFormValues(initialPubsubValue.nats)
      this.state.provider = providers.NATS
      this.state.natsUseCredentials = Boolean(password || username)
    } else if (update && 'mqtt' in initialPubsubValue) {
      this.state.provider = providers.MQTT
      this.state.mqttSecure = initialPubsubValue.mqtt.use_tls
      this.state.mqttUseCredentials = Boolean(
        initialPubsubValue.mqtt.username || initialPubsubValue.mqtt.password,
      )
    }
  }

  @bind
  async handleSubmit(values, { resetForm }) {
    const { appId, onSubmit, existCheck, update } = this.props

    const castedValues = validationSchema.cast(values)
    const pubsub = mapFormValuesToPubsub(castedValues, appId)

    this.setState({ error: '' })

    try {
      if (!update) {
        const pubsubId = pubsub.ids.pub_sub_id
        const exists = await existCheck(pubsubId)
        if (exists) {
          this.setState({ displayOverwriteModal: true, existingId: pubsubId })
          await new Promise((resolve, reject) => {
            this.modalResolve = resolve
            this.modalReject = reject
          })
        }
      }
      await onSubmit(pubsub)

      resetForm({ values })
    } catch (error) {
      resetForm({ values })

      this.setState({ error })
    }
  }

  @bind
  async handleDelete() {
    const { onDelete } = this.props
    try {
      await onDelete()
      this.form.current.resetForm()
    } catch (error) {
      this.setState({ error })
    }
  }

  @bind
  handleProviderSelect(event) {
    this.setState({ provider: event.target.value })
  }

  @bind
  handleUseCredentialsChangeNats(event) {
    this.setState({ natsUseCredentials: event.target.checked })
  }

  @bind
  handleMqttUseTlsChange(event) {
    this.setState({ mqttSecure: event.target.checked })
  }

  @bind
  handleUseCredentialsChangeMqtt(event) {
    this.setState({ mqttUseCredentials: event.target.checked })
  }

  @bind
  handleReplaceModalDecision(mayReplace) {
    if (mayReplace) {
      this.modalResolve()
    } else {
      this.modalReject()
    }
    this.setState({ displayOverwriteModal: false })
  }

  get natsSection() {
    const { natsUseCredentials } = this.state
    return (
      <>
        <Form.SubTitle title={m.natsConfig} />
        <Form.Field name="nats.secure" label={m.useSecureConnection} component={Checkbox} />
        <Form.Field
          name="nats._use_credentials"
          label={m.useCredentials}
          component={Checkbox}
          onChange={this.handleUseCredentialsChangeNats}
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
    )
  }

  get mqttSection() {
    const { mqttSecure, mqttUseCredentials } = this.state

    return (
      <>
        <Form.SubTitle title={m.mqttConfig} />
        <Form.Field
          name="mqtt.use_tls"
          label={m.useSecureConnection}
          component={Checkbox}
          onChange={this.handleMqttUseTlsChange}
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
          onChange={this.handleUseCredentialsChangeMqtt}
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
    )
  }

  get messageTypesSection() {
    return (
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
    )
  }

  render() {
    const { update, initialPubsubValue, mqttDisabled, natsDisabled } = this.props
    const { error, provider, displayOverwriteModal, existingId } = this.state
    let initialValues = blankValues
    if (update && initialPubsubValue) {
      initialValues = mapPubsubToFormValues(initialPubsubValue)
    }

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
          onSubmit={this.handleSubmit}
          validationSchema={validationSchema}
          initialValues={initialValues}
          error={error}
          formikRef={this.form}
        >
          <PortalledModal
            title={sharedMessages.idAlreadyExists}
            message={{ ...m.alreadyExistsModalMessage, values: { id: existingId } }}
            buttonMessage={m.replacePubsub}
            onComplete={this.handleReplaceModalDecision}
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
            disabled={natsDisabled || mqttDisabled}
          >
            <Radio label="NATS" value={providers.NATS} onChange={this.handleProviderSelect} />
            <Radio label="MQTT" value={providers.MQTT} onChange={this.handleProviderSelect} />
          </Form.Field>
          {provider === providers.NATS && this.natsSection}
          {provider === providers.MQTT && this.mqttSection}
          {this.messageTypesSection}
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
                onApprove={this.handleDelete}
              />
            )}
          </SubmitBar>
        </Form>
      </>
    )
  }
}
