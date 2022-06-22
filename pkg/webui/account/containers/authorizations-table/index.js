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
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getAuthorizationsList } from '@account/store/actions/authorizations'

import { selectUserId } from '@account/store/selectors/user'
import {
  selectAuthorizations,
  selectAuthorizationsTotalCount,
  selectAuthorizationsFetching,
} from '@account/store/selectors/authorizations'

const m = defineMessages({
  clientId: 'Client ID',
  tableTitle: 'OAuth Client Authorizations',
})

const getItemPathPrefix = item => `/${item.client_ids.client_id}`

const OAuthClientAuthorizationsTable = props => {
  const { userId, ...rest } = props

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'client_ids.client_id',
        displayName: m.clientId,
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

  const baseDataSelector = React.useCallback(
    state => ({
      authorizations: selectAuthorizations(state),
      totalCount: selectAuthorizationsTotalCount(state),
      fetching: selectAuthorizationsFetching(state),
      mayAdd: false,
    }),
    [],
  )

  const getItems = React.useCallback(() => getAuthorizationsList(userId, []), [userId])

  return (
    <FetchTable
      entity="authorizations"
      defaultOrder="-created_at"
      headers={headers}
      getItemsAction={getItems}
      baseDataSelector={baseDataSelector}
      getItemPathPrefix={getItemPathPrefix}
      tableTitle={<Message content={m.tableTitle} />}
      handlesSorting
      clickable
      {...rest}
    />
  )
}

OAuthClientAuthorizationsTable.propTypes = {
  userId: PropTypes.string.isRequired,
}

export default connect(state => ({
  userId: selectUserId(state),
}))(OAuthClientAuthorizationsTable)
