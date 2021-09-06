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
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Radio from '@ttn-lw/components/radio-button'

import Message from '@ttn-lw/lib/components/message'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  ACTIVATION_MODES,
  FRAME_WIDTH_COUNT,
  fCntWidthEncode,
  fCntWidthDecode,
  parseLorawanMacVersion,
} from '@console/lib/device-utils'

import style from './mac-settings-section.styl'

const m = defineMessages({
  delayValue: '{count, plural, one {{count} second} other {{count} seconds}}',
  factoryPresetFreqDescription: 'List of factory-preset frequencies. Note: order is respected.',
  factoryPresetFreqTitle: 'Factory preset frequencies',
  freqAdd: 'Add Frequency',
  frequencyPlaceholder: 'e.g. 869525000 for 869,525 MHz',
  advancedMacSettings: 'Advanced MAC settings',
  pingSlotFrequencyTitle: 'Ping slot frequency',
  desiredPingSlotFrequencyTitle: 'Desired ping slot frequency',
  pingSlotPeriodicityDescription: 'Periodicity of the class B ping slot',
  pingSlotPeriodicityTitle: 'Ping slot periodicity',
  pingSlotPeriodicityValue: '{count, plural, one {every second} other {every {count} seconds}}',
  pingSlotDataRateTitle: 'Ping slot data rate',
  desiredPingSlotDataRateTitle: 'Desired ping slot data rate',
  resetWarning: 'Resetting is insecure and makes your device susceptible for replay attacks',
  resetsFCnt: 'Resets frame counters',
  rx1DataRateOffsetTitle: 'Rx1 data rate offset',
  desiredRx1DataRateOffsetTitle: 'Desired Rx1 data rate offset',
  rx1DelayTitle: 'Rx1 delay',
  desiredRx1DelayTitle: 'Desired Rx1 delay',
  rx2DataRateIndexTitle: 'Rx2 data rate index',
  desiredRx2DataRateIndexTitle: 'Desired Rx2 data rate index',
  desiredRx2FrequencyTitle: 'Desired Rx2 frequency',
  rx2FrequencyTitle: 'Rx2 frequency',
  updateSuccess: 'The MAC settings updated',
  beaconFrequency: 'Beacon frequency',
  desiredBeaconFrequency: 'Desired beacon frequency',
  classBTimeout: 'Class B timeout',
  classCTimeout: 'Class C timeout',
  maxDutyCycle: 'Maximum duty cycle',
  desiredMaxDutyCycle: 'Desired maximum duty cycle',
  useAdr: 'Use ADR',
  adrMargin: 'ADR margin',
  desiredAdrAckLimit: 'Desired ADR ack limit',
  desiredAdrAckDelay: 'Desired ADR ack delay',
  adrAckValue: '{count, plural, one {every message} other {every {count} messages}}',
  statusCountPeriodicity: 'Status count periodicity',
  statusTimePeriodicity: 'Status time periodicity',
})

const timeoutEncode = value => (Boolean(value) ? `${value}s` : value)
const timeoutDecode = value => (Boolean(value) ? RegExp(/\d+/).exec(value)[0] : value)

// 0...7
const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={m.pingSlotPeriodicityValue} values={{ count: value }} />,
  }
})
// 0...15
const adrAckLimitOptions = Array.from({ length: 16 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `ADR_ACK_LIMIT_${value}`,
    label: <Message content={m.adrAckValue} values={{ count: value }} />,
  }
})
// 0...15
const adrAckDelayOptions = Array.from({ length: 16 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `ADR_ACK_DELAY_${value}`,
    label: <Message content={m.adrAckValue} values={{ count: value }} />,
  }
})
const maxDutyCycleOptions = [
  { value: 'DUTY_CYCLE_1', label: '100%' },
  { value: 'DUTY_CYCLE_16', label: '6.25%' },
  { value: 'DUTY_CYCLE_128', label: '0.781%' },
  { value: 'DUTY_CYCLE_1024', label: '0.098%' },
  { value: 'DUTY_CYCLE_16384', label: '0.006%' },
]

