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
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { Switch, Route } from 'react-router'
import bind from 'autobind-decorator'

import BlankWebhookImg from '@assets/misc/blank-webhook.svg'

import PageTitle from '@ttn-lw/components/page-title'
import Link from '@ttn-lw/components/link'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import Message from '@ttn-lw/lib/components/message'

import ApplicationWebhookAddForm from '@console/views/application-integrations-webhook-add-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { listWebhookTemplates } from '@console/store/actions/webhook-templates'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectWebhookTemplates,
  selectWebhookTemplatesFetching,
} from '@console/store/selectors/webhook-templates'

import style from './application-integrations-webhook-add-choose.styl'

const m = defineMessages({
  chooseTemplate: 'Choose webhook template',
  customTileDescription: 'Create a custom webhook without template',
})

const WebhookTile = ({ ids, name, description, logo_url }) => (
  <Col xl={3} lg={4} sm={6} xs={6} key={`tile-${ids.template_id}`} className={style.tileColumn}>
    <Link to={`template/${ids.template_id}`} className={style.webhookTile}>
      <img className={style.logo} alt={name} src={logo_url} />
      <span className={style.name}>{name}</span>
      <span className={style.description}>
        {typeof description === 'string' ? description : <Message content={description} />}
      </span>
    </Link>
  </Col>
)

WebhookTile.propTypes = {
  description: PropTypes.message.isRequired,
  ids: PropTypes.shape({
    template_id: PropTypes.isRequired,
  }).isRequired,
  logo_url: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
}

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    webhookTemplates: selectWebhookTemplates(state),
    fetching: selectWebhookTemplatesFetching(state),
  }),
  { listWebhookTemplates },
)
@withBreadcrumb('apps.single.integrations.webhooks.add.from-template', ({ appId }) => (
  <Breadcrumb
    path={`/applications/${appId}/integrations/webhooks/add/template`}
    content={sharedMessages.add}
  />
))
export default class ApplicationWebhookAddChooser extends Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
    webhookTemplates: PropTypes.webhookTemplates,
  }

  static defaultProps = {
    webhookTemplates: undefined,
  }

  @bind
  chooser() {
    const { webhookTemplates } = this.props

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <PageTitle title={m.chooseTemplate} />
          </Col>
        </Row>
        <Row gutterWidth={15} className={style.tileRow}>
          {webhookTemplates.map(WebhookTile)}
          <WebhookTile
            ids={{ template_id: 'custom' }}
            name="Custom webhook"
            description={m.customTileDescription}
            logo_url={BlankWebhookImg}
          />
        </Row>
      </Container>
    )
  }

  render() {
    const { match } = this.props

    return (
      <Switch>
        <Route exact path={match.path} component={this.chooser} />
        <Route exact path={`${match.path}/:templateId`} component={ApplicationWebhookAddForm} />
      </Switch>
    )
  }
}
