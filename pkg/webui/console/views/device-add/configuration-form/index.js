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
import { merge } from 'lodash'

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Select from '@ttn-lw/components/select'
import Radio from '@ttn-lw/components/radio-button'
import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { ACTIVATION_MODES, LORAWAN_VERSIONS } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const messages = defineMessages({
  activationModeWarning: 'Activation mode selection unavailable',
  nsActivationModeWarning: 'ABP and multicast activation mode selection unavailable',
  jsActivationModeWarning: 'OTAA mode selection unavailable',
  activationModeNone: 'Do not configure activation',
  start: 'Start',
  preparation: 'Preparation',
})

const ConfigurationForm = React.memo(props => {
  const { onSubmit, nsConfig, jsConfig, asConfig, initialValues, mayEditKeys } = props

  const asEnabled = asConfig.enabled
  const jsEnabled = jsConfig.enabled
  const nsEnabled = nsConfig.enabled
  const asUrl = asEnabled ? asConfig.base_url : undefined
  const jsUrl = jsEnabled ? jsConfig.base_url : undefined
  const nsUrl = nsEnabled ? nsConfig.base_url : undefined

  const canCreateJs = jsConfig.enabled
  // There is no sensible way to create ABP or multicast end devices without the right to
  // set session keys, or if the Network Server is unavailable.
  const canCreateNs = nsConfig.enabled && mayEditKeys

  const formRef = React.useRef(null)

  const validationContext = React.useMemo(
    () => ({
      asUrl,
      nsUrl,
      jsUrl,
      asEnabled,
      jsEnabled,
      nsEnabled,
      mayEditKeys,
    }),
    [asEnabled, asUrl, jsEnabled, jsUrl, mayEditKeys, nsEnabled, nsUrl],
  )

  const initialActivationMode = canCreateJs
    ? ACTIVATION_MODES.OTAA
    : canCreateNs
    ? ACTIVATION_MODES.ABP
    : ACTIVATION_MODES.NONE

  const [activationMode, setActivationMode] = React.useState(initialActivationMode)
  const handleActivationModeChange = React.useCallback(
    mode => {
      const { setValues, values } = formRef.current

      setActivationMode(mode)
      setValues(
        validationSchema.cast(
          {
            ...values,
            _activation_mode: mode,
          },
          {
            context: validationContext,
          },
        ),
      )
    },
    [validationContext],
  )

  const [externalJs, setExternalJs] = React.useState(!jsEnabled)
  const handleExternalJsChange = React.useCallback(
    evt => {
      const { checked } = evt.target
      const { setValues, values } = formRef.current

      setExternalJs(checked)
      setValues(
        validationSchema.cast(
          {
            ...values,
            _external_js: checked,
          },
          {
            context: validationContext,
          },
        ),
      )
    },
    [validationContext],
  )

  const formInitialValues = React.useMemo(() => {
    const { jsEnabled } = validationContext

    return validationSchema.cast(
      merge(
        {
          _external_js: !jsEnabled,
          _activation_mode: initialActivationMode,
          application_server_address: undefined,
          network_server_address: undefined,
          join_server_address: undefined,
          lorawan_version: '',
        },
        initialValues,
      ),
      {
        context: validationContext,
      },
    )
  }, [initialActivationMode, initialValues, validationContext])

  const onFormSubmit = React.useCallback(
    (values, formikBag) => {
      const { _activation_mode, _external_js, ...configuration } = validationSchema.cast(values, {
        context: validationContext,
      })

      return onSubmit(configuration, formikBag)
    },
    [onSubmit, validationContext],
  )

  let activationModeWarning
  if (!canCreateNs && !canCreateJs) {
    activationModeWarning = messages.activationModeWarning
  } else if (!canCreateJs) {
    activationModeWarning = messages.jsActivationModeWarning
  } else if (!canCreateNs) {
    activationModeWarning = messages.nsActivationModeWarning
  }

  // We do not want to show the external JS option if the user is on JS only
  // deployment.
  // See https://github.com/TheThingsNetwork/lorawan-stack/issues/2119#issuecomment-597736420
  const showExternalJs = !(jsEnabled && !nsEnabled && !asEnabled)

  return (
    <Form
      validationSchema={validationSchema}
      validationContext={validationContext}
      initialValues={formInitialValues}
      onSubmit={onFormSubmit}
      formikRef={formRef}
    >
      <Message component="h4" content={messages.preparation} />
      <Form.Field
        title={sharedMessages.activationMode}
        name="_activation_mode"
        component={Radio.Group}
        horizontal={false}
        disabled={!nsConfig.enabled && !jsConfig.enabled}
        required={nsConfig.enabled || jsConfig.enabled}
        warning={activationModeWarning}
        onChange={handleActivationModeChange}
      >
        <Radio
          label={sharedMessages.otaa}
          value={ACTIVATION_MODES.OTAA}
          disabled={!canCreateJs}
          autoFocus={canCreateJs}
        />
        <Radio
          label={sharedMessages.abp}
          value={ACTIVATION_MODES.ABP}
          disabled={!canCreateNs}
          autoFocus={!canCreateJs && canCreateNs}
        />
        <Radio
          label={sharedMessages.multicast}
          value={ACTIVATION_MODES.MULTICAST}
          disabled={!canCreateNs}
        />
        <Radio label={messages.activationModeNone} value={ACTIVATION_MODES.NONE} />
      </Form.Field>
      {activationMode !== ACTIVATION_MODES.NONE && (
        <>
          <Form.Field
            required
            autoFocus={!canCreateJs && !canCreateNs}
            title={sharedMessages.macVersion}
            description={sharedMessages.macVersionDescription}
            name="lorawan_version"
            component={Select}
            options={LORAWAN_VERSIONS}
          />
          <Form.Field
            title={sharedMessages.networkServerAddress}
            placeholder={sharedMessages.addressPlaceholder}
            name="network_server_address"
            component={Input}
          />
          <Form.Field
            title={sharedMessages.applicationServerAddress}
            placeholder={sharedMessages.addressPlaceholder}
            name="application_server_address"
            component={Input}
          />
        </>
      )}
      {activationMode === ACTIVATION_MODES.OTAA && (
        <>
          {showExternalJs && (
            <Form.Field
              title={sharedMessages.externalJoinServer}
              description={sharedMessages.externalJoinServerDescription}
              name="_external_js"
              onChange={handleExternalJsChange}
              component={Checkbox}
            />
          )}
          <Form.Field
            title={sharedMessages.joinServerAddress}
            placeholder={externalJs ? sharedMessages.external : sharedMessages.addressPlaceholder}
            name="join_server_address"
            component={Input}
            disabled={externalJs}
          />
        </>
      )}
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={messages.start} />
      </SubmitBar>
    </Form>
  )
})

ConfigurationForm.propTypes = {
  asConfig: PropTypes.stackComponent.isRequired,
  initialValues: PropTypes.shape({
    application_server_address: PropTypes.string,
    network_server_address: PropTypes.string,
    join_server_address: PropTypes.string,
    lorawan_version: PropTypes.string,
  }),
  jsConfig: PropTypes.stackComponent.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  nsConfig: PropTypes.stackComponent.isRequired,
  onSubmit: PropTypes.func.isRequired,
}

ConfigurationForm.defaultProps = {
  initialValues: {},
}

export default ConfigurationForm
