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

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import PhyVersionInput from '@console/components/phy-version-input'
import JoinEUIPRefixesInput from '@console/components/join-eui-prefixes-input'

import DevAddrInput from '@console/containers/dev-addr-input'
import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  LORAWAN_VERSIONS,
  ACTIVATION_MODES,
  parseLorawanMacVersion,
  generate16BytesKey,
  DEVICE_CLASSES,
} from '@console/lib/device-utils'

import { REGISTRATION_TYPES } from '../../utils'
import messages from '../../messages'

import AdvancedSettingsSection from './advanced-settings'
import validationSchema from './validation-schema'

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

  const [deviceClass, setDeviceClass] = React.useState(
    initialValues._activation_mode === ACTIVATION_MODES.OTAA ? DEVICE_CLASSES.CLASS_A : undefined,
  )
  const handleDeviceClassChange = React.useCallback(devClass => {
    setDeviceClass(devClass)
  }, [])

  const lwVersion = parseLorawanMacVersion(lorawanVersion)
  const isOTAA = activationMode === ACTIVATION_MODES.OTAA
  const isABP = activationMode === ACTIVATION_MODES.ABP
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST
  const isNone = activationMode === ACTIVATION_MODES.NONE

  const handleSubmit = React.useCallback(
    async (values, { setSubmitting, resetForm }) => {
      try {
        handleSetError(undefined)

        const {
          _activation_mode,
          _device_class,
          _registration,
          _external_servers,
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

  let appKeyPlaceholder = undefined
  let nwkKeyPlaceholder = undefined
  if (!mayEditKeys) {
    appKeyPlaceholder = sharedMessages.insufficientAppKeyRights
    nwkKeyPlaceholder = sharedMessages.insufficientNwkKeyRights
  }

  return (
    <Form
      onSubmit={handleSubmit}
      validationSchema={validationSchema}
      validationContext={validationContext}
      initialValues={initialValues}
      error={error}
      formikRef={formRef}
    >
      {!isNone && (
        <>
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
          />
        </>
      )}
      {!isNone && <hr />}
      <AdvancedSettingsSection
        jsEnabled={jsConfig.enabled}
        nsEnabled={nsConfig.enabled}
        activationMode={activationMode}
        onActivationModeChange={handleActivationModeChange}
        deviceClass={deviceClass}
        onDeviceClassChange={handleDeviceClassChange}
      />
      <hr />
      {!isNone && (
        <>
          {!isMulticast && (
            <Form.Field
              title={sharedMessages.devEUI}
              name="ids.dev_eui"
              type="byte"
              min={8}
              max={8}
              required={isOTAA}
              component={Input}
              tooltipId={tooltipIds.DEV_EUI}
            />
          )}
          {(isABP || isMulticast) && (
            <>
              <DevAddrInput title={sharedMessages.devAddr} name="session.dev_addr" required />
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
        </>
      )}
      <Form.Field
        required
        title={sharedMessages.devID}
        name="ids.device_id"
        placeholder={sharedMessages.deviceIdPlaceholder}
        component={Input}
        tooltipId={tooltipIds.DEVICE_ID}
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
  asConfig: PropTypes.stackComponent.isRequired,
  createDevice: PropTypes.func.isRequired,
  createDeviceSuccess: PropTypes.func.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  nsConfig: PropTypes.stackComponent.isRequired,
  prefixes: PropTypes.euiPrefixes.isRequired,
}

export default withBreadcrumb('devices.add.manually', props => (
  <Breadcrumb path={`/applications/${props.appId}/devices/add/repository`} content={m.register} />
))(ManualForm)
