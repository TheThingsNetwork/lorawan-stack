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

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { defineMessages } from 'react-intl'
import { isPlainObject } from 'lodash'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Notification from '@ttn-lw/components/notification'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'
import LocationMap from '@ttn-lw/components/map'
import Overlay from '@ttn-lw/components/overlay'
import Checkbox from '@ttn-lw/components/checkbox'

import Message from '@ttn-lw/lib/components/message'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { latitude as latitudeRegexp, longitude as longitudeRegexp } from '@console/lib/regexp'

const m = defineMessages({
  deleteAllLocations: 'Delete all location data',
  deleteFailure: 'An error occurred and the location could not be deleted',
  deleteLocation: 'Remove location data',
  deleteSuccess: 'Location deleted',
  deleteAllInfo:
    'You can use the checkbox below to also remove location data that was set automatically (e.g. via frame payloads or status messages).',
  deleteWarning:
    'Are you sure you want to delete location data? This will delete the manual location entry from this entity.',
  deleteAllWarning:
    'Are you sure you want to delete location data? This will delete all location entries from this entity.',
  loadingLocation: 'Loading location…',
  mapDescription: 'Click into the map to set a location',
  mapDescriptionDisabled: 'The location will appear on the map once it has been set automatically',
  noLocationSetInfo: 'There is currently no manual location information set',
  updateSuccess: 'Location updated',
})

const validationSchema = Yup.object()
  .shape({
    latitude: Yup.number()
      .test('is-valid-latitude', sharedMessages.validateLatitude, value =>
        latitudeRegexp.test(String(value)),
      )
      .required(sharedMessages.validateRequired),
    longitude: Yup.number()
      .test('is-valid-longitude', sharedMessages.validateLongitude, value =>
        longitudeRegexp.test(String(value)),
      )
      .required(sharedMessages.validateRequired),
    altitude: Yup.number()
      .integer(sharedMessages.validateInt32)
      .required(sharedMessages.validateRequired),
    source: Yup.string().default('SOURCE_REGISTRY'),
  })
  .noUnknown()

// We consider location of an entity set iff at least one coordinate is set,
// i.e. longitude, altitude, latitude.
const hasLocationSet = location =>
  isPlainObject(location) &&
  (typeof location.altitude !== 'undefined' ||
    typeof location.latitude !== 'undefined' ||
    typeof location.longitude !== 'undefined')

const defaultLocation = [38.43745529233546, -5.089416503906251]

const emptyLocation = {
  latitude: undefined,
  longitude: undefined,
  altitude: undefined,
}

