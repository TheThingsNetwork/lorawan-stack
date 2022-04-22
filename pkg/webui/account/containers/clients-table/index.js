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
import { defineMessages, useIntl } from 'react-intl'
import { bindActionCreators } from 'redux'

import toast from '@ttn-lw/components/toast'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayPerformAdminActions } from '@account/lib/feature-checks'

import { deleteClient, restoreClient, getClientsList } from '@account/store/actions/clients'

import { selectUserIsAdmin } from '@account/store/selectors/user'
import {
  selectOAuthClients,
  selectOAuthClientsTotalCount,
  selectOAuthClientsFetching,
} from '@account/store/selectors/clients'

const m = defineMessages({
  ownedTabTitle: 'Owned OAuth clients',
  restoreSuccess: 'OAuth client restored',
  restoreFail: 'There was an error and OAuth client could not be restored',
  purgeSuccess: 'OAuth client purged',
  purgeFail: 'There was an error and the OAuth client could not be purged',
  addClient: 'Add OAuth Client',
})

const OWNED_TAB = 'owned'
const ALL_TAB = 'all'
const DELETED_TAB = 'deleted'
const tabs = [
  {
    title: m.ownedTabTitle,
    name: OWNED_TAB,
  },
  {
    title: sharedMessages.allAdmin,
    name: ALL_TAB,
  },
  { title: sharedMessages.deleted, name: DELETED_TAB },
]

const capitalize = str => str.charAt(0).toUpperCase() + str.slice(1)

const ClientsTable = props => {
  const { isAdmin, restoreClient, purgeClient, ...rest } = props
  const { formatMessage } = useIntl()

  const [tab, setTab] = React.useState(OWNED_TAB)
  const isDeletedTab = tab === DELETED_TAB

  const handleRestore = React.useCallback(
    async id => {
      try {
        await restoreClient(id)
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
    [restoreClient],
  )

  const handlePurge = React.useCallback(
    async id => {
      try {
        await purgeClient(id)
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
    [purgeClient],
  )

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'ids.client_id',
        displayName: sharedMessages.id,
        width: 25,
        sortable: true,
        sortKey: 'client_id',
      },
      {
        name: 'name',
        displayName: sharedMessages.name,
        width: 25,
        sortable: true,
      },
      {
        name: 'description',
        displayName: sharedMessages.description,
        width: 50,
      },
    ]

    if (tab === DELETED_TAB) {
      baseHeaders.push({
        name: 'actions',
        displayName: sharedMessages.actions,
        width: 50,
        getValue: row => ({
          id: row.ids.client_id,
          name: row.name,
          restore: handleRestore.bind(null, row.ids.client_id),
          purge: handlePurge.bind(null, row.ids.client_id),
        }),
        render: details => (
          <ButtonGroup align="end">
            <Button message={sharedMessages.restore} onClick={details.restore} />
            <DeleteModalButton
              entityId={details.id}
              entityName={name}
              message={sharedMessages.purge}
              onApprove={details.purge}
              onlyPurge
            />
          </ButtonGroup>
        ),
      })
    } else {
      baseHeaders.push({
        name: 'state',
        displayName: sharedMessages.state,
        width: 50,
        render: state => capitalize(formatMessage({ id: `enum:${state}` })),
      })
    }

    return baseHeaders
  }, [tab, handlePurge, handleRestore, formatMessage])

  const baseDataSelector = React.useCallback(
    state => ({
      clients: selectOAuthClients(state),
      totalCount: selectOAuthClientsTotalCount(state),
      fetching: selectOAuthClientsFetching(state),
      mayAdd: checkFromState(mayPerformAdminActions, state),
    }),
    [],
  )

  const getItems = React.useCallback(filters => {
    const { tab, query } = filters
    const isDeletedTab = tab === DELETED_TAB

    setTab(tab)
    return getClientsList({ ...filters, deleted: isDeletedTab }, ['name', 'description', 'state'], {
      isSearch: tab === ALL_TAB || isDeletedTab || query.length > 0,
    })
  }, [])

  return (
    <FetchTable
      entity="clients"
      defaultOrder="client_id"
      addMessage={m.addClient}
      headers={headers}
      getItemsAction={getItems}
      baseDataSelector={baseDataSelector}
      tableTitle={<Message content={sharedMessages.oauthClients} />}
      searchable
      clickable={!isDeletedTab}
      tabs={isAdmin ? tabs : []}
      {...rest}
    />
  )
}

ClientsTable.propTypes = {
  isAdmin: PropTypes.bool.isRequired,
  purgeClient: PropTypes.func.isRequired,
  restoreClient: PropTypes.func.isRequired,
}

export default connect(
  state => ({
    isAdmin: selectUserIsAdmin(state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        purgeClient: attachPromise(deleteClient),
        restoreClient: attachPromise(restoreClient),
      },
      dispatch,
    ),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    purgeClient: id => dispatchProps.purgeClient(id, { purge: true }),
    restoreClient: id => dispatchProps.restoreClient(id),
  }),
)(ClientsTable)
