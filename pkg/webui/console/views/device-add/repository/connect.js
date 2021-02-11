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

import api from '@console/api'

import {
  selectNsConfig,
  selectAsConfig,
  selectJsConfig,
  selectSupportLinkConfig,
} from '@ttn-lw/lib/selectors/env'

import { checkFromState, mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { getTemplate } from '@console/store/actions/device-repository'

import {
  selectDeviceTemplate,
  selectDeviceTemplateFetching,
} from '@console/store/selectors/device-repository'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectJoinEUIPrefixes } from '@console/store/selectors/join-server'

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
  prefixes: selectJoinEUIPrefixes(state),
  template: selectDeviceTemplate(state),
  templateFetching: selectDeviceTemplateFetching(state),
  createDevice: (appId, device) => api.device.create(appId, device),
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
})

export default DeviceRepository => connect(mapStateToProps, mapDispatchToProps)(DeviceRepository)
