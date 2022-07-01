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
import { bindActionCreators } from 'redux'
import { defineMessages } from 'react-intl'

import Status from '@ttn-lw/components/status'
import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { getUserId } from '@ttn-lw/lib/selectors/id'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayManageUsers, maySendInvites } from '@console/lib/feature-checks'

import { getUsersList, deleteInvite, getUserInvitations } from '@console/store/actions/users'

import { selectUserId, selectUserIsAdmin } from '@console/store/selectors/logout'
import {
  selectUsers,
  selectUsersTotalCount,
  selectUsersFetching,
  selectUserInvitations,
  selectUserInvitationsTotalCount,
  selectUserInvitationsFetching,
} from '@console/store/selectors/users'

import style from './users-table.styl'

const m = defineMessages({
  invite: 'Invite user',
  revokeInvitation: 'Revoke this invitation',
  sentAt: 'Sent',
  revokeSuccess: 'Invite removed successfully',
  revokeError: 'There was an error and the invite could not be revoked',
})

const USERS_TAB = 'users'
const INVITATIONS_TAB = 'invitations'
const tabs = [
  {
    title: sharedMessages.users,
    name: 'users',
  },
  {
    title: sharedMessages.userInvitations,
    name: 'invitations',
  },
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
  const { pageSize, currentUserId, maySendInvites, handleDeleteInvitation } = props
  const dispatch = useDispatch()

  const [tab, setTab] = React.useState(USERS_TAB)
  const isInvitationsTab = tab === INVITATIONS_TAB

  const onDeleteInvite = React.useCallback(
    async email => {
      try {
        await handleDeleteInvitation(email)
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
    [handleDeleteInvitation, dispatch],
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
              icon="delete"
              danger
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
                  <Message className={style.hint} content={sharedMessages.currentUserIndicator} />
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
        ...state,
        {
          name: 'admin',
          displayName: sharedMessages.admin,
          width: 7,
          render: isAdmin => {
            if (isAdmin) {
              return <Icon className={style.icon} icon="check" />
            }

            return null
          },
        },
      )
    }

    return baseHeaders
  }, [currentUserId, tab, onDeleteInvite])

  const getItems = React.useCallback(params => {
    const { tab } = params
    setTab(tab)

    if (tab === INVITATIONS_TAB) {
      return getUserInvitations(params, ['state'])
    }
    return getUsersList(params, ['name', 'primary_email_address', 'state', 'admin'])
  }, [])

  const baseDataSelector = React.useCallback(
    state => {
      if (tab === INVITATIONS_TAB) {
        return {
          invitations: selectUserInvitations(state),
          totalCount: selectUserInvitationsTotalCount(state),
          fetching: selectUserInvitationsFetching(state),
          mayAdd: maySendInvites,
          mayLink: false,
        }
      }

      return {
        users: selectUsers(state),
        totalCount: selectUsersTotalCount(state),
        fetching: selectUsersFetching(state),
        mayAdd: checkFromState(mayManageUsers, state),
      }
    },
    [tab, maySendInvites],
  )

  const getItemPathPrefix = item => `/${item.email}`

  return (
    <FetchTable
      entity={isInvitationsTab ? 'invitations' : 'users'}
      defaultOrder={isInvitationsTab ? '' : '-user_id'}
      headers={headers}
      addMessage={isInvitationsTab ? m.invite : sharedMessages.userAdd}
      itemPathPrefix={isInvitationsTab ? '/invitations' : ''}
      getItemPathPrefix={isInvitationsTab ? getItemPathPrefix : undefined}
      tableTitle={<Message content={sharedMessages.users} />}
      getItemsAction={getItems}
      searchItemsAction={getItems}
      baseDataSelector={baseDataSelector}
      pageSize={pageSize}
      tabs={maySendInvites ? tabs : []}
      searchable={!isInvitationsTab}
    />
  )
}

UsersTable.propTypes = {
  currentUserId: PropTypes.string.isRequired,
  handleDeleteInvitation: PropTypes.func.isRequired,
  maySendInvites: PropTypes.bool.isRequired,
  pageSize: PropTypes.number.isRequired,
}

export default connect(
  state => ({
    currentUserId: selectUserId(state),
    isAdmin: selectUserIsAdmin(state),
    maySendInvites: checkFromState(maySendInvites, state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        handleDeleteInvitation: attachPromise(deleteInvite),
      },
      dispatch,
    ),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    handleDeleteSession: email => dispatchProps.handleDeleteSession(email),
  }),
)(UsersTable)
