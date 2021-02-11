// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { omit } from 'lodash'

import STACK_COMPONENTS_MAP from '@console/constants/stack-components'

import Wizard from '@ttn-lw/components/wizard'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import PropTypes from '@ttn-lw/lib/prop-types'

import { ACTIVATION_MODES } from '@console/lib/device-utils'

import m from '../../messages'
import {
  getActivationMode,
  getLorawanVersion,
  getApplicationServerAddress,
  getNetworkServerAddress,
  getJoinServerAddress,
} from '../../utils'

import BasicSettingsForm from './basic-settings-form'
import NetworkSettingsForm from './network-settings-form'
import ApplicationSettingsForm from './application-settings-form'
import JoinSettingsForm from './join-settings-form'
import Prompt from './prompt'

const WIZARD_STEPS_MAP = {
  BASIC: STACK_COMPONENTS_MAP.is,
  NETWORK: STACK_COMPONENTS_MAP.ns,
  APPLICATION: STACK_COMPONENTS_MAP.as,
  JOIN: STACK_COMPONENTS_MAP.js,
}

const excludePaths = ['_device_classes']

const DeviceWizard = props => {
  const {
    configuration,
    asConfig,
    nsConfig,
    jsConfig,
    mayEditKeys,
    match,
    createDevice,
    createDeviceSuccess,
    rollbackProgress,
    prefixes,
  } = props

  const [completed, setCompleted] = React.useState(false)

  const wizardRef = React.useRef(null)

  const handleWizardComplete = React.useCallback(
    async values => {
      try {
        await createDevice(omit(values, excludePaths))
        setCompleted(true)
        return createDeviceSuccess(values)
      } catch (error) {
        const { steps, onError } = wizardRef.current
        const { request_details = {} } = error

        const { id } = steps.find(({ id }) => id === request_details.stack_component) || {}

        onError(id, error)
      }
    },
    [createDevice, createDeviceSuccess],
  )

  const handleBlockNavigation = React.useCallback(
    location => !location.pathname.endsWith('/steps'),
    [],
  )

  const handlePromptApprove = React.useCallback(
    location => {
      rollbackProgress(location.pathname)
    },
    [rollbackProgress],
  )

  return (
    <Wizard
      initialValues={configuration}
      initialStepId={WIZARD_STEPS_MAP.BASIC}
      onComplete={handleWizardComplete}
      completeMessage={sharedMessages.addDevice}
      ref={wizardRef}
    >
      {({ snapshot }) => {
        const activationMode = getActivationMode(snapshot)
        const lorawanVersion = getLorawanVersion(snapshot)

        const showNetworkStep =
          nsConfig.enabled &&
          getNetworkServerAddress(snapshot) === getHostnameFromUrl(nsConfig.base_url) &&
          activationMode !== ACTIVATION_MODES.NONE
        const showApplicationStep =
          asConfig.enabled &&
          getApplicationServerAddress(snapshot) === getHostnameFromUrl(asConfig.base_url) &&
          (activationMode === ACTIVATION_MODES.ABP || activationMode === ACTIVATION_MODES.MULTICAST)
        const showJoinStep =
          getJoinServerAddress(snapshot) === getHostnameFromUrl(jsConfig.base_url) &&
          activationMode === ACTIVATION_MODES.OTAA

        return (
          <>
            <Prompt
              when={!completed}
              shouldBlockNavigation={handleBlockNavigation}
              onApprove={handlePromptApprove}
            />
            <Wizard.Stepper>
              <Wizard.Stepper.Step title={m.basicTitle} description={m.basicDescription} />
              {showNetworkStep && (
                <Wizard.Stepper.Step title={m.networkTitle} description={m.networkDescription} />
              )}
              {showApplicationStep && (
                <Wizard.Stepper.Step title={m.appTitle} description={m.appDescription} />
              )}
              {showJoinStep && (
                <Wizard.Stepper.Step title={m.joinTitle} description={m.joinDescription} />
              )}
            </Wizard.Stepper>
            <Wizard.Steps>
              <Wizard.Step title={m.basicTitle} id={WIZARD_STEPS_MAP.BASIC}>
                <BasicSettingsForm
                  activationMode={activationMode}
                  lorawanVersion={lorawanVersion}
                  match={match}
                  title={m.basicTitle}
                  prefixes={prefixes}
                />
              </Wizard.Step>
              {showNetworkStep && (
                <Wizard.Step title={m.networkTitle} id={WIZARD_STEPS_MAP.NETWORK}>
                  <NetworkSettingsForm
                    activationMode={activationMode}
                    lorawanVersion={lorawanVersion}
                    mayEditKeys={mayEditKeys}
                    match={match}
                    title={m.networkTitle}
                  />
                </Wizard.Step>
              )}
              {showApplicationStep && (
                <Wizard.Step title={m.appTitle} id={WIZARD_STEPS_MAP.APPLICATION}>
                  <ApplicationSettingsForm
                    mayEditKeys={mayEditKeys}
                    match={match}
                    title={m.appTitle}
                  />
                </Wizard.Step>
              )}
              {showJoinStep && (
                <Wizard.Step title={m.joinTitle} id={WIZARD_STEPS_MAP.JOIN}>
                  <JoinSettingsForm
                    lorawanVersion={lorawanVersion}
                    mayEditKeys={mayEditKeys}
                    match={match}
                    title={m.joinTitle}
                  />
                </Wizard.Step>
              )}
            </Wizard.Steps>
          </>
        )
      }}
    </Wizard>
  )
}

DeviceWizard.propTypes = {
  asConfig: PropTypes.stackComponent.isRequired,
  configuration: PropTypes.shape({
    lorawan_version: PropTypes.string,
    supports_join: PropTypes.bool,
    multicast: PropTypes.bool,
    application_server_address: PropTypes.string,
    join_server_address: PropTypes.string,
    network_server_address: PropTypes.string,
  }).isRequired,
  createDevice: PropTypes.func.isRequired,
  createDeviceSuccess: PropTypes.func.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  match: PropTypes.match.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  nsConfig: PropTypes.stackComponent.isRequired,
  prefixes: PropTypes.euiPrefixes.isRequired,
  rollbackProgress: PropTypes.func.isRequired,
}

export default DeviceWizard
