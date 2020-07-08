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
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Notification from '@ttn-lw/components/notification'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'
import LocationMap from '@ttn-lw/components/map'
import Overlay from '@ttn-lw/components/overlay'

import Message from '@ttn-lw/lib/components/message'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { latitude as latitudeRegexp, longitude as longitudeRegexp } from '@console/lib/regexp'

import style from './location-form.styl'

const m = defineMessages({
  deleteWarning: 'Are you sure you want to delete this location entry?',
  deleteLocation: 'Remove location entry',
  noLocationSet: 'There is currently no location information set',
  updateSuccess: 'Location updated',
  deleteFailure: 'An error occurred and the location could not be deleted',
  deleteSuccess: 'Location deleted',
  loadingLocation: 'Loading location...',
})

const validationSchema = Yup.object().shape({
  latitude: Yup.number()
    .test('is-valid-latitude', sharedMessages.validateLat, value =>
      latitudeRegexp.test(String(value)),
    )
    .required(sharedMessages.validateRequired),
  longitude: Yup.number()
    .test('is-valid-longitude', sharedMessages.validateLong, value =>
      longitudeRegexp.test(String(value)),
    )
    .required(sharedMessages.validateRequired),
  altitude: Yup.number()
    .integer(sharedMessages.validateInt32)
    .required(sharedMessages.validateRequired),
})

// We consider location of an entity set iff at least one coordinate is set,
// i.e. longitude, altitude, latitude.
const hasLocationSet = location =>
  typeof location.altitude !== 'undefined' ||
  typeof location.latitude !== 'undefined' ||
  typeof location.longitude !== 'undefined'

const defaultLocation = [0, 0]

class LocationForm extends Component {
  static propTypes = {
    /** Flag specifying whether the delete location button can be enabled. */
    allowDelete: PropTypes.bool,
    /** Additional fields to be passed as children. */
    children: PropTypes.node,
    entityId: PropTypes.string.isRequired,
    /** The title message shown at the top of the form. */
    formTitle: PropTypes.message.isRequired,
    /** The initial values of the form. */
    initialValues: PropTypes.shape({
      latitude: PropTypes.number,
      longitude: PropTypes.number,
      altitude: PropTypes.number,
    }),
    locationFieldsDisabled: PropTypes.bool,
    /** The handler for the delete function of the form. */
    onDelete: PropTypes.func.isRequired,
    /** The handler for the submit function of the form. */
    onSubmit: PropTypes.func.isRequired,
    validationSchema: PropTypes.shape({
      cast: PropTypes.func.isRequired,
    }),
  }

  static defaultProps = {
    children: null,
    initialValues: {
      latitude: undefined,
      longitude: undefined,
      altitude: undefined,
    },
    validationSchema,
    locationFieldsDisabled: false,
    allowDelete: true,
  }

  constructor(props) {
    super(props)

    this.form = React.createRef()

    this.state = {
      latitude: props.initialValues.latitude,
      longitude: props.initialValues.longitude,
      error: '',
      mapCenter: hasLocationSet(props.initialValues)
        ? [Number(props.initialValues.latitude), Number(props.initialValues.longitude)]
        : undefined,
      loading: !hasLocationSet(props.initialValues),
    }
  }

  @bind
  componentDidMount() {
    const { initialValues } = this.props
    if (!hasLocationSet(initialValues)) {
      if ('geolocation' in navigator) {
        navigator.geolocation.getCurrentPosition(
          position => {
            this.setState({
              mapCenter: [position.coords.latitude, position.coords.longitude],
              loading: false,
            })
          },
          error => {
            this.setState({ mapCenter: defaultLocation, loading: false })
          },
        )
      } else {
        this.setState({ mapCenter: defaultLocation, loading: false })
      }
    }
  }

