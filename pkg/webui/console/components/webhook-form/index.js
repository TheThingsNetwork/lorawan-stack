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

import React, { Component } from 'react'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Button from '@ttn-lw/components/button'
import Notification from '@ttn-lw/components/notification'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import ModalButton from '@ttn-lw/components/button/modal-button'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import Checkbox from '@ttn-lw/components/checkbox'

import WebhookTemplateInfo from '@console/components/webhook-template-info'

import WebhookFormatSelector from '@console/containers/webhook-formats-select'

import Yup from '@ttn-lw/lib/yup'
import { url as urlRegexp, id as webhookIdRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { apiKey as webhookAPIKeyRegexp } from '@console/lib/regexp'

import {
  blankValues,
  encodeBasicAuthRequest,
  decodeBasicAuthRequest,
  decodeBasicAuthHeaderUsername,
  decodeBasicAuthHeaderPassword,
  encodeBasicAuthUsername,
  encodeBasicAuthPassword,
  encodeHeaders,
  decodeHeaders,
  decodeMessageType,
  encodeMessageType,
  hasBasicAuth,
  isBasicAuth,
} from './mapping'

const pathPlaceholder = '/path/to/webhook'

const m = defineMessages({
  idPlaceholder: 'my-new-webhook',
  messageInfo:
    'For each enabled message type, an optional path can be defined which will be appended to the base URL',
  deleteWebhook: 'Delete Webhook',
  modalWarning:
    'Are you sure you want to delete webhook "{webhookId}"? Deleting a webhook cannot be undone.',
  additionalHeaders: 'Additional headers',
  headersKeyPlaceholder: 'Authorization',
  headersValuePlaceholder: 'Bearer my-auth-token',
  headersAdd: 'Add header entry',
  downlinkAPIKey: 'Downlink API key',
  downlinkAPIKeyDesc:
    'The API key will be provided to the endpoint using the "X-Downlink-Apikey" header',
  templateInformation: 'Template information',
  enabledMessages: 'Enabled messages',
  endpointSettings: 'Endpoint settings',
  updateErrorTitle: 'Could not update webhook',
  createErrorTitle: 'Could not create webhook',
  reactivateButtonMessage: 'Reactivate',
  suspendedWebhookMessage:
    'This webhook has been deactivated due to several unsuccessful forwarding attempts. It will be automatically reactivated after 24 hours. If you wish to reactivate right away, you can use the button below.',
  pendingInfo:
    'This webhook is currently pending until attempting its first regular request attempt. Note that webhooks can be deactivated if they encounter too many request failures.',
  messagePathValidateTooLong: 'Enabled message path must be at most 64 characters',
  basicAuthCheckbox: 'Use basic access authentication (basic auth)',
  requestBasicAuth: 'Request authentication',
})

const messageCheck = message => {
  if (message && message.enabled) {
    const { value } = message

    return value ? value.length <= 64 : true
  }

  return true
}

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    webhook_id: Yup.string()
      .min(3, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .matches(webhookIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
      .required(sharedMessages.validateRequired),
  }),
  format: Yup.string().required(sharedMessages.validateRequired),
  _headers: Yup.array()
    .of(
      Yup.object({
        value: Yup.string(),
        key: Yup.string(),
      }),
    )
    .default([]),
  base_url: Yup.string()
    .matches(urlRegexp, Yup.passValues(sharedMessages.validateUrl))
    .required(sharedMessages.validateRequired),
  downlink_api_key: Yup.string().matches(
    webhookAPIKeyRegexp,
    Yup.passValues(sharedMessages.validateApiKey),
  ),
  uplink_message: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      path: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  join_accept: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  downlink_ack: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  downlink_nack: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  downlink_sent: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  downlink_failed: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  downlink_queued: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  downlink_queue_invalidated: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  location_solved: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
  service_data: Yup.object()
    .shape({
      enabled: Yup.boolean(),
      value: Yup.string(),
    })
    .test('has path length at most 64 characters', m.messagePathValidateTooLong, messageCheck),
})

