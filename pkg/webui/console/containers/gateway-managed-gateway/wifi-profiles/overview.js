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
import { defineMessages } from 'react-intl'

import Link from '@ttn-lw/components/link'
import Form from '@ttn-lw/components/form'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import toast from '@ttn-lw/components/toast'
import PageTitle from '@ttn-lw/components/page-title'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import ShowProfilesSelect from '@console/containers/gateway-managed-gateway/shared/show-profiles-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  deleteConnectionProfile,
  getConnectionProfilesList,
} from '@console/store/actions/connection-profiles'

import {
  selectConnectionProfiles,
  selectConnectionProfilesTotalCount,
} from '@console/store/selectors/connection-profiles'

const m = defineMessages({
  information:
    'Connection profiles are setup to allow for multiple gateways to connect via the same settings. You can use this view to manage all your profiles or create new ones, after which you can assign them to your gateway.<br></br> <link>Learn more about gateway network connection profiles.</link>',
  profileId: 'Profile ID',
  accessPoint: 'Access point',
  deleteSuccess: 'Connection profile deleted',
  deleteFail: 'There was an error and the connection profile could not be deleted',
})

const GatewayWifiProfilesOverview = () => {
  const { type } = useParams()
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const onAddProfile = useCallback(
    profileOf => {
      navigate(`add?profileOf=${profileOf}`)
    },
    [navigate],
  )

  const handleEdit = React.useCallback(
    (profileId, profileOf) => {
      navigate(`edit/${profileId}?profileOf=${profileOf}`)
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
      <PageTitle title={sharedMessages.wifiProfiles} />
      <Message
        className="d-block mb-cs-l"
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
          profileOf: 'yourself',
        }}
      >
        {({ values }) => (
          <>
            <div className="d-flex j-between al-end gap-cs-m">
              <ShowProfilesSelect name="profileOf" />
              <Button
                className="mb-cs-m"
                primary
                onClick={() => onAddProfile(values.profileOf)}
                message={sharedMessages.addWifiProfile}
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

export default GatewayWifiProfilesOverview