const MacSettingsSection = props => {
  const {
    activationMode,
    resetsFCnt: initialFCnt,
    initiallyCollapsed,
    lorawanVersion,
    isClassB,
    isClassC,
    isUseAdr,
  } = props

  const isNewLorawanVersion = parseLorawanMacVersion(lorawanVersion) >= 110
  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const isOTAA = activationMode === ACTIVATION_MODES.OTAA

  const [resetsFCnt, setResetsFCnt] = React.useState(isABP && initialFCnt)
  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  const pingPeriodicityRequired = isClassB && (isABP || isMulticast)

  const [isCollapsed, setIsCollapsed] = React.useState(initiallyCollapsed)
  const handleIsCollapsedChange = React.useCallback(() => {
    if (!isCollapsed && pingPeriodicityRequired) {
      // Do not close section if `ping_slot_perdiodicity` is required.
      return
    }

    setIsCollapsed(isCollapsed => !isCollapsed)
  }, [isCollapsed, pingPeriodicityRequired])

  const isUseAdrRef = React.useRef(isUseAdr)
  const [useAdr, setUseAdr] = React.useState(isUseAdr)
  const handleUseAdrChange = React.useCallback(evt => {
    const checked = evt.target.checked

    setUseAdr(checked)
  }, [])

  React.useEffect(() => {
    if (isCollapsed && pingPeriodicityRequired) {
      // Expand section if `ping_slot_periodicity` is required.
      setIsCollapsed(false)
    }
  }, [handleIsCollapsedChange, isABP, isClassB, isCollapsed, isMulticast, pingPeriodicityRequired])

  React.useEffect(() => {
    if (isUseAdrRef.current !== isUseAdr) {
      isUseAdrRef.current = isUseAdr
      setUseAdr(isUseAdr)
    }
  }, [isUseAdr])

  return (
    <Form.CollapseSection
      id="mac-settings"
      title={m.advancedMacSettings}
      initiallyCollapsed={initiallyCollapsed}
      onCollapse={handleIsCollapsedChange}
      isCollapsed={isCollapsed}
    >
      <Form.Field
        title={sharedMessages.frameCounterWidth}
        name="mac_settings.supports_32_bit_f_cnt"
        component={Radio.Group}
        encode={fCntWidthEncode}
        decode={fCntWidthDecode}
        tooltipId={tooltipIds.FRAME_COUNTER_WIDTH}
      >
        <Radio label={sharedMessages['16Bit']} value={FRAME_WIDTH_COUNT.SUPPORTS_16_BIT} />
        <Radio label={sharedMessages['32Bit']} value={FRAME_WIDTH_COUNT.SUPPORTS_32_BIT} />
      </Form.Field>
      {(isABP || isOTAA) && (
        <>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                title={m.rx1DelayTitle}
                type="number"
                tooltipId={tooltipIds.RX1_DELAY}
                name="mac_settings.rx1_delay"
                component={Input}
                min={1}
                max={15}
                inputWidth="xxs"
              />
            )}
            <Form.Field
              title={m.desiredRx1DelayTitle}
              type="number"
              name="mac_settings.desired_rx1_delay"
              tooltipId={tooltipIds.RX1_DELAY}
              component={Input}
              min={1}
              max={15}
              inputWidth="xxs"
            />
          </Form.FieldContainer>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                title={m.rx1DataRateOffsetTitle}
                type="number"
                name="mac_settings.rx1_data_rate_offset"
                inputWidth="xxs"
                component={Input}
                min={0}
                max={7}
                tooltipId={tooltipIds.DATA_RATE_OFFSET}
              />
            )}
            <Form.Field
              title={m.desiredRx1DataRateOffsetTitle}
              type="number"
              inputWidth="xxs"
              name="mac_settings.desired_rx1_data_rate_offset"
              component={Input}
              min={0}
              max={7}
              tooltipId={tooltipIds.DATA_RATE_OFFSET}
            />
          </Form.FieldContainer>
          {!isOTAA && (
            <Form.Field
              title={sharedMessages.resetsFCnt}
              onChange={handleResetsFCntChange}
              warning={resetsFCnt ? m.resetWarning : undefined}
              name="mac_settings.resets_f_cnt"
              tooltipId={tooltipIds.RESETS_F_CNT}
              component={Checkbox}
            />
          )}
        </>
      )}
      <Form.FieldContainer horizontal>
        {!isOTAA && (
          <Form.Field
            title={m.rx2DataRateIndexTitle}
            type="number"
            name="mac_settings.rx2_data_rate_index"
            component={Input}
            min={0}
            max={15}
            tooltipId={tooltipIds.RX2_DATA_RATE_INDEX}
            inputWidth="xxs"
          />
        )}
        <Form.Field
          title={m.desiredRx2DataRateIndexTitle}
          type="number"
          name="mac_settings.desired_rx2_data_rate_index"
          component={Input}
          min={0}
          max={15}
          inputWidth="xxs"
          tooltipId={tooltipIds.RX2_DATA_RATE_INDEX}
        />
      </Form.FieldContainer>
      <Form.FieldContainer horizontal>
        {!isOTAA && (
          <Form.Field
            type="number"
            min={100000}
            step={100}
            title={m.rx2FrequencyTitle}
            placeholder={m.frequencyPlaceholder}
            name="mac_settings.rx2_frequency"
            component={Input}
            tooltipId={tooltipIds.RX2_FREQUENCY}
          />
        )}
        <Form.Field
          type="number"
          min={100000}
          step={100}
          title={m.desiredRx2FrequencyTitle}
          placeholder={m.frequencyPlaceholder}
          name="mac_settings.desired_rx2_frequency"
          component={Input}
          tooltipId={tooltipIds.RX2_FREQUENCY}
        />
      </Form.FieldContainer>
      <Form.FieldContainer horizontal>
        {!isOTAA && (
          <Form.Field
            title={m.maxDutyCycle}
            name="mac_settings.max_duty_cycle"
            component={Select}
            options={maxDutyCycleOptions}
            inputWidth="xs"
            tooltipId={tooltipIds.MAX_DUTY_CYCLE}
          />
        )}
        <Form.Field
          title={m.desiredMaxDutyCycle}
          name="mac_settings.desired_max_duty_cycle"
          component={Select}
          options={maxDutyCycleOptions}
          inputWidth="xs"
          tooltipId={tooltipIds.MAX_DUTY_CYCLE}
        />
      </Form.FieldContainer>
      <Form.Field
        indexAsKey
        name="mac_settings.factory_preset_frequencies"
        component={KeyValueMap}
        title={m.factoryPresetFreqTitle}
        description={m.factoryPresetFreqDescription}
        addMessage={m.freqAdd}
        valuePlaceholder={m.frequencyPlaceholder}
        tooltipId={tooltipIds.FACTORY_PRESET_FREQUENCIES}
      />
      {isClassC && (
        <Form.Field
          className={style.smallField}
          title={m.classCTimeout}
          name="mac_settings.class_c_timeout"
          tooltipId={tooltipIds.CLASS_C_TIMEOUT}
          encode={timeoutEncode}
          decode={timeoutDecode}
          component={Input}
          type="number"
          inputWidth="xxs"
        />
      )}
      {(isClassB || isMulticast) && (
        <>
          <Form.Field
            className={style.smallField}
            title={m.classBTimeout}
            name="mac_settings.class_b_timeout"
            tooltipId={tooltipIds.CLASS_B_TIMEOUT}
            encode={timeoutEncode}
            decode={timeoutDecode}
            component={Input}
            type="number"
            inputWidth="xxs"
          />
          <Form.Field
            title={m.pingSlotPeriodicityTitle}
            description={m.pingSlotPeriodicityDescription}
            name="mac_settings.ping_slot_periodicity"
            component={Select}
            options={pingSlotPeriodicityOptions}
            required={pingPeriodicityRequired}
            menuPlacement="top"
          />
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                type="number"
                min={100000}
                title={m.beaconFrequency}
                placeholder={m.frequencyPlaceholder}
                name="mac_settings.beacon_frequency"
                tooltipId={tooltipIds.BEACON_FREQUENCY}
                component={Input}
              />
            )}
            <Form.Field
              type="number"
              min={100000}
              title={m.desiredBeaconFrequency}
              placeholder={m.frequencyPlaceholder}
              name="mac_settings.desired_beacon_frequency"
              tooltipId={tooltipIds.BEACON_FREQUENCY}
              component={Input}
            />
          </Form.FieldContainer>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                type="number"
                min={100000}
                step={100}
                title={m.pingSlotFrequencyTitle}
                placeholder={m.frequencyPlaceholder}
                name="mac_settings.ping_slot_frequency"
                tooltipId={tooltipIds.PING_SLOT_FREQUENCY}
                component={Input}
              />
            )}
            <Form.Field
              type="number"
              min={100000}
              step={100}
              title={m.desiredPingSlotFrequencyTitle}
              placeholder={m.frequencyPlaceholder}
              name="mac_settings.desired_ping_slot_frequency"
              tooltipId={tooltipIds.PING_SLOT_FREQUENCY}
              component={Input}
            />
          </Form.FieldContainer>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                title={m.pingSlotDataRateTitle}
                name="mac_settings.ping_slot_data_rate_index"
                tooltipId={tooltipIds.PING_SLOT_DATA_RATE_INDEX}
                component={Input}
                type="number"
                min={0}
                max={15}
              />
            )}
            <Form.Field
              title={m.desiredPingSlotDataRateTitle}
              name="mac_settings.desired_ping_slot_data_rate_index"
              tooltipId={tooltipIds.PING_SLOT_DATA_RATE_INDEX}
              component={Input}
              type="number"
              min={0}
              max={15}
            />
          </Form.FieldContainer>
        </>
      )}
      <Form.FieldContainer horizontal>
        <Form.Field
          className={style.smallField}
          title={m.statusCountPeriodicity}
          name="mac_settings.status_count_periodicity"
          component={Input}
          type="number"
          inputWidth="xxs"
          tooltipId={tooltipIds.STATUS_COUNT_PERIODICITY}
        />
        <Form.Field
          className={style.smallField}
          title={m.statusTimePeriodicity}
          name="mac_settings.status_time_periodicity"
          encode={timeoutEncode}
          decode={timeoutDecode}
          component={Input}
          type="number"
          inputWidth="xxs"
          tooltipId={tooltipIds.STATUS_TIME_PERIODICITY}
        />
      </Form.FieldContainer>
      <Form.Field
        name="mac_settings.use_adr"
        component={Checkbox}
        label={m.useAdr}
        onChange={handleUseAdrChange}
        tooltipId={tooltipIds.ADR_USE}
      />
      {useAdr && (
        <>
          <Form.Field
            title={m.adrMargin}
            name="mac_settings.adr_margin"
            component={Input}
            type="number"
            tooltipId={tooltipIds.ADR_MARGIN}
          />
          {isNewLorawanVersion && (
            <>
              <Form.Field
                title={m.desiredAdrAckLimit}
                name="mac_settings.desired_adr_ack_limit_exponent"
                component={Select}
                options={adrAckLimitOptions}
                tooltipId={tooltipIds.ADR_ACK_LIMIT}
              />
              <Form.Field
                title={m.desiredAdrAckDelay}
                name="mac_settings.desired_adr_ack_delay_exponent"
                component={Select}
                options={adrAckDelayOptions}
                tooltipId={tooltipIds.ADR_ACK_DELAY}
              />
            </>
          )}
        </>
      )}
    </Form.CollapseSection>
  )
}

MacSettingsSection.propTypes = {
  activationMode: PropTypes.oneOf(Object.values(ACTIVATION_MODES)).isRequired,
  initiallyCollapsed: PropTypes.bool,
  isClassB: PropTypes.bool,
  isClassC: PropTypes.bool,
  isUseAdr: PropTypes.bool,
  lorawanVersion: PropTypes.string.isRequired,
  resetsFCnt: PropTypes.bool,
}

MacSettingsSection.defaultProps = {
  resetsFCnt: false,
  initiallyCollapsed: true,
  isClassB: false,
  isClassC: false,
  isUseAdr: false,
}

export default MacSettingsSection
