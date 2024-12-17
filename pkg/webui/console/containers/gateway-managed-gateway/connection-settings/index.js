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

import React, { useCallback, useState } from 'react'
import { defineMessages } from 'react-intl'
import { useParams, useSearchParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import { isEmpty, isEqual } from 'lodash'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Notification from '@ttn-lw/components/notification'
import Form from '@ttn-lw/components/form'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import validationSchema from '@console/containers/gateway-managed-gateway/connection-settings/validation-schema'
import {
  CONNECTION_TYPES,
  initialWifiProfile,
  initialEthernetProfile,
  normalizeWifiProfile,
  revertEthernetProfile,
} from '@console/containers/gateway-managed-gateway/shared/utils'
import WifiSettingsFormFields from '@console/containers/gateway-managed-gateway/connection-settings/wifi-settings-form-fields'
import EthernetSettingsFormFields from '@console/containers/gateway-managed-gateway/connection-settings/ethernet-settings-form-fields'
import ManagedGatewayConnections from '@console/containers/gateway-managed-gateway/connection-settings/connections'
import useConnectionsData from '@console/containers/gateway-managed-gateway/connection-settings/use-connections-data'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import diff from '@ttn-lw/lib/diff'
import { isNotFoundError } from '@ttn-lw/lib/errors/utils'

import { checkFromState, mayViewOrEditGatewayCollaborators } from '@console/lib/feature-checks'

import {
  createConnectionProfile,
  GET_ACCESS_POINTS_BASE,
  getConnectionProfile,
  updateConnectionProfile,
} from '@console/store/actions/connection-profiles'
import { updateManagedGateway } from '@console/store/actions/gateways'

import {
  selectSelectedGateway,
  selectSelectedManagedGateway,
} from '@console/store/selectors/gateways'
import { selectUserId } from '@account/store/selectors/user'

const m = defineMessages({
  firstNotification:
    'You have just claimed a managed gateway. To connect it to WiFi or ethernet you can configure those connections here. The preprovisioned cellular backhaul typically connects automatically.',
  updateSuccess: 'Connection settings updated',
  updateFailure: 'There was an error updating these connection settings',
})

const GatewayConnectionSettings = () => {
  const { gtwId } = useParams()
  const selectedGateway = useSelector(selectSelectedGateway)
  const [searchParams] = useSearchParams()
  const isFirstClaim = Boolean(searchParams.get('claimed'))
  const [error, setError] = useState(undefined)
  const isLoading = useSelector(state => selectFetchingEntry(state, GET_ACCESS_POINTS_BASE))
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditGatewayCollaborators, state),
  )
  const selectedManagedGateway = useSelector(selectSelectedManagedGateway)
  const userId = useSelector(selectUserId)
  const dispatch = useDispatch()
  const [nonSharedWifiProfileId, setNonSharedWifiProfileId] = useState(null)
  const [saveFormClicked, setSaveFormClicked] = useState(false)

  const connectionsData = useConnectionsData()

  const [initialValues, setInitialValues] = useState({
    wifi_profile: { ...initialWifiProfile },
    ethernet_profile: { ...initialEthernetProfile },
  })

  const hasWifiProfileSet = Boolean(selectedManagedGateway.wifi_profile_id)
  const hasEthernetProfileSet = Boolean(selectedManagedGateway.ethernet_profile_id)

  useBreadcrumbs(
    'gtws.single.managed-gateway.connection-settings',
    <Breadcrumb
      path={`/gateways/${gtwId}/managed-gateway/connection-settings`}
      content={sharedMessages.connectionSettings}
    />,
  )

  const fetchWifiProfile = useCallback(
    async collaborators => {
      let wifiProfile
      let entityId

      if (hasWifiProfileSet) {
        setError(undefined)
        try {
          // If WiFi profile is set, first check if that profile belongs to the user
          wifiProfile = await dispatch(
            attachPromise(
              getConnectionProfile(
                userId,
                selectedManagedGateway.wifi_profile_id,
                CONNECTION_TYPES.WIFI,
              ),
            ),
          )
          entityId = userId
        } catch (e) {
          if (!isNotFoundError(e)) {
            setError(e)
          }
        }
        if (!wifiProfile) {
          const orgCollaborators = collaborators
            .filter(({ ids }) => 'organization_ids' in ids)
            .map(({ ids }) => getCollaboratorId({ ids }))

          setError(undefined)
          // If the WiFi profile doesn't belong to the user, iterate through collaborators until found
          for (const orgId of orgCollaborators) {
            try {
              // eslint-disable-next-line no-await-in-loop
              wifiProfile = await dispatch(
                attachPromise(
                  getConnectionProfile(
                    orgId,
                    selectedManagedGateway.wifi_profile_id,
                    CONNECTION_TYPES.WIFI,
                  ),
                ),
              )
              entityId = orgId
              break
            } catch (e) {
              if (!isNotFoundError(e)) {
                setError(e)
                break
              }
            }
          }
        }
      }
      return { wifiProfile, entityId }
    },
    [dispatch, hasWifiProfileSet, selectedManagedGateway.wifi_profile_id, userId],
  )

  const fetchEthernetProfile = useCallback(async () => {
    let ethernetProfile
    if (hasEthernetProfileSet) {
      setError(undefined)
      try {
        ethernetProfile = await dispatch(
          attachPromise(
            getConnectionProfile(
              undefined,
              selectedManagedGateway.ethernet_profile_id,
              CONNECTION_TYPES.ETHERNET,
            ),
          ),
        )
      } catch (e) {
        if (!isNotFoundError(e)) {
          setError(e)
        }
      }
    }
    return { ethernetProfile }
  }, [dispatch, hasEthernetProfileSet, selectedManagedGateway.ethernet_profile_id])

  const updateInitialWifiProfile = useCallback(
    (values, profile, entityId, isNonSharedProfile) => ({
      ...values.wifi_profile,
      ...(isNonSharedProfile && { ...profile.data }),
      profile_id: isNonSharedProfile
        ? 'non-shared'
        : (selectedManagedGateway.wifi_profile_id ?? ''),
      _override: !Boolean(profile) && hasWifiProfileSet,
      _enable_wifi_connection: Boolean(profile),
      _profile_of: entityId ?? '',
    }),
    [hasWifiProfileSet, selectedManagedGateway.wifi_profile_id],
  )

  const updateInitialEthernetProfile = useCallback(
    (values, profile) => ({
      ...values.ethernet_profile,
      ...revertEthernetProfile(profile?.data ?? {}),
      profile_id: selectedManagedGateway.ethernet_profile_id ?? '',
    }),
    [selectedManagedGateway.ethernet_profile_id],
  )

  const loadData = useCallback(
    async dispatch => {
      let collaborators = []
      if (mayViewCollaborators) {
        const { entities } = await dispatch(attachPromise(getCollaboratorsList('gateway', gtwId)))
        collaborators = entities
      }
      const { wifiProfile, entityId } = await fetchWifiProfile(collaborators)
      const { ethernetProfile } = await fetchEthernetProfile()
      const isNonSharedProfile = Boolean(wifiProfile) && !Boolean(wifiProfile.data.shared)
      if (isNonSharedProfile) {
        setNonSharedWifiProfileId(selectedManagedGateway.wifi_profile_id)
      }
      setInitialValues(oldValues => ({
        ...oldValues,
        wifi_profile: updateInitialWifiProfile(
          oldValues,
          wifiProfile,
          entityId,
          isNonSharedProfile,
        ),
        ethernet_profile: updateInitialEthernetProfile(oldValues, ethernetProfile),
      }))
    },
    [
      fetchEthernetProfile,
      fetchWifiProfile,
      gtwId,
      mayViewCollaborators,
      selectedManagedGateway.wifi_profile_id,
      updateInitialEthernetProfile,
      updateInitialWifiProfile,
    ],
  )

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }, cleanValues) => {
      const getWifiProfileId = async (profile, shouldUpdateNonSharedWifiProfile) => {
        const { profile_id, _profile_of, _override, _enable_wifi_connection, ...wifiProfile } =
          profile
        let wifiProfileId = profile_id
        // If the WiFi profile id contains 'shared', create/update that profile.
        // The id could be either shared or non-shared.
        if (wifiProfileId.includes('shared')) {
          const normalizedWifiProfile = normalizeWifiProfile(
            wifiProfile,
            wifiProfileId === 'shared',
          )
          if (shouldUpdateNonSharedWifiProfile) {
            if (
              !isEmpty(
                diff(initialValues.wifi_profile, profile, {
                  exclude: ['_access_point'],
                }),
              )
            ) {
              const profileDiff = diff(initialValues.wifi_profile, normalizedWifiProfile)

              await dispatch(
                attachPromise(
                  updateConnectionProfile(
                    _profile_of,
                    nonSharedWifiProfileId,
                    CONNECTION_TYPES.WIFI,
                    profileDiff,
                  ),
                ),
              )
            }
            wifiProfileId = nonSharedWifiProfileId
          } else {
            const { data } = await dispatch(
              attachPromise(
                createConnectionProfile(_profile_of, CONNECTION_TYPES.WIFI, normalizedWifiProfile),
              ),
            )
            wifiProfileId = data.profile_id
          }
        }
        return wifiProfileId
      }

      const getEthernetProfileId = async (profile, cleanProfile) => {
        if (!isEqual(profile, initialValues.ethernet_profile) || !hasEthernetProfileSet) {
          const { _use_static_ip, ...rest } = cleanProfile
          const body = { ...rest, profile_name: Date.now().toString() }
          const {
            data: { profile_id: ethernet_profile_id },
          } = await dispatch(
            attachPromise(createConnectionProfile(undefined, CONNECTION_TYPES.ETHERNET, body)),
          )

          return ethernet_profile_id
        }

        return initialValues.ethernet_profile.profile_id
      }

      setError(undefined)
      try {
        const { wifi_profile, ethernet_profile } = values

        const shouldUpdateNonSharedWifiProfile =
          wifi_profile.profile_id === 'non-shared' &&
          Boolean(nonSharedWifiProfileId) &&
          wifi_profile._enable_wifi_connection

        const wifiProfileId = await getWifiProfileId(wifi_profile, shouldUpdateNonSharedWifiProfile)
        const ethernetProfileId = await getEthernetProfileId(
          ethernet_profile,
          cleanValues.ethernet_profile,
        )
        const body = {
          wifi_profile_id:
            !wifi_profile._enable_wifi_connection && !wifi_profile._override ? null : wifiProfileId,
          ethernet_profile_id: ethernetProfileId,
        }

        await dispatch(attachPromise(updateManagedGateway(gtwId, body)))

        // Reset the form and the initial values
        let resetValues = { ...values }
        if (wifi_profile.profile_id !== 'non-shared') {
          resetValues = {
            ...values,
            wifi_profile: {
              ...values.wifi_profile,
              ...initialWifiProfile,
              profile_id: wifiProfileId,
              _profile_of: wifi_profile._profile_of,
            },
          }
        }

        setInitialValues(resetValues)
        resetForm({
          values: resetValues,
        })

        if (resetValues.wifi_profile._enable_wifi_connection) {
          setSaveFormClicked(true)
        }

        toast({
          title: selectedGateway.name,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
        toast({
          title: selectedGateway.name,
          message: m.updateFailure,
          type: toast.types.ERROR,
        })
      }
    },
    [
      dispatch,
      gtwId,
      hasEthernetProfileSet,
      initialValues.ethernet_profile,
      initialValues.wifi_profile,
      nonSharedWifiProfileId,
      selectedGateway.name,
    ],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <div className="item-12 d-flex gap-ls-s md:direction-column">
        <div className="w-full">
          <PageTitle title={sharedMessages.connectionSettings} />
          {isFirstClaim && <Notification info small content={m.firstNotification} />}
          <Form
            error={error}
            onSubmit={handleSubmit}
            initialValues={initialValues}
            validationSchema={validationSchema}
          >
            <>
              <WifiSettingsFormFields
                initialValues={initialValues}
                isWifiConnected={connectionsData.isWifiConnected}
                saveFormClicked={saveFormClicked}
              />
              <hr />
              <EthernetSettingsFormFields />

              <SubmitBar className="mb-cs-l">
                <Form.Submit
                  component={SubmitButton}
                  message={sharedMessages.saveChanges}
                  disabled={isLoading}
                />
              </SubmitBar>
            </>
          </Form>
        </div>
        <ManagedGatewayConnections connectionsData={connectionsData} />
      </div>
    </RequireRequest>
  )
}

export default GatewayConnectionSettings
