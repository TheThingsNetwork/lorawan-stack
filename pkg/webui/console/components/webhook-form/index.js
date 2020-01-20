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
import * as Yup from 'yup'
import bind from 'autobind-decorator'

import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import Notification from '../../../components/notification'
import Message from '../../../lib/components/message'
import KeyValueMap from '../../../components/key-value-map'
import ModalButton from '../../../components/button/modal-button'
import WebhookFormatSelector from '../../containers/webhook-formats-select'
import sharedMessages from '../../../lib/shared-messages'
import { id as webhookIdRegexp, apiKey as webhookAPIKeyRegexp } from '../../lib/regexp'
import PropTypes from '../../../lib/prop-types'

import { mapWebhookToFormValues, mapFormValuesToWebhook, blankValues } from './mapping'

const pathPlaceholder = '/path/to/webhook'

const m = defineMessages({
  idPlaceholder: 'my-new-webhook',
  messageInfo:
    'For each enabled message type, you can set an optional path that will be appended to the base URL.',
  deleteWebhook: 'Delete Webhook',
  modalWarning:
    'Are you sure you want to delete webhook "{webhookId}"? Deleting a webhook cannot be undone!',
  headers: 'Headers',
  headersKeyPlaceholder: 'Authorization',
  headersValuePlaceholder: 'Bearer my-auth-token',
  headersAdd: 'Add header entry',
  headersValidateRequired: 'All header entry values are required. Please remove empty entries.',
  downlinkAPIKey: 'Downlink API Key',
  downlinkAPIKeyDesc:
    'The API Key will be provided to the endpoint using the X-Downlink-Apikey header',
})

const headerCheck = headers =>
  headers === undefined ||
  (headers instanceof Array &&
    (headers.length === 0 || headers.every(header => header.key !== '' && header.value !== '')))

const validationSchema = Yup.object().shape({
  webhook_id: Yup.string()
    .matches(webhookIdRegexp, sharedMessages.validateIdFormat)
    .min(2, sharedMessages.validateTooShort)
    .max(25, sharedMessages.validateTooLong)
    .required(sharedMessages.validateRequired),
  format: Yup.string().required(sharedMessages.validateRequired),
  headers: Yup.array().test('has no empty string values', m.headersValidateRequired, headerCheck),
  base_url: Yup.string()
    .url(sharedMessages.validateUrl)
    .required(sharedMessages.validateRequired),
  downlink_api_key: Yup.string().matches(webhookAPIKeyRegexp, sharedMessages.validateFormat),
})

@bind
export default class WebhookForm extends Component {
  constructor(props) {
    super(props)

    this.form = React.createRef()
  }

  state = {
    error: '',
  }

  async handleSubmit(values, { setSubmitting, resetForm }) {
    const { appId, onSubmit, onSubmitSuccess, onSubmitFailure } = this.props
    const webhook = mapFormValuesToWebhook(values, appId)

    await this.setState({ error: '' })

    try {
      const result = await onSubmit(webhook)

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

  render() {
    const { update, initialWebhookValue } = this.props
    const { error } = this.state
    let initialValues = blankValues
    if (update && initialWebhookValue) {
      initialValues = mapWebhookToFormValues(initialWebhookValue)
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
          name="webhook_id"
          title={sharedMessages.webhookId}
          placeholder={m.idPlaceholder}
          component={Input}
          required
          autoFocus
          disabled={update}
        />
        <WebhookFormatSelector horizontal name="format" required />
        <Form.Field
          name="headers"
          title={m.headers}
          keyPlaceholder={m.headersKeyPlaceholder}
          valuePlaceholder={m.headersValuePlaceholder}
          addMessage={m.headersAdd}
          component={KeyValueMap}
        />
        <Form.Field
          name="base_url"
          title={sharedMessages.webhookBaseUrl}
          placeholder="http://example.com/webhooks"
          component={Input}
          required
        />
        <Form.Field
          name="downlink_api_key"
          title={m.downlinkAPIKey}
          component={Input}
          description={m.downlinkAPIKeyDesc}
          code
        />
        <Message component="h4" content={sharedMessages.messageTypes} />
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
    )
  }
}

WebhookForm.propTypes = {
  initialWebhookValue: PropTypes.object,
  onDelete: PropTypes.func,
  onDeleteFailure: PropTypes.func,
  onDeleteSuccess: PropTypes.func,
  onSubmit: PropTypes.func.isRequired,
  onSubmitFailure: PropTypes.func,
  onSubmitSuccess: PropTypes.func,
  update: PropTypes.bool.isRequired,
}

WebhookForm.defaultProps = {
  onSubmitSuccess: () => null,
  onSubmitFailure: () => null,
  onDeleteSuccess: () => null,
  onDeleteFailure: () => null,
  onDelete: () => null,
}
