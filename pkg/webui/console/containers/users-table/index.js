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

import Icon, { IconTrash, IconCheck } from '@ttn-lw/components/icon'
import Status from '@ttn-lw/components/status'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'
import ButtonGroup from '@ttn-lw/components/button/group'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { getUserId } from '@ttn-lw/lib/selectors/id'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayManageUsers, maySendInvites } from '@console/lib/feature-checks'

import {
  getUsersList,
  deleteInvite,
  getUserInvitations,
  deleteUser,
  restoreUser,
} from '@console/store/actions/users'

import { selectUserId } from '@console/store/selectors/user'
import {
  selectUsers,
  selectUsersTotalCount,
  selectUserInvitations,
  selectUserInvitationsTotalCount,
} from '@console/store/selectors/users'

const m = defineMessages({
  invite: 'Invite user',
  revokeInvitation: 'Revoke this invitation',
  sentAt: 'Sent',
  revokeSuccess: 'Invite removed successfully',
  revokeError: 'There was an error and the invite could not be revoked',
})

const USERS_TAB = 'users'
const DELETED_TAB = 'deleted'
const INVITATIONS_TAB = 'invitations'
const tabs = [
  {
    title: sharedMessages.users,
    name: USERS_TAB,
  },
  { title: sharedMessages.deleted, name: DELETED_TAB },
]

const state = {
  name: 'state',
  displayName: sharedMessages.state,
  width: 15,
  render: state => {
    let indicator = 'unknown'
    let label = sharedMessages.notSet
    switch (state) {
      case 'STATE_APPROVED':
        indicator = 'good'
        label = sharedMessages.stateApproved
        break
      case 'STATE_REQUESTED':
        indicator = 'mediocre'
        label = sharedMessages.stateRequested
        break
      case 'STATE_REJECTED':
        indicator = 'bad'
        label = sharedMessages.stateRejected
        break
      case 'STATE_FLAGGED':
        indicator = 'bad'
        label = sharedMessages.stateFlagged
        break
      case 'STATE_SUSPENDED':
        indicator = 'bad'
        label = sharedMessages.stateSuspended
        break
    }

    return <Status status={indicator} label={label} pulse={false} />
  },
}

