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
import { createSelector } from 'reselect'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getAuthorizationsList } from '@account/store/actions/authorizations'

import {
  selectAuthorizations,
  selectAuthorizationsTotalCount,
} from '@account/store/selectors/authorizations'
import { selectUserId } from '@account/store/selectors/user'

const getItemPathPrefix = item => `${item.client_ids.client_id}`

const OAuthClientAuthorizationsTable = () => {
  const userId = useSelector(selectUserId)

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'client_ids.client_id',
        displayName: sharedMessages.clientId,
        width: 20,
      },
      {
        name: 'user_ids.user_id',
        displayName: sharedMessages.userId,
        width: 20,
      },
      {
        name: 'created_at',
        displayName: sharedMessages.created,
        width: 40,
        sortable: true,
        render: created_at => <DateTime.Relative value={created_at} />,
      },
    ]
    return baseHeaders
  }, [])

  const baseDataSelector = createSelector(
    [selectAuthorizations, selectAuthorizationsTotalCount],
    (authorizations, totalCount) => ({
      authorizations,
      totalCount,
      mayAdd: false,
    }),
  )

  const getItems = React.useCallback(filters => getAuthorizationsList(userId, filters), [userId])

  return (
    <FetchTable
      entity="authorizations"
      defaultOrder="-created_at"
      headers={headers}
      getItemsAction={getItems}
      baseDataSelector={baseDataSelector}
      getItemPathPrefix={getItemPathPrefix}
      tableTitle={<Message content={sharedMessages.oauthClientAuthorizations} />}
      clickable
    />
  )
}

export default OAuthClientAuthorizationsTable
