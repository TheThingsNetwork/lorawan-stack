// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useEffect, useState } from 'react'
import { isEmpty, isUndefined, merge, omitBy } from 'lodash'
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import tts from '@console/api/tts'

import toast from '@ttn-lw/components/toast'
import Form, { useFormContext } from '@ttn-lw/components/form'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import UnitInput from '@ttn-lw/components/unit-input'

import Message from '@ttn-lw/lib/components/message'

import WarningTooltip from '@console/views/device-add-old/manual/form/warning-tooltip'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import { getBackendErrorName, isBackend } from '@ttn-lw/lib/errors/utils'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import { ACTIVATION_MODES, hasCFListTypeChMask } from '@console/lib/device-utils'
import { checkFromState } from '@account/lib/feature-checks'
import { mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { REGISTRATION_TYPES, DEVICE_CLASS_MAP } from '../../utils'
import messages from '../../messages'

const m = defineMessages({
  advancedSectionTitle: 'Show advanced activation, LoRaWAN class and cluster settings',
  classA: 'None (class A only)',
  classB: 'Class B (Beaconing)',
  classC: 'Class C (Continuous)',
  classBandC: 'Class B and class C',
  skipJsRegistration: 'Skip registration on Join Server',
  multicastClassCapabilities: 'LoRaWAN class for multicast downlinks',
  factoryFreqWarning:
    'In LoRaWAN, factory preset frequencies are only supported for bands with a CFList type of frequencies',
  register: 'Register manually',
  macSettingsError:
    'There was an error and the default MAC settings for the <code>{freqPlan}</code> frequency plan could not be loaded',
  fpNotFoundError:
    'The LoRaWAN version <code>{lorawanVersion}</code> does not support the <code>{freqPlan}</code> frequency plan. Please choose a different MAC version or frequency plan.',
  disabledNetworkSettings:
    'Please select frequency plan and LoRaWAN versions to manage MAC settings',
})

const allClassOptions = [
  { label: m.classA, value: DEVICE_CLASS_MAP.CLASS_A },
  { label: m.classB, value: DEVICE_CLASS_MAP.CLASS_B },
  { label: m.classC, value: DEVICE_CLASS_MAP.CLASS_C },
  { label: m.classBandC, value: DEVICE_CLASS_MAP.CLASS_B_C },
]
const multicastClassOptions = allClassOptions.filter(
  ({ value }) => value !== DEVICE_CLASS_MAP.CLASS_A,
)
const asUrl = selectAsConfig().base_url
const jsUrl = selectJsConfig().base_url
const nsUrl = selectNsConfig().base_url
const jsHost = getHostFromUrl(jsUrl)
const nsHost = getHostFromUrl(nsUrl)
const asHost = getHostFromUrl(asUrl)

const factoryPresetFreqEncoder = value => (isEmpty(value) ? undefined : value)
const factoryPresetFreqDecoder = value => (value === undefined ? [] : value)

const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={messages.pingSlotPeriodicityValue} values={{ count: value }} />,
  }
})

const initialValues = {
  multicast: false,
  mac_settings: {
    // Adding just values that don't have defaults.
    ping_slot_periodicity: '',
    beacon_frequency: '',
  },
  supports_class_b: false,
  supports_class_c: false,
  join_server_address: jsHost,
  network_server_address: nsHost,
  application_server_address: asHost,
  supports_join: true,
  _device_class: DEVICE_CLASS_MAP.CLASS_A,
  _activation_mode: ACTIVATION_MODES.OTAA,
  _registration: REGISTRATION_TYPES.SINGLE,
  _default_ns_settings: true,
  _skip_js_registration: false,
}

