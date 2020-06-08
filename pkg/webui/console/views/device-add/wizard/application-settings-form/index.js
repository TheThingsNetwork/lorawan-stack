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

import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Wizard from '@ttn-lw/components/wizard'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { generate16BytesKey } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const messages = defineMessages({
  skipCryptoTitle: 'Skip payload encryption and decryption',
  skipCryptoDescription: 'Skip decryption of uplink payloads and encryption of downlink payloads',
  skipCryptoPlaceholder: 'Encryption/decryption disabled',
})

const defaultFormValues = {
  skip_payload_crypto: false,
  session: {
    keys: {
      app_s_key: {
        key: '',
      },
    },
  },
}

const ApplicationSettingsForm = React.memo(props => {
  const { mayEditKeys, error } = props

  const validationContext = React.useMemo(() => ({ mayEditKeys }), [mayEditKeys])

  const formRef = React.useRef(null)

  const [skipCrypto, setSkipCrypto] = React.useState(false)
  const handleSkipCryptoChange = React.useCallback(
    evt => {
      const { checked } = evt.target
      const { setValues, values } = formRef.current

      setSkipCrypto(checked)
      setValues(
        validationSchema.cast(
          {
            ...values,
            skip_payload_crypto: checked,
            session: {
              keys: {
                app_s_key: {
                  key: '',
                },
              },
            },
          },
          { context: validationContext },
        ),
      )
    },
    [validationContext],
  )

  return (
    <Wizard.Form
      initialValues={defaultFormValues}
      ref={formRef}
      validationSchema={validationSchema}
      validationContext={validationContext}
      error={error}
    >
      <Form.Field
        autoFocus
        title={messages.skipCryptoTitle}
        name="skip_payload_crypto"
        description={messages.skipCryptoDescription}
        component={Checkbox}
        onChange={handleSkipCryptoChange}
      />
      {mayEditKeys && (
        <Form.Field
          required={mayEditKeys && !skipCrypto}
          title={sharedMessages.appSKey}
          name="session.keys.app_s_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={skipCrypto ? messages.skipCryptoPlaceholder : undefined}
          disabled={skipCrypto}
          description={sharedMessages.appSKeyDescription}
          component={Input.Generate}
          mayGenerateValue={mayEditKeys && !skipCrypto}
          onGenerateValue={generate16BytesKey}
        />
      )}
    </Wizard.Form>
  )
})

ApplicationSettingsForm.propTypes = {
  error: PropTypes.error,
  mayEditKeys: PropTypes.bool.isRequired,
}

ApplicationSettingsForm.defaultProps = {
  error: undefined,
}

const WrappedApplicationSettingsForm = withBreadcrumb('device.add.steps.app', props => (
  <Breadcrumb path={props.match.url} content={props.title} />
))(ApplicationSettingsForm)

WrappedApplicationSettingsForm.propTypes = {
  match: PropTypes.match.isRequired,
  title: PropTypes.message.isRequired,
}

export default WrappedApplicationSettingsForm
