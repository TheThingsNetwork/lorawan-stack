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

import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import IntlHelmet from '../../../lib/components/intl-helmet'
import ModalButton from '../../../components/button/modal-button'
import Notification from '../../../components/notification'
import toast from '../../../components/toast'

import { updateDevice } from '../../store/actions/devices'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '../../store/selectors/devices'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

const m = defineMessages({
  deleteCode: 'Delete claim authentication code',
  deleteWarning: 'Are you sure you want to delete this claim authentication code?',
  noCodeSet: 'There is currently no claim authentication code set',
  validateToDate: 'Invalid code expiration date',
  deleteFailure: 'There was a problem deleting the claim authentication code',
  deleteSuccess: 'The claim authentication code has been deleted successfully',
  updateSuccess: 'The claim authentication code has been updated successfully',
  validateCode: 'Only uppercase letters and numbers are allowed',
})

const validationSchema = Yup.object({
  claim_authentication_code: Yup.object({
    value: Yup.string()
      .matches(/^[A-Z0-9]{1,32}$/, m.validateCode)
      .required(sharedMessages.validateRequired),
    valid_from: Yup.date().required(sharedMessages.validateRequired),
    valid_to: Yup.date()
      .required(sharedMessages.validateRequired)
      .when('valid_from', (validFrom, schema) => {
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

      // Convert ISO 8601 date time representation to just yyyy-MM-dd required by the
      // date input. So, we pick only the data part from YYYY-MM-DDTHH:mm:ss.sssZ which is
      // known to be fixed.
      const validFrom = new Date(valid_from).toISOString().slice(0, 10)
      const validTo = new Date(valid_to).toISOString().slice(0, 10)

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

      try {
        await updateDevice(appId, devId, validationSchema.cast(values))

        resetForm(values)
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
              required
            />
            <Form.Field
              title={sharedMessages.validTo}
              name="claim_authentication_code.valid_to"
              type="date"
              component={Input}
              required
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
