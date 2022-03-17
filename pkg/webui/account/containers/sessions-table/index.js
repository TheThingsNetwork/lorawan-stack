// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import Button from '@ttn-lw/components/button'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getUserSessionsList } from '@account/store/actions/user'

import {
  selectUserId,
  selectUserSessions,
  selectUserSessionsTotalCount,
  selectUserSessionsFetching,
} from '@account/store/selectors/user'

const headers = [
  {
    name: 'session_id',
    displayName: 'ID',
    width: 24,
    sortable: true,
  },
  {
    name: 'updated_at',
    displayName: 'Last seen',
    width: 22,
    sortable: true,
    render: updated_at => <LastSeen status="none" lastSeen={updated_at} short />,
  },
  {
    name: 'created_at',
    displayName: '',
    width: 22,
    render: () => (
      <Button type="button" message="Remove this session" icon="delete" title={'Delete'} />
    ),
  },
]

const UserSessionsTable = props => {
  const { selectTableData, pageSize, user } = props

  const getSessions = React.useCallback(filters => getUserSessionsList(filters, user), [user])

  return (
    <FetchTable
      entity="sessions"
      headers={headers}
      getItemsAction={getSessions}
      baseDataSelector={selectTableData}
      tableTitle={<Message content={'Sessions'} />}
      pageSize={pageSize}
    />
  )
}

UserSessionsTable.propTypes = {
  pageSize: PropTypes.number.isRequired,
  selectTableData: PropTypes.func.isRequired,
}

export default connect(
  state => ({
    user: selectUserId(state),
  }),
  null,
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    selectTableData: state => ({
      sessions: selectUserSessions(state),
      totalCount: selectUserSessionsTotalCount(state),
      fetching: selectUserSessionsFetching(state),
    }),
  }),
)(UserSessionsTable)
