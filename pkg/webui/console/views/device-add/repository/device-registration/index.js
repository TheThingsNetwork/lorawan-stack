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
import classnames from 'classnames'

import Spinner from '@ttn-lw/components/spinner'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Overlay from '@ttn-lw/components/overlay'
import Radio from '@ttn-lw/components/radio-button'

import Message from '@ttn-lw/lib/components/message'

import JoinEUIPRefixesInput from '@console/components/join-eui-prefixes-input'

import DevAddrInput from '@console/containers/dev-addr-input'
import FreqPlansSelect from '@console/containers/device-freq-plans-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import env from '@ttn-lw/lib/env'

import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'

import { useRepositoryContext } from '../context'
import { selectBand } from '../reducer'
import { REGISTRATION_TYPES } from '../../utils'
import messages from '../../messages'
import style from '../../device-add.styl'

const m = defineMessages({
  fetching: 'Fetching template…',
})

const Registration = props => {
  const {
    template,
    fetching,
    prefixes,
    mayEditKeys,
    onIdPrefill,
    onIdSelect,
    generateDevEUI,
    applicationDevEUICounter,
    nsEnabled,
    asEnabled,
    jsEnabled,
  } = props
  const state = useRepositoryContext()
  const hasTemplate = Boolean(template)
  const idInputRef = React.useRef(null)
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

  const devEUIGenerateDisabled =
    applicationDevEUICounter === env.devEUIConfig.applicationLimit ||
    !env.devEUIConfig.devEUIIssuingEnabled ||
    devEUIGenerated

  const handleIdFocus = React.useCallback(() => {
    onIdSelect(idInputRef)
  }, [idInputRef, onIdSelect])

  const handleGenerate = React.useCallback(async () => {
    try {
      const result = await generateDevEUI()
      setDevEUIGenerated(true)
      euiInputRef.current.focus()
      setErrorMessage(undefined)
      return result
    } catch (error) {
      if (error.details[0].name === 'global_eui_limit_reached') {
        setErrorMessage(sharedMessages.devEUIBlockLimitReached)
      } else setErrorMessage(sharedMessages.unknownError)
      setDevEUIGenerated(true)
    }
  }, [generateDevEUI])

  if (!hasTemplate || (fetching && !hasTemplate)) {
    return (
      <Spinner center after={0}>
        <Message content={m.fetching} />
      </Spinner>
    )
  }

  const indicatorCls = classnames(style.indicator, {
    [style.error]:
      applicationDevEUICounter === env.devEUIConfig.applicationLimit || Boolean(errorMessage),
  })

  const band = selectBand(state)
  const { end_device } = template
  const { supports_join, lorawan_version } = end_device

  const isOTAA = supports_join
  const lwVersion = parseLorawanMacVersion(lorawan_version)

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
      onBlur={onIdPrefill}
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
      onBlur={onIdPrefill}
    />
  )

  return (
    <Overlay visible={fetching} loading={fetching} spinnerMessage={m.fetching}>
      <div data-test-id="device-registration">
        {nsEnabled && (
          <FreqPlansSelect
            required
            tooltipId={tooltipIds.FREQUENCY_PLAN}
            name="frequency_plan_id"
            bandId={band}
          />
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
            {devEUIComponent}
            {jsEnabled && (
              <>
                <Form.Field
                  required
                  title={sharedMessages.appKey}
                  name="root_keys.app_key.key"
                  type="byte"
                  min={16}
                  max={16}
                  component={Input.Generate}
                  disabled={!mayEditKeys}
                  mayGenerateValue={mayEditKeys}
                  onGenerateValue={generate16BytesKey}
                  tooltipId={tooltipIds.APP_KEY}
                  placeholder={appKeyPlaceholder}
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
                    disabled={!mayEditKeys}
                    mayGenerateValue={mayEditKeys}
                    onGenerateValue={generate16BytesKey}
                    placeholder={nwkKeyPlaceholder}
                    tooltipId={tooltipIds.NETWORK_KEY}
                  />
                )}
              </>
            )}
          </>
        )}
        {!isOTAA && (
          <>
            {nsEnabled && (
              <DevAddrInput title={sharedMessages.devAddr} name="session.dev_addr" required />
            )}
            {lwVersion === 104 && devEUIComponent}
            {asEnabled && (
              <Form.Field
                required={mayEditKeys}
                title={sharedMessages.appSKey}
                name="session.keys.app_s_key.key"
                type="byte"
                min={16}
                max={16}
                component={Input.Generate}
                mayGenerateValue={mayEditKeys}
                onGenerateValue={generate16BytesKey}
                tooltipId={tooltipIds.APP_SESSION_KEY}
              />
            )}
            {nsEnabled && (
              <>
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
          </>
        )}
        <Form.Field
          required
          title={sharedMessages.devID}
          name="ids.device_id"
          placeholder={sharedMessages.deviceIdPlaceholder}
          component={Input}
          onFocus={handleIdFocus}
          inputRef={idInputRef}
          tooltipId={tooltipIds.DEVICE_ID}
          description={messages.deviceIdDescription}
        />
        <Form.Field title={messages.afterRegistration} name="_registration" component={Radio.Group}>
          <Radio label={messages.singleRegistration} value={REGISTRATION_TYPES.SINGLE} />
          <Radio label={messages.multipleRegistration} value={REGISTRATION_TYPES.MULTIPLE} />
        </Form.Field>
      </div>
    </Overlay>
  )
}

Registration.propTypes = {
  applicationDevEUICounter: PropTypes.number.isRequired,
  asEnabled: PropTypes.bool.isRequired,
  fetching: PropTypes.bool,
  generateDevEUI: PropTypes.func.isRequired,
  jsEnabled: PropTypes.bool.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  nsEnabled: PropTypes.bool.isRequired,
  onIdPrefill: PropTypes.func.isRequired,
  onIdSelect: PropTypes.func.isRequired,
  prefixes: PropTypes.euiPrefixes.isRequired,
  template: PropTypes.shape({
    end_device: PropTypes.shape({
      supports_join: PropTypes.bool,
      lorawan_version: PropTypes.string.isRequired,
    }),
  }),
}

Registration.defaultProps = {
  fetching: false,
  template: undefined,
}

export default Registration
