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
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import Form from '../../../components/form'
import Input from '../../../components/input'
import Select from '../../../components/select'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import IntlHelmet from '../../../lib/components/intl-helmet'
import Notification from '../../../components/notification'
import toast from '../../../components/toast'
import Message from '../../../lib/components/message'
import Checkbox from '../../../components/checkbox'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import {
  isDeviceABP,
  isDeviceMulticast,
  isDeviceOTAA,
  ACTIVATION_MODES,
} from '../../lib/device-utils'

import validationSchema from './validation-schema'

const m = defineMessages({
  dataRate: 'Data Rate {index}',
  resetsFCnt: 'Resets Frame Counters',
  resetWarning: 'Reseting is insecure and makes your device susceptible for replay attacks',
  rx2DataDateIndexDescription: 'The default RX2 data rate index value device uses after reset',
  rx2DataRateIndexTitle: 'Rx2 Data Rate Index',
  setMacSettings: 'Set End Device MAC Settings',
  updateSuccess: 'The MAC settings have been updated successfully',
})

const rx2DataRateIndexes = Array.from({ length: 15 }, (_, index) => ({
  value: index.toString(),
  label: <Message content={m.dataRate} values={{ index }} />,
}))

const DeviceMacSettings = props => {
  const { updateDevice, device, appId, devId } = props
  const { mac_settings = {} } = device

  const isABP = isDeviceABP(device)
  const isMulticast = isDeviceMulticast(device)
  const isOTAA = isDeviceOTAA(device)

  const [error, setError] = React.useState('')

  const [resetsFCnt, setResetsFCnt] = React.useState((isABP && mac_settings.resets_f_cnt) || false)
  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  const initialValues = React.useMemo(() => {
    let activation_mode = ACTIVATION_MODES.ABP
    if (isOTAA) {
      activation_mode = ACTIVATION_MODES.OTAA
    } else if (isMulticast) {
      activation_mode = ACTIVATION_MODES.MULTICAST
    }

    return validationSchema.cast(device, {
      context: {
        activation_mode,
      },
    })
  }, [device, isMulticast, isOTAA])

  const handleSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError('')

      let activation_mode = ACTIVATION_MODES.ABP
      if (isOTAA) {
        activation_mode = ACTIVATION_MODES.OTAA
      } else if (isMulticast) {
        activation_mode = ACTIVATION_MODES.MULTICAST
      }

      console.log(
        validationSchema.cast(values, {
          context: {
            activation_mode,
          },
        }),
      )

      try {
        await updateDevice(
          appId,
          devId,
          validationSchema.cast(values, {
            context: {
              activation_mode,
            },
          }),
        )

        resetForm(values)
        toast({
          title: devId,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        console.dir(error)
        setError(error)
        setSubmitting(false)
      }

      return
    },
    [appId, devId, isMulticast, isOTAA, updateDevice],
  )

  return (
    <Container>
      <IntlHelmet title={sharedMessages.macSettings} />
      <Row>
        <Col lg={8} md={12}>
          <Form
            horizontal
            validateOnChange
            enableReinitialize
            error={error}
            initialValues={initialValues}
            validationSchema={validationSchema}
            onSubmit={handleSubmit}
          >
            <Message component="h4" content={m.setMacSettings} />
            <Form.Field
              title={m.rx2DataRateIndexTitle}
              description={m.rx2DataDateIndexDescription}
              name="mac_settings.rx2_data_rate_index.value"
              component={Select}
              options={rx2DataRateIndexes}
            />
            {isABP && (
              <Form.Field
                title={m.resetsFCnt}
                onChange={handleResetsFCntChange}
                warning={resetsFCnt ? m.resetWarning : undefined}
                name="mac_settings.resets_f_cnt"
                component={Checkbox}
              />
            )}
            <SubmitBar>
              <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
            </SubmitBar>
          </Form>
        </Col>
      </Row>
    </Container>
  )
}

DeviceMacSettings.propTypes = {
  appId: PropTypes.string.isRequired,
  devId: PropTypes.string.isRequired,
  device: PropTypes.device.isRequired,
  updateDevice: PropTypes.func.isRequired,
}

export default DeviceMacSettings
