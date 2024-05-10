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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import { createSelector } from 'reselect'
import { useSelector } from 'react-redux'
import { get, set } from 'lodash'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Radio from '@ttn-lw/components/radio-button'
import UnitInput from '@ttn-lw/components/unit-input'
import Button from '@ttn-lw/components/button'
import Icon from '@ttn-lw/components/icon'

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
import getDataRate from '@console/lib/data-rate-utils'

import { selectDataRates } from '@console/store/selectors/configuration'

const m = defineMessages({
  delayValue: '{count, plural, one {{count} second} other {{count} seconds}}',
  factoryPresetFreqDescription: 'List of factory-preset frequencies. Note: order is respected.',
  advancedMacSettings: 'Advanced MAC settings',
  desiredPingSlotFrequencyTitle: 'Desired ping slot frequency',
  pingSlotPeriodicityDescription: 'Periodicity of the class B ping slot',
  pingSlotDataRateTitle: 'Ping slot data rate index',
  desiredPingSlotDataRateTitle: 'Desired ping slot data rate',
  resetWarning: 'Resetting is insecure and makes your device susceptible for replay attacks',
  desiredRx1DataRateOffsetTitle: 'Desired Rx1 data rate offset',
  desiredRx1DelayTitle: 'Desired Rx1 delay',
  rx2DataRateIndexTitle: 'Rx2 data rate index',
  desiredRx2DataRateIndexTitle: 'Desired Rx2 data rate index',
  desiredRx2FrequencyTitle: 'Desired Rx2 frequency',
  updateSuccess: 'The MAC settings updated',
  desiredBeaconFrequency: 'Desired beacon frequency',
  maxDutyCycle: 'Maximum duty cycle',
  desiredMaxDutyCycle: 'Desired maximum duty cycle',
  adrMargin: 'ADR margin',
  adrUplinks: 'ADR number of transmissions',
  adrAdaptiveDataRate: 'Adaptive data rate (ADR)',
  adrDataRate: 'ADR data rate index',
  adrTransPower: 'ADR transmission power index',
  adrDynamic: 'Dynamic mode',
  adrStatic: 'Static mode',
  desiredAdrAckLimit: 'Desired ADR ack limit',
  desiredAdrAckDelay: 'Desired ADR ack delay',
  adrAckValue: '{count, plural, one {every message} other {every {count} messages}}',
  statusCountPeriodicity: 'Status count periodicity',
  statusTimePeriodicity: 'Status time periodicity',
  dataRate: 'Data Rate {n}',
  dataRatePlaceholder: 'Data Rate',
  minNbTrans: 'Min. NbTrans',
  maxNbTrans: 'Max. NbTrans',
  useDefaultNbTrans: 'Use default settings for number of retransmissions',
  adrNbTrans: 'ADR number of retransmissions (NbTrans)',
  overrideNbTrans: 'Override server defaults for NbTrans (all data rates)',
  defaultForAllRates: '(Default for all data rates)',
  defaultNbTransMessage:
    'Overriding the default is not required for using data rate overrides (below)',
  specificOverrides: 'Data rate specific overrides',
  addSpecificOverride: 'Add data rate specific override',
})

// 0...7
const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={sharedMessages.secondInterval} values={{ count: value }} />,
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

const encodeAdrMode = value => ({ [value]: {} })
const decodeAdrMode = value => (value !== undefined ? Object.keys(value)[0] : null)

