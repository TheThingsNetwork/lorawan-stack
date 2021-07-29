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
import { merge } from 'lodash'
import classnames from 'classnames'

import api from '@console/api'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import Message from '@ttn-lw/lib/components/message'

import PhyVersionInput from '@console/components/phy-version-input'
import JoinEUIPRefixesInput from '@console/components/join-eui-prefixes-input'

import DevAddrInput from '@console/containers/dev-addr-input'
import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import env from '@ttn-lw/lib/env'

import {
  LORAWAN_VERSIONS,
  ACTIVATION_MODES,
  parseLorawanMacVersion,
  generate16BytesKey,
  DEVICE_CLASSES,
} from '@console/lib/device-utils'

import { REGISTRATION_TYPES } from '../../utils'
import messages from '../../messages'
import style from '../../device-add.styl'

import AdvancedSettingsSection from './advanced-settings'
import validationSchema, { devEUISchema } from './validation-schema'

const m = defineMessages({
  register: 'Register manually',
})

const defaultValues = {
  ids: {
    dev_eui: '',
    join_eui: '',
    device_id: '',
  },
  lorawan_version: '',
  lorawan_phy_version: '',
  frequency_plan_id: '',
  root_keys: {
    app_key: {
      key: '',
    },
    nwk_key: {
      key: '',
    },
  },
  session: {
    dev_addr: '',
    keys: {
      f_nwk_s_int_key: { key: '' },
      s_nwk_s_int_key: { key: '' },
      nwk_s_enc_key: { key: '' },
    },
  },
  multicast: false,
  supports_join: false,
  supports_class_b: false,
  supports_class_c: false,
  mac_settings: {},
  _activation_mode: '',
  _device_class: undefined,
  _external_servers: false,
  _registration: REGISTRATION_TYPES.SINGLE,
  _default_ns_settings: true,
}

const generateDeviceId = (device = {}) => {
  const { ids: idsValues = {} } = device

  try {
    devEUISchema.validateSync(idsValues.dev_eui)
    return `eui-${idsValues.dev_eui.toLowerCase()}`
  } catch (e) {
    // We dont want to use invalid `dev_eui` as `device_id`.
  }

  return defaultValues.ids.device_id || ''
}

