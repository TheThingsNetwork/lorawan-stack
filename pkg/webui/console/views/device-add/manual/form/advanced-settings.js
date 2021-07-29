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

import { ACTIVATION_MODES, hasCFListTypeChMask } from '@console/lib/device-utils'

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
  multicastClassCapabilities: 'LoRaWAN class for multicast downlinks',
  factoryFreqWarning:
    'In LoRaWAN, factory preset frequencies are only supported for bands with a CFList type of frequencies',
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
    onDefaultNsSettingsChange,
    defaultNsSettings,
    freqPlan,
  } = props

  const isOTAA = activationMode === ACTIVATION_MODES.OTAA
  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const isClassB =
    deviceClass === DEVICE_CLASS_MAP.CLASS_B_C || deviceClass === DEVICE_CLASS_MAP.CLASS_B
  const isClassC =
    deviceClass === DEVICE_CLASS_MAP.CLASS_B_C || deviceClass === DEVICE_CLASS_MAP.CLASS_C

  // The technical difference between bands that do support factory preset frequencies
  // and bands that do not support them, is that the former uses a CFList type of Frequencies,
  // and the latter uses a CFList type of ChMask (channel mask).
  // When there is a channel mask, the frequencies aren't configured by frequency in Hertz,
  // but by index. The factory preset frequencies is really the frequencies in Hertz,
  // so it requires bands with a CFList type of Frequencies.
  const disableFactoryPresetFreq = hasCFListTypeChMask(freqPlan)

  const [externalServers, setExternalServer] = React.useState(false)
  const handleExternalServers = React.useCallback(
    () => setExternalServer(external => !external),
    [],
  )

  const handleDefaultNsSettings = React.useCallback(
    evt => onDefaultNsSettingsChange(evt.target.checked),
    [onDefaultNsSettingsChange],
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
      </Form.Field>
      <Form.Field
        title={isMulticast ? m.multicastClassCapabilities : messages.classCapabilities}
        required={isMulticast}
        name="_device_class"
        component={Select}
        onChange={onDeviceClassChange}
        options={isMulticast ? multicastClassOptions : allClassOptions}
        tooltipId={tooltipIds.CLASSES}
      />
      <Form.Field
        title={messages.networkDefaults}
        label={messages.defaultNetworksSettings}
        name="_default_ns_settings"
        component={Checkbox}
        onChange={handleDefaultNsSettings}
        tooltipId={tooltipIds.NETWORK_RX_DEFAULTS}
      />
      {!defaultNsSettings && nsEnabled && (
        <>
          {isABP && (
            <>
              <Form.FieldContainer horizontal>
                <Form.Field
                  className={style.smallField}
                  title={messages.rx1DataRateOffsetTitle}
                  type="number"
                  name="mac_settings.rx1_data_rate_offset"
                  component={Input}
                  min={0}
                  max={7}
                  tooltipId={tooltipIds.DATA_RATE_OFFSET}
                  inputWidth="xxs"
                />
                <Form.Field
                  title={messages.rx1DelayTitle}
                  type="number"
                  description={m.rx1DelayDescription}
                  name="mac_settings.rx1_delay"
                  tooltipId={tooltipIds.RX1_DELAY}
                  component={Input}
                  min={1}
                  max={15}
                  inputWidth="xxs"
                  autoWidth
                />
              </Form.FieldContainer>
            </>
          )}
          {(isClassB || isMulticast) && (
            <Form.FieldContainer horizontal>
              <Form.Field
                className={style.smallField}
                title={messages.classBTimeout}
                name="mac_settings.class_b_timeout"
                tooltipId={tooltipIds.CLASS_B_TIMEOUT}
                encode={timeoutEncode}
                decode={timeoutDecode}
                component={Input}
                type="number"
                inputWidth="xxs"
              />
              <Form.Field
                title={messages.pingSlotPeriodicityTitle}
                name="mac_settings.ping_slot_periodicity"
                tooltipId={tooltipIds.PING_SLOT_PERIODICITY}
                component={Select}
                options={pingSlotPeriodicityOptions}
                required={isClassB && (isMulticast || isABP)}
                menuPlacement="top"
                autoWidth
              />
            </Form.FieldContainer>
          )}
          <Form.FieldContainer horizontal>
            {isClassC && (
              <Form.Field
                className={style.smallField}
                title={messages.classCTimeout}
                name="mac_settings.class_c_timeout"
                encode={timeoutEncode}
                decode={timeoutDecode}
                tooltipId={tooltipIds.CLASS_C_TIMEOUT}
                component={Input}
                type="number"
                inputWidth="xxs"
              />
            )}
            <Form.Field
              className={style.smallField}
              title={messages.rx2DataRateIndexTitle}
              type="number"
              name="mac_settings.rx2_data_rate_index"
              tooltipId={tooltipIds.RX2_DATA_RATE_INDEX}
              component={Input}
              min={0}
              max={15}
            />
            <Form.Field
              type="number"
              min={100000}
              step={100}
              title={messages.rx2FrequencyTitle}
              placeholder={messages.frequencyPlaceholder}
              name="mac_settings.rx2_frequency"
              tooltipId={tooltipIds.RX2_FREQUENCY}
              component={Input}
              autoWidth
            />
          </Form.FieldContainer>
          {!isOTAA && (
            <Form.Field
              indexAsKey
              disabled={disableFactoryPresetFreq}
              name="mac_settings.factory_preset_frequencies"
              description={disableFactoryPresetFreq ? m.factoryFreqWarning : undefined}
              component={KeyValueMap}
              title={messages.factoryPresetFreqTitle}
              addMessage={messages.freqAdd}
              valuePlaceholder={messages.frequencyPlaceholder}
              tooltipId={tooltipIds.FACTORY_PRESET_FREQUENCIES}
            />
          )}
        </>
      )}
      <Form.Field
        title={messages.clusterSettings}
        label={m.useExternalServers}
        name="_external_servers"
        component={Checkbox}
        onChange={handleExternalServers}
        tooltipId={tooltipIds.CLUSTER_SETTINGS}
      />
      {externalServers && (
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
  defaultNsSettings: PropTypes.bool.isRequired,
  deviceClass: PropTypes.string,
  freqPlan: PropTypes.string,
  jsEnabled: PropTypes.bool.isRequired,
  nsEnabled: PropTypes.bool.isRequired,
  onActivationModeChange: PropTypes.func.isRequired,
  onDefaultNsSettingsChange: PropTypes.func.isRequired,
  onDeviceClassChange: PropTypes.func.isRequired,
}

AdvancedSettingsSection.defaultProps = {
  deviceClass: undefined,
  freqPlan: undefined,
}

export default AdvancedSettingsSection
