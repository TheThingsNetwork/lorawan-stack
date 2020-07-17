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
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Notification from '@ttn-lw/components/notification'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import ModalButton from '@ttn-lw/components/button/modal-button'

import Message from '@ttn-lw/lib/components/message'

import WebhookTemplateInfo from '@console/components/webhook-template-info'

import WebhookFormatSelector from '@console/containers/webhook-formats-select'

import Yup from '@ttn-lw/lib/yup'
import { url as urlRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { id as webhookIdRegexp, apiKey as webhookAPIKeyRegexp } from '@console/lib/regexp'

import { mapWebhookToFormValues, mapFormValuesToWebhook, blankValues } from './mapping'

const pathPlaceholder = '/path/to/webhook'

const m = defineMessages({
  idPlaceholder: 'my-new-webhook',
  messageInfo:
    'For each enabled message type, an optional path can be defined which will be appended to the base URL',
  deleteWebhook: 'Delete Webhook',
  modalWarning:
    'Are you sure you want to delete webhook "{webhookId}"? Deleting a webhook cannot be undone.',
  headers: 'Headers',
  headersKeyPlaceholder: 'Authorization',
  headersValuePlaceholder: 'Bearer my-auth-token',
  headersAdd: 'Add header entry',
  headersValidateRequired: 'All header entry values are required. Please remove empty entries.',
  downlinkAPIKey: 'Downlink API key',
  downlinkAPIKeyDesc:
    'The API key will be provided to the endpoint using the "X-Downlink-Apikey" header',
  templateInformation: 'Template information',
})

const headerCheck = headers =>
  headers === undefined ||
  (headers instanceof Array &&
    (headers.length === 0 || headers.every(header => header.key !== '' && header.value !== '')))

const validationSchema = Yup.object().shape({
  webhook_id: Yup.string()
    .matches(webhookIdRegexp, sharedMessages.validateIdFormat)
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(25, Yup.passValues(sharedMessages.validateTooLong))
    .required(sharedMessages.validateRequired),
  format: Yup.string().required(sharedMessages.validateRequired),
  headers: Yup.array().test('has no empty string values', m.headersValidateRequired, headerCheck),
  base_url: Yup.string()
    .matches(urlRegexp, sharedMessages.validateUrl)
    .required(sharedMessages.validateRequired),
  downlink_api_key: Yup.string().matches(webhookAPIKeyRegexp, sharedMessages.validateApiKey),
})

export default class WebhookForm extends Component {
  static propTypes = {
    appId: PropTypes.string,
    initialWebhookValue: PropTypes.shape({
      ids: PropTypes.shape({
        webhook_id: PropTypes.string,
      }),
    }),
    onDelete: PropTypes.func,
    onDeleteFailure: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func.isRequired,
    update: PropTypes.bool.isRequired,
    webhookTemplate: PropTypes.webhookTemplate,
  }

  static defaultProps = {
    appId: undefined,
    initialWebhookValue: undefined,
    onSubmitFailure: () => null,
    onDeleteFailure: () => null,
    onDeleteSuccess: () => null,
    onDelete: () => null,
    webhookTemplate: undefined,
  }

  form = React.createRef()

  state = {
    error: '',
  }

  @bind
  async handleSubmit(values, { setSubmitting, resetForm }) {
    const { appId, onSubmit, onSubmitSuccess, onSubmitFailure } = this.props
    const webhook = mapFormValuesToWebhook(values, appId)

    await this.setState({ error: '' })

    try {
      const result = await onSubmit(webhook)

      resetForm({ values })
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

  render() {
    const { update, initialWebhookValue, webhookTemplate } = this.props
    const { error } = this.state
    let initialValues = blankValues
    if (update && initialWebhookValue) {
      initialValues = mapWebhookToFormValues(initialWebhookValue)
    }

    return (
      <>
        {Boolean(webhookTemplate) && <WebhookTemplateInfo webhookTemplate={webhookTemplate} />}
        <Form
          onSubmit={this.handleSubmit}
          validationSchema={validationSchema}
          initialValues={initialValues}
          error={error}
          formikRef={this.form}
        >
          <Message component="h4" content={sharedMessages.generalSettings} />
          <Form.Field
            name="webhook_id"
            title={sharedMessages.webhookId}
            placeholder={m.idPlaceholder}
            component={Input}
            required
            autoFocus
            disabled={update}
          />
          <Form.Field
            name="headers"
            title={m.headers}
            keyPlaceholder={m.headersKeyPlaceholder}
            valuePlaceholder={m.headersValuePlaceholder}
            addMessage={m.headersAdd}
            component={KeyValueMap}
          />
          <Form.Field
            name="downlink_api_key"
            title={m.downlinkAPIKey}
            component={Input}
            description={m.downlinkAPIKeyDesc}
            code
          />
          <Message component="h4" content={sharedMessages.messageTypes} />
          <WebhookFormatSelector horizontal name="format" required />
          <Form.Field
            name="base_url"
            title={sharedMessages.webhookBaseUrl}
            placeholder="https://example.com/webhooks"
            component={Input}
            autoComplete="url"
            required
          />
          <Notification info content={m.messageInfo} small />
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
