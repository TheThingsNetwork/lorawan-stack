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
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { createSelector } from 'reselect'

import Status from '@ttn-lw/components/status'
import toast from '@ttn-lw/components/toast'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import SafeInspector from '@ttn-lw/components/safe-inspector'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { checkFromState, mayCreateGateways } from '@console/lib/feature-checks'

import { getGatewaysList, restoreGateway, deleteGateway } from '@console/store/actions/gateways'

import { selectUserIsAdmin } from '@console/store/selectors/logout'
import { selectGateways, selectGatewaysTotalCount } from '@console/store/selectors/gateways'

const m = defineMessages({
  ownedTabTitle: 'Owned gateways',
  restoreSuccess: 'Gateway restored',
  restoreFail: 'There was an error and the gateway could not be restored',
  purgeSuccess: 'Gateway purged',
  purgeFail: 'There was an error and the gateway could not be purged',
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
const mayAddSelector = state => checkFromState(mayCreateGateways, state)

const GatewaysTable = () => {
  const dispatch = useDispatch()
  const isAdmin = useSelector(selectUserIsAdmin)
  const [tab, setTab] = React.useState(OWNED_TAB)
  const isDeletedTab = tab === DELETED_TAB

  const handleRestore = React.useCallback(
    async id => {
      try {
        await dispatch(attachPromise(restoreGateway(id)))
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
        await dispatch(attachPromise(deleteGateway(id, { purge: true })))
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

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'ids.gateway_id',
        displayName: sharedMessages.id,
        width: 25,
        sortable: true,
        sortKey: 'gateway_id',
      },
      {
        name: 'name',
        displayName: sharedMessages.name,
        width: 33,
        sortable: true,
      },
      {
        name: 'ids.eui',
        displayName: sharedMessages.gatewayEUI,
        width: 22,
        sortable: true,
        sortKey: 'gateway_eui',
        render: gatewayEui =>
          !Boolean(gatewayEui) ? (
            <Message
              className="c-text-neutral-light"
              component="i"
              content={sharedMessages.none}
              firstToLower
            />
          ) : (
            <SafeInspector data={gatewayEui} noTransform noCopyPopup small hideable={false} />
          ),
      },
    ]

    if (tab === DELETED_TAB) {
      baseHeaders.push({
        name: 'actions',
        displayName: sharedMessages.actions,
        width: 25,
        getValue: row => ({
          id: row.ids.gateway_id,
          name: row.name,
          restore: handleRestore.bind(null, row.ids.gateway_id),
          purge: handlePurge.bind(null, row.ids.gateway_id),
        }),
        render: details => (
          <ButtonGroup align="end">
            <Button message={sharedMessages.restore} onClick={details.restore} secondary />
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
      baseHeaders.push(
        {
          name: 'status',
          width: 10,
          displayName: sharedMessages.status,
          render: status => {
            let indicator = 'unknown'
            let label = sharedMessages.unknown

            if (status === 'connected') {
              indicator = 'good'
              label = sharedMessages.connected
            } else if (status === 'disconnected') {
              indicator = 'bad'
              label = sharedMessages.disconnected
            } else if (status === 'other-cluster') {
              indicator = 'unknown'
              label = sharedMessages.otherCluster
            } else if (status === 'unknown') {
              indicator = 'mediocre'
              label = sharedMessages.unknown
            }

            return <Status status={indicator} label={label} flipped />
          },
        },
        {
          name: 'created_at',
          width: 10,
          displayName: sharedMessages.createdAt,
          align: 'right',
          sortable: true,
          render: date => <DateTime.Relative value={date} />,
        },
      )
    }

    return baseHeaders
  }, [handlePurge, handleRestore, tab])

  const baseDataSelector = createSelector(
    selectGateways,
    selectGatewaysTotalCount,
    mayAddSelector,
    (gateways, totalCount, mayAdd) => ({
      gateways,
      totalCount,
      mayAdd,
    }),
  )

  const getGateways = React.useCallback(filters => {
    const { tab, query } = filters
    const isDeletedTab = tab === DELETED_TAB

    setTab(tab)

    return getGatewaysList(
      { ...filters, deleted: isDeletedTab },
      ['name', 'description', 'frequency_plan_ids', 'gateway_server_address'],
      {
        withStatus: !isDeletedTab,
        isSearch: tab === ALL_TAB || isDeletedTab || query.length > 0,
      },
    )
  }, [])

  return (
    <FetchTable
      entity="gateways"
      defaultOrder="-created_at"
      addMessage={sharedMessages.registerGateway}
      headers={headers}
      getItemsAction={getGateways}
      baseDataSelector={baseDataSelector}
      tableTitle={<Message content={sharedMessages.gateways} />}
      searchable
      clickable={!isDeletedTab}
      tabs={isAdmin ? tabs : []}
    />
  )
}

export default GatewaysTable
