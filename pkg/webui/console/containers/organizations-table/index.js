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
import { bindActionCreators } from 'redux'
import { defineMessages } from 'react-intl'

import toast from '@ttn-lw/components/toast'
import Button from '@ttn-lw/components/button'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import ButtonGroup from '@ttn-lw/components/button/group'
import Spinner from '@ttn-lw/components/spinner'

import FetchTable from '@ttn-lw/containers/fetch-table'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { getOrganizationId } from '@ttn-lw/lib/selectors/id'

import { checkFromState, mayCreateOrganizations } from '@console/lib/feature-checks'

import {
  getOrganizationsList,
  deleteOrganization,
  restoreOrganization,
} from '@console/store/actions/organizations'

import { selectUserIsAdmin } from '@console/store/selectors/logout'
import {
  selectOrganizations,
  selectOrganizationsTotalCount,
  selectOrganizationsFetching,
  selectOrganizationCollaboratorCount,
} from '@console/store/selectors/organizations'

const m = defineMessages({
  restoreSuccess: 'Organization restored',
  restoreFail: 'There was an error and the organization could not be restored',
  purgeSuccess: 'Organization purged',
  purgeFail: 'There was an error and the organization could not be purged',
})

const OWNED_TAB = 'owned'
const ALL_TAB = 'all'
const DELETED_TAB = 'deleted'
const tabs = [
  {
    title: sharedMessages.organizations,
    name: OWNED_TAB,
  },
  {
    title: sharedMessages.allAdmin,
    name: ALL_TAB,
  },
  { title: sharedMessages.deleted, name: DELETED_TAB },
]

const OrganizationsTable = props => {
  const { pageSize, isAdmin, purgeOrganization, restoreOrganization } = props

  const [tab, setTab] = React.useState(OWNED_TAB)
  const isDeletedTab = tab === DELETED_TAB

  const handleRestore = React.useCallback(
    async id => {
      try {
        await restoreOrganization(id)
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
    [restoreOrganization],
  )

  const handlePurge = React.useCallback(
    async id => {
      try {
        await purgeOrganization(id)
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
    [purgeOrganization],
  )

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'ids.organization_id',
        displayName: sharedMessages.id,
        width: 20,
        sortable: true,
        sortKey: 'organization_id',
      },
      {
        name: 'name',
        displayName: sharedMessages.name,
        width: 20,
        sortable: true,
      },
      {
        name: 'description',
        displayName: sharedMessages.description,
        width: 30,
      },
    ]

    if (isDeletedTab) {
      baseHeaders.push({
        name: 'actions',
        displayName: sharedMessages.actions,
        width: 30,
        getValue: row => ({
          id: row.ids.organization_id,
          name: row.name,
          restore: handleRestore.bind(null, row.ids.organization_id),
          purge: handlePurge.bind(null, row.ids.organization_id),
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
      baseHeaders.push(
        {
          name: '_collaboratorCount',
          displayName: sharedMessages.collaborators,
          width: 10,
          align: 'center',
          render: deviceCount =>
            typeof deviceCount !== 'number' ? (
              <Spinner micro center inline after={100} className="c-subtle-gray" />
            ) : (
              deviceCount
            ),
        },
        {
          name: 'created_at',
          displayName: sharedMessages.createdAt,
          sortable: true,
          width: 20,
          render: date => <DateTime.Relative value={date} />,
        },
      )
    }

    return baseHeaders
  }, [handlePurge, handleRestore, isDeletedTab])

  const baseDataSelector = React.useCallback(state => {
    const organizations = selectOrganizations(state)
    const decoratedOrganizations = []

    for (const org of organizations) {
      decoratedOrganizations.push({
        ...org,
        _collaboratorCount: selectOrganizationCollaboratorCount(state, getOrganizationId(org)),
      })
    }

    return {
      organizations: decoratedOrganizations,
      totalCount: selectOrganizationsTotalCount(state),
      fetching: selectOrganizationsFetching(state),
      mayAdd: checkFromState(mayCreateOrganizations, state),
    }
  }, [])

  const getOrganizations = React.useCallback(filters => {
    const { tab, query } = filters
    const isDeletedTab = tab === DELETED_TAB

    setTab(tab)

    return getOrganizationsList({ ...filters, deleted: isDeletedTab }, ['name', 'description'], {
      isSearch: tab === ALL_TAB || isDeletedTab || query.length > 0,
      withCollaboratorCount: true,
    })
  }, [])

  return (
    <FetchTable
      entity="organizations"
      defaultOrder="-created_at"
      headers={headers}
      addMessage={sharedMessages.addOrganization}
      tableTitle={<Message content={sharedMessages.organizations} />}
      getItemsAction={getOrganizations}
      baseDataSelector={baseDataSelector}
      pageSize={pageSize}
      searchable
      clickable={!isDeletedTab}
      tabs={isAdmin ? tabs : []}
    />
  )
}

OrganizationsTable.propTypes = {
  isAdmin: PropTypes.bool.isRequired,
  pageSize: PropTypes.number.isRequired,
  purgeOrganization: PropTypes.func.isRequired,
  restoreOrganization: PropTypes.func.isRequired,
}

export default connect(
  state => ({
    isAdmin: selectUserIsAdmin(state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        purgeOrganization: attachPromise(deleteOrganization),
        restoreOrganization: attachPromise(restoreOrganization),
      },
      dispatch,
    ),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    purgeOrganization: id => dispatchProps.purgeOrganization(id, { purge: true }),
    restoreOrganization: id => dispatchProps.restoreOrganization(id),
  }),
)(OrganizationsTable)
