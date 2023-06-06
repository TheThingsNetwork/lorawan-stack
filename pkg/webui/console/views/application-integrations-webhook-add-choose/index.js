// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { Routes, Route, useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import BlankWebhookImg from '@assets/misc/blank-webhook.svg'

import PageTitle from '@ttn-lw/components/page-title'
import Link from '@ttn-lw/components/link'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import Message from '@ttn-lw/lib/components/message'

import ApplicationWebhookAddForm from '@console/views/application-integrations-webhook-add-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectWebhookTemplates } from '@console/store/selectors/webhook-templates'

import style from './application-integrations-webhook-add-choose.styl'

const m = defineMessages({
  chooseTemplate: 'Choose webhook template',
  customTileDescription: 'Create a custom webhook without template',
})

const WebhookTile = ({ ids, name, description, logo_url }) => (
  <Col xl={3} lg={4} sm={6} xs={6} key={`tile-${ids.template_id}`} className={style.tileColumn}>
    <Link to={ids.template_id} className={style.webhookTile}>
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
    template_id: PropTypes.string.isRequired,
  }).isRequired,
  logo_url: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
}

const WebhookChooser = () => {
  const webhookTemplates = useSelector(selectWebhookTemplates)

  return (
    <Container>
      <Row>
        <Col lg={8} md={12}>
          <PageTitle title={m.chooseTemplate} />
        </Col>
      </Row>
      <Row gutterWidth={15} className={style.tileRow}>
        <WebhookTile
          ids={{ template_id: 'custom' }}
          name="Custom webhook"
          description={m.customTileDescription}
          logo_url={BlankWebhookImg}
        />
        {webhookTemplates.map(WebhookTile)}
      </Row>
    </Container>
  )
}
const ApplicationWebhookAddChooser = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.integrations.webhooks.add.from-template',
    <Breadcrumb
      path={`/applications/${appId}/integrations/webhooks/add/template`}
      content={sharedMessages.add}
    />,
  )

  return (
    <Routes>
      <Route index Component={WebhookChooser} />
      <Route path=":templateId" Component={ApplicationWebhookAddForm} />
    </Routes>
  )
}

export default ApplicationWebhookAddChooser
