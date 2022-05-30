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

import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import tts from '@console/api/tts'

import {
  selectNsConfig,
  selectAsConfig,
  selectJsConfig,
  selectSupportLinkConfig,
} from '@ttn-lw/lib/selectors/env'

import { checkFromState, mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { getTemplate } from '@console/store/actions/device-repository'
import { getApplicationDevEUICount, issueDevEUI } from '@console/store/actions/applications'

import {
  selectDeviceTemplate,
  selectDeviceTemplateFetching,
  selectDeviceTemplateError,
} from '@console/store/selectors/device-repository'
import {
  selectApplicationDevEUICount,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'
import { selectJoinEUIPrefixes } from '@console/store/selectors/join-server'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
  applicationDevEUICounter: selectApplicationDevEUICount(state),
  prefixes: selectJoinEUIPrefixes(state),
  template: selectDeviceTemplate(state),
  templateFetching: selectDeviceTemplateFetching(state),
  templateError: selectDeviceTemplateError(state),
  createDevice: (appId, device) => tts.Applications.Devices.create(appId, device),
  mayEditKeys: checkFromState(mayEditApplicationDeviceKeys, state),
  jsConfig: selectJsConfig(),
  nsConfig: selectNsConfig(),
  asConfig: selectAsConfig(),
  supportLink: selectSupportLinkConfig(),
})

const mapDispatchToProps = dispatch => ({
  createDeviceSuccess: (appId, deviceId) =>
    dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
  getRegistrationTemplate: (appId, version) => dispatch(getTemplate(appId, version)),
  fetchDevEUICounter: appId => dispatch(getApplicationDevEUICount(appId)),
  issueDevEUI: appId => dispatch(attachPromise(issueDevEUI(appId))),
})

export default DeviceRepository => connect(mapStateToProps, mapDispatchToProps)(DeviceRepository)
