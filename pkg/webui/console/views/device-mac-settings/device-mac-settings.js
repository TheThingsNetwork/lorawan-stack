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
import Message from '../../../lib/components/message'
import Checkbox from '../../../components/checkbox'
import KeyValueMap from '../../../components/key-value-map'
import toast from '../../../components/toast'

import diff from '../../../lib/diff'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import { isDeviceABP, isDeviceMulticast, ACTIVATION_MODES } from '../../lib/device-utils'

import validationSchema from './validation-schema'

const m = defineMessages({
  dataRateValue: 'Data Rate {index}',
  delayValue: '{count, plural, one {{count} second} other {{count} seconds}}',
  pingSlotPeriodicityDescription: 'Periodicity of the class B ping slot',
  pingSlotPeriodicityTitle: 'Ping Slot Periodicity',
  pingSlotPeriodicityValue: '{count, plural, one {every second} other {every {count} seconds}}',
  resetsFCnt: 'Resets Frame Counters',
  resetWarning: 'Reseting is insecure and makes your device susceptible for replay attacks',
  rx1DelayTitle: 'RX1 Delay',
  rx1DelayDescription: 'Class A RX1 delay. RX2 delay is RX1 delay + 1 second.',
  rx1DataRateOffsetTitle: 'RX1 Data Rate Offset',
  rx2DataDateIndexDescription: 'The default RX2 data rate index value device uses after reset',
  rx2DataRateIndexTitle: 'RX2 Data Rate Index',
  rx2FrequencyTitle: 'RX2 Frequency',
  rx2FrequencyDescription: 'Frequency for RX2 (Hz)',
  frequencyPlaceholder: 'e.g. 869525000 for 869,525 MHz',
  setMacSettings: 'Set End Device MAC Settings',
  updateSuccess: 'The MAC settings have been updated successfully',
  factoryPresetFreqTitle: 'Factory Preset Frequencies',
  factoryPresetFreqDescription: 'List of factory-preset frequencies',
  freqAdd: 'Add Frequency',
})

// Inconsistent mac_settings enum values see https://github.com/TheThingsNetwork/lorawan-stack/issues/2258

// 0...15
const dataRateIndexOptions = Array.from({ length: 15 }, (_, index) => ({
  value: index,
  label: <Message content={m.dataRateValue} values={{ index }} />,
}))

// 1...15
const rx1DelayOptions = Array.from({ length: 14 }, (_, index) => ({
  value: index + 1,
  label: <Message content={m.delayValue} values={{ count: index + 1 }} />,
}))

// 0...7
const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={m.pingSlotPeriodicityValue} values={{ count: value }} />,
  }
})

const DeviceMacSettings = props => {
  const { updateDevice, device, appId, devId } = props
  const { mac_settings = {} } = device

  const isABP = isDeviceABP(device)
  const isMulticast = isDeviceMulticast(device)
  const isClassB = device.supports_class_b || false
  const activationMode = isMulticast ? ACTIVATION_MODES.MULTICAST : ACTIVATION_MODES.ABP

  const [error, setError] = React.useState('')

  const [resetsFCnt, setResetsFCnt] = React.useState((isABP && mac_settings.resets_f_cnt) || false)
  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  const initialValues = React.useMemo(() => {
    return validationSchema.cast(device, {
      context: {
        activation_mode: activationMode,
        class_b: isClassB,
        is_init: true,
      },
    })
  }, [device, isClassB, activationMode])

  const handleSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError('')

      const casted = validationSchema.cast(values, {
        context: {
          activation_mode: activationMode,
          class_b: isClassB,
          is_init: false,
        },
      })

      const updatedValues = diff(device, casted)

      // Check if is there is a diff for `factory_preset_frequencies`. If so,
      // set the actual array instead of diff.
      if ('factory_preset_frequencies' in updatedValues.mac_settings) {
        updatedValues.mac_settings.factory_preset_frequencies =
          casted.mac_settings.factory_preset_frequencies
      }

      try {
        await updateDevice(appId, devId, updatedValues)

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
    [appId, devId, device, isClassB, updateDevice, activationMode],
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
            {isABP && (
              <>
                <Form.Field
                  title={m.rx1DelayTitle}
                  description={m.rx1DelayDescription}
                  name="mac_settings.rx1_delay.value"
                  component={Select}
                  options={rx1DelayOptions}
                />
                <Form.Field
                  type="number"
                  min={0}
                  max={7}
                  step={1}
                  title={m.rx1DataRateOffsetTitle}
                  name="mac_settings.rx1_data_rate_offset"
                  component={Input}
                />
                <Form.Field
                  title={m.resetsFCnt}
                  onChange={handleResetsFCntChange}
                  warning={resetsFCnt ? m.resetWarning : undefined}
                  name="mac_settings.resets_f_cnt"
                  component={Checkbox}
                />
              </>
            )}
            <Form.Field
              title={m.rx2DataRateIndexTitle}
              description={m.rx2DataDateIndexDescription}
              name="mac_settings.rx2_data_rate_index.value"
              component={Select}
              options={dataRateIndexOptions}
            />
            <Form.Field
              type="number"
              min={100000}
              step={1}
              title={m.rx2FrequencyTitle}
              description={m.rx2FrequencyDescription}
              placeholder={m.frequencyPlaceholder}
              name="mac_settings.rx2_frequency"
              component={Input}
            />
            <Form.Field
              indexAsKey
              name="mac_settings.factory_preset_frequencies"
              component={KeyValueMap}
              title={m.factoryPresetFreqTitle}
              description={m.factoryPresetFreqDescription}
              addMessage={m.freqAdd}
              valuePlaceholder={m.frequencyPlaceholder}
            />
            {isClassB && (
              <Form.Field
                title={m.pingSlotPeriodicityTitle}
                description={m.pingSlotPeriodicityDescription}
                name="mac_settings.ping_slot_periodicity.value"
                component={Select}
                options={pingSlotPeriodicityOptions}
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
