// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { replace } from 'connected-react-router'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import diff from '../../../lib/diff'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import IntlHelmet from '../../../lib/components/intl-helmet'
import getHostnameFromUrl from '../../../lib/host-from-url'
import PropTypes from '../../../lib/prop-types'
import toast from '../../../components/toast'
import api from '../../api'

import { updateDevice } from '../../store/actions/devices'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '../../store/selectors/devices'
import {
  selectIsConfig,
  selectAsConfig,
  selectJsConfig,
  selectNsConfig,
} from '../../../lib/selectors/env'
import {
  mayEditApplicationDeviceKeys,
  mayReadApplicationDeviceKeys,
} from '../../lib/feature-checks'

import { isDeviceOTAA, isDeviceJoined } from './utils'
import m from './messages'

import IdentityServerForm from './identity-server-form'
import ApplicationServerForm from './application-server-form'
import JoinServerForm from './join-server-form'
import NetworkServerForm from './network-server-form'
import Collapse from './collapse'

import style from './device-general-settings.styl'

@connect(
  state => ({
    device: selectSelectedDevice(state),
    devId: selectSelectedDeviceId(state),
    appId: selectSelectedApplicationId(state),
    isConfig: selectIsConfig(),
    asConfig: selectAsConfig(),
    jsConfig: selectJsConfig(),
    nsConfig: selectNsConfig(),
    mayReadKeys: mayReadApplicationDeviceKeys.check(
      mayReadApplicationDeviceKeys.rightsSelector(state),
    ),
    mayEditKeys: mayEditApplicationDeviceKeys.check(
      mayEditApplicationDeviceKeys.rightsSelector(state),
    ),
  }),
  dispatch => ({
    ...bindActionCreators({ updateDevice: attachPromise(updateDevice) }, dispatch),
    onDeleteSuccess: appId => dispatch(replace(`/applications/${appId}/devices`)),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    onDeleteSuccess: () => dispatchProps.onDeleteSuccess(stateProps.appId),
  }),
)
@withBreadcrumb('device.single.general-settings', function(props) {
  const { devId, appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/general-settings`}
      content={sharedMessages.generalSettings}
    />
  )
})
export default class DeviceGeneralSettings extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    asConfig: PropTypes.stackComponent.isRequired,
    device: PropTypes.device.isRequired,
    isConfig: PropTypes.stackComponent.isRequired,
    jsConfig: PropTypes.stackComponent.isRequired,
    mayEditKeys: PropTypes.bool.isRequired,
    mayReadKeys: PropTypes.bool.isRequired,
    nsConfig: PropTypes.stackComponent.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    updateDevice: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(values) {
    const { device, appId, updateDevice } = this.props
    const { activation_mode, ...updatedDevice } = values

    const {
      ids: { device_id: deviceId },
    } = device

    const changed = diff(device, updatedDevice, ['updated_at', 'created_at'])

    return updateDevice(appId, deviceId, changed)
  }

  @bind
  async handleSubmitSuccess() {
    const { device } = this.props

    const {
      ids: { device_id: deviceId },
    } = device

    toast({
      title: deviceId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDelete() {
    const { appId, device } = this.props
    const {
      ids: { device_id: deviceId },
    } = device

    return api.device.delete(appId, deviceId)
  }

  @bind
  async handleDeleteSuccess() {
    const { device, onDeleteSuccess } = this.props
    const {
      ids: { device_id: deviceId },
    } = device

    onDeleteSuccess()
    toast({
      title: deviceId,
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDeleteFailure() {
    const { device } = this.props
    const {
      ids: { device_id: deviceId },
    } = device

    toast({
      title: deviceId,
      message: m.deleteFailure,
      type: toast.types.ERROR,
    })
  }

  render() {
    const { device, isConfig, asConfig, jsConfig, nsConfig, mayEditKeys, mayReadKeys } = this.props

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

    return (
      <Container>
        <IntlHelmet title={sharedMessages.generalSettings} />
        <Row>
          <Col lg={8} md={12} className={style.container}>
            <Collapse
              title={m.isTitle}
              description={isDescription}
              disabled={isDisabled}
              initialCollapsed={false}
            >
              <IdentityServerForm
                device={device}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                onDelete={this.handleDelete}
                onDeleteSuccess={this.handleDeleteSuccess}
                onDeleteFailure={this.handleDeleteFailure}
                jsConfig={jsConfig}
                mayReadKeys={mayReadKeys}
              />
            </Collapse>
            <Collapse title={m.nsTitle} description={nsDescription} disabled={nsDisabled}>
              <NetworkServerForm
                device={device}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                mayEditKeys={mayEditKeys}
                mayReadKeys={mayReadKeys}
              />
            </Collapse>
            <Collapse title={m.asTitle} description={asDescription} disabled={asDisabled}>
              <ApplicationServerForm
                device={device}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                mayEditKeys={mayEditKeys}
                mayReadKeys={mayReadKeys}
              />
            </Collapse>
            <Collapse title={m.jsTitle} description={jsDescription} disabled={jsDisabled}>
              <JoinServerForm
                device={device}
                onSubmit={this.handleSubmit}
                onSubmitSuccess={this.handleSubmitSuccess}
                mayEditKeys={mayEditKeys}
                mayReadKeys={mayReadKeys}
              />
            </Collapse>
          </Col>
        </Row>
      </Container>
    )
  }
}
