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

import React, { useEffect } from 'react'
import { useFormikContext } from 'formik'
import { isEmpty, isPlainObject, isUndefined, merge } from 'lodash'
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import toast from '@ttn-lw/components/toast'
import Form, { useFormContext } from '@ttn-lw/components/form'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import UnitInput from '@ttn-lw/components/unit-input'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import { getBackendErrorName, isBackend } from '@ttn-lw/lib/errors/utils'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { ACTIVATION_MODES } from '@console/lib/device-utils'
import { checkFromState } from '@account/lib/feature-checks'
import { mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { getDefaultMacSettings } from '@console/store/actions/network-server'

import { selectDefaultMacSettings } from '@console/store/selectors/network-server'

import { initialValues as provisioningInitialValues } from '../../provisioning-form-section'
import WarningTooltip from '../../warning-tooltip'
import { DEVICE_CLASS_MAP } from '../../utils'
import messages from '../../messages'

const m = defineMessages({
  advancedSectionTitle: 'Show advanced activation, LoRaWAN class and cluster settings',
  classA: 'None (class A only)',
  classB: 'Class B (Beaconing)',
  classC: 'Class C (Continuous)',
  classBandC: 'Class B and class C',
  skipJsRegistration: 'Skip registration on Join Server',
  multicastClassCapabilities: 'LoRaWAN class for multicast downlinks',
  register: 'Register manually',
})

const emptyDefaultMacSettings = {}

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

const activationModeEncoder = activationMode => ({
  supports_join: activationMode === ACTIVATION_MODES.OTAA,
  multicast: activationMode === ACTIVATION_MODES.MULTICAST,
})
const activationModeDecoder = ({ supports_join, multicast }) => {
  if (multicast) {
    return ACTIVATION_MODES.MULTICAST
  } else if (supports_join) {
    return ACTIVATION_MODES.OTAA
  }
  return ACTIVATION_MODES.ABP
}

const activationModeValueSetter = ({ setValues }, { value, value: { multicast } }) => {
  setValues(({ supports_class_b, supports_class_c, _claim, ...values }) => {
    const isClassA = supports_class_b === false && supports_class_c === false
    return {
      ...values,
      ...value,
      // (Re)set class capabilities when choosing multicast while
      // retaining the last value if possible.
      supports_class_b:
        multicast && isClassA
          ? undefined
          : !multicast && supports_class_b === undefined
            ? false
            : supports_class_b,
      supports_class_c:
        multicast && isClassA
          ? undefined
          : !multicast && supports_class_c === undefined
            ? false
            : supports_class_c,
      // Reset provisioning data if activation mode changed.
      ...(values.supports_join !== value.supports_join ? provisioningInitialValues : {}),
      // Skip JoinEUI check if the device is ABP/Multicast.
      _claim: !value.supports_join ? false : values._withQRdata ? values._claim : null,
    }
  })
}
const deviceClassEncoder = deviceClass => ({
  supports_class_b:
    deviceClass === DEVICE_CLASS_MAP.CLASS_B || deviceClass === DEVICE_CLASS_MAP.CLASS_B_C,
  supports_class_c:
    deviceClass === DEVICE_CLASS_MAP.CLASS_C || deviceClass === DEVICE_CLASS_MAP.CLASS_B_C,
})

const deviceClassDecoder = ({ supports_class_b, supports_class_c }) => {
  if (isUndefined(supports_class_b || supports_class_c)) {
    return ''
  }
  if (supports_class_b && supports_class_c) {
    return DEVICE_CLASS_MAP.CLASS_B_C
  } else if (supports_class_b) {
    return DEVICE_CLASS_MAP.CLASS_B
  } else if (supports_class_c) {
    return DEVICE_CLASS_MAP.CLASS_C
  }
  return DEVICE_CLASS_MAP.CLASS_A
}

const skipJsRegistrationEncoder = skip => (skip ? undefined : jsHost)
const skipJsRegistrationDecoder = isUndefined

const pingSlotPeriodicityOptions = Array.from({ length: 8 }, (_, index) => {
  const value = Math.pow(2, index)

  return {
    value: `PING_EVERY_${value}S`,
    label: <Message content={sharedMessages.secondInterval} values={{ count: value }} />,
  }
})

const initialValues = {
  multicast: false,
  mac_settings: {
    // Adding just values that don't have defaults.
    ping_slot_periodicity: undefined,
    beacon_frequency: undefined,
  },
  supports_class_b: false,
  supports_class_c: false,
  join_server_address: jsHost,
  network_server_address: nsHost,
  application_server_address: asHost,
  supports_join: true,
  _default_ns_settings: true,
}

const AdvancedSettingsSection = () => {
  const { addToFieldRegistry, removeFromFieldRegistry } = useFormikContext()
  const {
    setValues,
    setValidationContext,
    values: {
      mac_settings,
      _default_ns_settings,
      _claim,
      frequency_plan_id,
      lorawan_phy_version,
      lorawan_version,
      multicast: isMulticast,
      supports_join,
      supports_class_b: isClassB,
      supports_class_c: isClassC,
    },
  } = useFormContext()

  // Set common flags based on current form fields.
  const isOTAA = supports_join
  const isABP = !supports_join && !isMulticast
  // Network settings should not be modified when the type context is not yet known.
  const canManageNetworkSettings =
    Boolean(frequency_plan_id) && Boolean(lorawan_phy_version) && Boolean(lorawan_version)
  // Disallow using default settings when there is a required field within.
  const mayChangeToDefaultSettings = !((isABP || isMulticast) && isClassB)

  const dispatch = useDispatch()

  const defaultMacSettings =
    useSelector(state => selectDefaultMacSettings(state, frequency_plan_id, lorawan_phy_version)) ||
    emptyDefaultMacSettings
  const mayEditKeys = useSelector(state => checkFromState(mayEditApplicationDeviceKeys, state))

  // Fetch and apply default MAC settings, when FP or PHY version changes.
  useEffect(() => {
    let isMounted = true
    const updateMacSettings = async () => {
      try {
        const result = await dispatch(
          attachPromise(getDefaultMacSettings(frequency_plan_id, lorawan_phy_version)),
        )
        if (isMounted && isPlainObject(result) && 'defaultMacSettings' in result) {
          const { defaultMacSettings: settings } = result
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
            message: sharedMessages.fpNotFoundError,
            messageValues: {
              lorawanVersion: lorawan_phy_version,
              freqPlan: frequency_plan_id,
              code: msg => <code>{msg}</code>,
            },
          })
        } else {
          toast({
            type: toast.types.ERROR,
            message: sharedMessages.macSettingsError,
            messageValues: {
              freqPlan: frequency_plan_id,
              code: msg => <code>{msg}</code>,
            },
          })
        }
      }
    }

    if (frequency_plan_id && lorawan_phy_version) {
      updateMacSettings()
    }
    return () => {
      isMounted = false
    }
  }, [
    frequency_plan_id,
    lorawan_version,
    setValues,
    lorawan_phy_version,
    setValidationContext,
    dispatch,
    defaultMacSettings,
  ])

  // Manage mac settings in form.
  useEffect(() => {
    setValues(values => ({
      ...values,
      // Reset mac settings to defaults when toggled.
      mac_settings: _default_ns_settings
        ? merge({}, initialValues.mac_settings, values.mac_settings, defaultMacSettings)
        : merge({}, initialValues.mac_settings, values.mac_settings),
      // Unset default settings when required field is present (`ping_slot_periodicity`).
      _default_ns_settings: mayChangeToDefaultSettings ? _default_ns_settings : false,
    }))
  }, [_default_ns_settings, defaultMacSettings, mayChangeToDefaultSettings, setValues])

  // Register hidden fields so they don't get cleaned.
  useEffect(() => {
    const hiddenFields = ['network_server_address', 'application_server_address']
    addToFieldRegistry(...hiddenFields)
    return () => removeFromFieldRegistry(...hiddenFields)
  }, [addToFieldRegistry, removeFromFieldRegistry])

  // Do not render advanced settings when FP, MAC and PHY version is unknown.
  if (!canManageNetworkSettings) {
    return null
  }

  return (
    <>
      <hr />
      <Form.CollapseSection id="advanced-settings" title={m.advancedSectionTitle}>
        <Form.Field
          title={sharedMessages.activationMode}
          name="supports_join,multicast"
          component={Radio.Group}
          tooltipId={tooltipIds.ACTIVATION_MODE}
          encode={activationModeEncoder}
          decode={activationModeDecoder}
          valueSetter={activationModeValueSetter}
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
          name="supports_class_b,supports_class_c"
          component={Select}
          options={isMulticast ? multicastClassOptions : allClassOptions}
          tooltipId={tooltipIds.CLASSES}
          encode={deviceClassEncoder}
          decode={deviceClassDecoder}
        />
        <Form.Field
          title={messages.networkDefaults}
          label={messages.defaultNetworksSettings}
          name="_default_ns_settings"
          component={Checkbox}
          tooltipId={tooltipIds.NETWORK_RX_DEFAULTS}
          disabled={!mayChangeToDefaultSettings || !canManageNetworkSettings}
        />
        <div style={{ display: !_default_ns_settings ? 'block' : 'none' }}>
          {isABP && (
            <>
              <Form.FieldContainer horizontal>
                <Form.Field
                  required={!isUndefined(mac_settings.rx1_data_rate_offset)}
                  title={sharedMessages.rx1DataRateOffset}
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
                  title={sharedMessages.classBTimeout}
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
                  title={sharedMessages.pingSlotPeriodicity}
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
                  title={sharedMessages.beaconFrequency}
                  placeholder={sharedMessages.frequencyPlaceholder}
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
                  title={sharedMessages.pingSlotFrequency}
                  placeholder={sharedMessages.frequencyPlaceholder}
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
                title={sharedMessages.classCTimeout}
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
              title={sharedMessages.rx2Frequency}
              placeholder={sharedMessages.frequencyPlaceholder}
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
              name="mac_settings.factory_preset_frequencies"
              component={KeyValueMap}
              title={sharedMessages.factoryPresetFrequencies}
              addMessage={messages.freqAdd}
              valuePlaceholder={sharedMessages.frequencyPlaceholder}
              tooltipId={tooltipIds.FACTORY_PRESET_FREQUENCIES}
              encode={factoryPresetFreqEncoder}
              decode={factoryPresetFreqDecoder}
            />
          )}
        </div>
        <Form.Field
          title={messages.clusterSettings}
          label={m.skipJsRegistration}
          name="join_server_address"
          encode={skipJsRegistrationEncoder}
          decode={skipJsRegistrationDecoder}
          component={Checkbox}
          tooltipId={tooltipIds.SKIP_JOIN_SERVER_REGISTRATION}
          disabled={_claim}
        />
      </Form.CollapseSection>
      <hr />
    </>
  )
}

export { AdvancedSettingsSection as default, initialValues }
