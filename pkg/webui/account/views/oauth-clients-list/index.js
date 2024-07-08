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

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import PageTitle from '@ttn-lw/components/page-title'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import ClientsTable from '@account/containers/clients-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getUserRights } from '@account/store/actions/user'

import { selectUserId } from '@account/store/selectors/user'

const ClientsList = () => {
  const userId = useSelector(selectUserId)
  return (
    <RequireRequest requestAction={getUserRights(userId)}>
      <div className="container container--xxl grid">
        <IntlHelmet title={sharedMessages.oauthClients} />
        <div className="item-12">
          <PageTitle title={sharedMessages.oauthClients} hideHeading />
          <ClientsTable pageSize={PAGE_SIZES.REGULAR} />
        </div>
      </div>
    </RequireRequest>
  )
}

export default ClientsList
