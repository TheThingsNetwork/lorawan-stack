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

import React, { Component } from 'react'

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import Status from '../../../components/status'
import Icon from '../../../components/icon'

import FetchTable from '../fetch-table'

import { getUsersList } from '../../../console/store/actions/users'
import {
  selectUsers,
  selectUsersTotalCount,
  selectUsersFetching,
  selectUsersError,
} from '../../store/selectors/users'

import style from './users-table.styl'

const headers = [
  {
    name: 'ids.user_id',
    displayName: sharedMessages.id,
    width: 28,
  },
  {
    name: 'name',
    displayName: sharedMessages.name,
    width: 22,
  },
  {
    name: 'primary_email_address',
    displayName: sharedMessages.email,
    width: 28,
  },
  {
    name: 'state',
    displayName: sharedMessages.state,
    width: 15,
    render(state) {
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
  },
  {
    name: 'admin',
    displayName: sharedMessages.admin,
    width: 7,
    render(isAdmin) {
      if (isAdmin) {
        return <Icon className={style.icon} icon="check" />
      }

      return null
    },
  },
]

export default class UsersTable extends Component {
  static propTypes = {
    pageSize: PropTypes.number.isRequired,
  }

  constructor(props) {
    super(props)

    this.getUsersList = params =>
      getUsersList(params, ['name', 'primary_email_address', 'state', 'admin'])
  }

  baseDataSelector(state) {
    return {
      users: selectUsers(state),
      totalCount: selectUsersTotalCount(state),
      fetching: selectUsersFetching(state),
      error: selectUsersError(state),
      mayAdd: false,
    }
  }

  render() {
    const { pageSize } = this.props

    return (
      <FetchTable
        entity="users"
        headers={headers}
        addMessage={sharedMessages.addOrganization}
        tableTitle={<Message content={sharedMessages.users} />}
        getItemsAction={this.getUsersList}
        searchItemsAction={this.getUsersList}
        baseDataSelector={this.baseDataSelector}
        pageSize={pageSize}
        searchable
      />
    )
  }
}
