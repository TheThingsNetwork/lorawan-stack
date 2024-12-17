// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { useDispatch, useSelector } from 'react-redux'
import { useNavigate, useParams } from 'react-router-dom'
import classNames from 'classnames'

import tts from '@console/api/tts'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import toast from '@ttn-lw/components/toast'
import Collapse from '@ttn-lw/components/collapse'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import {
  selectAsConfig,
  selectIsConfig,
  selectJsConfig,
  selectNsConfig,
} from '@ttn-lw/lib/selectors/env'
import { isBackend, getBackendErrorName } from '@ttn-lw/lib/errors/utils'

import {
  mayEditApplicationDeviceKeys,
  mayReadApplicationDeviceKeys,
} from '@console/lib/feature-checks'

import {
  updateDevice,
  resetDevice,
  resetUsedDevNonces,
  deleteDevice,
} from '@console/store/actions/devices'
import { unclaimDevice } from '@console/store/actions/claim'
import { getBandsList, getNsFrequencyPlans } from '@console/store/actions/configuration'

import {
  selectSelectedDevice,
  selectSelectedDeviceClaimable,
} from '@console/store/selectors/devices'
import { selectNsFrequencyPlans } from '@console/store/selectors/configuration'

import IdentityServerForm from './identity-server-form'
import ApplicationServerForm from './application-server-form'
import JoinServerForm from './join-server-form'
import NetworkServerForm from './network-server-form'
import { isDeviceOTAA, isDeviceJoined } from './utils'
import m from './messages'

import style from './device-general-settings.styl'

