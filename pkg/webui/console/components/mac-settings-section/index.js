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

import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  ACTIVATION_MODES,
  FRAME_WIDTH_COUNT,
  fCntWidthEncode,
  fCntWidthDecode,
} from '@console/lib/device-utils'

const m = defineMessages({
  delayValue: '{count, plural, one {{count} second} other {{count} seconds}}',
  factoryPresetFreqDescription: 'List of factory-preset frequencies. Note: order is respected.',
  factoryPresetFreqTitle: 'Factory Preset Frequencies',
  freqAdd: 'Add Frequency',
  frequencyPlaceholder: 'e.g. 869525000 for 869,525 MHz',
  macSettings: 'MAC settings',
  pingSlotFrequencyDescription: 'Frequency of the class B ping slot (Hz)',
  pingSlotFrequencyTitle: 'Ping Slot Frequency',
  pingSlotPeriodicityDescription: 'Periodicity of the class B ping slot',
  pingSlotPeriodicityTitle: 'Ping Slot Periodicity',
  pingSlotPeriodicityValue: '{count, plural, one {every second} other {every {count} seconds}}',
  resetWarning: 'Resetting is insecure and makes your device susceptible for replay attacks',
  resetsFCnt: 'Resets Frame Counters',
  rx1DataRateOffsetTitle: 'RX1 Data Rate Offset',
  rx1DelayDescription: 'Class A RX1 delay in seconds. RX2 delay is RX1 delay + 1 second.',
  rx1DelayTitle: 'RX1 Delay',
  rx2DataRateIndexTitle: 'RX2 Data Rate Index',
  rx2FrequencyDescription: 'Frequency for RX2 (Hz)',
  rx2FrequencyTitle: 'RX2 Frequency',
  updateSuccess: 'The MAC settings updated',
})

// 0...7
const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={m.pingSlotPeriodicityValue} values={{ count: value }} />,
  }
})

const MacSettingsSection = props => {
  const { activationMode, resetsFCnt: initialFCnt, initiallyCollapsed, isClassB } = props

  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST

  const [resetsFCnt, setResetsFCnt] = React.useState(isABP && initialFCnt)
  const handleResetsFCntChange = React.useCallback(evt => {
    const { checked } = evt.target

    setResetsFCnt(checked)
  }, [])

  return (
    <Form.CollapseSection
      id="mac-settings"
      title={sharedMessages.advancedSettings}
      initiallyCollapsed={initiallyCollapsed}
    >
      <Form.SubTitle title={m.macSettings} />
      <Form.Field
        title={sharedMessages.frameCounterWidth}
        name="mac_settings.supports_32_bit_f_cnt"
        component={Radio.Group}
        encode={fCntWidthEncode}
        decode={fCntWidthDecode}
      >
        <Radio label={sharedMessages['16Bit']} value={FRAME_WIDTH_COUNT.SUPPORTS_16_BIT} />
        <Radio label={sharedMessages['32Bit']} value={FRAME_WIDTH_COUNT.SUPPORTS_32_BIT} />
      </Form.Field>
      {isABP && (
        <>
          <Form.Field
            title={m.rx1DelayTitle}
            type="number"
            description={m.rx1DelayDescription}
            name="mac_settings.rx1_delay"
            component={Input}
            min={1}
            max={15}
          />
          <Form.Field
            title={m.rx1DataRateOffsetTitle}
            type="number"
            name="mac_settings.rx1_data_rate_offset"
            component={Input}
            min={0}
            max={7}
            glossaryId={glossaryIds.DATA_RATE_OFFSET}
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
        type="number"
        name="mac_settings.rx2_data_rate_index"
        component={Input}
        min={0}
        max={15}
        glossaryId={glossaryIds.DATA_RATE_INDEX}
      />
      <Form.Field
        type="number"
        min={100000}
        step={100}
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
      {(isClassB || isMulticast) && (
        <>
          <Form.Field
            title={m.pingSlotPeriodicityTitle}
            description={m.pingSlotPeriodicityDescription}
            name="mac_settings.ping_slot_periodicity"
            component={Select}
            options={pingSlotPeriodicityOptions}
            required={isClassB && isMulticast}
            menuPlacement="top"
          />
          <Form.Field
            type="number"
            min={100000}
            step={100}
            title={m.pingSlotFrequencyTitle}
            description={m.pingSlotFrequencyDescription}
            placeholder={m.frequencyPlaceholder}
            name="mac_settings.ping_slot_frequency"
            component={Input}
          />
        </>
      )}
    </Form.CollapseSection>
  )
}

MacSettingsSection.propTypes = {
  activationMode: PropTypes.oneOf(Object.values(ACTIVATION_MODES)).isRequired,
  initiallyCollapsed: PropTypes.bool,
  isClassB: PropTypes.bool,
  resetsFCnt: PropTypes.bool,
}

MacSettingsSection.defaultProps = {
  resetsFCnt: false,
  initiallyCollapsed: true,
  isClassB: false,
}

export default MacSettingsSection
