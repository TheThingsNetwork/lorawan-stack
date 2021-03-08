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

import glossaryId from '@console/constants/glossary-ids'

import Spinner from '@ttn-lw/components/spinner'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Overlay from '@ttn-lw/components/overlay'
import Radio from '@ttn-lw/components/radio-button'

import Message from '@ttn-lw/lib/components/message'

import JoinEUIPRefixesInput from '@console/components/join-eui-prefixes-input'

import DevAddrInput from '@console/containers/dev-addr-input'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'

import { useRepositoryContext } from '../context'
import { selectBand } from '../reducer'
import { REGISTRATION_TYPES } from '../../utils'

import FreqPlansSelect from './freq-plans-select'

const m = defineMessages({
  fetching: 'Fetching template…',
  afterRegistration: 'After registration',
  singleRegistration: 'View registered end device',
  multipleRegistration: 'Register another end device of this type',
})

const Registration = props => {
  const { template, fetching, prefixes, mayEditKeys, onIdPrefill, onIdSelect } = props
  const state = useRepositoryContext()
  const hasTemplate = Boolean(template)
  const idInputRef = React.useRef(null)

  const handleIdFocus = React.useCallback(() => {
    onIdSelect(idInputRef)
  }, [idInputRef, onIdSelect])

  if (!hasTemplate || (fetching && !hasTemplate)) {
    return (
      <Spinner center after={0}>
        <Message content={m.fetching} />
      </Spinner>
    )
  }

  const band = selectBand(state)
  const { end_device } = template
  const { supports_join, lorawan_version } = end_device

  const isOTAA = supports_join
  const lwVersion = parseLorawanMacVersion(lorawan_version)

  return (
    <Overlay visible={fetching} loading={fetching} spinnerMessage={m.fetching}>
      <div data-test-id="device-registration">
        <FreqPlansSelect
          required
          glossaryId={glossaryId.FREQUENCY_PLAN}
          name="frequency_plan_id"
          bandId={band}
        />
        {isOTAA && (
          <>
            <Form.Field
              title={lwVersion < 104 ? sharedMessages.appEUI : sharedMessages.joinEUI}
              component={JoinEUIPRefixesInput}
              name="ids.join_eui"
              description={
                lwVersion < 104
                  ? sharedMessages.appEUIDescription
                  : sharedMessages.joinEUIDescription
              }
              prefixes={prefixes}
              required
              showPrefixes
              glossaryId={lwVersion < 104 ? glossaryId.APP_EUI : glossaryId.JOIN_EUI}
            />
            <Form.Field
              title={sharedMessages.devEUI}
              name="ids.dev_eui"
              type="byte"
              min={8}
              max={8}
              description={sharedMessages.deviceEUIDescription}
              required
              component={Input}
              glossaryId={glossaryId.DEV_EUI}
              onBlur={onIdPrefill}
            />
            <Form.Field
              required
              title={sharedMessages.appKey}
              name="root_keys.app_key.key"
              type="byte"
              min={16}
              max={16}
              description={
                lwVersion >= 110
                  ? sharedMessages.appKeyNewDescription
                  : sharedMessages.appKeyDescription
              }
              component={Input.Generate}
              disabled={!mayEditKeys}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
              glossaryId={glossaryId.APP_KEY}
            />
            {lwVersion >= 110 && (
              <Form.Field
                required
                title={sharedMessages.nwkKey}
                name="root_keys.nwk_key.key"
                type="byte"
                min={16}
                max={16}
                description={sharedMessages.nwkKeyDescription}
                component={Input.Generate}
                disabled={!mayEditKeys}
                mayGenerateValue={mayEditKeys}
                onGenerateValue={generate16BytesKey}
              />
            )}
          </>
        )}
        {!isOTAA && (
          <>
            <DevAddrInput
              title={sharedMessages.devAddr}
              name="session.dev_addr"
              description={sharedMessages.deviceAddrDescription}
              required
            />
            {lwVersion === 104 && (
              <Form.Field
                title={sharedMessages.devEUI}
                name="ids.dev_eui"
                type="byte"
                min={8}
                max={8}
                description={sharedMessages.deviceEUIDescription}
                required
                component={Input}
                glossaryId={glossaryId.DEV_EUI}
              />
            )}
            <Form.Field
              required={mayEditKeys}
              title={sharedMessages.appSKey}
              name="session.keys.app_s_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.appSKeyDescription}
              component={Input.Generate}
              mayGenerateValue={mayEditKeys}
              onGenerateValue={generate16BytesKey}
            />
            <Form.Field
              mayGenerateValue
              title={lwVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey}
              name="session.keys.f_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              required
              description={
                lwVersion >= 110
                  ? sharedMessages.fNwkSIntKeyDescription
                  : sharedMessages.nwkSKeyDescription
              }
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
              glossaryId={
                lwVersion >= 110
                  ? glossaryId.NETWORK_SESSION_KEY
                  : glossaryId.FORWARDING_NETWORK_SESSION_INTEGRITY_KEY
              }
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
                glossaryId={glossaryId.SERVING_NETWORK_SESSION_INTEGRITY_KEY}
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
                glossaryId={glossaryId.NETWORK_SESSION_ENCRYPTION_KEY}
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
          onFocus={handleIdFocus}
          inputRef={idInputRef}
        />
        <Form.Field title={m.afterRegistration} name="_registration" component={Radio.Group}>
          <Radio label={m.singleRegistration} value={REGISTRATION_TYPES.SINGLE} />
          <Radio label={m.multipleRegistration} value={REGISTRATION_TYPES.MULTIPLE} />
        </Form.Field>
      </div>
    </Overlay>
  )
}

Registration.propTypes = {
  fetching: PropTypes.bool,
  mayEditKeys: PropTypes.bool.isRequired,
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
