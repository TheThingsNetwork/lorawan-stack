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
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import LocationForm from '@console/components/location-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { updateDevice } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

const m = defineMessages({
  setDeviceLocation: 'Set end device location',
})

const getRegistryLocation = function (locations) {
  let registryLocation
  if (locations) {
    for (const key of Object.keys(locations)) {
      if (locations[key].source === 'SOURCE_REGISTRY') {
        registryLocation = { location: locations[key], key }
        break
      }
    }
  }
  return registryLocation
}

@connect(
  state => ({
    device: selectSelectedDevice(state),
    appId: selectSelectedApplicationId(state),
    devId: selectSelectedDeviceId(state),
  }),
  { updateDevice: attachPromise(updateDevice) },
)
@withBreadcrumb('device.single.location', function (props) {
  const { devId, appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/location`}
      content={sharedMessages.location}
    />
  )
})
export default class DeviceGeneralSettings extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    devId: PropTypes.string.isRequired,
    device: PropTypes.device.isRequired,
    updateDevice: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(values) {
    const { device, appId, devId, updateDevice } = this.props

    const patch = {
      locations: {
        ...device.locations,
      },
    }

    const registryLocation = getRegistryLocation(device.locations)
    if (registryLocation) {
      // Update old location value.
      patch.locations[registryLocation.key] = {
        ...registryLocation.location,
        ...values,
      }
    } else {
      // Create new location value.
      patch.locations.user = {
        ...values,
        accuracy: 0,
        source: 'SOURCE_REGISTRY',
      }
    }

    await updateDevice(appId, devId, patch)
  }

  @bind
  async handleDelete() {
    const { device, devId, appId, updateDevice } = this.props
    const registryLocation = getRegistryLocation(device.locations)

    const patch = {
      locations: { ...device.location },
    }
    delete patch.locations[registryLocation.key]

    return updateDevice(appId, devId, patch)
  }

  render() {
    const { device, devId } = this.props
    const registryLocation = getRegistryLocation(device.locations)

    return (
      <Container>
        <IntlHelmet title={sharedMessages.location} />
        <Row>
          <Col lg={8} md={12}>
            <LocationForm
              entityId={devId}
              formTitle={m.setDeviceLocation}
              initialValues={registryLocation ? registryLocation.location : undefined}
              onSubmit={this.handleSubmit}
              onDelete={this.handleDelete}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
