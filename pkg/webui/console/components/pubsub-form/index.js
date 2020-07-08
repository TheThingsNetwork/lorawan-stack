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
import bind from 'autobind-decorator'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import FileInput from '@ttn-lw/components/file-input'
import Radio from '@ttn-lw/components/radio-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Select from '@ttn-lw/components/select'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Notification from '@ttn-lw/components/notification'
import ModalButton from '@ttn-lw/components/button/modal-button'

import PubsubFormatSelector from '@console/containers/pubsub-formats-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

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
    initialPubsubValue: PropTypes.pubsub,
    onDelete: PropTypes.func,
    onDeleteFailure: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func,
    update: PropTypes.bool.isRequired,
  }

  static defaultProps = {
    initialPubsubValue: undefined,
    onSubmitSuccess: () => null,
    onSubmitFailure: () => null,
    onDeleteSuccess: () => null,
    onDeleteFailure: () => null,
    onDelete: () => null,
  }

  constructor(props) {
    super(props)

    this.form = React.createRef()

    const { initialPubsubValue, update } = this.props

    this.state = {
      error: '',
      provider: blankValues._provider,
      mqttUseCredentials: true,
      natsUseCredentials: true,
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
  async handleSubmit(values, { setSubmitting, resetForm }) {
    const { appId, onSubmit, onSubmitSuccess, onSubmitFailure } = this.props

    const castedValues = validationSchema.cast(values)
    const pubsub = mapFormValuesToPubsub(castedValues, appId)

    await this.setState({ error: '' })

    try {
      const result = await onSubmit(pubsub)

      resetForm({ values: castedValues })
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm({ values: castedValues })

      await this.setState({ error })
      await onSubmitFailure(error)
    }
  }

  @bind
  async handleDelete() {
    const { onDelete, onDeleteSuccess, onDeleteFailure } = this.props
    try {
      await onDelete()
      this.form.current.resetForm()
      onDeleteSuccess()
    } catch (error) {
      await this.setState({ error })
      onDeleteFailure()
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

  get natsSection() {
    const { natsUseCredentials } = this.state
    return (
      <React.Fragment>
        <Form.SubTitle title={m.natsConfig} />
        <Form.Field name="nats.secure" title={sharedMessages.secure} component={Checkbox} />
        <Form.Field
          name="nats._use_credentials"
          title={m.useCredentials}
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
      </React.Fragment>
    )
  }

  get mqttSection() {
    const { mqttSecure, mqttUseCredentials } = this.state

    return (
      <React.Fragment>
        <Form.SubTitle title={m.mqttConfig} />
        <Form.Field
          name="mqtt.use_tls"
          title={sharedMessages.secure}
          component={Checkbox}
          onChange={this.handleMqttUseTlsChange}
        />
        {mqttSecure && (
          <React.Fragment>
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
          </React.Fragment>
        )}
        <Form.Field
          name="mqtt.server_url"
          title={m.serverUrl}
          placeholder={m.mqttServerUrlPlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="mqtt.client_id"
          title={m.clientId}
          placeholder={m.mqttClientIdPlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="mqtt._use_credentials"
          title={m.useCredentials}
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
      </React.Fragment>
    )
  }

  get messageTypesSection() {
    return (
      <React.Fragment>
        <Form.SubTitle title={sharedMessages.messageTypes} />
        <PubsubFormatSelector horizontal name="format" required />
        <Form.Field
          name="base_topic"
          title={sharedMessages.pubsubBaseTopic}
          placeholder="base-topic"
          component={Input}
          required
        />
        <Notification content={m.messageInfo} info small />
        <Form.Field
          name="uplink_message"
          type="toggled-input"
          title={sharedMessages.uplinkMessage}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="join_accept"
          type="toggled-input"
          title={sharedMessages.joinAccept}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_ack"
          type="toggled-input"
          title={sharedMessages.downlinkAck}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_nack"
          type="toggled-input"
          title={sharedMessages.downlinkNack}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_sent"
          type="toggled-input"
          title={sharedMessages.downlinkSent}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_failed"
          type="toggled-input"
          title={sharedMessages.downlinkFailed}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_queued"
          type="toggled-input"
          title={sharedMessages.downlinkQueued}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="location_solved"
          type="toggled-input"
          title={sharedMessages.locationSolved}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_push"
          type="toggled-input"
          title={sharedMessages.downlinkPush}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
        <Form.Field
          name="downlink_replace"
          type="toggled-input"
          title={sharedMessages.downlinkReplace}
          placeholder={pathPlaceholder}
          component={Input.Toggled}
        />
      </React.Fragment>
    )
  }

  render() {
    const { update, initialPubsubValue } = this.props
    const { error, provider } = this.state
    let initialValues = blankValues
    if (update && initialPubsubValue) {
      initialValues = mapPubsubToFormValues(initialPubsubValue)
    }

    return (
      <Form
        onSubmit={this.handleSubmit}
        validationSchema={validationSchema}
        initialValues={initialValues}
        error={error}
        formikRef={this.form}
      >
        <Form.SubTitle title={sharedMessages.generalInformation} />
        <Form.Field
          name="pub_sub_id"
          title={sharedMessages.pubsubId}
          placeholder={m.idPlaceholder}
          component={Input}
          required
          autoFocus
          disabled={update}
        />
        <Form.Field title={sharedMessages.provider} name="_provider" component={Radio.Group}>
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
    )
  }
}
