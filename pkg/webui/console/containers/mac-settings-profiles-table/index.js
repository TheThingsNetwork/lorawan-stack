// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { useDispatch } from 'react-redux'
import { useNavigate, useParams } from 'react-router-dom'

import ButtonGroup from '@ttn-lw/components/button/group'
import Button from '@ttn-lw/components/button'
import PageTitle from '@ttn-lw/components/page-title'
import { IconPencil } from '@ttn-lw/components/icon'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import toast from '@ttn-lw/components/toast'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { deleteMacSettingsProfile } from '@console/store/actions/mac-settings-profiles'

const m = defineMessages({
  information:
    'MAC settings profiles are used to set MAC settings of end devices more conveniently.',
  deleteSuccess: 'MAC settings profile deleted',
  deleteFail: 'There was an error and the MAC settings profile could not be deleted',
  deleteModalMessage: 'The profile will not be applicable to end devices anymore.',
})

const MacSettingsProfilesTable = props => {
  const navigate = useNavigate()
  const dispatch = useDispatch()
  const { appId } = useParams()
  const { baseDataSelector, getItemsAction } = props

  const handleEdit = useCallback(
    profileId => {
      navigate(`/applications/${appId}/mac-settings-profiles/${profileId}`)
    },
    [navigate, appId],
  )

  const handleDelete = useCallback(
    async id => {
      try {
        await dispatch(attachPromise(deleteMacSettingsProfile(appId, id)))
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
    [appId, dispatch],
  )

  const headers = [
    {
      name: 'ids.profile_id',
      displayName: sharedMessages.nameAndId,
      getValue: row => ({
        id: row.ids.profile_id,
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
      sortKey: 'ids.profile_id',
    },
    {
      name: 'actions',
      displayName: sharedMessages.actions,
      width: 50,
      getValue: row => ({
        id: row.ids.profile_id,
        name: row.name,
        edit: handleEdit.bind(null, row.ids.profile_id),
        delete: handleDelete.bind(null, row.ids.profile_id),
      }),
      render: details => (
        <ButtonGroup align="end">
          <Button
            message={sharedMessages.edit}
            icon={IconPencil}
            onClick={details.edit}
            secondary
            small
          />
          <DeleteModalButton
            message={sharedMessages.delete}
            entityId={details.id}
            entityName={details.name}
            onApprove={details.delete}
            defaultMessage={m.deleteModalMessage}
            small
          />
        </ButtonGroup>
      ),
    },
  ]

  return (
    <>
      <PageTitle title={sharedMessages.macSettingsProfiles} className="mb-0" />
      <Message content={m.information} />
      <FetchTable
        entity="mac_settings_profiles"
        defaultOrder="-ids.profile_id"
        headers={headers}
        addMessage={'Create MAC settings profile'}
        baseDataSelector={baseDataSelector}
        getItemsAction={getItemsAction}
        clickable={false}
      />
    </>
  )
}

MacSettingsProfilesTable.propTypes = {
  baseDataSelector: PropTypes.func.isRequired,
  getItemsAction: PropTypes.func.isRequired,
}

export default MacSettingsProfilesTable
