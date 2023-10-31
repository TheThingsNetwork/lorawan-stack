// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { useDispatch } from 'react-redux'

import Link from '@ttn-lw/components/link'
import ModalButton from '@ttn-lw/components/button/modal-button'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Radio from '@ttn-lw/components/radio-button'
import Form from '@ttn-lw/components/form'
import Notification from '@ttn-lw/components/notification'
import Checkbox from '@ttn-lw/components/checkbox'
import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'

import PhyVersionInput from '@console/components/phy-version-input'
import LorawanVersionInput from '@console/components/lorawan-version-input'
import MacSettingsSection from '@console/components/mac-settings-section'

import FreqPlansSelect from '@console/containers/device-freq-plans-select'
import DevAddrInput from '@console/containers/dev-addr-input'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import { isBackend, getBackendErrorName } from '@ttn-lw/lib/errors/utils'
import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  parseLorawanMacVersion,
  ACTIVATION_MODES,
  generate16BytesKey,
} from '@console/lib/device-utils'

import messages from '../messages'
import {
  isDeviceABP,
  isDeviceMulticast,
  hasExternalJs,
  isDeviceJoined,
  isDeviceOTAA,
} from '../utils'

import validationSchema from './validation-schema'

const m = defineMessages({
  resetTitle: 'Session and MAC state reset',
  resetButtonTitle: 'Reset session and MAC state',
  resetSuccess: 'End device reset',
  resetFailure: 'There was an error and the end device session and MAC state could not be reset',
  modalMessage:
    'Are you sure you want to reset the session context and MAC state of this end device?{break}{break}This will have the following consequences:<ul><li>For <OTAADocLink>OTAA</OTAADocLink>-activated end devices <b>all session and MAC data is wiped and the end device MUST rejoin</b></li><li>For <ABPDocLink>ABP</ABPDocLink>-activated end devices, session keys, device address and downlink queue are preserved, while <b>the MAC state is reset</b></li></ul>',
})

const defaultValues = {
  mac_settings: {
    ping_slot_periodicity: '',
  },
}

