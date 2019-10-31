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

import Form from '../form'
import Input from '../input'
import Message from '../../lib/components/message'
import SubmitBar from '../submit-bar'
import SubmitButton from '../submit-button'
import Notification from '../notification'
import ModalButton from '../button/modal-button'
import toast from '../toast'

import sharedMessages from '../../lib/shared-messages'
import PropTypes from '../../lib/prop-types'

const m = defineMessages({
  deleteWarning: 'Are you sure you want to delete this location entry?',
  deleteLocation: 'Remove location entry',
  noLocationSet: 'There is currently no location information set',
  updateSuccess: 'The location has been updated successfully',
  deleteFailure: 'There was a problem removing the location',
  deleteSuccess: 'The location has been removed successfully',
})

const defaultValues = {
  longitude: undefined,
  latitude: undefined,
  altitude: undefined,
}

@bind
class LocationForm extends Component {
  constructor(props) {
    super(props)

    this.form = React.createRef()
  }

  state = {
    error: '',
  }

  async onSubmit(values, { resetForm, setSubmitting }) {
    const { onSubmit, entityId } = this.props

    this.setState({ error: '' })

    try {
      await onSubmit(values)
      resetForm()
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

  async onDelete() {
    const { onDelete, entityId } = this.props

    try {
      await onDelete()
      this.form.current.resetForm()
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
    const { initialValues, formTitle, validationSchema } = this.props

    const { error } = this.state

    const entryExists = Boolean(initialValues)

    return (
      <React.Fragment>
        {!entryExists && <Notification content={m.noLocationSet} info small />}
        <Form
          error={error}
          horizontal
          validateOnChange
          initialValues={initialValues || defaultValues}
          validationSchema={validationSchema}
          onSubmit={this.onSubmit}
          formikRef={this.form}
        >
          <Message component="h4" content={formTitle} />
          <Form.Field
            type="number"
            title={sharedMessages.latitude}
            description={sharedMessages.latitudeDesc}
            name="latitude"
            component={Input}
            required
          />
          <Form.Field
            type="number"
            title={sharedMessages.longitude}
            description={sharedMessages.longitudeDesc}
            name="longitude"
            component={Input}
            required
          />
          <Form.Field
            type="number"
            title={sharedMessages.altitude}
            description={sharedMessages.altitudeDesc}
            name="altitude"
            component={Input}
            required
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
              disabled={!entryExists}
              danger
              naked
            />
          </SubmitBar>
        </Form>
      </React.Fragment>
    )
  }
}

LocationForm.propTypes = {
  /** The initial values of the form */
  initialValues: PropTypes.object,
  /** The handler for the submit function of the form */
  onSubmit: PropTypes.func.isRequired,
  /** The handler for the delete function of the form */
  onDelete: PropTypes.func.isRequired,
  /** The title message shown at the top of the form */
  formTitle: PropTypes.message,
  /** The validation schema of the form */
  validationSchema: PropTypes.object,
}

export default LocationForm