export default class WebhookForm extends Component {
  static propTypes = {
    existCheck: PropTypes.func,
    initialWebhookValue: PropTypes.shape({
      ids: PropTypes.shape({
        webhook_id: PropTypes.string,
      }),
      health_status: PropTypes.shape({
        unhealthy: PropTypes.shape({}),
      }),
      headers: PropTypes.shape({
        Authorization: PropTypes.string,
      }),
    }),
    onDelete: PropTypes.func,
    onDeleteFailure: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onReactivate: PropTypes.func,
    onReactivateSuccess: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func.isRequired,
    update: PropTypes.bool.isRequired,
    webhookTemplate: PropTypes.webhookTemplate,
  }

  static defaultProps = {
    initialWebhookValue: undefined,
    onReactivate: () => null,
    onReactivateSuccess: () => null,
    onSubmitFailure: () => null,
    onDeleteFailure: () => null,
    onDeleteSuccess: () => null,
    onDelete: () => null,
    webhookTemplate: undefined,
    existCheck: () => false,
  }

  form = React.createRef()
  modalResolve = () => null
  modalReject = () => null

  constructor(props) {
    super(props)
    const { initialWebhookValue } = this.props

    this.state = {
      error: undefined,
      displayOverwriteModal: false,
      existingId: undefined,
      shouldShowCredentialsInput: Boolean(
        initialWebhookValue?.headers?.Authorization?.startsWith('Basic '),
      ),
    }
  }

