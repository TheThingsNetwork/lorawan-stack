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
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'

import PropTypes from '../../../lib/prop-types'
import PageTitle from '../../../components/page-title'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import WebhookForm from '../../components/webhook-form'
import WebhookTemplateForm from '../../components/webhook-template-form'
import sharedMessages from '../../../lib/shared-messages'
import NotFoundRoute from '../../../lib/components/not-found-route'
import { getWebhookTemplate } from '../../store/actions/webhook-templates'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectWebhookTemplateById } from '../../store/selectors/webhook-templates'
import api from '../../api'

const m = defineMessages({
  addCustomWebhook: 'Add custom webhook',
  addWebhookViaTemplate: 'Add {templateName} webhook',
  customWebhook: 'Custom webhook',
})

@connect(
  (state, props) => {
    const templateId = props.match.params.templateId
    return {
      appId: selectSelectedApplicationId(state),
      templateId,
      isCustom: !templateId || templateId === 'custom',
      isSimpleAdd: !templateId,
      webhookTemplate: selectWebhookTemplateById(state, templateId),
    }
  },
  dispatch => ({
    getWebhookTemplate: (templateId, selector) =>
      dispatch(getWebhookTemplate(templateId, selector)),
    navigateToList: appId => dispatch(push(`/applications/${appId}/integrations/webhooks`)),
  }),
)
@withBreadcrumb('apps.single.integrations.webhooks.various.add', function(props) {
  const { appId, templateId, webhookTemplate: { name } = {}, isCustom } = props
  let breadcrumbContent = m.customWebhook
  if (!templateId) {
    breadcrumbContent = sharedMessages.add
  } else if (!isCustom && name) {
    breadcrumbContent = name
  }

  return (
    <Breadcrumb
      path={`/applications/${appId}/integrations/webhooks/add/template/${templateId}`}
      content={breadcrumbContent}
    />
  )
})
export default class ApplicationWebhookAddForm extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    isCustom: PropTypes.bool.isRequired,
    navigateToList: PropTypes.func.isRequired,
    templateId: PropTypes.string,
    webhookTemplate: PropTypes.webhookTemplate,
  }

  static defaultProps = {
    templateId: undefined,
    webhookTemplate: undefined,
  }

  @bind
  async handleSubmit(webhook) {
    const { appId } = this.props

    await api.application.webhooks.create(appId, webhook)
  }

  @bind
  handleSubmitSuccess() {
    const { navigateToList, appId } = this.props

    navigateToList(appId)
  }

  render() {
    const { templateId, isCustom, webhookTemplate, appId } = this.props
    let pageTitle = m.addCustomWebhook

    if (!webhookTemplate) {
      pageTitle = sharedMessages.addWebhook
    } else if (isCustom) {
      pageTitle = {
        ...m.addWebhookViaTemplate,
        values: {
          templateName: webhookTemplate.name,
        },
      }
    }

    // Render Not Found error, when the template was not found
    if (!isCustom && templateId && !webhookTemplate) {
      return <NotFoundRoute />
    }

    return (
      <Container>
        <PageTitle title={pageTitle} />
        <Row>
          <Col lg={8} md={12}>
            {isCustom ? (
              <WebhookForm
                update={false}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
              />
            ) : (
              <WebhookTemplateForm
                appId={appId}
                templateId={templateId}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                webhookTemplate={webhookTemplate}
              />
            )}
          </Col>
        </Row>
      </Container>
    )
  }
}
