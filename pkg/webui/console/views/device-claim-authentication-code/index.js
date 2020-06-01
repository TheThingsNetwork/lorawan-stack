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
import * as Yup from 'yup'
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'
import Notification from '@ttn-lw/components/notification'
import toast from '@ttn-lw/components/toast'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import toInputDate from '@ttn-lw/lib/to-input-date'

import { updateDevice } from '@console/store/actions/devices'
import { attachPromise } from '@console/store/actions/lib'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

const m = defineMessages({
  deleteCode: 'Delete claim authentication code',
  deleteWarning: 'Are you sure you want to delete this claim authentication code?',
  noCodeSet: 'There is currently no claim authentication code set',
  validateToDate: 'Expiration date must be after the validity date',
  deleteFailure: 'There was an error and the claim authentication code could not be deleted',
  deleteSuccess: 'Claim authentication deleted',
  updateSuccess: 'Claim authentication updated',
  validateCode: 'Claim authentication code must consist only of numbers and letters',
})

const validationSchema = Yup.object({
  claim_authentication_code: Yup.object({
    value: Yup.string()
      .matches(/^[A-Z0-9]{1,32}$/, m.validateCode)
      .required(sharedMessages.validateRequired),
    valid_from: Yup.date(),
    valid_to: Yup.date().when('valid_from', (validFrom, schema) => {
      if (validFrom) {
        return schema.min(validFrom, m.validateToDate)
      }

      return schema
    }),
  }),
})

const DeviceClaimAuthenticationCode = props => {
  const { appId, devId, device, updateDevice } = props

  const [error, setError] = React.useState('')
  const initialValues = React.useMemo(() => {
    const { claim_authentication_code } = device

    if (claim_authentication_code) {
      const { value, valid_from, valid_to } = claim_authentication_code

      const validFrom = toInputDate(new Date(valid_from))
      const validTo = toInputDate(new Date(valid_to))

      return {
        claim_authentication_code: {
          value,
          valid_from: validFrom,
          valid_to: validTo,
        },
      }
    }

    return {
      claim_authentication_code: {
        value: undefined,
        valid_from: undefined,
        valid_to: undefined,
      },
    }
  }, [device])

  const handleSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError('')

      // Convert any false value to undefined.
      for (const [key, value] of Object.entries(values.claim_authentication_code)) {
        values.claim_authentication_code[key] = value ? value : undefined
      }

      try {
        await updateDevice(appId, devId, validationSchema.cast(values))

        resetForm({ values })
        toast({
          title: devId,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
        setSubmitting(false)
      }
    },
    [appId, devId, updateDevice],
  )

  const handleDelete = React.useCallback(async () => {
    try {
      await updateDevice(appId, devId, { claim_authentication_code: null })

      toast({
        title: devId,
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      toast({
        title: devId,
        message: m.deleteFailure,
        type: toast.types.ERROR,
      })
    }
  }, [appId, devId, updateDevice])

  const entryExists = Boolean(device.claim_authentication_code)

  return (
    <Container>
      <IntlHelmet title={sharedMessages.location} />
      <Row>
        <Col lg={8} md={12}>
          {!entryExists && <Notification content={m.noCodeSet} info small />}
          <Form
            horizontal
            validateOnChange
            enableReinitialize
            error={error}
            initialValues={initialValues}
            validationSchema={validationSchema}
            onSubmit={handleSubmit}
          >
            <Form.Field
              title={sharedMessages.claimAuthCode}
              type="password"
              name="claim_authentication_code.value"
              component={Input}
              required
            />
            <Form.Field
              title={sharedMessages.validFrom}
              name="claim_authentication_code.valid_from"
              component={Input}
              type="date"
            />
            <Form.Field
              title={sharedMessages.validTo}
              name="claim_authentication_code.valid_to"
              type="date"
              component={Input}
            />
            <SubmitBar>
              <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
              <ModalButton
                type="button"
                icon="delete"
                message={m.deleteCode}
                modalData={{
                  message: m.deleteWarning,
                }}
                onApprove={handleDelete}
                danger
                naked
                disabled={!entryExists}
              />
            </SubmitBar>
          </Form>
        </Col>
      </Row>
    </Container>
  )
}

DeviceClaimAuthenticationCode.propTypes = {
  appId: PropTypes.string.isRequired,
  devId: PropTypes.string.isRequired,
  device: PropTypes.device.isRequired,
  updateDevice: PropTypes.func.isRequired,
}

export default connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    devId: selectSelectedDeviceId(state),
    device: selectSelectedDevice(state),
  }),
  { updateDevice: attachPromise(updateDevice) },
)(DeviceClaimAuthenticationCode)
