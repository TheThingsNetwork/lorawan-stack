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

import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Wizard, { useWizardContext } from '@ttn-lw/components/wizard'

import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { generate16BytesKey } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const defaultFormValues = {
  skip_payload_crypto: false,
}

const ApplicationSettingsForm = props => {
  const { mayEditKeys } = props
  const { snapshot } = useWizardContext()

  const validationContext = React.useMemo(() => ({ mayEditKeys }), [mayEditKeys])

  const formRef = React.useRef(null)

  const [skipCrypto, setSkipCrypto] = React.useState(snapshot.skip_payload_crypto)
  const handleSkipCryptoChange = React.useCallback(
    evt => {
      const { setValues, values } = formRef.current

      setSkipCrypto(evt.target.checked)
      setValues(
        validationSchema.cast(
          { ...values, skip_payload_crypto: evt.target.checked },
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
    >
      <Form.Field
        autoFocus
        title={sharedMessages.skipCryptoTitle}
        name="skip_payload_crypto"
        description={sharedMessages.skipCryptoDescription}
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
          placeholder={skipCrypto ? sharedMessages.skipCryptoPlaceholder : undefined}
          disabled={skipCrypto}
          component={Input.Generate}
          mayGenerateValue={mayEditKeys && !skipCrypto}
          onGenerateValue={generate16BytesKey}
          glossaryId={glossaryIds.APP_SESSION_KEY}
        />
      )}
    </Wizard.Form>
  )
}

ApplicationSettingsForm.propTypes = {
  mayEditKeys: PropTypes.bool.isRequired,
}

const WrappedApplicationSettingsForm = withBreadcrumb('device.add.steps.app', props => (
  <Breadcrumb path={props.match.url} content={props.title} />
))(ApplicationSettingsForm)

WrappedApplicationSettingsForm.propTypes = {
  match: PropTypes.match.isRequired,
  title: PropTypes.message.isRequired,
}

export default WrappedApplicationSettingsForm
