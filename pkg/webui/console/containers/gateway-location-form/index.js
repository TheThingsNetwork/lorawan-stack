// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Radio from '@ttn-lw/components/radio-button'

import LocationForm, { hasLocationSet } from '@console/components/location-form'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { latitude as latitudeRegexp, longitude as longitudeRegexp } from '@console/lib/regexp'

import { updateGateway } from '@console/store/actions/gateways'

import { selectSelectedGateway, selectSelectedGatewayId } from '@console/store/selectors/gateways'

const m = defineMessages({
  updateLocationFromStatus: 'Update from status messages',
  updateLocationFromStatusDescription:
    'Update the location of this gateway based on incoming status messages',
  setGatewayLocation: 'Gateway antenna location settings',
  locationSource: 'Location source',
  locationPrivacy: 'Location privacy',
  placement: 'Placement',
  indoor: 'Indoor',
  outdoor: 'Outdoor',
  locationFromStatusMessage: 'Location set automatically from status messages',
  setLocationManually: 'Set location manually',
  noLocationSetInfo: 'This gateway has no location information set',
})

const validationSchema = Yup.object().shape({
  latitude: Yup.number().when('update_location_from_status', {
    is: false,
    then: schema =>
      schema
        .required(sharedMessages.validateRequired)
        .test('is-valid-latitude', sharedMessages.validateLatitude, value =>
          latitudeRegexp.test(String(value)),
        ),
    otherwise: schema => schema.strip(),
  }),
  longitude: Yup.number().when('update_location_from_status', {
    is: false,
    then: schema =>
      schema
        .required(sharedMessages.validateRequired)
        .test('is-valid-longitude', sharedMessages.validateLongitude, value =>
          longitudeRegexp.test(String(value)),
        ),
    otherwise: schema => schema.strip(),
  }),
  altitude: Yup.number().when('update_location_from_status', {
    is: false,
    then: schema =>
      schema.integer(sharedMessages.validateInt32).required(sharedMessages.validateRequired),
    otherwise: schema => schema.strip(),
  }),
  location_public: Yup.bool(),
  update_location_from_status: Yup.bool(),
  placement: Yup.string().oneOf(['PLACEMENT_UNKNOWN', 'INDOOR', 'OUTDOOR']),
})

const getRegistryLocation = antennas => {
  let registryLocation
  if (antennas) {
    for (const key of Object.keys(antennas)) {
      if (
        antennas[key].location !== null &&
        typeof antennas[key].location === 'object' &&
        antennas[key].location.source === 'SOURCE_REGISTRY'
      ) {
        registryLocation = { antenna: antennas[key], key }
        break
      } else {
        registryLocation = { antenna: antennas[key], key }
      }
    }
  }
  return registryLocation
}

const GatewayLocationForm = () => {
  const gateway = useSelector(selectSelectedGateway)
  const gatewayId = useSelector(selectSelectedGatewayId)
  const dispatch = useDispatch()
  const registryLocation = getRegistryLocation(gateway.antennas)
  const initialValues = {
    placement:
      registryLocation && registryLocation.antenna.placement
        ? registryLocation.antenna.placement
        : 'PLACEMENT_UNKNOWN',
    location_public: gateway.location_public || false,
    update_location_from_status: gateway.update_location_from_status || false,
    ...(hasLocationSet(registryLocation?.antenna?.location)
      ? registryLocation.antenna.location
      : {
          latitude: undefined,
          longitude: undefined,
          altitude: undefined,
        }),
  }

  const handleSubmit = useCallback(
    async values => {
      const { update_location_from_status, location_public, placement, ...location } = values
      const patch = {
        location_public,
        update_location_from_status,
      }

      const registryLocation = getRegistryLocation(gateway.antennas)
      if (!values.update_location_from_status) {
        if (registryLocation) {
          // Update old location value.
          patch.antennas = [...gateway.antennas]
          patch.antennas[registryLocation.key].location = {
            ...registryLocation.antenna.location,
            ...location,
          }
          patch.antennas[registryLocation.key].placement = placement
        } else {
          // Create new location value.
          patch.antennas = [
            {
              gain: 0,
              location: {
                ...values,
                accuracy: 0,
                source: 'SOURCE_REGISTRY',
              },
              placement,
            },
          ]
        }
      } else if (registryLocation) {
        patch.antennas = gateway.antennas.map(antenna => {
          const { location, ...rest } = antenna
          return rest
        })
        patch.antennas[registryLocation.key].placement = values.placement
      } else {
        patch.antennas = [{ gain: 0, placement: values.placement }]
      }

      return dispatch(attachPromise(updateGateway(gatewayId, patch)))
    },
    [dispatch, gateway.antennas, gatewayId],
  )

  const handleDelete = useCallback(
    async deleteAll => {
      const registryLocation = getRegistryLocation(gateway.antennas)

      if (deleteAll) {
        return dispatch(attachPromise(updateGateway(gatewayId, { antennas: [] })))
      }

      const patch = {
        antennas: [...gateway.antennas],
      }
      patch.antennas.splice(registryLocation.key, 1)

      return dispatch(attachPromise(updateGateway(gatewayId, patch)))
    },
    [dispatch, gateway.antennas, gatewayId],
  )

  const [updateLocationFromStatus, setUpdateLocationFromStatus] = useState(
    initialValues.update_location_from_status,
  )

  const handleUpdateLocationFromStatusChange = useCallback(useAutomaticUpdates => {
    setUpdateLocationFromStatus(useAutomaticUpdates)
  }, [])

  return (
    <LocationForm
      entityId={gatewayId}
      initialValues={initialValues}
      validationSchema={validationSchema}
      formTitle={m.setGatewayLocation}
      onSubmit={handleSubmit}
      onDelete={handleDelete}
      updatesDisabled={updateLocationFromStatus}
      disabledInfo={m.locationFromStatusMessage}
      noLocationSetInfo={m.noLocationSetInfo}
    >
      <Form.Field
        title={m.locationPrivacy}
        name="location_public"
        component={Checkbox}
        label={sharedMessages.gatewayLocationPublic}
        description={sharedMessages.locationDescription}
        tooltipId={tooltipIds.GATEWAY_LOCATION_PUBLIC}
      />
      <Form.Field
        title={m.locationSource}
        name="update_location_from_status"
        component={Radio.Group}
        tooltipId={tooltipIds.UPDATE_LOCATION_FROM_STATUS}
        onChange={handleUpdateLocationFromStatusChange}
      >
        <Radio label={m.setLocationManually} value={false} />
        <Radio label={m.updateLocationFromStatus} value />
      </Form.Field>
      <Form.Field
        title={m.placement}
        name="placement"
        component={Radio.Group}
        horizontal
        tooltipId={tooltipIds.GATEWAY_PLACEMENT}
      >
        <Radio label={sharedMessages.unknown} value="PLACEMENT_UNKNOWN" />
        <Radio label={m.indoor} value="INDOOR" />
        <Radio label={m.outdoor} value="OUTDOOR" />
      </Form.Field>
    </LocationForm>
  )
}

export default GatewayLocationForm
