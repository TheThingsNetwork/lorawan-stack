// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { createSelector } from 'reselect'
import { useDispatch } from 'react-redux'

import Tabs from '@ttn-lw/components/tabs'
import Link from '@ttn-lw/components/link'
import Form from '@ttn-lw/components/form'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import toast from '@ttn-lw/components/toast'
import PageTitle from '@ttn-lw/components/page-title'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import {
  CONNECTION_TYPES,
  getFormTypeMessage,
} from '@console/containers/gateway-the-things-station/utils'
import ShowProfilesSelect from '@console/containers/gateway-the-things-station/show-profiles-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { deleteConnectionProfile, getConnectionProfilesList } from '@console/store/actions/gateways'

import {
  selectConnectionProfiles,
  selectConnectionProfilesTotalCount,
} from '@console/store/selectors/gateways'

import m from './messages'

const GatewayConnectionProfilesOverview = () => {
  const { gtwId, type } = useParams()
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const tabs = [
    {
      title: m.wifiProfiles,
      name: CONNECTION_TYPES.WIFI,
      link: `/gateways/${gtwId}/the-things-station/connection-profiles/${CONNECTION_TYPES.WIFI}?shared=`,
    },
    {
      title: m.ethernetProfiles,
      name: CONNECTION_TYPES.ETHERNET,
      link: `/gateways/${gtwId}/the-things-station/connection-profiles/${CONNECTION_TYPES.ETHERNET}`,
    },
  ]

  const onAddProfile = useCallback(
    shared => {
      navigate(`add?shared=${shared}`)
    },
    [navigate],
  )

  const handleEdit = React.useCallback(
    (profileId, shared) => {
      navigate(`edit/${profileId}?shared=${shared}`)
    },
    [navigate],
  )

  const handleDelete = React.useCallback(
    async id => {
      try {
        await dispatch(attachPromise(deleteConnectionProfile(id)))
        toast({
          title: id,
          message: m.deleteSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (err) {
        toast({
          title: id,
          message: m.deleteFail,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch],
  )

  const headers = React.useMemo(
    () => [
      {
        name: 'ids.profile_id',
        displayName: m.profileId,
        width: 25,
        sortable: true,
      },
      {
        name: 'access_point',
        displayName: m.accessPoint,
        width: 25,
        sortable: true,
      },
      {
        name: 'created_at',
        displayName: sharedMessages.created,
        width: 25,
      },
      {
        name: 'actions',
        displayName: sharedMessages.actions,
        width: 25,
        getValue: row => ({
          id: row.id,
          name: row.access_point,
          edit: handleEdit.bind(null, row.id),
          delete: handleDelete.bind(null, row.id),
        }),
        render: details => (
          <ButtonGroup align="end">
            <Button icon="edit" onClick={details.edit} />
            <DeleteModalButton
              entityId={details.id}
              entityName={details.name}
              onApprove={details.delete}
            />
          </ButtonGroup>
        ),
      },
    ],
    [handleDelete, handleEdit],
  )

  const baseDataSelector = createSelector(
    [selectConnectionProfiles, selectConnectionProfilesTotalCount],
    (connectionProfiles, totalCount) => ({
      connectionProfiles,
      totalCount,
      mayAdd: false,
      mayLink: false,
    }),
  )

  const getItems = React.useCallback(() => getConnectionProfilesList({ type }), [type])

  return (
    <>
      <PageTitle title={m.theThingsStationConnectionProfiles} />
      <Tabs tabs={tabs} divider />
      <Message
        className="d-block mt-cs-l mb-cs-l"
        content={m.information}
        values={{
          link: txt => (
            <Link secondary to="#">
              {txt}
            </Link>
          ),
          br: () => <br />,
        }}
      />
      <Form
        initialValues={{
          shared: false,
        }}
      >
        {({ values }) => (
          <>
            <div className="d-flex j-between al-end">
              <ShowProfilesSelect name="shared" />
              <Button
                className="mb-cs-m"
                primary
                onClick={() => onAddProfile(values.shared)}
                message={getFormTypeMessage(type)}
                icon="add"
              />
            </div>
            <FetchTable
              entity="connectionProfiles"
              defaultOrder="-created_at"
              headers={headers}
              getItemsAction={getItems}
              baseDataSelector={baseDataSelector}
              filtersClassName="d-none"
            />
          </>
        )}
      </Form>
    </>
  )
}

export default GatewayConnectionProfilesOverview