const LocationForm = props => {
  const {
    formTitle,
    validationSchema,
    children,
    updatesDisabled,
    additionalMarkers,
    initialValues,
    disabledInfo,
    noLocationSetInfo,
    onSubmit,
    onDelete,
    entityId,
  } = props

  const form = useRef(null)
  const [latitude, setLatitude] = useState(props.initialValues.latitude)
  const [longitude, setLongitude] = useState(props.initialValues.longitude)
  const [zoom, setZoom] = useState(14)
  const [error, setError] = useState(undefined)
  const [mapCenter, setMapCenter] = useState(undefined)
  const [loading, setLoading] = useState(true)
  const [deleteAll, setDeleteAll] = useState(
    !hasLocationSet(initialValues) && Object.keys(additionalMarkers).length !== 0,
  )

  const entryExists = useMemo(() => hasLocationSet(initialValues), [initialValues])
  const automaticExists = Object.keys(additionalMarkers).length !== 0
  const anyEntryExists = entryExists || automaticExists
  const onlyAutomaticExists = !entryExists && anyEntryExists
  const markers = [...additionalMarkers]
  if (typeof longitude === 'number' && typeof latitude === 'number') {
    markers.push({ position: { longitude, latitude } })
  }

  const getCurrentLocation = useCallback(async () => {
    let newState = { mapCenter: defaultLocation }
    if (
      !hasLocationSet(initialValues) &&
      additionalMarkers.length === 0 &&
      'geolocation' in navigator
    ) {
      newState = await new Promise(resolve => {
        navigator.geolocation.getCurrentPosition(
          position => {
            resolve({
              mapCenter: [
                isNaN(position.coords.latitude) ? defaultLocation[0] : position.coords.latitude,
                isNaN(position.coords.longitude) ? defaultLocation[1] : position.coords.longitude,
              ],
            })
          },
          () => {
            resolve({ mapCenter: defaultLocation, zoom: 2 })
          },
        )
      })
    }

    return newState
  }, [additionalMarkers.length, initialValues])

  useEffect(() => {
    getCurrentLocation().then(res => {
      setMapCenter(res.mapCenter)
      setZoom(res.zoom ?? 2)
      setLoading(false)
    })
  }, [getCurrentLocation])

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      setError(undefined)

      const castedValues = validationSchema.cast(values)

      try {
        await onSubmit(castedValues)
        setDeleteAll(false)
        toast({
          title: entityId,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
        setSubmitting(false)
      }
    },
    [entityId, onSubmit, validationSchema],
  )

  const handleClick = useCallback(
    event => {
      const { setValues, values } = form.current
      const latitude = isNaN(event.latlng.lat) ? defaultLocation[0] : event.latlng.lat
      const longitude = isNaN(event.latlng.lng) ? defaultLocation[1] : event.latlng.lng

      if (updatesDisabled) {
        return
      }
      setLatitude(latitude)
      setLongitude(longitude)
      setValues({
        ...values,
        latitude,
        longitude,
        altitude: values.altitude ? values.altitude : 0,
      })
    },
    [updatesDisabled],
  )

  const handleLatitudeChange = useCallback(
    event => {
      const latitude = event.currentTarget.value
      setLatitude(latitude)
      if (longitude) {
        setMapCenter([Number(latitude), Number(longitude)])
      }
    },
    [longitude],
  )

  const handleLongitudeChange = useCallback(
    event => {
      const longitude = event.currentTarget.value
      setLongitude(longitude)
      if (latitude) {
        setMapCenter([Number(latitude), Number(longitude)])
      }
    },
    [latitude],
  )

  const handleDeleteAllCheck = useCallback(event => {
    if (!event || !event.target) {
      return
    }
    setDeleteAll(event.target.checked)
  }, [])

  const handleDelete = useCallback(async () => {
    try {
      await onDelete(deleteAll)
      form.current.resetForm({ values: emptyLocation })
      setLatitude(undefined)
      setLongitude(undefined)
      toast({
        title: entityId,
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      setError(error)
    }
  }, [deleteAll, entityId, onDelete])

  return (
    <Form
      error={error}
      validateOnChange
      initialValues={initialValues}
      validationSchema={validationSchema}
      onSubmit={handleSubmit}
      formikRef={form}
    >
      <Form.SubTitle title={formTitle} />
      {children}
      <Message content={sharedMessages.location} component="h4" className="mb-cs-xs mt-0" />
      {!entryExists && <Notification content={noLocationSetInfo} info small />}
      <Overlay loading={loading} visible={loading} spinnerMessage={m.loadingLocation}>
        <LocationMap
          widget
          leafletConfig={{ zoom, minZoom: 1 }}
          mapCenter={mapCenter}
          markers={markers}
          onClick={handleClick}
          clickable
          centerOnMarkers
        />
      </Overlay>
      <Message
        content={updatesDisabled ? m.mapDescriptionDisabled : m.mapDescription}
        component="p"
        className="p-0 mt-cs-xs mb-cs-l tc-subtle-gray"
      />
      {updatesDisabled && disabledInfo && <Notification content={disabledInfo} info small />}
      <Form.Field
        type="number"
        step="any"
        title={sharedMessages.latitude}
        description={sharedMessages.latitudeDesc}
        name="latitude"
        component={Input}
        required={!updatesDisabled}
        disabled={updatesDisabled}
        onBlur={handleLatitudeChange}
      />
      <Form.Field
        type="number"
        step="any"
        title={sharedMessages.longitude}
        description={sharedMessages.longitudeDesc}
        name="longitude"
        component={Input}
        required={!updatesDisabled}
        disabled={updatesDisabled}
        onBlur={handleLongitudeChange}
      />
      <Form.Field
        type="number"
        step="1"
        title={sharedMessages.altitude}
        description={sharedMessages.altitudeDesc}
        name="altitude"
        component={Input}
        required={!updatesDisabled}
        disabled={updatesDisabled}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
        <ModalButton
          type="button"
          icon="delete"
          message={m.deleteLocation}
          modalData={{
            children: (
              <div>
                <Message
                  content={onlyAutomaticExists ? m.deleteAllWarning : m.deleteWarning}
                  component="span"
                />
                {entryExists && automaticExists && (
                  <>
                    <br />
                    <br />
                    <Message content={m.deleteAllInfo} component="span" />
                    <Checkbox
                      name="delete-all"
                      label={m.deleteAllLocations}
                      onChange={handleDeleteAllCheck}
                      className="mt-cs-m"
                      value={deleteAll}
                    />
                  </>
                )}
              </div>
            ),
          }}
          onApprove={handleDelete}
          disabled={updatesDisabled || !anyEntryExists}
          naked
          danger
        />
      </SubmitBar>
    </Form>
  )
}

LocationForm.propTypes = {
  additionalMarkers: PropTypes.markers,
  /** Additional fields to be passed as children. */
  children: PropTypes.node,
  disabledInfo: PropTypes.message,
  entityId: PropTypes.string.isRequired,
  /** The title message shown at the top of the form. */
  formTitle: PropTypes.message.isRequired,
  initialValues: PropTypes.entityLocation,
  noLocationSetInfo: PropTypes.message,
  /** The handler for the delete function of the form. */
  onDelete: PropTypes.func.isRequired,
  /** The handler for the submit function of the form. */
  onSubmit: PropTypes.func.isRequired,
  updatesDisabled: PropTypes.bool,
  validationSchema: PropTypes.shape({
    cast: PropTypes.func.isRequired,
  }),
}

LocationForm.defaultProps = {
  additionalMarkers: [],
  children: null,
  disabledInfo: undefined,
  initialValues: emptyLocation,
  validationSchema,
  updatesDisabled: false,
  noLocationSetInfo: m.noLocationSetInfo,
}

export { LocationForm as default, hasLocationSet }