const MacSettingsSection = props => {
  const {
    activationMode,
    resetsFCnt: initialFCnt,
    initiallyCollapsed,
    lorawanVersion,
    isClassB,
    isClassC,
    bandId,
  } = props

  const { values, setFieldValue, setFieldTouched } = useFormContext()
  const { mac_settings } = values
  const alreadySelectedDataRates = Object.keys(mac_settings?.adr?.dynamic?.overrides || [])
  const dataRateOverrideOptions = useSelector(
    createSelector(
      state => selectDataRates(state, bandId, values.lorawan_phy_version),
      dataRates =>
        Object.keys(dataRates).reduce(
          (result, key) =>
            result.concat({
              label: getDataRate({ settings: { data_rate: dataRates[key].rate } }),
              value: `data_rate_${key}`,
            }),
          [],
        ),
    ),
  )
  // Filter out the already selected data rate indices.
  const dataRateFilterOption = useCallback(
    option => !alreadySelectedDataRates.includes(option.value),
    [alreadySelectedDataRates],
  )

  const isNewLorawanVersion = parseLorawanMacVersion(lorawanVersion) >= 110
  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const isOTAA = activationMode === ACTIVATION_MODES.OTAA
  const isDynamicAdr = mac_settings.adr && 'dynamic' in mac_settings.adr
  const isStaticAdr = mac_settings.adr && 'static' in mac_settings.adr
  const [resetsFCnt, setResetsFCnt] = React.useState(isABP && initialFCnt)
  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  const pingPeriodicityRequired = isClassB && (isABP || isMulticast)

  const [isCollapsed, setIsCollapsed] = React.useState(initiallyCollapsed)
  const handleIsCollapsedChange = React.useCallback(() => {
    if (!isCollapsed && pingPeriodicityRequired) {
      // Do not close section if `ping_slot_periodicity` is required.
      return
    }

    setIsCollapsed(isCollapsed => !isCollapsed)
  }, [isCollapsed, pingPeriodicityRequired])

  React.useEffect(() => {
    if (isCollapsed && pingPeriodicityRequired) {
      // Expand section if `ping_slot_periodicity` is required.
      setIsCollapsed(false)
    }
  }, [handleIsCollapsedChange, isABP, isClassB, isCollapsed, isMulticast, pingPeriodicityRequired])

  const adrOverrides = mac_settings.adr.dynamic?.overrides
  const showEditNbTrans = !values.mac_settings.adr.dynamic?._use_default_nb_trans
  const defaultNbTransDisabled = !values.mac_settings.adr.dynamic?._override_nb_trans_defaults
  const addOverride = React.useCallback(() => {
    const newOverride = { _data_rate_index: '', min_nb_trans: '', max_nb_trans: '' }
    setFieldValue(
      'mac_settings.adr.dynamic.overrides',
      adrOverrides
        ? { ...adrOverrides, [`_empty-${Date.now()}`]: newOverride }
        : { [`_empty-${Date.now()}`]: newOverride },
    )
    setFieldTouched('mac_settings.adr.dynamic._overrides', true)
  }, [setFieldValue, adrOverrides, setFieldTouched])
  const handleRemoveButtonClick = useCallback(
    (_, index) => {
      setFieldValue(
        'mac_settings.adr.dynamic.overrides',
        Object.keys(adrOverrides)
          .filter(key => key !== index)
          .reduce((acc, key) => ({ ...acc, [key]: adrOverrides[key] }), {}),
      )
    },
    [adrOverrides, setFieldValue],
  )

  // Define a value setter for the data rate index field which
  // handles setting the object keys correctly, since the index
  // is set as the object key in the API schema.
  // A similar result could be done without pseudo values, purely
  // with decoder/encoder, but it would make error mapping
  // more complex.
  const dataRateValueSetter = useCallback(
    ({ setValues }, { name, value }) => {
      const index = name.split('.').slice(-2)[0] // Would be: data_rate_{x}.
      const oldOverride = get(values, `mac_settings.adr.dynamic.overrides.${index}`, {})
      const overrides = { ...get(values, 'mac_settings.adr.dynamic.overrides', {}) }
      // Empty data rate index objects, are stored with a pseudo key. Remove it.
      delete overrides[index]
      // Move the existing values to the new data rate key.
      overrides[value] = { ...oldOverride, _data_rate_index: value }
      setValues(values => set(values, 'mac_settings.adr.dynamic.overrides', overrides))
    },
    [values],
  )

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
        horizontal
      >
        <Radio label={sharedMessages['16Bit']} value={FRAME_WIDTH_COUNT.SUPPORTS_16_BIT} />
        <Radio label={sharedMessages['32Bit']} value={FRAME_WIDTH_COUNT.SUPPORTS_32_BIT} />
      </Form.Field>
      {(isABP || isOTAA) && (
        <>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                title={sharedMessages.rx1Delay}
                type="number"
                tooltipId={tooltipIds.RX1_DELAY}
                append={<Message content={sharedMessages.secondsAbbreviated} />}
                name="mac_settings.rx1_delay"
                component={Input}
                min={1}
                max={15}
                inputWidth="xs"
                fieldWidth="xs"
              />
            )}
            {!isMulticast && (
              <Form.Field
                title={m.desiredRx1DelayTitle}
                type="number"
                name="mac_settings.desired_rx1_delay"
                append={<Message content={sharedMessages.secondsAbbreviated} />}
                tooltipId={tooltipIds.RX1_DELAY}
                component={Input}
                min={1}
                max={15}
                inputWidth="xs"
                fieldWidth="xs"
              />
            )}
          </Form.FieldContainer>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                title={sharedMessages.rx1DataRateOffset}
                type="number"
                name="mac_settings.rx1_data_rate_offset"
                inputWidth="xxs"
                fieldWidth="xs"
                component={Input}
                min={0}
                max={7}
                tooltipId={tooltipIds.DATA_RATE_OFFSET}
              />
            )}
            {!isMulticast && (
              <Form.Field
                title={m.desiredRx1DataRateOffsetTitle}
                type="number"
                inputWidth="xxs"
                fieldWidth="xs"
                name="mac_settings.desired_rx1_data_rate_offset"
                component={Input}
                min={0}
                max={7}
                tooltipId={tooltipIds.DATA_RATE_OFFSET}
              />
            )}
          </Form.FieldContainer>
          {!isOTAA && (
            <Form.Field
              label={sharedMessages.resetsFCnt}
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
            fieldWidth="xs"
          />
        )}
        {!isMulticast && (
          <Form.Field
            title={m.desiredRx2DataRateIndexTitle}
            type="number"
            name="mac_settings.desired_rx2_data_rate_index"
            component={Input}
            min={0}
            max={15}
            inputWidth="xxs"
            tooltipId={tooltipIds.RX2_DATA_RATE_INDEX}
            fieldWidth="xs"
          />
        )}
      </Form.FieldContainer>
      <Form.FieldContainer horizontal>
        {!isOTAA && (
          <Form.Field
            type="number"
            min={100000}
            step={100}
            title={sharedMessages.rx2Frequency}
            name="mac_settings.rx2_frequency"
            component={UnitInput.Hertz}
            tooltipId={tooltipIds.RX2_FREQUENCY}
            fieldWidth="xs"
          />
        )}
        {!isMulticast && (
          <Form.Field
            type="number"
            min={100000}
            step={100}
            title={m.desiredRx2FrequencyTitle}
            name="mac_settings.desired_rx2_frequency"
            component={UnitInput.Hertz}
            tooltipId={tooltipIds.RX2_FREQUENCY}
            fieldWidth="xs"
          />
        )}
      </Form.FieldContainer>
      <Form.FieldContainer horizontal>
        {!isOTAA && (
          <Form.Field
            title={m.maxDutyCycle}
            name="mac_settings.max_duty_cycle"
            component={Select}
            options={maxDutyCycleOptions}
            fieldWidth="xs"
            tooltipId={tooltipIds.MAX_DUTY_CYCLE}
          />
        )}
        {!isMulticast && (
          <Form.Field
            title={m.desiredMaxDutyCycle}
            name="mac_settings.desired_max_duty_cycle"
            component={Select}
            options={maxDutyCycleOptions}
            fieldWidth="xs"
            tooltipId={tooltipIds.MAX_DUTY_CYCLE}
          />
        )}
      </Form.FieldContainer>
      <Form.Field
        indexAsKey
        name="mac_settings.factory_preset_frequencies"
        component={KeyValueMap}
        title={sharedMessages.factoryPresetFrequencies}
        description={m.factoryPresetFreqDescription}
        addMessage={sharedMessages.freqAdd}
        valuePlaceholder={sharedMessages.frequencyPlaceholder}
        tooltipId={tooltipIds.FACTORY_PRESET_FREQUENCIES}
      />
      {isClassC && (
        <Form.Field
          title={sharedMessages.classCTimeout}
          name="mac_settings.class_c_timeout"
          tooltipId={tooltipIds.CLASS_C_TIMEOUT}
          component={UnitInput.Duration}
          unitSelector={['ms', 's']}
          type="number"
          fieldWidth="xs"
        />
      )}
      {(isClassB || isMulticast) && (
        <>
          <Form.Field
            title={sharedMessages.classBTimeout}
            name="mac_settings.class_b_timeout"
            tooltipId={tooltipIds.CLASS_B_TIMEOUT}
            component={UnitInput.Duration}
            unitSelector={['ms', 's']}
            type="number"
            fieldWidth="xs"
          />
          <Form.Field
            title={sharedMessages.pingSlotPeriodicity}
            description={m.pingSlotPeriodicityDescription}
            name="mac_settings.ping_slot_periodicity"
            component={Select}
            options={pingSlotPeriodicityOptions}
            required={pingPeriodicityRequired}
            menuPlacement="top"
            fieldWidth="xs"
          />
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                type="number"
                min={100000}
                title={sharedMessages.beaconFrequency}
                placeholder={sharedMessages.frequencyPlaceholder}
                name="mac_settings.beacon_frequency"
                tooltipId={tooltipIds.BEACON_FREQUENCY}
                component={UnitInput.Hertz}
                fieldWidth="xs"
              />
            )}
            {!isMulticast && (
              <Form.Field
                type="number"
                min={100000}
                title={m.desiredBeaconFrequency}
                placeholder={sharedMessages.frequencyPlaceholder}
                name="mac_settings.desired_beacon_frequency"
                tooltipId={tooltipIds.BEACON_FREQUENCY}
                component={UnitInput.Hertz}
                fieldWidth="xs"
              />
            )}
          </Form.FieldContainer>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                type="number"
                min={100000}
                step={100}
                title={sharedMessages.pingSlotFrequency}
                placeholder={sharedMessages.frequencyPlaceholder}
                name="mac_settings.ping_slot_frequency"
                tooltipId={tooltipIds.PING_SLOT_FREQUENCY}
                component={UnitInput.Hertz}
                fieldWidth="xs"
              />
            )}
            {!isMulticast && (
              <Form.Field
                type="number"
                min={100000}
                step={100}
                title={m.desiredPingSlotFrequencyTitle}
                placeholder={sharedMessages.frequencyPlaceholder}
                name="mac_settings.desired_ping_slot_frequency"
                tooltipId={tooltipIds.PING_SLOT_FREQUENCY}
                component={UnitInput.Hertz}
                fieldWidth="xs"
              />
            )}
          </Form.FieldContainer>
          <Form.FieldContainer horizontal>
            {!isOTAA && (
              <Form.Field
                title={m.pingSlotDataRateTitle}
                name="mac_settings.ping_slot_data_rate_index"
                tooltipId={tooltipIds.PING_SLOT_DATA_RATE_INDEX}
                component={Input}
                type="number"
                inputWidth="xxs"
                fieldWidth="xs"
                min={0}
                max={15}
              />
            )}
            {!isMulticast && (
              <Form.Field
                title={m.desiredPingSlotDataRateTitle}
                name="mac_settings.desired_ping_slot_data_rate_index"
                tooltipId={tooltipIds.PING_SLOT_DATA_RATE_INDEX}
                component={Input}
                type="number"
                fieldWidth="xs"
                inputWidth="xxs"
                min={0}
                max={15}
              />
            )}
          </Form.FieldContainer>
        </>
      )}
      <Form.FieldContainer horizontal>
        <Form.Field
          title={m.statusCountPeriodicity}
          name="mac_settings.status_count_periodicity"
          component={Input}
          append={<Message content={sharedMessages.messages} />}
          type="number"
          inputWidth="s"
          fieldWidth="xs"
          tooltipId={tooltipIds.STATUS_COUNT_PERIODICITY}
        />
        <Form.Field
          title={m.statusTimePeriodicity}
          name="mac_settings.status_time_periodicity"
          component={UnitInput.Duration}
          unitSelector={['ms', 's']}
          type="number"
          tooltipId={tooltipIds.STATUS_TIME_PERIODICITY}
          fieldWidth="xs"
        />
      </Form.FieldContainer>
      <Form.Field
        name="mac_settings.adr"
        component={Radio.Group}
        title={m.adrAdaptiveDataRate}
        tooltipId={tooltipIds.ADR_USE}
        encode={encodeAdrMode}
        decode={decodeAdrMode}
      >
        <Radio label={m.adrDynamic} value="dynamic" />
        <Radio label={m.adrStatic} value="static" />
        <Radio label={sharedMessages.disabled} value="disabled" />
      </Form.Field>
      {isDynamicAdr && (
        <>
          <Form.Field
            title={m.adrMargin}
            name="mac_settings.adr.dynamic.margin"
            component={Input}
            type="number"
            tooltipId={tooltipIds.ADR_MARGIN}
            min={-100}
            max={100}
            inputWidth="xs"
            append="dB"
          />
          <Form.Field
            label={m.useDefaultNbTrans}
            name="mac_settings.adr.dynamic._use_default_nb_trans"
            component={Checkbox}
            tooltipId={tooltipIds.USE_DEFAULT_NB_TRANS}
          />
          {showEditNbTrans && (
            <>
              <Form.Field
                title={m.adrNbTrans}
                name="mac_settings.adr.dynamic._override_nb_trans_defaults"
                component={Checkbox}
                label={m.overrideNbTrans}
              />
              <Form.FieldContainer horizontal className="al-end mb-cs-xs">
                <Form.Field
                  title={m.minNbTrans}
                  name="mac_settings.adr.dynamic.min_nb_trans"
                  component={Input}
                  type="number"
                  min={1}
                  max={3}
                  disabled={defaultNbTransDisabled}
                  inputWidth="xs"
                  className="d-flex direction-column"
                />
                <Form.Field
                  title={m.maxNbTrans}
                  name="mac_settings.adr.dynamic.max_nb_trans"
                  component={Input}
                  type="number"
                  min={1}
                  max={3}
                  disabled={defaultNbTransDisabled}
                  inputWidth="xs"
                  className="d-flex direction-column"
                />
                <Message content={m.defaultForAllRates} className="mt-cs-xl" />
              </Form.FieldContainer>
              {!defaultNbTransDisabled && (
                <div>
                  <Icon icon="info" nudgeUp className="mr-cs-xxs" />
                  <Message content={m.defaultNbTransMessage} />
                </div>
              )}
              <Form.InfoField
                title={m.specificOverrides}
                tooltipId={tooltipIds.DATA_RATE_SPECIFIC_OVERRIDES}
                className="mt-cs-m"
              >
                {adrOverrides &&
                  Object.keys(adrOverrides).map(index => (
                    <Form.FieldContainer horizontal className="al-end" key={index}>
                      <Form.Field
                        title={m.dataRatePlaceholder}
                        name={`mac_settings.adr.dynamic.overrides.${index}._data_rate_index`}
                        valueSetter={dataRateValueSetter}
                        component={Select}
                        options={dataRateOverrideOptions}
                        filterOption={dataRateFilterOption}
                        inputWidth="s"
                        fieldWidth="xxs"
                        className="d-flex direction-column"
                      />
                      <Form.Field
                        title={m.minNbTrans}
                        name={`mac_settings.adr.dynamic.overrides.${index}.min_nb_trans`}
                        component={Input}
                        fieldWidth="xxs"
                        className="d-flex direction-column"
                        type="number"
                        min={1}
                        max={3}
                      />
                      <Form.Field
                        title={m.maxNbTrans}
                        name={`mac_settings.adr.dynamic.overrides.${index}.max_nb_trans`}
                        component={Input}
                        fieldWidth="xxs"
                        className="d-flex direction-column"
                        type="number"
                        min={1}
                        max={3}
                      />
                      <Button
                        type="button"
                        onClick={handleRemoveButtonClick}
                        icon="delete"
                        message={sharedMessages.remove}
                        value={index}
                      />
                    </Form.FieldContainer>
                  ))}
                <Button
                  type="button"
                  message={m.addSpecificOverride}
                  onClick={addOverride}
                  icon="add"
                />
              </Form.InfoField>
            </>
          )}
        </>
      )}
      {isStaticAdr && (
        <>
          <Form.Field
            title={m.adrDataRate}
            name="mac_settings.adr.static.data_rate_index"
            component={Input}
            type="number"
            inputWidth="xs"
          />
          <Form.Field
            title={m.adrTransPower}
            name="mac_settings.adr.static.tx_power_index"
            component={Input}
            type="number"
            inputWidth="xs"
          />
          <Form.Field
            title={m.adrUplinks}
            name="mac_settings.adr.static.nb_trans"
            component={Input}
            type="number"
            inputWidth="xs"
          />
        </>
      )}
      {isNewLorawanVersion && !isMulticast && (
        <>
          <Form.Field
            title={m.desiredAdrAckLimit}
            name="mac_settings.desired_adr_ack_limit_exponent"
            component={Select}
            options={adrAckLimitOptions}
            tooltipId={tooltipIds.ADR_ACK_LIMIT}
            fieldWidth="xs"
          />
          <Form.Field
            title={m.desiredAdrAckDelay}
            name="mac_settings.desired_adr_ack_delay_exponent"
            component={Select}
            options={adrAckDelayOptions}
            tooltipId={tooltipIds.ADR_ACK_DELAY}
            fieldWidth="xs"
          />
        </>
      )}
    </Form.CollapseSection>
  )
}

MacSettingsSection.propTypes = {
  activationMode: PropTypes.oneOf(Object.values(ACTIVATION_MODES)).isRequired,
  bandId: PropTypes.string.isRequired,
  initiallyCollapsed: PropTypes.bool,
  isClassB: PropTypes.bool,
  isClassC: PropTypes.bool,
  lorawanVersion: PropTypes.string.isRequired,
  resetsFCnt: PropTypes.bool,
}

MacSettingsSection.defaultProps = {
  resetsFCnt: false,
  initiallyCollapsed: true,
  isClassB: false,
  isClassC: false,
}

export default MacSettingsSection
