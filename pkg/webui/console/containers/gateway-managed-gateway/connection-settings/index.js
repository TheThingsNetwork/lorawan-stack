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

import React, { useCallback, useMemo, useRef, useState } from 'react'
import { defineMessages } from 'react-intl'
import { useParams, useSearchParams } from 'react-router-dom'
import { Col, Row } from 'react-grid-system'
import { useDispatch, useSelector } from 'react-redux'

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
  normalizeEthernetProfile,
} from '@console/containers/gateway-managed-gateway/shared/utils'
import WifiSettingsFormFields from '@console/containers/gateway-managed-gateway/connection-settings/wifi-settings-form-fields'
import EthernetSettingsFormFields from '@console/containers/gateway-managed-gateway/connection-settings/ethernet-settings-form-fields'
import ManagedGatewayConnections from '@console/containers/gateway-managed-gateway/connection-settings/connections'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'

import { checkFromState, mayViewOrEditGatewayCollaborators } from '@console/lib/feature-checks'

import {
  createConnectionProfile,
  GET_ACCESS_POINTS_BASE,
  getConnectionProfile,
} from '@console/store/actions/connection-profiles'

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

  const hasWifiProfileSet = Boolean(selectedManagedGateway.wifi_profile_id)

  useBreadcrumbs(
    'gtws.single.managed-gateway.connection-settings',
    <Breadcrumb
      path={`/gateways/${gtwId}/managed-gateway/connection-settings`}
      content={sharedMessages.connectionSettings}
    />,
  )

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      setError(undefined)
      try {
        const { wifi_profile, ethernet_profile } = values
        if (wifi_profile.profile.includes('shared')) {
          const { profile, _profileOf, _connection_type, ...wifiProfile } = wifi_profile
          const normalizedWifiProfile = normalizeWifiProfile(wifiProfile, profile === 'shared')
          const {
            data: { profile_id: wifi_profile_id },
          } = await dispatch(
            attachPromise(
              createConnectionProfile(_profileOf, CONNECTION_TYPES.WIFI, normalizedWifiProfile),
            ),
          )
          console.log(wifi_profile_id)
        }
        const { _connection_type, ...ethernetProfile } = ethernet_profile

        const normalizedEthernetProfile = normalizeEthernetProfile(ethernetProfile)
        const {
          data: { profile_id: ethernet_profile_id },
        } = await dispatch(
          attachPromise(
            createConnectionProfile(
              undefined,
              CONNECTION_TYPES.ETHERNET,
              normalizedEthernetProfile,
            ),
          ),
        )
        console.log(ethernet_profile_id)

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
    [dispatch, selectedGateway.name],
  )

  const initialValues = useMemo(
    () => ({
      wifi_profile: { ...initialWifiProfile },
      ethernet_profile: { ...initialEthernetProfile },
    }),
    [],
  )

  const loadData = useCallback(
    async dispatch => {
      let collaborators = []
      let wifiProfile
      let entityId
      if (mayViewCollaborators) {
        const { entities } = await dispatch(attachPromise(getCollaboratorsList('gateway', gtwId)))
        collaborators = entities
      }
      if (hasWifiProfileSet) {
        try {
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
        } catch (e) {}
        if (!wifiProfile) {
          const orgCollaborators = collaborators
            .filter(({ ids }) => 'organization_ids' in ids)
            .map(({ ids }) => getCollaboratorId({ ids }))

          for (const orgId of orgCollaborators) {
            try {
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
              console.log(e)
            }
          }
        }
      }
    },
    [
      gtwId,
      hasWifiProfileSet,
      mayViewCollaborators,
      selectedManagedGateway.wifi_profile_id,
      userId,
    ],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <PageTitle title={sharedMessages.connectionSettings} />
      <Row>
        <Col lg={8} md={6} sm={12}>
          {isFirstClaim && <Notification info small content={m.firstNotification} />}
          <Form
            error={error}
            onSubmit={handleSubmit}
            initialValues={initialValues}
            validationSchema={validationSchema}
          >
            <>
              <WifiSettingsFormFields />
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
        </Col>
        <Col lg={4} md={6} sm={12}>
          <ManagedGatewayConnections />
        </Col>
      </Row>
    </RequireRequest>
  )
}

export default GatewayConnectionSettings