const NetworkServerForm = React.memo(props => {
  const {
    device,
    onSubmit,
    onSubmitSuccess,
    mayEditKeys,
    mayReadKeys,
    onMacReset,
    getDefaultMacSettings,
  } = props
  const {
    multicast = false,
    supports_join = false,
    supports_class_b = false,
    supports_class_c = false,
    version_ids = {},
  } = device

  const isABP = isDeviceABP(device)
  const isMulticast = isDeviceMulticast(device)
  const isJoinedOTAA = isDeviceOTAA(device) && isDeviceJoined(device)
  const bandId = version_ids.band_id

  const validationContext = React.useMemo(
    () => ({
      mayEditKeys,
      mayReadKeys,
      isJoined: isDeviceOTAA(device) && isDeviceJoined(device),
      externalJs: hasExternalJs(device),
    }),
    [device, mayEditKeys, mayReadKeys],
  )

  const formRef = React.useRef(null)

  const [macSettings, setMacSettings] = React.useState({})

  const [phyVersion, setPhyVersion] = React.useState(device.lorawan_phy_version)
  const phyVersionRef = React.useRef()
  const handlePhyVersionChange = React.useCallback(setPhyVersion, [])

  const [error, setError] = React.useState('')

  const [lorawanVersion, setLorawanVersion] = React.useState(device.lorawan_version)
  const lwVersion = parseLorawanMacVersion(lorawanVersion)

  const freqPlanRef = React.useRef(device.frequency_plan_id)
  const [freqPlan, setFreqPlan] = React.useState(device.frequency_plan_id)
  const handleFreqPlanChange = React.useCallback(plan => {
    setFreqPlan(plan.value)
  }, [])

  const [isClassB, setClassB] = React.useState(supports_class_b)
  const handleClassBChange = React.useCallback(evt => {
    const { checked } = evt.target

    setClassB(checked)
  }, [])
  const [isClassC, setClassC] = React.useState(supports_class_c)
  const handleClassCChange = React.useCallback(evt => {
    const { checked } = evt.target

    setClassC(checked)
  }, [])

  React.useEffect(() => {
    const getMacSettings = async (freqPlan, phyVersion) => {
      try {
        const settings = await getDefaultMacSettings(freqPlan, phyVersion)
        setMacSettings(settings)
        if (formRef.current) {
          const { setValues, values } = formRef.current
          setValues(
            validationSchema.cast(
              {
                ...values,
                mac_settings: {
                  ...settings,
                  ...values.mac_settings,
                  // To use `adr_margin` as initial value of `adr.dynamic.margin`.
                  // And to make sure that, if there is already a value set for `adr`, it is not overwritten
                  // by the default mac settings.
                  adr:
                    'dynamic' in values.mac_settings.adr
                      ? {
                          dynamic: {
                            margin: values.mac_settings.adr.dynamic?.margin ?? settings.adr_margin,
                          },
                        }
                      : values.mac_settings.adr,
                },
              },
              { context: validationContext },
            ),
          )
        }
      } catch (err) {
        if (isBackend(err) && getBackendErrorName(err) === 'no_band_version') {
          toast({
            type: toast.types.ERROR,
            message: sharedMessages.fpNotFoundError,
            messageValues: {
              lorawanVersion,
              freqPlan,
              code: msg => <code>{msg}</code>,
            },
          })
        } else {
          toast({
            type: toast.types.ERROR,
            message: messages.macSettingsError,
            messageValues: {
              freqPlan,
              code: msg => <code>{msg}</code>,
            },
          })
        }
      }
    }

    if (freqPlan && phyVersion) {
      if (freqPlanRef.current !== freqPlan || phyVersionRef.current !== phyVersion) {
        freqPlanRef.current = freqPlan
        phyVersionRef.current = phyVersion

        getMacSettings(freqPlan, phyVersion)
      }
    }
  }, [freqPlan, getDefaultMacSettings, lorawanVersion, phyVersion, validationContext])

  const initialActivationMode = supports_join
    ? ACTIVATION_MODES.OTAA
    : multicast
    ? ACTIVATION_MODES.MULTICAST
    : ACTIVATION_MODES.ABP

  const initialValues = React.useMemo(
    () =>
      validationSchema.cast(
        {
          ...defaultValues,
          ...device,
          _activation_mode: initialActivationMode,
          _device_classes: { class_b: isClassB, class_c: isClassC },
          supports_class_b: isClassB,
          supports_class_c: isClassC,
          mac_settings: {
            ...defaultValues.mac_settings,
            ...macSettings,
            ...device.mac_settings,
          },
        },
        { context: validationContext, stripUnknown: true },
      ),
    [device, initialActivationMode, isClassB, isClassC, macSettings, validationContext],
  )

  const dispatch = useDispatch()
  const appId = device.ids.application_ids.application_id
  const devId = device.ids.device_id
  const handleMacReset = React.useCallback(async () => {
    try {
      await dispatch(attachPromise(onMacReset(appId, devId)))
      toast({
        message: m.resetSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (err) {
      toast({
        message: m.resetFailure,
        type: toast.types.ERROR,
      })
    }
  }, [onMacReset, dispatch, devId, appId])

  const handleSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values, {
        context: validationContext,
        stripUnknown: true,
      })

      const updatedValues = diff(device, castedValues, {
        exclude: [
          '_activation_mode',
          '_device_classes',
          'class_b',
          'class_c',
          'mac_settings',
          'f_nwk_s_int_key',
          's_nwk_s_int_key',
          'nwk_s_enc_key',
          'app_s_key',
        ],
      })

      const patch = updatedValues
      // Always submit current `mac_settings` values to avoid overwriting nested entries.
      patch.mac_settings = castedValues.mac_settings

      const isOTAA = values._activation_mode === ACTIVATION_MODES.OTAA
      // Do not update session for joined OTAA end devices.
      if (!isOTAA && castedValues.session && castedValues.session.keys) {
        const { app_s_key, ...keys } = castedValues.session.keys
        patch.session = {
          ...updatedValues.session,
          keys,
        }
      }

      if (patch.session && patch.session.keys && Object.keys(patch.session.keys).length === 0) {
        delete patch.session.keys
      }

      if (patch.session && Object.keys(patch.session).length === 0) {
        delete patch.session
      }

      if (patch.mac_settings.adr) {
        patch.mac_settings.adr_margin = null
        patch.mac_settings.use_adr = null
      }

      setError('')
      try {
        await onSubmit(patch)
        resetForm({ values })
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [device, onSubmit, onSubmitSuccess, validationContext],
  )

  const handleDeviceClassChange = React.useCallback(
    deviceClasses => {
      const { setValues, values } = formRef.current
      setValues(
        validationSchema.cast(
          {
            ...defaultValues,
            ...values,
            _device_classes: deviceClasses,
            mac_settings: {
              ...defaultValues.mac_settings,
              ...macSettings,
              ...values.mac_settings,
            },
          },
          { context: validationContext, stripUnknown: true },
        ),
      )
    },
    [macSettings, validationContext],
  )

  const handleVersionChange = React.useCallback(
    version => {
      const isABP = initialValues._activation_mode === ACTIVATION_MODES.ABP
      const lwVersion = parseLorawanMacVersion(version)
      setLorawanVersion(version)
      const { setValues, values: formValues } = formRef.current
      const { session = {} } = formValues
      const { session: initialSession } = initialValues
      if (lwVersion >= 110) {
        const updatedSession = isABP
          ? {
              dev_addr: session.dev_addr,
              keys: {
                ...session.keys,
                s_nwk_s_int_key:
                  session.keys.s_nwk_s_int_key || initialSession.keys.s_nwk_s_int_key,
                nwk_s_enc_key: session.keys.nwk_s_enc_key || initialSession.keys.nwk_s_enc_key,
              },
            }
          : session
        setValues({
          ...formValues,
          lorawan_version: version,
          session: updatedSession,
        })
      } else {
        const updatedSession = isABP
          ? {
              dev_addr: session.dev_addr,
              keys: {
                f_nwk_s_int_key: session.keys.f_nwk_s_int_key,
              },
            }
          : session
        setValues({
          ...formValues,
          lorawan_version: version,
          session: updatedSession,
        })
      }
    },
    [initialValues],
  )

  // Notify the user that the session keys might be there, but since there are
  // no rights to read the keys we cannot display them.
  const showResetNotification = !mayReadKeys && mayEditKeys && !Boolean(device.session)

  return (
    <Form
      validationSchema={validationSchema}
      validationContext={validationContext}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      error={error}
      formikRef={formRef}
    >
      <FreqPlansSelect
        name="frequency_plan_id"
        required
        tooltipId={tooltipIds.FREQUENCY_PLAN}
        onChange={handleFreqPlanChange}
        bandId={bandId}
      />
      <Form.Field
        title={sharedMessages.macVersion}
        name="lorawan_version"
        component={LorawanVersionInput}
        required
        onChange={handleVersionChange}
        tooltipId={tooltipIds.LORAWAN_VERSION}
        frequencyPlan={freqPlan}
      />
      <Form.Field
        title={sharedMessages.phyVersion}
        name="lorawan_phy_version"
        component={PhyVersionInput}
        required
        tooltipId={tooltipIds.REGIONAL_PARAMETERS}
        lorawanVersion={lorawanVersion}
        onChange={handlePhyVersionChange}
      />
      <Form.Field
        title={sharedMessages.lorawanClassCapabilities}
        name="_device_classes"
        component={Checkbox.Group}
        required={isMulticast}
        tooltipId={tooltipIds.CLASSES}
        onChange={handleDeviceClassChange}
      >
        <Checkbox
          name="class_b"
          label={sharedMessages.supportsClassB}
          onChange={handleClassBChange}
        />
        <Checkbox
          name="class_c"
          label={sharedMessages.supportsClassC}
          onChange={handleClassCChange}
        />
      </Form.Field>
      <Form.Field
        title={sharedMessages.activationMode}
        disabled
        required
        name="_activation_mode"
        component={Radio.Group}
        tooltipId={tooltipIds.ACTIVATION_MODE}
      >
        <Radio label={sharedMessages.otaa} value={ACTIVATION_MODES.OTAA} />
        <Radio label={sharedMessages.abp} value={ACTIVATION_MODES.ABP} />
        <Radio label={sharedMessages.multicast} value={ACTIVATION_MODES.MULTICAST} />
      </Form.Field>
      {(isABP || isMulticast || isJoinedOTAA) && (
        <>
          {showResetNotification && <Notification content={messages.keysResetWarning} info small />}
          <DevAddrInput
            title={sharedMessages.devAddr}
            name="session.dev_addr"
            disabled={!mayEditKeys}
            required={mayReadKeys && mayEditKeys}
          />
          <Form.Field
            required
            title={lwVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
            name="session.keys.f_nwk_s_int_key.key"
            type="byte"
            min={16}
            max={16}
            disabled={!mayEditKeys}
            component={Input.Generate}
            mayGenerateValue={mayEditKeys}
            onGenerateValue={generate16BytesKey}
            tooltipId={lwVersion >= 110 ? undefined : tooltipIds.NETWORK_SESSION_KEY}
            sensitive
          />
          {lwVersion >= 110 && (
            <Form.Field
              required
              title={sharedMessages.sNwkSIKey}
              name="session.keys.s_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.sNwkSIKeyDescription}
              disabled={!mayEditKeys}
              component={Input.Generate}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
              sensitive
            />
          )}
          {lwVersion >= 110 && (
            <Form.Field
              required
              title={sharedMessages.nwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.nwkSEncKeyDescription}
              disabled={!mayEditKeys}
              component={Input.Generate}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
              sensitive
            />
          )}
        </>
      )}
      <Form.InfoField title={m.resetTitle} tooltipId={tooltipIds.RESET_MAC}>
        <ModalButton
          type="button"
          warning
          message={m.resetButtonTitle}
          modalData={{
            children: (
              <div>
                <Message
                  content={m.modalMessage}
                  values={{
                    b: msg => <b>{msg}</b>,
                    ul: msg => <ul>{msg}</ul>,
                    li: msg => <li>{msg}</li>,
                    break: <br />,
                    OTAADocLink: msg => (
                      <Link.DocLink secondary path="/devices/abp-vs-otaa#otaa">
                        {msg}
                      </Link.DocLink>
                    ),
                    ABPDocLink: msg => (
                      <Link.DocLink secondary path="/devices/abp-vs-otaa#abp">
                        {msg}
                      </Link.DocLink>
                    ),
                  }}
                />
              </div>
            ),
          }}
          onApprove={handleMacReset}
        />
      </Form.InfoField>
      <MacSettingsSection
        activationMode={initialActivationMode}
        lorawanVersion={lorawanVersion}
        isClassB={isClassB}
        isClassC={isClassC}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

NetworkServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  getDefaultMacSettings: PropTypes.func.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  mayReadKeys: PropTypes.bool.isRequired,
  onMacReset: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default NetworkServerForm
