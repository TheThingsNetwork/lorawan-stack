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
import * as Yup from 'yup'

import sharedMessages from '../../../lib/shared-messages'
import { getDeviceId } from '../../../lib/selectors/id'

import LocationForm from '../../../components/location-form'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import IntlHelmet from '../../../lib/components/intl-helmet'

import { updateDevice } from '../../store/actions/device'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'

import {
  latitude as latitudeRegexp,
  longitude as longitudeRegexp,
  int32 as int32Regexp,
} from '../../lib/regexp'

const m = defineMessages({
  setDeviceLocation: 'Set Device Location',
})

const validationSchema = Yup.object().shape({
  latitude: Yup.string().required(sharedMessages.validateRequired)
    .matches(latitudeRegexp, sharedMessages.validateLatLong),
  longitude: Yup.string().required(sharedMessages.validateRequired)
    .matches(longitudeRegexp, sharedMessages.validateLatLong),
  altitude: Yup.string().matches(int32Regexp, sharedMessages.validateInt32).required(sharedMessages.validateRequired),
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
    device: state.device.device,
    appId: selectSelectedApplicationId(state),
    devId: getDeviceId(state.device.device),
  }),
  { updateDevice: attachPromise(updateDevice) }
)
@withBreadcrumb('device.single.data', function (props) {
  const { devId, appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/location`}
      icon="location"
      content={sharedMessages.location}
    />
  )
})
@bind
export default class DeviceGeneralSettings extends React.Component {
  async handleSubmit (values) {
    const { device, appId, devId, updateDevice } = this.props

    const patch = {
      locations: {
        ...device.locations,
      },
    }

    const registryLocation = getRegistryLocation(device.locations)
    if (registryLocation) {
      // Update old location value
      patch.locations[registryLocation.key] = {
        ...registryLocation.location,
        ...values,
      }
    } else {
      // Create new location value
      patch.locations.user = {
        ...values,
        accuracy: 0,
        source: 'SOURCE_REGISTRY',
      }
    }

    await updateDevice(appId, devId, patch)
  }

  async handleDelete () {
    const { device, devId, appId, updateDevice } = this.props
    const registryLocation = getRegistryLocation(device.locations)

    const patch = {
      locations: { ...device.location },
    }
    delete patch.locations[registryLocation.key]

    await updateDevice(appId, devId, patch)
  }

  render () {
    const { device, devId } = this.props
    const registryLocation = getRegistryLocation(device.locations)

    return (
      <Container>
        <IntlHelmet
          title={sharedMessages.location}
        />
        <Row>
          <Col sm={12} md={12} lg={8} xl={8}>
            <LocationForm
              entityId={devId}
              formTitle={m.setDeviceLocation}
              validationSchema={validationSchema}
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
