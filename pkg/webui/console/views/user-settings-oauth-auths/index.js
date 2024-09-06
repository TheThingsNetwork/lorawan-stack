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
import { defineMessages } from 'react-intl'
import { Routes, Route, useParams } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import Tabs from '@ttn-lw/components/tabs'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import TokensTable from '@console/containers/tokens-table'

import AuthorizationSettings from '@console/views/user-settings-oauth-auth-settings'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import { getClientsList } from '@console/store/actions/clients'
import { getAuthorizationsList } from '@console/store/actions/authorizations'

import { selectClientById } from '@console/store/selectors/clients'
import { selectUserId } from '@console/store/selectors/user'

import style from './authorization.styl'

const m = defineMessages({
  authorizationSettings: 'Authorization settings',
  accessTokens: 'Active access tokens',
})

const AuthorizationOverviewInner = () => {
  const { clientId } = useParams()
  const siteName = useSelector(selectApplicationSiteName)

  const client = useSelector(state => selectClientById(state, clientId))
  const clientName = client?.name || clientId

  const basePath = `/user-settings/authorizations/${clientId}`

  useBreadcrumbs(
    'user-settings.oauth-client-authorizations.single',
    <Breadcrumb path={basePath} content={clientId} />,
  )

  const tabs = [
    {
      title: m.authorizationSettings,
      name: 'overview',
      link: `${basePath}`,
    },
    { title: m.accessTokens, name: 'access-tokens', link: `${basePath}/access-tokens` },
  ]

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${clientId} - ${siteName}`} />
      <div className={style.titleSection}>
        <div className="container container--xl grid pb-0">
          <div className="item-12">
            <PageTitle title={clientName} className={style.pageTitle} />
            <Tabs className={style.tabs} narrow tabs={tabs} />
          </div>
        </div>
      </div>
      <Routes>
        <Route index Component={AuthorizationSettings} />
        <Route path="access-tokens" Component={TokensTable} />
        <Route path="*" Component={<GenericNotFound />} />
      </Routes>
    </>
  )
}

const AuthorizationOverview = () => {
  const userId = useSelector(selectUserId)

  return (
    <RequireRequest
      requestAction={[getAuthorizationsList(userId), getClientsList(undefined, ['name'])]}
    >
      <AuthorizationOverviewInner />
    </RequireRequest>
  )
}

export default AuthorizationOverview
