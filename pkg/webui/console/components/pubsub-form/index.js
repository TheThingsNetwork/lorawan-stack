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

import Form from '../../../components/form'
import Input from '../../../components/input'
import FileInput from '../../../components/file-input'
import Radio from '../../../components/radio-button'
import Checkbox from '../../../components/checkbox'
import Select from '../../../components/select'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import Notification from '../../../components/notification'
import Message from '../../../lib/components/message'
import ModalButton from '../../../components/button/modal-button'
import PubsubFormatSelector from '../../containers/pubsub-formats-select'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

import { mapPubsubToFormValues, mapFormValuesToPubsub, blankValues } from './mapping'
import m from './messages'
import { qosOptions } from './qos-options'
import validationSchema from './validation-schema'

const pathPlaceholder = '/sub-topic'

@bind
export default class PubsubForm extends Component {
  constructor(props) {
    super(props)

    this.form = React.createRef()

    const { initialPubsubValue, update } = this.props

    const initialIsMqtt = update && 'mqtt' in initialPubsubValue
    const initialMqttSecure = initialIsMqtt ? initialPubsubValue.mqtt.use_tls : false

    this.state = {
      error: '',
      isMqtt: initialIsMqtt,
      mqttSecure: initialMqttSecure,
    }
  }

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

  async handleSubmit(values, { setSubmitting, resetForm }) {
    const { appId, onSubmit, onSubmitSuccess, onSubmitFailure } = this.props
    const pubsub = mapFormValuesToPubsub(values, appId)

    await this.setState({ error: '' })

    try {
      const result = await onSubmit(pubsub)

      resetForm(values)
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm(values)

      await this.setState({ error })
      await onSubmitFailure(error)
    }
  }

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

  handleNatsSelect() {
    this.setState({ isMqtt: false })
  }

  handleMqttSelect() {
    this.setState({ isMqtt: true })
  }

  handleMqttUseTlsChange(event) {
    this.setState({ mqttSecure: event.target.checked })
  }

  get natsSection() {
    return (
      <React.Fragment>
        <Message component="h4" content={m.natsConfig} />
        <Form.Field
          name="nats.username"
          title={sharedMessages.username}
          placeholder={m.usernamePlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="nats.password"
          title={sharedMessages.password}
          placeholder={m.passwordPlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="nats.address"
          title={sharedMessages.address}
          placeholder={m.natsAddressPlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="nats.port"
          title={sharedMessages.port}
          placeholder={m.natsPortPlaceholder}
          component={Input}
          required
        />
        <Form.Field name="nats.secure" title={sharedMessages.secure} component={Checkbox} />
      </React.Fragment>
    )
  }

  get mqttSection() {
    const { mqttSecure } = this.state

    return (
      <React.Fragment>
        <Message component="h4" content={m.mqttConfig} />
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
          name="mqtt.username"
          title={sharedMessages.username}
          placeholder={m.usernamePlaceholder}
          component={Input}
          required
        />
        <Form.Field
          name="mqtt.password"
          title={sharedMessages.password}
          placeholder={m.passwordPlaceholder}
          component={Input}
          required
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
      </React.Fragment>
    )
  }

  render() {
    const { update, initialPubsubValue } = this.props
    const { error, isMqtt } = this.state
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
        <Message component="h4" content={sharedMessages.generalInformation} />
        <Form.Field
          name="pub_sub_id"
          title={sharedMessages.pubsubId}
          placeholder={m.idPlaceholder}
          component={Input}
          required
          autoFocus
          disabled={update}
        />
        <PubsubFormatSelector horizontal name="format" required />
        <Form.Field
          name="base_topic"
          title={sharedMessages.pubsubBaseTopic}
          placeholder="base-topic"
          component={Input}
          required
        />
        <Form.Field title={sharedMessages.provider} name="_provider" component={Radio.Group}>
          <Radio label="NATS" value="nats" onChange={this.handleNatsSelect} />
          <Radio label="MQTT" value="mqtt" onChange={this.handleMqttSelect} />
        </Form.Field>
        {isMqtt ? this.mqttSection : this.natsSection}
        <Message component="h4" content={sharedMessages.messageTypes} />
        <Notification info={m.messageInfo} small />
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
                  values: { pubsubId: initialPubsubValue.ids.pubsub_id },
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
