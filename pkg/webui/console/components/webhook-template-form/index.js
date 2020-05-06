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
import * as Yup from 'yup'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'
import urlTemplate from 'url-template'

import api from '@console/api'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import Message from '@ttn-lw/lib/components/message'

import WebhookTemplateInfo from '@console/components/webhook-template-info'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { id as webhookIdRegexp } from '@console/lib/regexp'

const m = defineMessages({
  createTemplate: 'Create {template} webhook',
  idPlaceholder: 'my-new-{templateId}-webhook',
  templateSettings: 'Template settings',
})

const pathExpand = (url, fields) =>
  Boolean(url) && url.path ? { path: urlTemplate.parse(url.path).expand(fields) } : url

export default class WebhookTemplateForm extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    onSubmit: PropTypes.func.isRequired,
    onSubmitSuccess: PropTypes.func.isRequired,
    templateId: PropTypes.string.isRequired,
    webhookTemplate: PropTypes.webhookTemplate.isRequired,
  }

  state = {
    error: '',
  }

  @bind
  async convertTemplateToWebhook(values) {
    const { webhookTemplate: template, appId } = this.props
    const { webhook_id, ...fields } = values

    const headers = Object.keys(template.headers).reduce((acc, key) => {
      const val = template.headers[key]
      acc[key] = val.replace(/{([a-z0-9_-]+)}/i, (_, field) => values[field])
      return acc
    }, {})

    const webhook = {
      ids: {
        webhook_id: values.webhook_id,
      },
      template_ids: template.ids,
      format: template.format,
      headers,
      template_fields: fields,
      base_url: urlTemplate.parse(template.base_url).expand(fields),
      uplink_message: pathExpand(template.uplink_message, fields),
      join_accept: pathExpand(template.join_accept, fields),
      downlink_ack: pathExpand(template.downlink_ack, fields),
      downlink_nack: pathExpand(template.downlink_nack, fields),
      downlink_sent: pathExpand(template.downlink_sent, fields),
      downlink_failed: pathExpand(template.downlink_failed, fields),
      downlink_queued: pathExpand(template.downlink_queued, fields),
      location_solved: pathExpand(template.location_solved, fields),
    }

    if (template.create_downlink_api_key) {
      const key = {
        name: `${webhook_id} downlink API key`,
        rights: ['RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE'],
      }
      const { key: downlink_api_key } = await api.application.apiKeys.create(appId, key)
      webhook.downlink_api_key = downlink_api_key
    }

    return webhook
  }

  @bind
  async handleSubmit(values, { setSubmitting, resetForm }) {
    const { onSubmit, onSubmitSuccess } = this.props

    await this.setState({ error: '' })
    try {
      const webhook = await this.convertTemplateToWebhook(values)
      const result = await onSubmit(webhook)
      resetForm(values)
      onSubmitSuccess(result)
    } catch (error) {
      setSubmitting(false)
      await this.setState({ error })
    }
  }

  render() {
    const { templateId, webhookTemplate } = this.props
    const { name, fields } = webhookTemplate
    const { error } = this.state
    const validationSchema = Yup.object({
      ...fields.reduce(
        (acc, field) => ({
          ...acc,
          [field.id]: Yup.string().required(sharedMessages.validateRequired),
        }),
        {
          webhook_id: Yup.string()
            .matches(webhookIdRegexp, sharedMessages.validateIdFormat)
            .min(2, Yup.passValues(sharedMessages.validateTooShort))
            .max(25, Yup.passValues(sharedMessages.validateTooLong))
            .required(sharedMessages.validateRequired),
        },
      ),
    })

    const initialValues = fields.reduce((acc, field) => ({ ...acc, [field.id]: '' }), {
      webhook_id: '',
    })
    return (
      <div>
        <WebhookTemplateInfo webhookTemplate={webhookTemplate} />
        <Form
          onSubmit={this.handleSubmit}
          validationSchema={validationSchema}
          initialValues={initialValues}
          error={error}
          formikRef={this.form}
        >
          <Message component="h4" content={m.templateSettings} />
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
              required
            />
          ))}
          <SubmitBar>
            <Form.Submit
              component={SubmitButton}
              message={{
                ...m.createTemplate,
                values: { template: name.toLowerCase() },
              }}
            />
          </SubmitBar>
        </Form>
      </div>
    )
  }
}
