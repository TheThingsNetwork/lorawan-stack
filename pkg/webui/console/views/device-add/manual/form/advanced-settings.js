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

import React from 'react'
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

import { ACTIVATION_MODES } from '@console/lib/device-utils'

import messages from '../../messages'

import { DEVICE_CLASS_MAP } from './constants'

import style from './form.styl'

const m = defineMessages({
  advancedSectionTitle: 'Show advanced activation, LoRaWAN class and cluster settings',
  classA: 'None (class A only)',
  classB: 'Class B (Beaconing)',
  classC: 'Class C (Continuous)',
  classBandC: 'Class B and class C',
  useExternalServers: 'Use external LoRaWAN backend servers',
})

const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={messages.pingSlotPeriodicityValue} values={{ count: value }} />,
  }
})

const timeoutEncode = value => (Boolean(value) ? `${value}s` : value)
const timeoutDecode = value => (Boolean(value) ? RegExp(/\d+/).exec(value)[0] : value)

const allClassOptions = [
  { label: m.classA, value: DEVICE_CLASS_MAP.CLASS_A },
  { label: m.classB, value: DEVICE_CLASS_MAP.CLASS_B },
  { label: m.classC, value: DEVICE_CLASS_MAP.CLASS_C },
  { label: m.classBandC, value: DEVICE_CLASS_MAP.CLASS_B_C },
]
const multicastClassOptions = allClassOptions.filter(
  ({ value }) => value !== DEVICE_CLASS_MAP.CLASS_A,
)

const AdvancedSettingsSection = props => {
  const {
    nsEnabled,
    jsEnabled,
    onActivationModeChange,
    onDeviceClassChange,
    deviceClass,
    activationMode,
  } = props

  const isOTAA = activationMode === ACTIVATION_MODES.OTAA
  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const isNone = activationMode === ACTIVATION_MODES.NONE
  const isClassB = deviceClass === DEVICE_CLASS_MAP.CLASS_B

  const [externalServers, setExternalServer] = React.useState(false)
  const handleExternalServers = React.useCallback(
    () => setExternalServer(external => !external),
    [],
  )

  return (
    <Form.CollapseSection
      className={style.advancesSection}
      id="advanced-settings"
      title={m.advancedSectionTitle}
    >
      <Form.Field
        title={sharedMessages.activationMode}
        name="_activation_mode"
        component={Radio.Group}
        disabled={!nsEnabled && !jsEnabled}
        required={nsEnabled || jsEnabled}
        tooltipId={tooltipIds.ACTIVATION_MODE}
        onChange={onActivationModeChange}
      >
        <Radio label={sharedMessages.otaa} value={ACTIVATION_MODES.OTAA} disabled={!jsEnabled} />
        <Radio label={sharedMessages.abp} value={ACTIVATION_MODES.ABP} disabled={!nsEnabled} />
        <Radio
          label={sharedMessages.multicast}
          value={ACTIVATION_MODES.MULTICAST}
          disabled={!nsEnabled}
        />
        <Radio label={messages.activationModeNone} value={ACTIVATION_MODES.NONE} />
      </Form.Field>
      {!isNone && (
        <Form.Field
          title={messages.classCapabilities}
          required={isMulticast}
          name="_device_class"
          component={Select}
          onChange={onDeviceClassChange}
          options={isMulticast ? multicastClassOptions : allClassOptions}
          tooltipId={tooltipIds.CLASSES}
        />
      )}
      {nsEnabled && (
        <>
          {isABP && (
            <>
              <Form.FieldContainer horizontal>
                <Form.Field
                  title={messages.rx1DataRateOffsetTitle}
                  type="number"
                  name="mac_settings.rx1_data_rate_offset"
                  component={Input}
                  min={0}
                  max={7}
                  tooltipId={tooltipIds.DATA_RATE_OFFSET}
                  inputWidth="xxs"
                  autoWidth
                />
                <Form.Field
                  title={messages.rx1DelayTitle}
                  type="number"
                  description={m.rx1DelayDescription}
                  name="mac_settings.rx1_delay"
                  component={Input}
                  min={1}
                  max={15}
                  inputWidth="xxs"
                  autoWidth
                />
              </Form.FieldContainer>
            </>
          )}
          {((isClassB && !isNone) || isMulticast) && (
            <Form.FieldContainer horizontal>
              <Form.Field
                title="Class B timeout"
                name="mac_settings.class_b_timeout"
                encode={timeoutEncode}
                decode={timeoutDecode}
                component={Input}
                type="number"
                inputWidth="xxs"
                autoWidth
              />
              <Form.Field
                title={messages.pingSlotPeriodicityTitle}
                description={messages.pingSlotPeriodicityDescription}
                name="mac_settings.ping_slot_periodicity"
                component={Select}
                options={pingSlotPeriodicityOptions}
                required={isClassB && (isMulticast || isABP)}
                menuPlacement="top"
                autoWidth
              />
            </Form.FieldContainer>
          )}
          {!isNone && (
            <>
              <Form.FieldContainer horizontal>
                <Form.Field
                  title={messages.rx2DataRateIndexTitle}
                  type="number"
                  name="mac_settings.rx2_data_rate_index"
                  component={Input}
                  min={0}
                  max={15}
                  autoWidth
                />
                <Form.Field
                  type="number"
                  min={100000}
                  step={100}
                  title={messages.rx2FrequencyTitle}
                  description={messages.rx2FrequencyDescription}
                  placeholder={messages.frequencyPlaceholder}
                  name="mac_settings.rx2_frequency"
                  tooltipId={tooltipIds.RX2_FREQUENCY}
                  component={Input}
                  autoWidth
                />
              </Form.FieldContainer>
              <Form.Field
                indexAsKey
                name="mac_settings.factory_preset_frequencies"
                component={KeyValueMap}
                title={messages.factoryPresetFreqTitle}
                description={messages.factoryPresetFreqDescription}
                addMessage={messages.freqAdd}
                valuePlaceholder={messages.frequencyPlaceholder}
              />
            </>
          )}
        </>
      )}
      {!isNone && (
        <Form.Field
          title={m.useExternalServers}
          name="_external_servers"
          component={Checkbox}
          onChange={handleExternalServers}
        />
      )}
      {!isNone && externalServers && (
        <>
          <Form.Field
            title={sharedMessages.networkServerAddress}
            name="network_server_address"
            component={Input}
          />
          {isOTAA && (
            <Form.Field
              title={sharedMessages.joinServerAddress}
              name="join_server_address"
              component={Input}
            />
          )}
          {(isABP || isMulticast) && (
            <Form.Field
              title={sharedMessages.applicationServerAddress}
              name="application_server_address"
              component={Input}
            />
          )}
        </>
      )}
    </Form.CollapseSection>
  )
}

AdvancedSettingsSection.propTypes = {
  activationMode: PropTypes.oneOf(Object.values(ACTIVATION_MODES)).isRequired,
  deviceClass: PropTypes.string,
  jsEnabled: PropTypes.bool.isRequired,
  nsEnabled: PropTypes.bool.isRequired,
  onActivationModeChange: PropTypes.func.isRequired,
  onDeviceClassChange: PropTypes.func.isRequired,
}

AdvancedSettingsSection.defaultProps = {
  deviceClass: undefined,
}

export default AdvancedSettingsSection
