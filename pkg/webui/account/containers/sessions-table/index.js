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
import { connect, useDispatch } from 'react-redux'
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'
import Status from '@ttn-lw/components/status'
import toast from '@ttn-lw/components/toast'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getUserSessionsList, deleteUserSession } from '@account/store/actions/user'

import {
  selectUserId,
  selectUserSessions,
  selectUserSessionsTotalCount,
  selectUserSessionsFetching,
  selectSessionId,
} from '@account/store/selectors/user'

const m = defineMessages({
  deleteSessionSuccess: 'Session deleted successfully',
  deleteSession: 'Delete session',
  deleteSessionError: 'There was an error and this session could not be deleted',
  sessionsTableTitle: 'Sessions',
  removeButtonMessage: 'Remove this session',
})

const UserSessionsTable = props => {
  const { selectTableData, pageSize, user, handleDeleteSession } = props
  const dispatch = useDispatch()

  const getSessions = React.useCallback(filters => getUserSessionsList(filters, user), [user])

  const onDeleteSuccess = React.useCallback(() => {
    toast({
      title: m.deleteSession,
      message: m.deleteSessionSuccess,
      type: toast.types.SUCCESS,
    })

    dispatch(getUserSessionsList(user))
  }, [user, dispatch])

  const deleteSession = React.useCallback(
    async session_id => {
      try {
        const result = await handleDeleteSession(user, session_id)
        onDeleteSuccess(result)
      } catch {
        toast({
          title: m.deleteSession,
          message: m.deleteSessionError,
          type: toast.types.ERROR,
        })
      }
    },
    [user, handleDeleteSession, onDeleteSuccess],
  )

  const makeHeaders = React.useMemo(() => {
    const onDelete = session_id => () => deleteSession(session_id)

    const baseHeaders = [
      {
        name: 'session_id',
        displayName: 'ID',
        width: 5,
        sortable: true,
      },
      {
        name: 'status',
        displayName: '',
        width: 40,
        render: status => {
          if (status.currentSession) {
            return <Status flipped status="good" label="Current session" />
          }
        },
      },
      {
        name: 'updated_at',
        displayName: 'Last activity',
        width: 4,
        sortable: true,
        render: updated_at => <LastSeen status="none" lastSeen={updated_at} short />,
      },
      {
        name: 'status',
        displayName: '',
        width: 10,
        render: status => {
          const handleDeleteSession = onDelete(status._session_id)
          if (status.currentSession) {
            return null
          }

          return (
            <Button
              type="button"
              onClick={handleDeleteSession}
              message={m.removeButtonMessage}
              icon="delete"
            />
          )
        },
      },
    ]

    return baseHeaders
  }, [deleteSession])

  return (
    <FetchTable
      entity="sessions"
      headers={makeHeaders}
      getItemsAction={getSessions}
      baseDataSelector={selectTableData}
      tableTitle={<Message content={m.sessionsTableTitle} />}
      pageSize={pageSize}
    />
  )
}

UserSessionsTable.propTypes = {
  handleDeleteSession: PropTypes.func.isRequired,
  pageSize: PropTypes.number.isRequired,
  selectTableData: PropTypes.func.isRequired,
  user: PropTypes.string.isRequired,
}

export default connect(
  state => ({
    user: selectUserId(state),
    sessionId: selectSessionId(state),
  }),
  dispatch => ({
    handleDeleteSession: (user, session_id) =>
      dispatch(attachPromise(deleteUserSession(user, session_id))),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    selectTableData: state => {
      const sessions = selectUserSessions(state)
      const decoratedSessions = []

      if (sessions !== undefined) {
        for (const session of sessions) {
          decoratedSessions.push({
            ...session,
            id: session.session_id,
            status: {
              currentSession: session.session_id === stateProps.sessionId,
              _session_id: session.session_id,
            },
          })
        }
      }

      return {
        sessions: decoratedSessions,
        totalCount: selectUserSessionsTotalCount(state),
        fetching: selectUserSessionsFetching(state),
        mayAdd: false,
        mayLink: false,
      }
    },
  }),
)(UserSessionsTable)
