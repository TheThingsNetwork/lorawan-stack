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

import React from 'react'
import { Routes, Route } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'

import ApplicationWebhookChoose from '@console/views/application-integrations-webhook-add-choose'
import ApplicationWebhookEdit from '@console/views/application-integrations-webhook-edit'
import ApplicationWebhookAdd from '@console/views/application-integrations-webhook-add'
import ApplicationWebhooksList from '@console/views/application-integrations-webhooks-list'
import SubViewError from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

const ApplicationWebhooks = props => {
  const { match, appId } = props

  useBreadcrumbs(
    'apps.single.integrations.webhooks',
    <Breadcrumb
      path={`/applications/${appId}/integrations/webhooks`}
      content={sharedMessages.webhooks}
    />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route exact path={`${match.path}`} component={ApplicationWebhooksList} />
        <Route exact path={`${match.path}/add`} component={ApplicationWebhookAdd} />
        <Route
          exact
          path={`${match.path}/:webhookId${pathIdRegexp}`}
          component={ApplicationWebhookEdit}
          sensitive
        />
        <Route path={`${match.path}/add/template`} component={ApplicationWebhookChoose} />
      </Routes>
    </ErrorView>
  )
}

ApplicationWebhooks.propTypes = {
  appId: PropTypes.string.isRequired,
  match: PropTypes.match.isRequired,
}

export default ApplicationWebhooks
