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

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Notification from '@ttn-lw/components/notification'

import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'

import messages from '../messages'
import { hasExternalJs } from '../utils'

import validationSchema from './validation-schema'

// The Join Server can store end device fields while not exposing the root keys.
// This means that the `root_keys` object is present, same for
// `root_keys.nwk_key` and `root_keys.app_key`, while
// `root_keys.nwk_key.key` == nil or `root_keys.app_key.key == nil` must  hold.
// See https://github.com/TheThingsNetwork/lorawan-stack/issues/1473
const isNwkKeyHidden = ({ root_keys }) =>
  Boolean(root_keys) && Boolean(root_keys.nwk_key) && !Boolean(root_keys.nwk_key.key)
const isAppKeyHidden = ({ root_keys }) =>
  Boolean(root_keys) && Boolean(root_keys.app_key) && !Boolean(root_keys.app_key.key)

const JoinServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess, mayReadKeys, mayEditKeys } = props

  // Fallback to 1.1.0 in case NS is not available and lorawan version is not
  // set present.
  const isNewLorawanVersion = parseLorawanMacVersion(device.lorawan_version || '1.1.0') >= 110
  const externalJs = hasExternalJs(device) && mayReadKeys

  const formRef = React.useRef(null)
  const [error, setError] = React.useState('')
  const [resetsJoinNonces, setResetsJoinNonces] = React.useState(device.resets_join_nonces)

  // Setup and memoize initial form state.
  const initialValues = React.useMemo(() => {
    const values = {
      ...device,
      _external_js: hasExternalJs(device) && mayReadKeys,
      _lorawan_version: device.lorawan_version,
      _may_edit_keys: mayEditKeys,
      _may_read_keys: mayReadKeys,
    }

    return validationSchema.cast(values)
  }, [device, mayEditKeys, mayReadKeys])

  // Setup and memoize callbacks for changes to `resets_join_nonces` for
  // displaying the field warning.
  const handleResetsJoinNoncesChange = React.useCallback(
    evt => {
      setResetsJoinNonces(evt.target.checked)
    },
    [setResetsJoinNonces],
  )

  const onFormSubmit = React.useCallback(
    async (values, { setSubmitting, resetForm }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, [
        '_external_js',
        '_lorawan_version',
        '_may_edit_keys',
        '_may_read_keys',
      ])

      setError('')
      try {
        await onSubmit(updatedValues)
        resetForm(castedValues)
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit, onSubmitSuccess],
  )

  const nwkKeyHidden = isNwkKeyHidden(device)
  const appKeyHidden = isAppKeyHidden(device)

  let appKeyPlaceholder
  if (externalJs) {
    appKeyPlaceholder = sharedMessages.provisionedOnExternalJoinServer
  } else if (appKeyHidden) {
    appKeyPlaceholder = sharedMessages.unexposed
  }

  let nwkKeyPlaceholder
  if (externalJs) {
    nwkKeyPlaceholder = sharedMessages.provisionedOnExternalJoinServer
  } else if (nwkKeyHidden) {
    nwkKeyPlaceholder = sharedMessages.unexposed
  }

  // Notify the user that the root keys might be there, but since there are no
  // rights to read the keys we cannot display them.
  const showResetNotification = !mayReadKeys && mayEditKeys && !Boolean(device.root_keys)

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      formikRef={formRef}
      error={error}
      enableReinitialize
    >
      <Form.Field
        title={sharedMessages.homeNetID}
        description={sharedMessages.homeNetIDDescription}
        name="net_id"
        type="byte"
        min={3}
        max={3}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.asServerID}
        name="application_server_id"
        description={sharedMessages.asServerIDDescription}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.asServerKekLabel}
        name="application_server_kek_label"
        description={sharedMessages.asServerKekLabelDescription}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.nsServerKekLabel}
        name="network_server_kek_label"
        description={sharedMessages.nsServerKekLabelDescription}
        component={Input}
      />
      {isNewLorawanVersion && (
        <Form.Field
          title={sharedMessages.resetsJoinNonces}
          onChange={handleResetsJoinNoncesChange}
          warning={resetsJoinNonces ? sharedMessages.resetWarning : undefined}
          name="resets_join_nonces"
          component={Checkbox}
        />
      )}
      {showResetNotification && <Notification content={messages.keysResetWarning} info small />}
      <Form.Field
        title={sharedMessages.appKey}
        name="root_keys.app_key.key"
        type="byte"
        min={16}
        max={16}
        placeholder={appKeyPlaceholder}
        description={
          isNewLorawanVersion
            ? sharedMessages.appKeyNewDescription
            : sharedMessages.appKeyDescription
        }
        component={Input.Generate}
        disabled={appKeyHidden || !mayEditKeys}
        mayGenerateValue={mayEditKeys && !appKeyHidden}
        onGenerateValue={generate16BytesKey}
      />
      {isNewLorawanVersion && (
        <Form.Field
          title={sharedMessages.nwkKey}
          name="root_keys.nwk_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={nwkKeyPlaceholder}
          description={sharedMessages.nwkKeyDescription}
          component={Input.Generate}
          disabled={nwkKeyHidden || !mayEditKeys}
          mayGenerateValue={mayEditKeys && !nwkKeyHidden}
          onGenerateValue={generate16BytesKey}
        />
      )}
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

JoinServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  mayReadKeys: PropTypes.bool.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default JoinServerForm