  @bind
  async handleSubmit(values, { resetForm }) {
    const { onSubmit, onSubmitSuccess, onSubmitFailure, existCheck, update } = this.props
    const castedWebhookValues = validationSchema.cast(values)
    castedWebhookValues.headers = encodeHeaders(castedWebhookValues._headers)
    const { _headers, ...submitValues } = castedWebhookValues

    await this.setState({ error: '' })

    try {
      if (!update) {
        const webhookId = castedWebhookValues.ids.webhook_id
        const exists = await existCheck(webhookId)
        if (exists) {
          this.setState({ displayOverwriteModal: true, existingId: webhookId })
          await new Promise((resolve, reject) => {
            this.modalResolve = resolve
            this.modalReject = reject
          })
        }
      }
      const result = await onSubmit(submitValues)

      resetForm({ values: castedWebhookValues })
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm({ values })
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
  handleReplaceModalDecision(mayReplace) {
    if (mayReplace) {
      this.modalResolve()
    } else {
      this.modalReject()
    }
    this.setState({ displayOverwriteModal: false })
  }

  @bind
  async handleReactivate() {
    const { onReactivate, onReactivateSuccess } = this.props
    const healthStatus = {
      health_status: null,
    }

    try {
      await onReactivate(healthStatus)
      onReactivateSuccess()
    } catch (error) {
      this.setState({ error })
    }
  }

  @bind
  handleRequestAuthenticationChange(event) {
    this.setState({ shouldShowCredentialsInput: event.target.checked })
  }

  @bind
  handleHeadersChange(headers) {
    this.setState({ shouldShowCredentialsInput: hasBasicAuth(headers) })
  }

  render() {
    const { update, initialWebhookValue, webhookTemplate } = this.props
    const { error, displayOverwriteModal, existingId } = this.state
    const initialValues = initialWebhookValue || blankValues
    initialValues._headers = decodeHeaders(initialValues.headers)

    const mayReactivate =
      update &&
      initialWebhookValue &&
      initialWebhookValue.health_status &&
      initialWebhookValue.health_status.unhealthy

    const isPending =
      update &&
      initialWebhookValue &&
      (initialWebhookValue.health_status === null ||
        initialWebhookValue.health_status === undefined)

    return (
      <>
        {mayReactivate && (
          <Notification
            warning
            content={m.suspendedWebhookMessage}
            children={
              <Button
                onClick={this.handleReactivate}
                icon="refresh"
                message={m.reactivateButtonMessage}
                className="mt-cs-m"
              />
            }
            small
          />
        )}
        {isPending && <Notification info content={m.pendingInfo} small />}
        <PortalledModal
          title={sharedMessages.idAlreadyExists}
          message={{
            ...sharedMessages.webhookAlreadyExistsModalMessage,
            values: { id: existingId },
          }}
          buttonMessage={sharedMessages.replaceWebhook}
          onComplete={this.handleReplaceModalDecision}
          approval
          visible={displayOverwriteModal}
        />
        {Boolean(webhookTemplate) && <WebhookTemplateInfo webhookTemplate={webhookTemplate} />}
        <Form
          onSubmit={this.handleSubmit}
          validationSchema={validationSchema}
          initialValues={initialValues}
          error={error}
          errorTitle={update ? m.updateErrorTitle : m.createErrorTitle}
          formikRef={this.form}
        >
          <Form.SubTitle title={sharedMessages.generalSettings} />
          <Form.Field
            name="ids.webhook_id"
            title={sharedMessages.webhookId}
            placeholder={m.idPlaceholder}
            component={Input}
            required
            autoFocus
            disabled={update}
          />
          <WebhookFormatSelector name="format" required />
          <Form.SubTitle title={m.endpointSettings} />
          <Form.Field
            name="base_url"
            title={sharedMessages.webhookBaseUrl}
            placeholder="https://example.com/webhooks"
            component={Input}
            required
          />
          <Form.Field
            name="downlink_api_key"
            title={m.downlinkAPIKey}
            component={Input}
            description={m.downlinkAPIKeyDesc}
            sensitive
            code
          />
          <Form.Field
            title={m.requestBasicAuth}
            name="_headers"
            label={m.basicAuthCheckbox}
            onChange={this.handleRequestAuthenticationChange}
            decode={decodeBasicAuthRequest}
            encode={encodeBasicAuthRequest}
            component={Checkbox}
            tooltipId={tooltipIds.BASIC_AUTH}
          />
          {this.state.shouldShowCredentialsInput && (
            <Form.FieldContainer horizontal>
              <Form.Field
                required
                title={sharedMessages.username}
                name="_headers"
                decode={decodeBasicAuthHeaderUsername}
                encode={encodeBasicAuthUsername}
                component={Input}
              />
              <Form.Field
                required
                title={sharedMessages.password}
                name="_headers"
                decode={decodeBasicAuthHeaderPassword}
                encode={encodeBasicAuthPassword}
                component={Input}
                sensitive
              />
            </Form.FieldContainer>
          )}
          <Form.Field
            name="_headers"
            title={m.additionalHeaders}
            keyPlaceholder={m.headersKeyPlaceholder}
            valuePlaceholder={m.headersValuePlaceholder}
            addMessage={m.headersAdd}
            component={KeyValueMap}
            isReadOnly={isBasicAuth}
            onChange={this.handleHeadersChange}
          />
          <Form.SubTitle title={m.enabledMessages} />
          <Notification info content={m.messageInfo} small />
          <Form.Field
            name="uplink_message"
            type="toggled-input"
            title={sharedMessages.uplinkMessage}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="join_accept"
            type="toggled-input"
            title={sharedMessages.joinAccept}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="downlink_ack"
            type="toggled-input"
            title={sharedMessages.downlinkAck}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="downlink_nack"
            type="toggled-input"
            title={sharedMessages.downlinkNack}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="downlink_sent"
            type="toggled-input"
            title={sharedMessages.downlinkSent}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="downlink_failed"
            type="toggled-input"
            title={sharedMessages.downlinkFailed}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="downlink_queued"
            type="toggled-input"
            title={sharedMessages.downlinkQueued}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="downlink_queue_invalidated"
            type="toggled-input"
            title={sharedMessages.downlinkQueueInvalidated}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="location_solved"
            type="toggled-input"
            title={sharedMessages.locationSolved}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <Form.Field
            name="service_data"
            type="toggled-input"
            title={sharedMessages.serviceData}
            placeholder={pathPlaceholder}
            decode={decodeMessageType}
            encode={encodeMessageType}
            component={Input.Toggled}
          />
          <SubmitBar>
            <Form.Submit
              component={SubmitButton}
              message={update ? sharedMessages.saveChanges : sharedMessages.addWebhook}
            />
            {update && (
              <ModalButton
                type="button"
                icon="delete"
                danger
                naked
                message={m.deleteWebhook}
                modalData={{
                  message: {
                    values: { webhookId: initialWebhookValue.ids.webhook_id },
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
