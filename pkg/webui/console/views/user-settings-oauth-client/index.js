// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'
import { Routes, Route, useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Tabs from '@ttn-lw/components/tabs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import OAuthClientCollaboratorsList from '@console/views/user-settings-oauth-client-collaborators-list'
import OAuthClientCollaboratorAdd from '@console/views/user-settings-oauth-client-collaborator-add'
import OAuthClientGeneralSettings from '@console/views/user-settings-oauth-client-general-settings'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayPerformAllClientActions, checkFromState } from '@console/lib/feature-checks'

import { getClient, getClientRights } from '@console/store/actions/clients'

import { selectClientById } from '@console/store/selectors/clients'

import style from './user-settings-oauth-client.styl'

const OAuthClientInner = () => {
  const { clientId } = useParams()
  const oauthClient = useSelector(state => selectClientById(state, clientId))
  const siteName = useSelector(selectApplicationSiteName)
  const name = oauthClient.name || clientId
  const basePath = `/user-settings/oauth-clients/${clientId}`
  const canPerformAllActions = useSelector(state =>
    checkFromState(mayPerformAllClientActions, state),
  )

  const tabs = [
    {
      title: sharedMessages.generalSettings,
      name: 'oauth-client-settings',
      link: `${basePath}`,
    },
    {
      title: sharedMessages.collaborators,
      name: 'oauth-client-collaborators',
      link: `${basePath}/collaborators`,
      disabled: !canPerformAllActions,
      exact: false,
    },
  ]

  useBreadcrumbs(
    'user-settings.oauth-clients.single',
    <Breadcrumb path={`/user-settings/oauth-clients/${clientId}`} content={name} />,
  )

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <div className={style.titleSection}>
        <div className="container container--xl grid pb-0">
          <div className="item-12">
            <PageTitle title={name} className={style.pageTitle} />
            <Tabs className={style.tabs} narrow tabs={tabs} />
          </div>
        </div>
      </div>
      <Routes>
        <Route index Component={OAuthClientGeneralSettings} />
        <Route path="collaborators" Component={OAuthClientCollaboratorsList} />
        <Route path="collaborators/add" Component={OAuthClientCollaboratorAdd} />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </>
  )
}

const OAuthClient = () => {
  const { clientId } = useParams()
  const selector = [
    'name',
    'description',
    'state',
    'state_description',
    'redirect_uris',
    'logout_redirect_uris',
    'skip_authorization',
    'endorsed',
    'grants',
    'rights',
    'administrative_contact',
    'technical_contact',
  ]

  return (
    <RequireRequest requestAction={[getClientRights(clientId), getClient(clientId, selector)]}>
      <OAuthClientInner />
    </RequireRequest>
  )
}

export default OAuthClient
