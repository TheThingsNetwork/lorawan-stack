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

import React, { Component } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Notification from '@ttn-lw/components/notification'
import PageTitle from '@ttn-lw/components/page-title'

import DeviceImporter from '@console/containers/device-importer'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  selectDeviceTemplateFormats,
  selectDeviceTemplateFormatsFetching,
} from '@console/store/selectors/device-template-formats'

const m = defineMessages({
  noTemplatesTitle: 'No end device templates found',
  noTemplates:
    'There are currently no end device templates set up. Please set up an end device template to make use of the bulk device import feature. For more information please refer to the documentation.',
})

@connect(state => ({
  deviceTemplateFormats: selectDeviceTemplateFormats(state),
  deviceTemplateFormatsFetching: selectDeviceTemplateFormatsFetching(state),
}))
@withBreadcrumb('devices.import', function (props) {
  const { appId } = props
  return (
    <Breadcrumb path={`/applications/${appId}/devices/import`} content={sharedMessages.import} />
  )
})
export default class DeviceAddBulk extends Component {
  static propTypes = {
    deviceTemplateFormats: PropTypes.shape().isRequired,
    deviceTemplateFormatsFetching: PropTypes.bool.isRequired,
  }

  render() {
    const { deviceTemplateFormatsFetching, deviceTemplateFormats } = this.props
    const showEmptyWarning =
      !deviceTemplateFormatsFetching && Object.keys(deviceTemplateFormats).length === 0
    return (
      <Container>
        <PageTitle title={sharedMessages.importDevices} />
        <Row>
          <Col lg={8} md={12}>
            {showEmptyWarning && (
              <Notification warning title={m.noTemplatesTitle} content={m.noTemplates} />
            )}
            <DeviceImporter />
          </Col>
        </Row>
      </Container>
    )
  }
}
