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

import React from 'react'
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'

import LocationForm from '@console/components/location-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import locationToMarkers from '@console/lib/location-to-markers'

import { updateDevice } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

const m = defineMessages({
  locationInfoTitle: 'Understanding end device location settings',
  locationInfo:
    'The Things Stack is capable of storing locations from multiple sources at the same time. Next to automatic location updates sourced from frame payloads of the end device and various other means of resolving location, it is also possible to set a location manually. You can use this form to update this location-type.',
  setDeviceLocation: 'Set end device location manually',
})

@connect(
  state => ({
    device: selectSelectedDevice(state),
    appId: selectSelectedApplicationId(state),
    devId: selectSelectedDeviceId(state),
  }),
  { updateDevice: attachPromise(updateDevice) },
)
@withBreadcrumb('device.single.location', props => {
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
  async handleSubmit(location) {
    const { device, appId, devId, updateDevice } = this.props

    await updateDevice(appId, devId, { locations: { ...device.locations, user: location } })
  }

  @bind
  async handleDelete(deleteAll) {
    const { device, devId, appId, updateDevice } = this.props

    const { user, ...nonUserLocations } = device.locations || {}
    const newLocations = {
      ...(!deleteAll ? nonUserLocations : undefined),
    }

    return updateDevice(appId, devId, { locations: newLocations })
  }

  render() {
    const { device, devId } = this.props

    const { user, ...nonUserLocations } = device.locations || {}

    return (
      <Container>
        <IntlHelmet title={sharedMessages.location} />
        <Row>
          <Col lg={8} md={12}>
            <LocationForm
              entityId={devId}
              formTitle={m.setDeviceLocation}
              initialValues={user}
              additionalMarkers={locationToMarkers(nonUserLocations)}
              onSubmit={this.handleSubmit}
              onDelete={this.handleDelete}
              centerOnMarkers
            />
          </Col>
          <Col lg={4} md={12}>
            <Message content={m.locationInfoTitle} component="h4" className="mb-0 mt-ls-xl" />
            <Message content={m.locationInfo} component="p" />
          </Col>
        </Row>
      </Container>
    )
  }
}
