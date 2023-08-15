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
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import { createSelector } from 'reselect'

import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getUserSessionsList, deleteUserSession } from '@account/store/actions/sessions'

import { selectUserId, selectSessionId } from '@account/store/selectors/user'
import { selectUserSessions, selectUserSessionsTotalCount } from '@account/store/selectors/sessions'

const m = defineMessages({
  deleteSessionSuccess: 'Session removed successfully',
  deleteSessionError: 'There was an error and the session could not be deleted',
  sessionsTableTitle: 'Sessions',
  removeButtonMessage: 'Remove this session',
  noExpiryDate: 'No expiry date',
  endSession: 'Logout to end this session',
  currentSession: '(This is the current session)',
})

const getItemPathPrefix = item => `/${item.ids.client_id}`

const UserSessionsTable = () => {
  const userId = useSelector(selectUserId)
  const sessionId = useSelector(selectSessionId)
  const dispatch = useDispatch()

  const getSessions = React.useCallback(filters => getUserSessionsList(userId, filters), [userId])

  const deleteSession = React.useCallback(
    async sessionId => {
      try {
        await dispatch(attachPromise(deleteUserSession(userId, sessionId)))
        toast({
          message: m.deleteSessionSuccess,
          type: toast.types.SUCCESS,
        })
        dispatch(getUserSessionsList(userId))
      } catch {
        toast({
          message: m.deleteSessionError,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch, userId],
  )

  const baseDataSelector = createSelector(
    [selectUserSessions, selectUserSessionsTotalCount],
    (sessions, totalCount) => {
      const decoratedSessions = []

      if (sessions) {
        for (const session of sessions) {
          decoratedSessions.push({
            ...session,
            id: session.session_id,
            status: {
              currentSession: session.session_id === sessionId,
            },
          })
        }
      }

      return {
        sessions: decoratedSessions,
        totalCount,
        mayAdd: false,
        mayLink: false,
      }
    },
  )

  const makeHeaders = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'session_id',
        displayName: sharedMessages.id,
        width: 25,
        getValue: row => ({
          id: row.session_id,
          status: row.status,
        }),
        render: details => (
          <>
            {`${details.id.substr(0, 12)}... `}
            {details.status.currentSession && <Message content={m.currentSession} />}
          </>
        ),
      },
      {
        name: 'created_at',
        displayName: sharedMessages.createdAt,
        width: 25,
        sortable: true,
        render: created_at => (
          <>
            <DateTime value={created_at} />
            {' ('}
            <DateTime.Relative value={created_at} />
            {')'}
          </>
        ),
      },
      {
        name: 'expires_at',
        displayName: sharedMessages.expiry,
        width: 20,
        render: expires_at => {
          if (expires_at === undefined) {
            return <Message content={m.noExpiryDate} className="tc-subtle-gray" />
          }

          return (
            <>
              <DateTime value={expires_at} />
              {' ('}
              <DateTime.Relative value={expires_at} />
              {')'}
            </>
          )
        },
      },
      {
        name: 'actions',
        displayName: sharedMessages.actions,
        width: 20,
        getValue: row => ({
          id: row.session_id,
          status: row.status,
          delete: deleteSession.bind(null, row.session_id),
        }),
        render: details => {
          if (details.status.currentSession) {
            return <Message content={m.endSession} />
          }

          return (
            <Button
              type="button"
              onClick={details.delete}
              message={m.removeButtonMessage}
              icon="delete"
              danger
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
      baseDataSelector={baseDataSelector}
      tableTitle={<Message content={m.sessionsTableTitle} />}
      getItemPathPrefix={getItemPathPrefix}
    />
  )
}

export default UserSessionsTable
