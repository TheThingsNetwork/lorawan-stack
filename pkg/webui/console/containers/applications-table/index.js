// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages, FormattedNumber } from 'react-intl'
import { bindActionCreators } from 'redux'
import { createSelector } from 'reselect'

import Icon, { IconHelp } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Status from '@ttn-lw/components/status'
import toast from '@ttn-lw/components/toast'
import Spinner from '@ttn-lw/components/spinner'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayCreateApplications } from '@console/lib/feature-checks'
import { isOtherClusterApp } from '@console/lib/application-utils'

import {
  deleteApplication,
  restoreApplication,
  getApplicationsList,
} from '@console/store/actions/applications'

import { selectUserIsAdmin } from '@console/store/selectors/user'
import {
  selectApplicationsTotalCount,
  selectApplicationsWithDeviceCounts,
} from '@console/store/selectors/applications'

const m = defineMessages({
  restoreSuccess: 'Application restored',
  restoreFail: 'There was an error and application could not be restored',
  purgeSuccess: 'Application purged',
  purgeFail: 'There was an error and the application could not be purged',
  otherClusterTooltip:
    'This application is registered on a different cluster (`{host}`). To access this application, use the Console of the cluster that this application was registered on.',
})

const OWNED_TAB = 'owned'
const ALL_TAB = 'all'
const DELETED_TAB = 'deleted'
const tabs = [
  {
    title: sharedMessages.ownedByMe,
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
  // TODO: Add lastSeen column.

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
        displayName: sharedMessages.nameAndId,
        getValue: row => ({
          id: row.ids.application_id,
          name: row.name,
        }),
        render: ({ name, id }) =>
          Boolean(name) ? (
            <>
              <span className="mt-0 mb-cs-xxs p-0 fw-bold d-block">{name}</span>
              <span className="c-text-neutral-light d-block">{id}</span>
            </>
          ) : (
            <span className="mt-0 p-0 fw-bold d-block">{id}</span>
          ),
        sortable: true,
        sortKey: 'application_id',
      },
    ]

    if (tab === DELETED_TAB) {
      baseHeaders.push({
        name: 'actions',
        displayName: sharedMessages.actions,
        width: '13rem',
        getValue: row => ({
          id: row.ids.application_id,
          name: row.name,
          restore: handleRestore.bind(null, row.ids.application_id),
          purge: handlePurge.bind(null, row.ids.application_id),
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
      baseHeaders.push(
        {
          name: 'status',
          displayName: '',
          width: '10rem',
          render: status => {
            if (status.otherCluster) {
              const host = status.host
              return (
                <DocTooltip
                  docPath="/getting-started/cloud-hosted"
                  content={
                    <Message content={m.otherClusterTooltip} values={{ host }} convertBackticks />
                  }
                  placement="top-end"
                >
                  <Status status="unknown" label={sharedMessages.otherCluster}>
                    <Icon
                      icon={IconHelp}
                      textPaddedLeft
                      small
                      nudgeUp
                      className="c-text-neutral-light"
                    />
                  </Status>
                </DocTooltip>
              )
            }

            return null
          },
        },
        {
          name: '_devices',
          width: '7rem',
          displayName: sharedMessages.devicesShort,
          align: 'center',
          render: deviceCount =>
            typeof deviceCount !== 'number' ? (
              <Spinner micro center inline after={100} className="c-icon" />
            ) : (
              <FormattedNumber value={deviceCount} />
            ),
        },
        {
          name: 'created_at',
          width: '8rem',
          displayName: sharedMessages.created,
          sortable: true,
          render: date => <DateTime.Relative value={date} />,
        },
      )
    }

    return baseHeaders
  }, [handlePurge, handleRestore, tab])

  const selectDecoratedApplications = createSelector(
    selectApplicationsWithDeviceCounts,
    applications =>
      applications.map(app => ({
        ...app,
        status: {
          otherCluster: isOtherClusterApp(app),
          host:
            app.application_server_address || app.network_server_address || app.join_server_address,
        },
        _meta: {
          clickable: !isOtherClusterApp(app),
        },
      })),
  )

  const baseDataSelector = createSelector(
    selectDecoratedApplications,
    selectApplicationsTotalCount,
    state => checkFromState(mayCreateApplications, state),
    (applications, totalCount, mayAdd) => ({
      applications,
      totalCount,
      mayAdd,
    }),
  )

  const getApplications = React.useCallback(filters => {
    const { tab, query } = filters
    const isDeletedTab = tab === DELETED_TAB

    setTab(tab)

    return getApplicationsList(
      { ...filters, deleted: isDeletedTab },
      [
        'name',
        'description',
        'network_server_address',
        'application_server_address',
        'join_server_address',
      ],
      {
        isSearch: tab === ALL_TAB || isDeletedTab || query.length > 0,
        withDeviceCount: true,
      },
    )
  }, [])

  return (
    <FetchTable
      entity="applications"
      defaultOrder="-created_at"
      headers={headers}
      addMessage={sharedMessages.addApplication}
      tableTitle={<Message content={sharedMessages.applications} />}
      getItemsAction={getApplications}
      baseDataSelector={baseDataSelector}
      searchable
      clickable={!isDeletedTab}
      tabs={isAdmin ? tabs : []}
      searchPlaceholderMessage={sharedMessages.searchApplications}
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