  @bind
  async onSubmit(values, { resetForm, setSubmitting }) {
    const { onSubmit, entityId, validationSchema } = this.props

    this.setState({ error: '' })

    try {
      await onSubmit(validationSchema.cast(values))
      toast({
        title: entityId,
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      await this.setState({ error })
      setSubmitting(false)
    }
  }

  @bind
  handleClick(event) {
    const { setValues, values } = this.form.current
    this.setState({ latitude: event.latlng.lat, longitude: event.latlng.lng })
    setValues({
      ...values,
      latitude: event.latlng.lat,
      longitude: event.latlng.lng,
      altitude: values.altitude ? values.altitude : 0,
    })
  }

  @bind
  handleLatitudeChange(event) {
    const { longitude } = this.state
    const latitude = event.currentTarget.value
    if (longitude) {
      this.setState({ latitude, mapCenter: [Number(latitude), Number(longitude)] })
    } else {
      this.setState({ latitude })
    }
  }

  @bind
  handleLongitudeChange(event) {
    const { latitude } = this.state
    const longitude = event.currentTarget.value
    if (latitude) {
      this.setState({ longitude, mapCenter: [Number(latitude), Number(longitude)] })
    } else {
      this.setState({ longitude })
    }
  }

  @bind
  async onDelete() {
    const { onDelete, entityId } = this.props

    try {
      await onDelete()
      this.form.current.resetForm()
      this.setState({ latitude: undefined, longitude: undefined })
      toast({
        title: entityId,
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      await this.setState({ error })
    }
  }

  render() {
    const {
      initialValues,
      formTitle,
      validationSchema,
      children,
      locationFieldsDisabled,
      allowDelete,
    } = this.props
    const { error, latitude, longitude, mapCenter, loading } = this.state

    const entryExists = hasLocationSet(initialValues)

    let marker
    if (latitude && longitude) {
      marker = [
        {
          position: {
            latitude: Number(latitude),
            longitude: Number(longitude),
          },
        },
      ]
    }

    return (
      <React.Fragment>
        {!entryExists && <Notification content={m.noLocationSet} info small />}
        <Form
          enableReinitialize
          error={error}
          horizontal
          validateOnChange
          initialValues={initialValues}
          validationSchema={validationSchema}
          onSubmit={this.onSubmit}
          formikRef={this.form}
        >
          <Message component="h4" content={formTitle} />
          {children}
          <Overlay
            loading={loading}
            visible={loading}
            spinnerClassName={style.front}
            spinnerMessage={m.loadingLocation}
          >
            <LocationMap
              widget
              leafletConfig={{ zoom: 10, minZoom: 1 }}
              mapCenter={mapCenter}
              className={style.map}
              markers={marker}
              onClick={this.handleClick}
              clickable
              mapRef="map"
            />
          </Overlay>
          <Form.Field
            type="number"
            step="any"
            title={sharedMessages.latitude}
            description={sharedMessages.latitudeDesc}
            name="latitude"
            component={Input}
            required={!locationFieldsDisabled}
            disabled={locationFieldsDisabled}
            onBlur={this.handleLatitudeChange}
          />
          <Form.Field
            type="number"
            step="any"
            title={sharedMessages.longitude}
            description={sharedMessages.longitudeDesc}
            name="longitude"
            component={Input}
            required={!locationFieldsDisabled}
            disabled={locationFieldsDisabled}
            onBlur={this.handleLongitudeChange}
          />
          <Form.Field
            type="number"
            step="1"
            title={sharedMessages.altitude}
            description={sharedMessages.altitudeDesc}
            name="altitude"
            component={Input}
            required={!locationFieldsDisabled}
            disabled={locationFieldsDisabled}
          />
          <SubmitBar>
            <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
            <ModalButton
              type="button"
              icon="delete"
              message={m.deleteLocation}
              modalData={{
                message: m.deleteWarning,
              }}
              onApprove={this.onDelete}
              disabled={!entryExists || !allowDelete}
              danger
              naked
            />
          </SubmitBar>
        </Form>
      </React.Fragment>
    )
  }
}

export default LocationForm
