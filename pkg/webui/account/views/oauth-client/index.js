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
import { shallowEqual, useSelector } from 'react-redux'
import { Routes, Route, useParams } from 'react-router-dom'

import applicationIcon from '@assets/misc/application.svg'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import SideNavigation from '@ttn-lw/components/navigation/side'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import OAuthClientOverview from '@account/views/oauth-client-overview'
import OAuthClientGeneralSettings from '@account/views/oauth-client-general-settings'
import OAuthClientCollaboratorsList from '@account/views/oauth-client-collaborators-list'
import OAuthClientCollaboratorAdd from '@account/views/oauth-client-collaborator-add'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayPerformAllClientActions } from '@account/lib/feature-checks'

import { getClient, getClientRights } from '@account/store/actions/clients'

import { selectClientById, selectClientRights } from '@account/store/selectors/clients'

const OAuthClientInner = () => {
  const { clientId } = useParams()
  const rights = useSelector(selectClientRights, shallowEqual)
  const oauthClient = useSelector(state => selectClientById(state, clientId))
  const siteName = useSelector(selectApplicationSiteName)
  const name = oauthClient.name || clientId

  useBreadcrumbs(
    'clients.single',
    <Breadcrumb path={`/oauth-clients/${clientId}`} content={name} />,
  )

  return (
    <React.Fragment>
      <Breadcrumbs />
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <SideNavigation
        header={{
          icon: applicationIcon,
          iconAlt: sharedMessages.client,
          title: name,
          to: '',
        }}
      >
        {mayPerformAllClientActions.check(rights) && (
          <SideNavigation.Item title={sharedMessages.overview} path="" icon="overview" exact />
        )}
        {mayPerformAllClientActions.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path="collaborators"
            icon="organization"
          />
        )}
        {mayPerformAllClientActions.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            path="general-settings"
            icon="general_settings"
          />
        )}
      </SideNavigation>
      <Routes>
        <Route index Component={OAuthClientOverview} />
        <Route path="collaborators" Component={OAuthClientCollaboratorsList} />
        <Route path="collaborators/add" Component={OAuthClientCollaboratorAdd} />
        <Route path="general-settings" Component={OAuthClientGeneralSettings} />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </React.Fragment>
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
  ]

  return (
    <RequireRequest requestAction={[getClientRights(clientId), getClient(clientId, selector)]}>
      <OAuthClientInner />
    </RequireRequest>
  )
}

export default OAuthClient
