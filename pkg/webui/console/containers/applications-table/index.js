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
import { defineMessages } from 'react-intl'
import { bindActionCreators } from 'redux'

import toast from '@ttn-lw/components/toast'
import Button from '@ttn-lw/components/button'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import DeleteModalButton from '@console/components/delete-modal-button'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayCreateApplications } from '@console/lib/feature-checks'

import {
  deleteApplication,
  restoreApplication,
  getApplicationsList,
} from '@console/store/actions/applications'

import { selectUserIsAdmin } from '@console/store/selectors/user'
import {
  selectApplications,
  selectApplicationsTotalCount,
  selectApplicationsFetching,
  selectApplicationsError,
} from '@console/store/selectors/applications'

const m = defineMessages({
  ownedTabTitle: 'Owned applications',
  restoreSuccess: 'Application restored',
  restoreFail: 'There was an error and application could not be restored',
  purgeSuccess: 'Application purged',
  purgeFail: 'There was an error and the application could not be purged',
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

const ApplicationsTable = props => {
  const { isAdmin, restoreApplication, purgeApplication, ...rest } = props

  const [tab, setTab] = React.useState(OWNED_TAB)
  const isDeletedTab = tab === DELETED_TAB

  const handleRestore = React.useCallback(
    async id => {
      try {
        await restoreApplication(id)
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
    [restoreApplication],
  )

  const handlePurge = React.useCallback(
    async id => {
      try {
        await purgeApplication(id)
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
    [purgeApplication],
  )

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'ids.application_id',
        displayName: sharedMessages.id,
        width: 25,
        sortable: true,
        sortKey: 'application_id',
      },
      {
        name: 'name',
        displayName: sharedMessages.name,
        width: 25,
        sortable: true,
      },
    ]

    if (tab === DELETED_TAB) {
      baseHeaders.push({
        name: 'actions',
        displayName: sharedMessages.actions,
        width: 50,
        getValue: row => ({
          id: row.ids.application_id,
          name: row.name,
          restore: handleRestore.bind(null, row.ids.application_id),
          purge: handlePurge.bind(null, row.ids.application_id),
        }),
        render: details => (
          <>
            <Button message={sharedMessages.restore} secondary onClick={details.restore} />
            <DeleteModalButton
              entityId={details.id}
              entityName={name}
              message={sharedMessages.purge}
              onApprove={details.purge}
              onlyPurge
            />
          </>
        ),
      })
    } else {
      baseHeaders.push({
        name: 'description',
        displayName: sharedMessages.description,
        width: 50,
      })
    }

    return baseHeaders
  }, [handlePurge, handleRestore, tab])

  const baseDataSelector = React.useCallback(
    state => ({
      applications: selectApplications(state),
      totalCount: selectApplicationsTotalCount(state),
      fetching: selectApplicationsFetching(state),
      error: selectApplicationsError(state),
      mayAdd: checkFromState(mayCreateApplications, state),
    }),
    [],
  )

  const getApplications = React.useCallback(filters => {
    const { tab, query } = filters
    const isDeletedTab = tab === DELETED_TAB

    setTab(tab)

    return getApplicationsList({ ...filters, deleted: isDeletedTab }, ['name', 'description'], {
      isSearch: tab === ALL_TAB || isDeletedTab || query.length > 0,
    })
  }, [])

  return (
    <FetchTable
      entity="applications"
      headers={headers}
      addMessage={sharedMessages.addApplication}
      tableTitle={<Message content={sharedMessages.applications} />}
      getItemsAction={getApplications}
      baseDataSelector={baseDataSelector}
      searchable
      clickable={!isDeletedTab}
      tabs={isAdmin ? tabs : []}
      {...rest}
    />
  )
}

ApplicationsTable.propTypes = {
  isAdmin: PropTypes.bool.isRequired,
  purgeApplication: PropTypes.func.isRequired,
  restoreApplication: PropTypes.func.isRequired,
}

export default connect(
  state => ({
    isAdmin: selectUserIsAdmin(state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        purgeApplication: attachPromise(deleteApplication),
        restoreApplication: attachPromise(restoreApplication),
      },
      dispatch,
    ),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    purgeApplication: id => dispatchProps.purgeApplication(id, { purge: true }),
    restoreApplication: id => dispatchProps.restoreApplication(id),
  }),
)(ApplicationsTable)
