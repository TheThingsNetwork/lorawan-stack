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

import SubmitButton from '../../../../components/submit-button'
import SubmitBar from '../../../../components/submit-bar'
import Input from '../../../../components/input'
import Checkbox from '../../../../components/checkbox'
import Form from '../../../../components/form'

import diff from '../../../../lib/diff'
import m from '../../../components/device-data-form/messages'
import PropTypes from '../../../../lib/prop-types'
import sharedMessages from '../../../../lib/shared-messages'

import { parseLorawanMacVersion, hasExternalJs } from '../utils'
import validationSchema from './validation-schema'

// The Join Server can store end device fields while not exposing the root keys. This means
// that the `root_keys` object is present, same for `root_keys.nwk_key` and `root_keys.app_key`,
// while `root_keys.nwk_key.key` == nil or `root_keys.app_key.key == nil` must hold.
// See https://github.com/TheThingsNetwork/lorawan-stack/issues/1473
const isNwkKeyHidden = ({ root_keys }) =>
  Boolean(root_keys) && Boolean(root_keys.nwk_key) && !Boolean(root_keys.nwk_key.key)
const isAppKeyHidden = ({ root_keys }) =>
  Boolean(root_keys) && Boolean(root_keys.app_key) && !Boolean(root_keys.app_key.key)

const JoinServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess } = props

  // Fallback to 1.1.0 in case NS is not available and lorawan version is not set present.
  const isNewLorawanVersion = parseLorawanMacVersion(device.lorawan_version || '1.1.0') >= 110
  const externalJs = hasExternalJs(device)

  const formRef = React.useRef(null)
  const [error, setError] = React.useState('')
  const [resetsJoinNonces, setResetsJoinNonces] = React.useState(device.resets_join_nonces)

  // Setup and memoize initial form state.
  const initialValues = React.useMemo(() => {
    const values = {
      ...device,
      _external_js: hasExternalJs(device),
      _lorawan_version: device.lorawan_version,
    }

    return validationSchema.cast(values)
  }, [device])

  // Setup and memoize callbacks for changes to `resets_join_nonces` for displaying the field warning.
  const handleResetsJoinNoncesChange = React.useCallback(
    evt => {
      setResetsJoinNonces(evt.target.checked)
    },
    [setResetsJoinNonces],
  )

  const onFormSubmit = React.useCallback(
    async (values, { setSubmitting, resetForm }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, ['_external_js', '_lorawan_version'])

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

  let appKeyPlaceholder = m.leaveBlankPlaceholder
  if (externalJs) {
    appKeyPlaceholder = sharedMessages.provisionedOnExternalJoinServer
  } else if (appKeyHidden) {
    appKeyPlaceholder = m.unexposed
  }

  let nwkKeyPlaceholder = m.leaveBlankPlaceholder
  if (externalJs) {
    nwkKeyPlaceholder = sharedMessages.provisionedOnExternalJoinServer
  } else if (nwkKeyHidden) {
    nwkKeyPlaceholder = m.unexposed
  }

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
        title={sharedMessages.appKey}
        name="root_keys.app_key.key"
        type="byte"
        min={16}
        max={16}
        placeholder={appKeyPlaceholder}
        description={isNewLorawanVersion ? m.appKeyNewDescription : m.appKeyDescription}
        component={Input}
        disabled={externalJs || appKeyHidden}
      />
      {isNewLorawanVersion && (
        <Form.Field
          title={sharedMessages.nwkKey}
          name="root_keys.nwk_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={nwkKeyPlaceholder}
          description={m.nwkKeyDescription}
          component={Input}
          disabled={externalJs || nwkKeyHidden}
        />
      )}
      {isNewLorawanVersion && (
        <Form.Field
          title={m.resetsJoinNonces}
          onChange={handleResetsJoinNoncesChange}
          warning={resetsJoinNonces ? m.resetWarning : undefined}
          name="resets_join_nonces"
          component={Checkbox}
          disabled={externalJs}
        />
      )}
      <Form.Field
        title={m.homeNetID}
        name="net_id"
        type="byte"
        min={3}
        max={3}
        component={Input}
        disabled={externalJs}
      />
      <Form.Field
        title={m.asServerID}
        name="application_server_id"
        description={m.asServerIDDescription}
        component={Input}
        disabled={externalJs}
      />
      <Form.Field
        title={m.asServerKekLabel}
        name="application_server_kek_label"
        description={m.asServerKekLabelDescription}
        component={Input}
        disabled={externalJs}
      />
      <Form.Field
        title={m.nsServerKekLabel}
        name="network_server_kek_label"
        description={m.nsServerKekLabelDescription}
        component={Input}
        disabled={externalJs}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

JoinServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default JoinServerForm
