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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Spinner from '@ttn-lw/components/spinner'
import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'

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

import {
  selectAccessPoints,
  selectSelectedWifiConnectionProfile,
} from '@console/store/selectors/connection-profiles'

const m = defineMessages({
  updateWifiProfile: 'Update WiFi profile',
  createSuccess: 'Connection profile created',
  createFailure: 'There was an error and the connection profile could not be created',
  updateSuccess: 'Connection profile updated',
  updateFailure: 'There was an error updating this connection profile',
})

const GatewayWifiProfilesForm = () => {
  const [error, setError] = useState(undefined)
  const accessPoints = useSelector(selectAccessPoints)
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

  useEffect(() => {
    if (isEdit) {
      dispatch(attachPromise(getConnectionProfile(entityId, profileId, CONNECTION_TYPES.WIFI)))
        .then(res => {
          formRef.current.setValues(values => ({
            ...values,
            ...res.data,
            _default_network_interface: !Boolean(res.data.network_interface_addresses),
          }))
        })
        .catch(() => {
          navigate(baseUrl, { replace: true })
        })
    }
  }, [baseUrl, dispatch, entityId, isEdit, navigate, profileId, searchParams])

  useEffect(() => {
    if (isEdit && !isLoadingAccessPoints && Boolean(formRef.current)) {
      formRef.current.setValues(values => {
        const accessPoint = accessPoints.find(ap => ap.ssid === values.ssid)

        return {
          ...values,
          _access_point: {
            ...accessPoint,
            is_password_set: true,
            type: accessPoint ? 'all' : 'other',
          },
        }
      })
    }
  }, [accessPoints, isEdit, isLoadingAccessPoints, profileId])

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

          await dispatch(
            attachPromise(
              updateConnectionProfile(entityId, profileId, CONNECTION_TYPES.WIFI, profileDiff),
            ),
          )
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
    <>
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
          initialValues={initialWifiProfile}
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
    </>
  )
}

export default GatewayWifiProfilesForm