const UsersTable = props => {
  const { pageSize } = props
  const dispatch = useDispatch()
  const currentUserId = useSelector(selectUserId)
  const mayInvite = useSelector(state => checkFromState(maySendInvites, state))

  const tabsWithInvitations = [
    ...tabs,
    { title: sharedMessages.userInvitations, name: INVITATIONS_TAB },
  ]
  const [tab, setTab] = React.useState(USERS_TAB)
  const isInvitationsTab = tab === INVITATIONS_TAB
  const isDeletedTab = tab === DELETED_TAB

  const handleRestore = React.useCallback(
    async id => {
      try {
        await dispatch(attachPromise(restoreUser(id)))
        toast({
          title: id,
          message: m.restoreSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (err) {
        toast({
          title: id,
          message: m.restoreFail,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch],
  )

  const handlePurge = React.useCallback(
    async id => {
      try {
        await dispatch(attachPromise(deleteUser(id, { purge: true })))
        toast({
          title: id,
          message: m.purgeSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (err) {
        toast({
          title: id,
          message: m.purgeFail,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch],
  )

  const onDeleteInvite = React.useCallback(
    async email => {
      try {
        await dispatch(attachPromise(deleteInvite(email)))
        toast({
          message: m.revokeSuccess,
          type: toast.types.SUCCESS,
        })
        setTab(INVITATIONS_TAB)
        dispatch(getUserInvitations())
      } catch {
        toast({
          message: m.revokeError,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch],
  )

  const headers = React.useMemo(() => {
    const baseHeaders = []

    if (tab === INVITATIONS_TAB) {
      baseHeaders.push(
        {
          name: 'email',
          displayName: sharedMessages.email,
          width: 28,
        },
        {
          name: 'created_at',
          displayName: m.sentAt,
          width: 15,
          render: date => <DateTime.Relative value={date} />,
        },
        ...state,
        {
          name: 'actions',
          displayName: sharedMessages.actions,
          width: 42,
          getValue: row => ({
            email: row.email,
            delete: onDeleteInvite.bind(null, { email: row.email }),
          }),
          render: details => (
            <Button
              type="button"
              onClick={details.delete}
              message={m.revokeInvitation}
              icon={IconTrash}
              danger
              naked
            />
          ),
        },
      )
    } else {
      baseHeaders.push(
        {
          name: 'ids.user_id',
          displayName: sharedMessages.id,
          width: 28,
          sortable: true,
          sortKey: 'user_id',
          render: ids => {
            const userId = getUserId({ ids })
            if (userId === currentUserId) {
              return (
                <span>
                  {userId}{' '}
                  <Message
                    className="c-text-neutral-light"
                    content={sharedMessages.currentUserIndicator}
                  />
                </span>
              )
            }
            return userId
          },
        },
        {
          name: 'name',
          displayName: sharedMessages.name,
          width: 22,
          sortable: true,
        },
        {
          name: 'primary_email_address',
          displayName: sharedMessages.email,
          width: 28,
          sortable: true,
        },
      )
      if (tab === DELETED_TAB) {
        baseHeaders.push({
          name: 'actions',
          displayName: sharedMessages.actions,
          width: 45,
          getValue: row => ({
            id: row.ids.user_id,
            name: row.name,
            restore: handleRestore.bind(null, row.ids.user_id),
            purge: handlePurge.bind(null, row.ids.user_id),
          }),
          render: details => (
            <ButtonGroup align="end">
              <Button message={sharedMessages.restore} onClick={details.restore} secondary />
              <DeleteModalButton
                entityId={details.id}
                entityName={details.name}
                message={sharedMessages.purge}
                onApprove={details.purge}
                onlyPurge
              />
            </ButtonGroup>
          ),
        })
      } else {
        baseHeaders.push(state, {
          name: 'admin',
          displayName: sharedMessages.admin,
          width: 7,
          render: isAdmin => {
            if (isAdmin) {
              return <Icon className="c-text-brand-normal" icon={IconCheck} />
            }

            return null
          },
        })
      }
    }

    return baseHeaders
  }, [currentUserId, tab, onDeleteInvite, handleRestore, handlePurge])

  const getItems = React.useCallback(params => {
    const { tab, query } = params
    const isDeletedTab = tab === DELETED_TAB
    setTab(tab)

    if (tab === INVITATIONS_TAB) {
      return getUserInvitations(params, ['state'])
    }
    return getUsersList(
      { ...params, deleted: isDeletedTab },
      ['name', 'primary_email_address', 'state', 'admin'],
      {
        isSearch: isDeletedTab || query.length > 0,
      },
    )
  }, [])

  const invitationsBaseDataSelector = createSelector(
    selectUserInvitations,
    selectUserInvitationsTotalCount,
    (invitations, totalCount) => ({
      invitations,
      totalCount,
      mayAdd: mayInvite,
      mayLink: false,
    }),
  )
  const usersBaseDataSelector = createSelector(
    selectUsers,
    selectUsersTotalCount,
    (users, totalCount) => ({
      users,
      totalCount,
      mayAdd: mayManageUsers,
    }),
  )

  const getItemPathPrefix = item => `/${item.email}`

  return (
    <FetchTable
      entity={isInvitationsTab ? 'invitations' : 'users'}
      defaultOrder={isInvitationsTab ? '' : '-user_id'}
      headers={headers}
      addMessage={isInvitationsTab ? m.invite : sharedMessages.userAdd}
      itemPathPrefix={isInvitationsTab ? 'invitations/' : ''}
      getItemPathPrefix={isInvitationsTab ? getItemPathPrefix : undefined}
      tableTitle={<Message content={sharedMessages.users} />}
      getItemsAction={getItems}
      searchItemsAction={getItems}
      baseDataSelector={isInvitationsTab ? invitationsBaseDataSelector : usersBaseDataSelector}
      pageSize={pageSize}
      clickable={!isDeletedTab}
      tabs={mayInvite ? tabsWithInvitations : tabs}
      searchable={!isInvitationsTab}
    />
  )
}

UsersTable.propTypes = {
  pageSize: PropTypes.number,
}

UsersTable.defaultProps = {
  pageSize: undefined,
}

export default UsersTable
