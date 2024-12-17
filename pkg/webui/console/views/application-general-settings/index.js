// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import ApplicationGeneralSettingsContainer from '@console/containers/application-general-settings'

import Require from '@console/lib/components/require'

import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayEditBasicApplicationInfo,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayViewApplicationLink,
} from '@console/lib/feature-checks'

import { getAppPkgDefaultAssoc } from '@console/store/actions/application-packages'
import { getPubsubsList } from '@console/store/actions/pubsubs'
import { getWebhooksList } from '@console/store/actions/webhooks'
import { getApplicationLink } from '@console/store/actions/link'
import { getApiKeysList } from '@console/store/actions/api-keys'
import { getIsConfiguration } from '@console/store/actions/identity-server'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

const ApplicationGeneralSettings = () => {
  const appId = useSelector(selectSelectedApplicationId)
  const mayViewApiKeys = useSelector(state =>
    checkFromState(mayViewOrEditApplicationApiKeys, state),
  )
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditApplicationCollaborators, state),
  )
  const mayViewAppLink = useSelector(state => checkFromState(mayViewApplicationLink, state))

  const getApiKeysListRequest = mayViewApiKeys ? getApiKeysList('application', appId) : undefined
  const getApplicationLinkRequest = mayViewAppLink
    ? getApplicationLink(appId, 'skip_payload_crypto')
    : undefined
  const getCollaboratorsListRequest = mayViewCollaborators
    ? getCollaboratorsList('application', appId)
    : undefined
  const requestsList = [
    getAppPkgDefaultAssoc(appId, 202),
    getWebhooksList(appId),
    getPubsubsList(appId),
    getApiKeysListRequest,
    getApplicationLinkRequest,
    getCollaboratorsListRequest,
    getIsConfiguration(),
  ]

  useBreadcrumbs(
    `apps.single#${appId}.general-settings`,
    <Breadcrumb path={`/applications/${appId}`} content={sharedMessages.generalSettings} />,
  )

  return (
    <Require
      featureCheck={mayEditBasicApplicationInfo}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      {/* The request getApplicationLink returns 404 when there is no `skip_payload_crypto`. */}
      {/* This is expected behavior and should not be treated as an error. */}
      <RequireRequest requestAction={requestsList} handleErrors={false}>
        <div className="container container--xl grid">
          <PageTitle title={sharedMessages.generalSettings} />
          <div className="item-12">
            <ApplicationGeneralSettingsContainer appId={appId} />
          </div>
        </div>
      </RequireRequest>
    </Require>
  )
}

export default ApplicationGeneralSettings
