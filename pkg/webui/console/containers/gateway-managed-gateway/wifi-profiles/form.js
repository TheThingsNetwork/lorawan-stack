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

import React, { useCallback, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { isEmpty } from 'lodash'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Spinner from '@ttn-lw/components/spinner'
import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import {
  CONNECTION_TYPES,
  initialWifiProfile,
  normalizeWifiProfile,
} from '@console/containers/gateway-managed-gateway/shared/utils'
import GatewayWifiProfilesFormFields from '@console/containers/gateway-managed-gateway/shared/wifi-profiles-form-fields'
import { wifiValidationSchema } from '@console/containers/gateway-managed-gateway/shared/validation-schema'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import diff from '@ttn-lw/lib/diff'

import {
  createConnectionProfile,
  GET_ACCESS_POINTS_BASE,
  GET_CONNECTION_PROFILE_BASE,
  getConnectionProfile,
  updateConnectionProfile,
} from '@console/store/actions/connection-profiles'

import { selectSelectedWifiConnectionProfile } from '@console/store/selectors/connection-profiles'

const m = defineMessages({
  updateWifiProfile: 'Update WiFi profile',
  createSuccess: 'WiFi profile created',
  createFailure: 'There was an error and the WiFi profile could not be created',
  updateSuccess: 'WiFi profile updated',
  updateFailure: 'There was an error updating this WiFi profile',
})

const GatewayWifiProfilesForm = () => {
  const [error, setError] = useState(undefined)
  const [initialValues, setInitialValues] = useState(initialWifiProfile)
  const { gtwId, profileId } = useParams()
  const isLoadingAccessPoints = useSelector(state =>
    selectFetchingEntry(state, GET_ACCESS_POINTS_BASE),
  )
  const isLoadingProfile = useSelector(state =>
    selectFetchingEntry(state, GET_CONNECTION_PROFILE_BASE),
  )
  const selectedProfile = useSelector(selectSelectedWifiConnectionProfile)

  const formRef = useRef(null)
  const dispatch = useDispatch()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  const isEdit = Boolean(profileId)

  const entityId = searchParams.get('profileOf')

  const baseUrl = `/gateways/${gtwId}/managed-gateway/wifi-profiles`

  useBreadcrumbs(
    'gtws.single.managed-gateway.wifi-profiles.form',
    <Breadcrumb
      path={isEdit ? `${baseUrl}/edit/${profileId}` : `${baseUrl}/add`}
      content={isEdit ? m.updateWifiProfile : sharedMessages.addWifiProfile}
    />,
  )

  const loadData = useCallback(
    async dispatch => {
      if (isEdit) {
        try {
          const { data: profile } = await dispatch(
            attachPromise(getConnectionProfile(entityId, profileId, CONNECTION_TYPES.WIFI)),
          )

          setInitialValues(oldValues => ({
            ...oldValues,
            ...profile,
            _default_network_interface: !Boolean(profile.network_interface_addresses),
          }))
        } catch (e) {
          // Navigate(baseUrl, { replace: true })
        }
      }
    },
    [entityId, isEdit, profileId],
  )

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      setError(undefined)
      const profile = normalizeWifiProfile(values)
      try {
        if (!isEdit) {
          await dispatch(
            attachPromise(createConnectionProfile(entityId, CONNECTION_TYPES.WIFI, profile)),
          )
          navigate(baseUrl, { replace: true })
        } else {
          const profileDiff = diff(selectedProfile, profile)
          if (!isEmpty(profileDiff)) {
            await dispatch(
              attachPromise(
                updateConnectionProfile(entityId, profileId, CONNECTION_TYPES.WIFI, profileDiff),
              ),
            )
          }
        }

        toast({
          title: profile.profile_name,
          message: !isEdit ? m.createSuccess : m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
        toast({
          title: profile.profile_name,
          message: !isEdit ? m.createFailure : m.updateFailure,
          type: toast.types.ERROR,
        })
      }
    },
    [baseUrl, dispatch, entityId, isEdit, navigate, profileId, selectedProfile],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <PageTitle title={isEdit ? m.updateWifiProfile : sharedMessages.addWifiProfile} />
      {isLoadingProfile ? (
        <div className="pos-relative mt-cs-xl mb-cs-xl">
          <Spinner center>
            <Message content={sharedMessages.fetching} />
          </Spinner>
        </div>
      ) : (
        <Form
          error={error}
          onSubmit={handleSubmit}
          initialValues={initialValues}
          validationSchema={wifiValidationSchema}
          formikRef={formRef}
        >
          <GatewayWifiProfilesFormFields />

          <SubmitBar>
            <Form.Submit
              component={SubmitButton}
              message={sharedMessages.saveChanges}
              disabled={isLoadingAccessPoints}
            />
          </SubmitBar>
        </Form>
      )}
    </RequireRequest>
  )
}

export default GatewayWifiProfilesForm
