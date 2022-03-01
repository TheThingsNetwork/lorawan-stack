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
  templateSettings: 'Template settings',
})

export default class WebhookTemplateForm extends Component {
  static propTypes = {
    handleReplaceModalDecision: PropTypes.func.isRequired,
    onSubmit: PropTypes.func.isRequired,
    templateId: PropTypes.string.isRequired,
    webhookTemplate: PropTypes.webhookTemplate.isRequired,
  }

  state = {
    error: undefined,
    displayOverwriteModal: false,
    existingId: undefined,
  }

  render() {
    const { templateId, webhookTemplate, onSubmit, handleReplaceModalDecision } = this.props
    const { name, fields = [] } = webhookTemplate
    const { error, displayOverwriteModal, existingId } = this.state
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
            values: { id: existingId },
          }}
          buttonMessage={sharedMessages.replaceWebhook}
          onComplete={handleReplaceModalDecision}
          approval
          visible={displayOverwriteModal}
        />
        <WebhookTemplateInfo webhookTemplate={webhookTemplate} />
        <Form
          onSubmit={onSubmit}
          validationSchema={validationSchema}
          initialValues={initialValues}
          error={error}
          formikRef={this.form}
        >
          <Form.SubTitle title={m.templateSettings} />
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
                values: { template: name.toLowerCase() },
              }}
            />
          </SubmitBar>
        </Form>
      </div>
    )
  }
}