const AdvancedSettingsSection = () => {
  const {
    setValues,
    setValidationContext,
    values: {
      mac_settings,
      _activation_mode,
      _default_ns_settings,
      _device_class,
      _skip_js_registration,
      frequency_plan_id,
      lorawan_phy_version,
      lorawan_version,
    },
  } = useFormContext()

  // Set common flags based on current form fields.
  const isOTAA = _activation_mode === ACTIVATION_MODES.OTAA
  const isABP = _activation_mode === ACTIVATION_MODES.ABP
  const isMulticast = _activation_mode === ACTIVATION_MODES.MULTICAST
  const isClassB =
    _device_class === DEVICE_CLASS_MAP.CLASS_B_C || _device_class === DEVICE_CLASS_MAP.CLASS_B
  const isClassC =
    _device_class === DEVICE_CLASS_MAP.CLASS_B_C || _device_class === DEVICE_CLASS_MAP.CLASS_C
  // Network settings should not be modified when the type context is not yet known.
  const canManageNetworkSettings =
    Boolean(frequency_plan_id) && Boolean(lorawan_phy_version) && Boolean(lorawan_version)
  // Disallow using default settings when there is a required field within.
  const mayChangeToDefaultSettings = !((isABP || isMulticast) && isClassB)

  // The technical difference between bands that do support factory preset frequencies
  // and bands that do not support them, is that the former uses a CFList type of Frequencies,
  // and the latter uses a CFList type of ChMask (channel mask).
  // When there is a channel mask, the frequencies aren't configured by frequency in Hertz,
  // but by index. The factory preset frequencies is really the frequencies in Hertz,
  // so it requires bands with a CFList type of Frequencies.
  const disableFactoryPresetFreq = hasCFListTypeChMask(frequency_plan_id)

  const [defaultMacSettings, setDefaultMacSettings] = useState(initialValues.mac_settings)

  const mayEditKeys = useSelector(state => checkFromState(mayEditApplicationDeviceKeys, state))

  // Fetch and apply default MAC settings, when FP or PHY version changes.
  useEffect(() => {
    let isMounted = true
    const updateMacSettings = async () => {
      try {
        const settings = await tts.Ns.getDefaultMacSettings(frequency_plan_id, lorawan_phy_version)
        if (isMounted) {
          setDefaultMacSettings(settings)
          setValidationContext(context => ({ ...context, defaultMacSettings: settings }))

          setValues(values => ({
            ...values,
            mac_settings: merge({}, initialValues.mac_settings, values.mac_settings, settings),
          }))
        }
      } catch (err) {
        if (isBackend(err) && getBackendErrorName(err) === 'no_band_version') {
          toast({
            type: toast.types.ERROR,
            message: m.fpNotFoundError,
            messageValues: {
              lorawan_phy_version,
              frequency_plan_id,
              code: msg => <code>{msg}</code>,
            },
          })
        } else {
          toast({
            type: toast.types.ERROR,
            message: m.macSettingsError,
            messageValues: {
              frequency_plan_id,
              code: msg => <code>{msg}</code>,
            },
          })
        }
      }
    }
    const resetMacSettings = () => {
      setDefaultMacSettings({})
    }

    if (frequency_plan_id && lorawan_phy_version) {
      updateMacSettings()
    } else {
      resetMacSettings()
    }
    return () => {
      isMounted = false
    }
  }, [
    frequency_plan_id,
    lorawan_version,
    setValues,
    setDefaultMacSettings,
    lorawan_phy_version,
    setValidationContext,
  ])

  // Apply field inter-dependent modifications.
  // Note: Values set to `undefined`` will be stripped.
  useEffect(() => {
    setValues(values =>
      omitBy(
        {
          ...values,
          // Map device class selector to class flags.
          supports_class_b: isClassB,
          supports_class_c: isClassC,
          // Map activation mode selector to support join and multicast flag.
          supports_join: !_skip_js_registration ? isOTAA : undefined,
          multicast: isMulticast,
          // Reset mac settings to defaults when toggled.
          mac_settings: _default_ns_settings
            ? merge({}, initialValues.mac_settings, values.mac_settings, defaultMacSettings)
            : merge({}, initialValues.mac_settings, values.mac_settings),
          // Strip join server address when opting to skip JS registration.
          join_server_address: !_skip_js_registration ? jsHost : undefined,
          // Unset default settings when required field is present (`ping_slot_periodicity`).
          _default_ns_settings: mayChangeToDefaultSettings ? _default_ns_settings : false,
          // Reset device class selector, if multicast and set to class A.
          _device_class: isMulticast
            ? values._device_class === DEVICE_CLASS_MAP.CLASS_A
              ? ''
              : _device_class
            : _device_class,
        },
        isUndefined,
      ),
    )
  }, [
    _default_ns_settings,
    _device_class,
    _skip_js_registration,
    defaultMacSettings,
    isClassB,
    isClassC,
    isMulticast,
    isOTAA,
    mayChangeToDefaultSettings,
    setValues,
  ])

  return (
    <Form.CollapseSection id="advanced-settings" title={m.advancedSectionTitle}>
      <Form.Field
        title={sharedMessages.activationMode}
        name="_activation_mode"
        connectedFields={['supports_class_b', 'supports_class_c', 'multicast']}
        component={Radio.Group}
        tooltipId={tooltipIds.ACTIVATION_MODE}
        required
      >
        <Radio label={sharedMessages.otaa} value={ACTIVATION_MODES.OTAA} />
        {mayEditKeys && (
          <>
            <Radio label={sharedMessages.abp} value={ACTIVATION_MODES.ABP} />
            <Radio label={sharedMessages.multicast} value={ACTIVATION_MODES.MULTICAST} />
          </>
        )}
      </Form.Field>
      <Form.Field
        title={isMulticast ? m.multicastClassCapabilities : messages.classCapabilities}
        required={isMulticast}
        name="_device_class"
        connectedFields={['supports_join']}
        component={Select}
        options={isMulticast ? multicastClassOptions : allClassOptions}
        tooltipId={tooltipIds.CLASSES}
      />
      <Form.Field
        title={messages.networkDefaults}
        label={messages.defaultNetworksSettings}
        name="_default_ns_settings"
        component={Checkbox}
        tooltipId={tooltipIds.NETWORK_RX_DEFAULTS}
        disabled={!mayChangeToDefaultSettings || !canManageNetworkSettings}
        description={!canManageNetworkSettings ? m.disabledNetworkSettings : undefined}
      />
      <div style={{ display: !_default_ns_settings ? 'block' : 'none' }}>
        {isABP && (
          <>
            <Form.FieldContainer horizontal>
              <Form.Field
                required={!isUndefined(mac_settings.rx1_data_rate_offset)}
                title={messages.rx1DataRateOffsetTitle}
                type="number"
                name="mac_settings.rx1_data_rate_offset"
                component={Input}
                min={0}
                max={7}
                tooltipId={tooltipIds.DATA_RATE_OFFSET}
                inputWidth="xxs"
                fieldWidth="xs"
                titleChildren={
                  <WarningTooltip
                    desiredValue={defaultMacSettings.desired_rx1_data_rate_offset}
                    currentValue={mac_settings.rx1_data_rate_offset}
                  />
                }
              />
              <Form.Field
                title={messages.rx1DelayTitle}
                type="number"
                required={!isUndefined(defaultMacSettings.rx1_delay)}
                name="mac_settings.rx1_delay"
                append={<Message content={sharedMessages.secondsAbbreviated} />}
                tooltipId={tooltipIds.RX1_DELAY}
                component={Input}
                min={1}
                max={15}
                inputWidth="xs"
                fieldWidth={isClassB ? 'xs' : 'xxs'}
                titleChildren={
                  <WarningTooltip
                    desiredValue={defaultMacSettings.desired_rx1_delay}
                    currentValue={mac_settings.rx1_delay}
                  />
                }
              />
              <Form.Field
                title={sharedMessages.resetsFCnt}
                tooltipId={tooltipIds.RESETS_F_CNT}
                warning={mac_settings.resets_f_cnt ? sharedMessages.resetWarning : undefined}
                name="mac_settings.resets_f_cnt"
                component={Checkbox}
              />
            </Form.FieldContainer>
          </>
        )}
        {(isClassB || isMulticast) && (
          <>
            <Form.FieldContainer horizontal>
              <Form.Field
                required={!isUndefined(defaultMacSettings.class_b_timeout)}
                title={messages.classBTimeout}
                name="mac_settings.class_b_timeout"
                tooltipId={tooltipIds.CLASS_B_TIMEOUT}
                component={UnitInput.Duration}
                unitSelector={['ms', 's']}
                type="number"
                fieldWidth="xs"
                titleChildren={
                  <WarningTooltip
                    desiredValue={defaultMacSettings.desired_class_b_timeout}
                    currentValue={mac_settings.class_b_timeout}
                  />
                }
              />
              <Form.Field
                title={messages.pingSlotPeriodicityTitle}
                name="mac_settings.ping_slot_periodicity"
                tooltipId={tooltipIds.PING_SLOT_PERIODICITY}
                component={Select}
                options={pingSlotPeriodicityOptions}
                required={isClassB && (isMulticast || isABP)}
                menuPlacement="top"
                fieldWidth="xs"
              />
              <Form.Field
                title={messages.pingSlotDataRateTitle}
                name="mac_settings.ping_slot_data_rate_index"
                required={!isUndefined(defaultMacSettings.ping_slot_data_rate_index)}
                tooltipId={tooltipIds.PING_SLOT_DATA_RATE_INDEX}
                component={Input}
                type="number"
                min={0}
                max={15}
                inputWidth="xxs"
                titleChildren={
                  <WarningTooltip
                    desiredValue={defaultMacSettings.desired_ping_slot_data_rate_index}
                    currentValue={mac_settings.ping_slot_data_rate_index}
                  />
                }
              />
            </Form.FieldContainer>
            <Form.FieldContainer horizontal>
              <Form.Field
                type="number"
                min={100000}
                required={!isUndefined(defaultMacSettings.beacon_frequency)}
                title={messages.beaconFrequency}
                placeholder={messages.frequencyPlaceholder}
                name="mac_settings.beacon_frequency"
                tooltipId={tooltipIds.BEACON_FREQUENCY}
                component={UnitInput.Hertz}
                fieldWidth="xs"
                titleChildren={
                  <WarningTooltip
                    desiredValue={defaultMacSettings.desired_beacon_frequency}
                    currentValue={mac_settings.beacon_frequency}
                  />
                }
              />
              <Form.Field
                type="number"
                min={100000}
                required={!isUndefined(defaultMacSettings.ping_slot_frequency)}
                title={messages.pingSlotFrequencyTitle}
                placeholder={messages.frequencyPlaceholder}
                name="mac_settings.ping_slot_frequency"
                tooltipId={tooltipIds.PING_SLOT_FREQUENCY}
                component={UnitInput.Hertz}
                fieldWidth="xs"
                titleChildren={
                  <WarningTooltip
                    desiredValue={defaultMacSettings.desired_ping_slot_frequency}
                    currentValue={mac_settings.ping_slot_frequency}
                  />
                }
              />
            </Form.FieldContainer>
          </>
        )}
        <Form.FieldContainer horizontal>
          {isClassC && (
            <Form.Field
              required={!isUndefined(defaultMacSettings.class_c_timeout)}
              title={messages.classCTimeout}
              name="mac_settings.class_c_timeout"
              tooltipId={tooltipIds.CLASS_C_TIMEOUT}
              component={UnitInput.Duration}
              unitSelector={['ms', 's']}
              type="number"
              fieldWidth="xs"
              inputWidth="xxs"
              titleChildren={
                <WarningTooltip
                  desiredValue={defaultMacSettings.desired_class_c_timeout}
                  currentValue={mac_settings.class_c_timeout}
                />
              }
            />
          )}
          <Form.Field
            title={messages.rx2DataRateIndexTitle}
            type="number"
            name="mac_settings.rx2_data_rate_index"
            tooltipId={tooltipIds.RX2_DATA_RATE_INDEX}
            required={!isUndefined(defaultMacSettings.rx2_data_rate_index)}
            component={Input}
            min={0}
            max={15}
            inputWidth="xxs"
            fieldWidth={!isClassC || isMulticast ? 'xs' : 'xxs'}
            titleChildren={
              <WarningTooltip
                desiredValue={defaultMacSettings.desired_rx2_data_rate_index}
                currentValue={mac_settings.rx2_data_rate_index}
              />
            }
          />
          <Form.Field
            type="number"
            min={100000}
            step={100}
            required={!isUndefined(defaultMacSettings.rx2_frequency)}
            title={messages.rx2FrequencyTitle}
            placeholder={messages.frequencyPlaceholder}
            name="mac_settings.rx2_frequency"
            tooltipId={tooltipIds.RX2_FREQUENCY}
            component={UnitInput.Hertz}
            fieldWidth="xs"
            titleChildren={
              <WarningTooltip
                desiredValue={defaultMacSettings.desired_rx2_frequency}
                currentValue={mac_settings.rx2_frequency}
              />
            }
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
            encode={factoryPresetFreqEncoder}
            decode={factoryPresetFreqDecoder}
          />
        )}
      </div>
      <Form.Field
        title={messages.clusterSettings}
        label={m.skipJsRegistration}
        name="_skip_js_registration"
        connectedFields={[
          'network_server_address',
          'application_server_address',
          'join_server_address',
        ]}
        component={Checkbox}
        tooltipId={tooltipIds.SKIP_JOIN_SERVER_REGISTRATION}
      />
    </Form.CollapseSection>
  )
}

export { AdvancedSettingsSection as default, initialValues }
