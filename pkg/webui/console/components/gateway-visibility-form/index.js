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

import React, { useState, useCallback } from 'react'
import { Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Radio from '@ttn-lw/components/radio-button'

import PropTypes from '@ttn-lw/lib/prop-types'
import Yup from '@ttn-lw/lib/yup'

const m = defineMessages({
  saveDefaultGatewayVisibility: 'Save default gateway visibility',
})

const validationSchema = Yup.object({
  visibility: Yup.object({}),
})

const GatewayVisibilityForm = ({ onSubmit, initialValues, error, submitMessage }) => {
  const handleSubmit = useCallback(values => onSubmit(validationSchema.cast(values)), [onSubmit])

  return (
    <Form
      onSubmit={handleSubmit}
      initialValues={initialValues}
      error={error}
      validationSchema={validationSchema}
    >
      <Row>
        <Col md={6} xs={12}>
          <Form.Field
            name="visibility.location"
            component={Checkbox}
            label={'Location'}
            description={'Show location'}
          />
          <Form.Field
            name="visibility.antenna_placement"
            component={Checkbox}
            label={'Antenna placement'}
            description={'Show antenna placement (indoor/outdoor)'}
          />
          <Form.Field
            name="visibility.antenna_count"
            component={Checkbox}
            label={'Antenna count'}
            description={'Show antenna count'}
          />
          <Form.Field
            name="visibility.fine_timestamps"
            component={Checkbox}
            label={'Fine timestamps'}
            description={'Show whether the gateway produces fine timestamps'}
          />
        </Col>
        <Col sm={6} xs={12}>
          <Form.Field
            name="visibility.contact_info"
            component={Checkbox}
            label={'Contact information'}
            description={'Show contact information'}
          />
          <Form.Field
            name="visibility.status"
            component={Checkbox}
            label={'Status'}
            description={'Show status (online/offline)'}
          />
          <Form.Field
            name="visibility.frequency_plan"
            component={Checkbox}
            label={'Frequency plan'}
            description={'Show frequency plan'}
          />
          <Form.Field
            name="visibility.packet_rates"
            component={Checkbox}
            label={'Packet rates'}
            description={'Show receive and transmission packet rates'}
          />
        </Col>
      </Row>
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={submitMessage} />
      </SubmitBar>
    </Form>
  )
}

GatewayVisibilityForm.propTypes = {
  error: PropTypes.error,
  initialValues: PropTypes.shape({
    _use_default_gateway_visibility: PropTypes.bool,
    visibility: PropTypes.shape({}),
  }),
  onSubmit: PropTypes.func.isRequired,
  submitMessage: PropTypes.message,
}

GatewayVisibilityForm.defaultProps = {
  error: undefined,
  submitMessage: m.saveDefaultGatewayVisibility,
  initialValues: {
    _use_default_gateway_visibility: false,
    visibility: {
      location: false,
      antenna_placement: false,
      antenna_count: false,
      fine_timestamps: false,
      contact_info: false,
      status: false,
      frequency_plan: false,
      packet_rates: false,
    },
  },
}

export default GatewayVisibilityForm
