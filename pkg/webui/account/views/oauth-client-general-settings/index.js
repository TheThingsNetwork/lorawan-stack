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
import { useSelector } from 'react-redux'

import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import OAuthClientEdit from '@account/containers/oauth-client-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getUserRights } from '@account/store/actions/user'
import { getIsConfiguration } from '@account/store/actions/identity-server'

import {
  selectUserIsAdmin,
  selectUserId,
  selectUserRegularRights,
  selectUserPseudoRights,
} from '@account/store/selectors/user'
import { selectSelectedClient } from '@account/store/selectors/clients'

const OAuthClientGeneralSettingsInner = () => {
  const userId = useSelector(selectUserId)
  const isAdmin = useSelector(selectUserIsAdmin)
  const regularRights = useSelector(selectUserRegularRights)
  const pseudoRights = useSelector(selectUserPseudoRights)
  const oauthClient = useSelector(selectSelectedClient)

  return (
    <div className="container container--lg grid">
      <PageTitle title={sharedMessages.generalSettings} />
      <div className="item-12 lg:item-8">
        <OAuthClientEdit
          initialValues={oauthClient}
          isAdmin={isAdmin}
          userId={userId}
          rights={regularRights}
          pseudoRights={pseudoRights}
          update
        />
      </div>
    </div>
  )
}

const OAuthClientGeneralSettings = () => {
  const userId = useSelector(selectUserId)

  return (
    <RequireRequest requestAction={[getUserRights(userId), getIsConfiguration()]}>
      <OAuthClientGeneralSettingsInner />
    </RequireRequest>
  )
}

export default OAuthClientGeneralSettings