const DeviceGeneralSettings = () => {
  const dispatch = useDispatch()
  const { appId, devId } = useParams()
  const device = useSelector(selectSelectedDevice)
  const supportsClaiming = useSelector(selectSelectedDeviceClaimable)
  const mayReadKeys = useSelector(state =>
    mayReadApplicationDeviceKeys.check(mayReadApplicationDeviceKeys.rightsSelector(state)),
  )
  const mayEditKeys = useSelector(state =>
    mayEditApplicationDeviceKeys.check(mayEditApplicationDeviceKeys.rightsSelector(state)),
  )
  const navigate = useNavigate()
  const isConfig = selectIsConfig()
  const asConfig = selectAsConfig()
  const jsConfig = selectJsConfig()
  const nsConfig = selectNsConfig()
  const storeFrequencyPlans = useSelector(selectNsFrequencyPlans)
  const [defaultMacSettings, setMacSettings] = useState({})
  const [bandId, setBandId] = useState(undefined)

  useBreadcrumbs(
    'device.general',
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/general-settings`}
      content={sharedMessages.generalSettings}
    />,
  )

  const handleSubmit = useCallback(
    async patch => dispatch(attachPromise(updateDevice(appId, devId, patch))),
    [appId, devId, dispatch],
  )

  const handleSubmitSuccess = useCallback(async () => {
    toast({
      title: devId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [devId])

  const handleUnclaim = useCallback(async () => {
    const {
      ids: { dev_eui: devEui, join_eui: joinEui },
    } = device
    await dispatch(attachPromise(unclaimDevice(appId, devId, devEui, joinEui)))
  }, [appId, devId, device, dispatch])

  const handleUnclaimFailure = useCallback(async () => {
    toast({
      title: devId,
      message: m.unclaimFailure,
      type: toast.types.ERROR,
    })
  }, [devId])

  const handleDelete = useCallback(
    async () => await dispatch(attachPromise(deleteDevice(appId, devId))),
    [appId, devId, dispatch],
  )

  const handleDeleteSuccess = useCallback(async () => {
    const {
      ids: { device_id: deviceId },
    } = device
    navigate(`/applications/${appId}/devices`)
    toast({
      title: deviceId,
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
  }, [appId, device, navigate])

  const handleDeleteFailure = useCallback(async () => {
    const {
      ids: { device_id: deviceId },
    } = device
    toast({
      title: deviceId,
      message: m.deleteFailure,
      type: toast.types.ERROR,
    })
  }, [device])

  const isOTAA = isDeviceOTAA(device)
  const { enabled: isEnabled } = isConfig
  const { enabled: asEnabled, base_url: stackAsUrl } = asConfig
  const { enabled: jsEnabled, base_url: stackJsUrl } = jsConfig
  const { enabled: nsEnabled, base_url: stackNsUrl } = nsConfig

  // 1. Disable the section if IS is not in cluster.
  const isDisabled = !isEnabled
  let isDescription = m.isDescription
  if (isDisabled) {
    isDescription = m.isDescriptionMissing
  }

  // 1. Disable the section if AS is not in cluster.
  // 2. Disable the section if the device is OTAA and joined since no fields are stored in the AS.
  // 3. Disable the section if NS is not in cluster, since activation mode is unknown.
  // 4. Disable the seciton if `application_server_address` is not equal to the cluster AS address.
  const sameAsAddress = getHostnameFromUrl(stackAsUrl) === device.application_server_address
  const isJoined = isDeviceJoined(device)
  const asDisabled = !asEnabled || (isOTAA && !isJoined) || !nsEnabled || !sameAsAddress
  let asDescription = m.asDescription
  if (!asEnabled) {
    asDescription = m.asDescriptionMissing
  } else if (!nsEnabled) {
    asDescription = m.activationModeUnknown
  } else if (isOTAA && !isJoined) {
    asDescription = m.asDescriptionOTAA
  } else if (!sameAsAddress) {
    asDescription = m.notInCluster
  }

  // 1. Disable the section if JS is not in cluster.
  // 2. Disable the section if the device is ABP/Multicast, since JS does not store ABP/Multicast
  // devices.
  // 3. Disable the seciton if `join_server_address` is not equal to the cluster JS address.
  const sameJsAddress = getHostnameFromUrl(stackJsUrl) === device.join_server_address
  const jsDisabled = !jsEnabled || !isOTAA || !sameJsAddress
  let jsDescription = m.jsDescription
  if (!jsEnabled) {
    jsDescription = m.jsDescriptionMissing
  } else if (nsEnabled && !isOTAA) {
    jsDescription = m.jsDescriptionOTAA
  } else if (!sameJsAddress) {
    jsDescription = m.notInCluster
  }

  // 1. Disable the section if NS is not in cluster.
  // 2. Disable the seciton if `network_server_address` is not equal to the cluster NS address.
  const sameNsAddress = getHostnameFromUrl(stackNsUrl) === device.network_server_address
  const nsDisabled = !nsEnabled || !sameNsAddress
  let nsDescription = m.nsDescription
  if (!nsEnabled) {
    nsDescription = m.nsDescriptionMissing
  } else if (!sameNsAddress) {
    nsDescription = m.notInCluster
  }

  const fetchData = useCallback(
    async dispatch => {
      if (device.frequency_plan_id && device.lorawan_phy_version) {
        let frequencyPlans = storeFrequencyPlans
        if (frequencyPlans.length === 0) {
          frequencyPlans = await dispatch(attachPromise(getNsFrequencyPlans()))
        }
        const bandId = frequencyPlans.find(fp => fp.id === device.frequency_plan_id).band_id
        setBandId(bandId)
        await dispatch(getBandsList(bandId, device.lorawan_phy_version))
      }
      if (device.lorawan_phy_version && device.frequency_plan_id) {
        try {
          const settings = await tts.Ns.getDefaultMacSettings(
            device.frequency_plan_id,
            device.lorawan_phy_version,
          )
          setMacSettings(settings)
        } catch (err) {
          if (isBackend(err) && getBackendErrorName(err) === 'no_band_version') {
            toast({
              type: toast.types.ERROR,
              message: sharedMessages.fpNotFoundError,
              messageValues: {
                lorawanVersion: device.lorawan_phy_version,
                freqPlan: device.frequency_plan_id,
                code: msg => <code>{msg}</code>,
              },
            })
          } else {
            toast({
              type: toast.types.ERROR,
              message: m.macSettingsError,
              messageValues: {
                freqPlan: device.frequency_plan_id,
                code: msg => <code>{msg}</code>,
              },
            })
          }
        }
      }
    },
    [device.lorawan_phy_version, device.frequency_plan_id, storeFrequencyPlans],
  )

  return (
    <RequireRequest requestAction={fetchData}>
      <div className="container container--xl grid">
        <IntlHelmet title={sharedMessages.generalSettings} />

        <div className={classNames(style.container, 'item-12')}>
          <Collapse
            title={m.isTitle}
            description={isDescription}
            disabled={isDisabled}
            initialCollapsed={false}
          >
            <IdentityServerForm
              device={device}
              onSubmit={handleSubmit}
              onSubmitSuccess={handleSubmitSuccess}
              onDelete={handleDelete}
              onDeleteSuccess={handleDeleteSuccess}
              onDeleteFailure={handleDeleteFailure}
              onUnclaim={handleUnclaim}
              onUnclaimFailure={handleUnclaimFailure}
              jsConfig={jsConfig}
              nsConfig={nsConfig}
              asConfig={asConfig}
              supportsClaiming={supportsClaiming}
            />
          </Collapse>
          <Collapse title={m.nsTitle} description={nsDescription} disabled={nsDisabled}>
            <NetworkServerForm
              device={device}
              defaultMacSettings={defaultMacSettings}
              bandId={bandId}
              onSubmit={handleSubmit}
              onSubmitSuccess={handleSubmitSuccess}
              onMacReset={resetDevice}
              mayEditKeys={mayEditKeys}
              mayReadKeys={mayReadKeys}
              getDefaultMacSettings={tts.Ns.getDefaultMacSettings}
            />
          </Collapse>
          <Collapse title={m.asTitle} description={asDescription} disabled={asDisabled}>
            <ApplicationServerForm
              device={device}
              onSubmit={handleSubmit}
              onSubmitSuccess={handleSubmitSuccess}
              mayEditKeys={mayEditKeys}
              mayReadKeys={mayReadKeys}
            />
          </Collapse>
          <Collapse title={m.jsTitle} description={jsDescription} disabled={jsDisabled}>
            <JoinServerForm
              device={device}
              onSubmit={handleSubmit}
              onSubmitSuccess={handleSubmitSuccess}
              mayEditKeys={mayEditKeys}
              mayReadKeys={mayReadKeys}
              onUsedDevNoncesReset={resetUsedDevNonces}
            />
          </Collapse>
        </div>
      </div>
    </RequireRequest>
  )
}

export default DeviceGeneralSettings
