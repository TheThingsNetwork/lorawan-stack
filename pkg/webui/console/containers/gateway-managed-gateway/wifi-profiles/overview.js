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

import React, { useCallback, useMemo, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { createSelector } from 'reselect'
import { useDispatch, useSelector } from 'react-redux'
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
import RequireRequest from '@ttn-lw/lib/components/require-request'

import ShowProfilesSelect from '@console/containers/gateway-managed-gateway/shared/show-profiles-select'
import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/shared/utils'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

import { checkFromState, mayViewOrEditGatewayCollaborators } from '@console/lib/feature-checks'

import {
  deleteConnectionProfile,
  getConnectionProfilesList,
} from '@console/store/actions/connection-profiles'

import { selectConnectionProfilesByType } from '@console/store/selectors/connection-profiles'

const m = defineMessages({
  information:
    'Connection profiles are setup to allow for multiple gateways to connect via the same settings. You can use this view to manage all your profiles or create new ones, after which you can assign them to your gateway.<br></br> <link>Learn more about gateway network connection profiles.</link>',
  profileId: 'Profile ID',
  accessPoint: 'Access point',
  deleteProfile: 'Delete profile',
  deleteSuccess: 'Connection profile deleted',
  deleteFail: 'There was an error and the connection profile could not be deleted',
})

const GatewayWifiProfilesOverview = () => {
  const { gtwId } = useParams()
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const formRef = useRef(null)
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditGatewayCollaborators, state),
  )

  const onAddProfile = useCallback(() => {
    navigate(`add?profileOf=${formRef.current.values._profile_of}`)
  }, [navigate])

  const handleEdit = React.useCallback(
    (profileId, profileOf) => {
      navigate(`edit/${profileId}?profileOf=${profileOf}`)
    },
    [navigate],
  )

  const handleDelete = React.useCallback(
    async (profileId, name, profileOf) => {
      try {
        await dispatch(
          attachPromise(deleteConnectionProfile(profileOf, profileId, CONNECTION_TYPES.WIFI)),
        )
        toast({
          title: name,
          message: m.deleteSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (err) {
        toast({
          title: name,
          message: m.deleteFail,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch],
  )

  const headers = useMemo(
    () => [
      {
        name: 'profile_name',
        displayName: sharedMessages.profileName,
      },
      {
        name: 'ssid',
        displayName: m.accessPoint,
      },
      {
        name: 'actions',
        displayName: sharedMessages.actions,
        getValue: row => ({
          id: row.profile_id,
          name: row.profile_name,
          edit: handleEdit.bind(null, row.profile_id, formRef.current.values._profile_of),
          delete: handleDelete.bind(
            null,
            row.profile_id,
            row.profile_name,
            formRef.current.values._profile_of,
          ),
        }),
        render: details => (
          <ButtonGroup align="end">
            <Button icon="edit" onClick={details.edit} />
            <DeleteModalButton
              onlyIcon
              message={m.deleteProfile}
              defaultMessage=""
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
    state => selectConnectionProfilesByType(state, CONNECTION_TYPES.WIFI),
    connectionProfiles => ({
      connectionProfiles,
      totalCount: connectionProfiles?.length ?? 0,
      mayAdd: false,
      mayLink: false,
    }),
  )

  const getItemsAction = useCallback(
    () =>
      getConnectionProfilesList({
        entityId: formRef.current.values._profile_of,
        type: CONNECTION_TYPES.WIFI,
      }),
    [],
  )

  const handleChangeProfile = useCallback(
    async value => {
      await dispatch(
        attachPromise(
          getConnectionProfilesList({
            entityId: value,
            type: CONNECTION_TYPES.WIFI,
          }),
        ),
      )
    },
    [dispatch],
  )

  const loadData = useCallback(
    async dispatch => {
      if (mayViewCollaborators) {
        dispatch(getCollaboratorsList('gateway', gtwId))
      }
    },
    [gtwId, mayViewCollaborators],
  )

  return (
    <RequireRequest requestAction={loadData}>
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
          _profile_of: '',
        }}
        formikRef={formRef}
      >
        {({ values }) => (
          <>
            <div className="d-flex j-between al-end gap-cs-m">
              <ShowProfilesSelect name="_profile_of" onChange={handleChangeProfile} />
              <Button
                className="mb-cs-m"
                primary
                onClick={onAddProfile}
                message={sharedMessages.addWifiProfile}
                icon="add"
              />
            </div>
            {Boolean(values._profile_of) && (
              <FetchTable
                entity="connectionProfiles"
                defaultOrder="ssid"
                headers={headers}
                getItemsAction={getItemsAction}
                baseDataSelector={baseDataSelector}
                filtersClassName="d-none"
              />
            )}
          </>
        )}
      </Form>
    </RequireRequest>
  )
}

export default GatewayWifiProfilesOverview
