// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import PortalledModal from '@ttn-lw/components/modal/portalled'

import WebhookTemplateInfo from '@console/components/webhook-template-info'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { id as webhookIdRegexp } from '@ttn-lw/lib/regexp'

const m = defineMessages({
  createTemplate: 'Create {template} webhook',
  idPlaceholder: 'my-new-{templateId}-webhook',
})

export default class WebhookTemplateForm extends Component {
  static propTypes = {
    convertTemplateToWebhook: PropTypes.func.isRequired,
    error: PropTypes.string,
    existCheck: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    templateId: PropTypes.string.isRequired,
    webhookTemplate: PropTypes.webhookTemplate.isRequired,
  }

  static defaultProps = {
    error: undefined,
    existCheck: () => null,
  }

  modalResolve = () => null
  modalReject = () => null

  state = {
    displayOverwriteModal: false,
    existingId: undefined,
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
  async handleSubmit(values, { setSubmitting, resetForm }) {
    const { onSubmit, convertTemplateToWebhook, existCheck } = this.props
    const webhook = await convertTemplateToWebhook(values)
    const webhookId = webhook.ids.webhook_id
    const exists = await existCheck(webhookId)
    if (exists) {
      this.setState({ displayOverwriteModal: true, existingId: webhookId })
      await new Promise((resolve, reject) => {
        this.modalResolve = resolve
        this.modalReject = reject
      })
    }
    await onSubmit(values, webhook, { setSubmitting, resetForm })
  }

  render() {
    const { templateId, webhookTemplate, error } = this.props
    const { name, fields = [] } = webhookTemplate

    const validationSchema = Yup.object({
      ...fields.reduce(
        (acc, field) => ({
          ...acc,
          [field.id]: field.optional
            ? Yup.string()
            : Yup.string().required(sharedMessages.validateRequired),
        }),
        {
          webhook_id: Yup.string()
            .min(3, Yup.passValues(sharedMessages.validateTooShort))
            .max(36, Yup.passValues(sharedMessages.validateTooLong))
            .matches(webhookIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
            .required(sharedMessages.validateRequired),
        },
      ),
    })

    const initialValues = fields.reduce(
      (acc, field) => ({ ...acc, [field.id]: field.default_value || '' }),
      {
        webhook_id: '',
      },
    )
    return (
      <div>
        <PortalledModal
          title={sharedMessages.idAlreadyExists}
          message={{
            ...sharedMessages.webhookAlreadyExistsModalMessage,
            values: { id: this.state.existingId },
          }}
          buttonMessage={sharedMessages.replaceWebhook}
          onComplete={this.handleReplaceModalDecision}
          approval
          visible={this.state.displayOverwriteModal}
        />
        <WebhookTemplateInfo webhookTemplate={webhookTemplate} />
        <Form
          onSubmit={this.handleSubmit}
          validationSchema={validationSchema}
          initialValues={initialValues}
          error={error}
        >
          <Form.Field
            name="webhook_id"
            title={sharedMessages.webhookId}
            placeholder={{ ...m.idPlaceholder, values: { templateId } }}
            component={Input}
            required
            autoFocus
          />
          {fields.map(field => (
            <Form.Field
              component={Input}
              name={field.id}
              title={field.name}
              description={field.description}
              key={field.id}
              required={!field.optional}
              sensitive={field.secret}
            />
          ))}
          <SubmitBar>
            <Form.Submit
              component={SubmitButton}
              message={{
                ...m.createTemplate,
                values: { template: name },
              }}
            />
          </SubmitBar>
        </Form>
      </div>
    )
  }
}