const ManualForm = props => {
  const {
    appId,
    asConfig,
    jsConfig,
    nsConfig,
    mayEditKeys,
    prefixes,
    createDevice,
    createDeviceSuccess,
    applicationDevEUICounter,
    fetchDevEUICounter,
  } = props

  const asEnabled = asConfig.enabled
  const jsEnabled = jsConfig.enabled
  const nsEnabled = nsConfig.enabled
  const asUrl = asEnabled ? asConfig.base_url : undefined
  const jsUrl = jsEnabled ? jsConfig.base_url : undefined
  const nsUrl = nsEnabled ? nsConfig.base_url : undefined

  const validationContext = React.useMemo(
    () => ({
      jsUrl,
      jsEnabled,
      nsUrl,
      nsEnabled,
      asUrl,
      asEnabled,
      mayEditKeys,
    }),
    [asEnabled, asUrl, jsEnabled, jsUrl, mayEditKeys, nsEnabled, nsUrl],
  )
  const initialValues = React.useMemo(
    () =>
      validationSchema.cast(
        merge({}, defaultValues, {
          supports_join: jsEnabled,
          _activation_mode: jsEnabled
            ? ACTIVATION_MODES.OTAA
            : nsEnabled
            ? ACTIVATION_MODES.ABP
            : ACTIVATION_MODES.NONE,
        }),
        { context: validationContext },
      ),
    [jsEnabled, nsEnabled, validationContext],
  )
  const formRef = React.useRef(null)
  const deviceIdInputRef = React.useRef(null)
  const euiInputRef = React.useRef(null)
  const [devEUIGenerated, setDevEUIGenerated] = React.useState(false)
  const [errorMessage, setErrorMessage] = React.useState(undefined)

  const indicatorContent = Boolean(errorMessage)
    ? errorMessage
    : {
        ...sharedMessages.used,
        values: {
          currentValue: applicationDevEUICounter,
          maxValue: env.devEUIConfig.applicationLimit,
        },
      }

  const handleGenerate = React.useCallback(async () => {
    try {
      const result = await api.application.generateDevEUI(appId)
      setDevEUIGenerated(true)
      fetchDevEUICounter(appId)
      euiInputRef.current.focus()
      setErrorMessage(undefined)
      return result.dev_eui
    } catch (error) {
      if (error.details[0].name === 'global_eui_limit_reached') {
        setErrorMessage(sharedMessages.devEUIBlockLimitReached)
      } else setErrorMessage(sharedMessages.unknownError)
      setDevEUIGenerated(true)
    }
  }, [appId, fetchDevEUICounter])

  React.useEffect(() => {
    fetchDevEUICounter(appId)
  }, [appId, fetchDevEUICounter])

  const [error, setError] = React.useState(undefined)
  const handleSetError = React.useCallback(error => setError(error), [])

  const [lorawanVersion, setLorawanVersion] = React.useState(initialValues.lorawan_version)
  const handleLorawanVersionChange = React.useCallback(version => setLorawanVersion(version), [])

  const [activationMode, setActivationMode] = React.useState(initialValues._activation_mode)
  const handleActivationModeChange = React.useCallback(
    mode => {
      const { setValues, values } = formRef.current
      setValues(
        validationSchema.cast(
          { ...defaultValues, ...values, _activation_mode: mode },
          { context: validationContext },
        ),
      )

      return setActivationMode(mode)
    },
    [validationContext],
  )

  const [defaultNsSettings, setDefaultNsSettings] = React.useState(true)
  const handleDefaultNsSettings = React.useCallback(
    checked => {
      const { setValues, values } = formRef.current

      setValues(
        validationSchema.cast(
          {
            ...defaultValues,
            ...values,
            mac_settings: defaultValues.mac_settings,
          },
          { context: validationContext },
        ),
      )

      return setDefaultNsSettings(checked)
    },
    [validationContext],
  )

  const [deviceClass, setDeviceClass] = React.useState(
    initialValues._activation_mode === ACTIVATION_MODES.OTAA ? DEVICE_CLASSES.CLASS_A : undefined,
  )
  const handleDeviceClassChange = React.useCallback(devClass => {
    setDeviceClass(devClass)
  }, [])

  const [freqPlan, setFreqPlan] = React.useState()
  const handleFreqPlanChange = React.useCallback(band => {
    setFreqPlan(band.value)
  }, [])

  const handleIdPrefill = React.useCallback(() => {
    if (formRef.current) {
      const { values, setFieldValue } = formRef.current

      // Do not overwrite a value that the user has already set.
      if (values.ids.device_id === initialValues.ids.device_id) {
        const generatedId = generateDeviceId(values)
        setFieldValue('ids.device_id', generatedId)
      }
    }
  }, [initialValues.ids.device_id])
  const handleIdFocus = React.useCallback(() => {
    if (formRef.current && deviceIdInputRef.current) {
      const { current: inputElement } = deviceIdInputRef
      const { values } = formRef.current
      const generatedId = generateDeviceId(values)
      if (generatedId === values.ids.device_id) {
        // Select the value on focus if it was generated.
        inputElement.setSelectionRange(0, generatedId.length)
      }
    }
  }, [])

  const lwVersion = parseLorawanMacVersion(lorawanVersion)
  const isOTAA = activationMode === ACTIVATION_MODES.OTAA
  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const devEUIGenerateDisabled =
    applicationDevEUICounter === env.devEUIConfig.applicationLimit ||
    !env.devEUIConfig.devEUIIssuingEnabled ||
    devEUIGenerated

  const handleSubmit = React.useCallback(
    async (values, { setSubmitting, resetForm }) => {
      try {
        handleSetError(undefined)

        const {
          _activation_mode,
          _device_class,
          _registration,
          _external_servers,
          _default_ns_settings,
          ...castedValues
        } = validationSchema.cast(values, {
          context: validationContext,
        })
        const {
          ids,
          supports_join,
          lorawan_version,
          lorawan_phy_version,
          frequency_plan_id,
        } = castedValues

        if (Object.keys(castedValues.mac_settings).length === 0) {
          delete castedValues.mac_settings
        }

        await createDevice(appId, castedValues)

        switch (_registration) {
          case REGISTRATION_TYPES.MULTIPLE:
            toast({
              type: toast.types.SUCCESS,
              message: messages.createSuccess,
            })
            resetForm({
              errors: {},
              values: {
                ...castedValues,
                ...defaultValues,
                ids: {
                  ...defaultValues.ids,
                  join_eui: supports_join ? ids.join_eui : undefined,
                },
                lorawan_version,
                lorawan_phy_version,
                frequency_plan_id,
                _registration: REGISTRATION_TYPES.MULTIPLE,
              },
            })
            break
          case REGISTRATION_TYPES.SINGLE:
          default:
            createDeviceSuccess(appId, ids.device_id)
        }
      } catch (error) {
        handleSetError(error)
        setSubmitting(false)
      }
    },
    [appId, createDevice, createDeviceSuccess, handleSetError, validationContext],
  )

  const indicatorCls = classnames(style.indicator, {
    [style.error]:
      applicationDevEUICounter === env.devEUIConfig.applicationLimit || Boolean(errorMessage),
  })

  let appKeyPlaceholder = undefined
  let nwkKeyPlaceholder = undefined
  if (!mayEditKeys) {
    appKeyPlaceholder = sharedMessages.insufficientAppKeyRights
    nwkKeyPlaceholder = sharedMessages.insufficientNwkKeyRights
  }

  const devEUIComponent = env.devEUIConfig.devEUIIssuingEnabled ? (
    <Form.Field
      title={sharedMessages.devEUI}
      name="ids.dev_eui"
      type="byte"
      min={8}
      max={8}
      required
      component={Input.Generate}
      tooltipId={tooltipIds.DEV_EUI}
      onBlur={handleIdPrefill}
      onGenerateValue={handleGenerate}
      actionDisable={devEUIGenerateDisabled}
      inputRef={euiInputRef}
    >
      <Message className={indicatorCls} component="label" content={indicatorContent} />
    </Form.Field>
  ) : (
    <Form.Field
      title={sharedMessages.devEUI}
      name="ids.dev_eui"
      type="byte"
      min={8}
      max={8}
      required
      component={Input}
      tooltipId={tooltipIds.DEV_EUI}
      onBlur={handleIdPrefill}
    />
  )

  return (
    <Form
      onSubmit={handleSubmit}
      validationSchema={validationSchema}
      validationContext={validationContext}
      initialValues={initialValues}
      error={error}
      formikRef={formRef}
    >
      <Form.Field
        required
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={Select}
        options={LORAWAN_VERSIONS}
        tooltipId={tooltipIds.LORAWAN_VERSION}
        onChange={handleLorawanVersionChange}
      />
      <Form.Field
        required
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={PhyVersionInput}
        lorawanVersion={lorawanVersion}
        tooltipId={tooltipIds.REGIONAL_PARAMETERS}
      />
      <NsFrequencyPlansSelect
        required={nsEnabled}
        tooltipId={tooltipIds.FREQUENCY_PLAN}
        name="frequency_plan_id"
        onChange={handleFreqPlanChange}
      />
      <hr />
      <AdvancedSettingsSection
        jsEnabled={jsConfig.enabled}
        nsEnabled={nsConfig.enabled}
        activationMode={activationMode}
        onActivationModeChange={handleActivationModeChange}
        deviceClass={deviceClass}
        onDeviceClassChange={handleDeviceClassChange}
        onDefaultNsSettingsChange={handleDefaultNsSettings}
        defaultNsSettings={defaultNsSettings}
        freqPlan={freqPlan}
      />
      <hr />
      {!isMulticast && devEUIComponent}
      {(isABP || isMulticast) && (
        <>
          <DevAddrInput title={sharedMessages.devAddr} name="session.dev_addr" required />
          {asEnabled && (
            <Form.Field
              required
              title={sharedMessages.appSKey}
              name="session.keys.app_s_key.key"
              type="byte"
              min={16}
              max={16}
              disabled={!mayEditKeys}
              component={Input.Generate}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
              tooltipId={tooltipIds.APP_SESSION_KEY}
            />
          )}
          <Form.Field
            mayGenerateValue
            title={lwVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
            name="session.keys.f_nwk_s_int_key.key"
            type="byte"
            min={16}
            max={16}
            required
            component={Input.Generate}
            onGenerateValue={generate16BytesKey}
            tooltipId={lwVersion >= 110 ? undefined : tooltipIds.NETWORK_SESSION_KEY}
          />
          {lwVersion >= 110 && (
            <Form.Field
              mayGenerateValue
              title={sharedMessages.sNwkSIKey}
              name="session.keys.s_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              required
              description={sharedMessages.sNwkSIKeyDescription}
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
            />
          )}
          {lwVersion >= 110 && (
            <Form.Field
              mayGenerateValue
              title={sharedMessages.nwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              required
              description={sharedMessages.nwkSEncKeyDescription}
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
            />
          )}
        </>
      )}
      {isOTAA && (
        <>
          <Form.Field
            title={lwVersion < 104 ? sharedMessages.appEUI : sharedMessages.joinEUI}
            component={JoinEUIPRefixesInput}
            name="ids.join_eui"
            prefixes={prefixes}
            required
            showPrefixes
            tooltipId={tooltipIds.JOIN_EUI}
          />
          <Form.Field
            required
            disabled={!mayEditKeys}
            title={sharedMessages.appKey}
            name="root_keys.app_key.key"
            type="byte"
            min={16}
            max={16}
            component={Input.Generate}
            placeholder={appKeyPlaceholder}
            mayGenerateValue={mayEditKeys}
            onGenerateValue={generate16BytesKey}
            tooltipId={tooltipIds.APP_KEY}
          />
          {lwVersion >= 110 && (
            <Form.Field
              required
              title={sharedMessages.nwkKey}
              name="root_keys.nwk_key.key"
              type="byte"
              min={16}
              max={16}
              component={Input.Generate}
              placeholder={nwkKeyPlaceholder}
              disabled={!mayEditKeys}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
              tooltipId={tooltipIds.NETWORK_KEY}
            />
          )}
        </>
      )}
      <Form.Field
        required
        title={sharedMessages.devID}
        name="ids.device_id"
        placeholder={sharedMessages.deviceIdPlaceholder}
        component={Input}
        tooltipId={tooltipIds.DEVICE_ID}
        description={messages.deviceIdDescription}
        inputRef={deviceIdInputRef}
        onFocus={handleIdFocus}
      />
      <Form.Field title={messages.afterRegistration} name="_registration" component={Radio.Group}>
        <Radio label={messages.singleRegistration} value={REGISTRATION_TYPES.SINGLE} />
        <Radio label={messages.multipleRegistration} value={REGISTRATION_TYPES.MULTIPLE} />
      </Form.Field>
      <SubmitBar>
        <Form.Submit message={messages.submitTitle} component={SubmitButton} />
      </SubmitBar>
    </Form>
  )
}

ManualForm.propTypes = {
  appId: PropTypes.string.isRequired,
  applicationDevEUICounter: PropTypes.number.isRequired,
  asConfig: PropTypes.stackComponent.isRequired,
  createDevice: PropTypes.func.isRequired,
  createDeviceSuccess: PropTypes.func.isRequired,
  fetchDevEUICounter: PropTypes.func.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  nsConfig: PropTypes.stackComponent.isRequired,
  prefixes: PropTypes.euiPrefixes.isRequired,
}

export default withBreadcrumb('devices.add.manually', props => (
  <Breadcrumb path={`/applications/${props.appId}/devices/add/repository`} content={m.register} />
))(ManualForm)
