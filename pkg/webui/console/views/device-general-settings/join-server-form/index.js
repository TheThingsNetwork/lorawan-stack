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
// that the `root_keys` object is present while `root_keys.nwk_key` == nil or `root_keys.app_key == nil`
// must hold. See https://github.com/TheThingsNetwork/lorawan-stack/issues/1473
const isNwkKeyHidden = ({ root_keys }) => Boolean(root_keys) && !Boolean(root_keys.nwk_key)
const isAppKeyHidden = ({ root_keys }) => Boolean(root_keys) && !Boolean(root_keys.app_key)

const actionTypes = Object.freeze({
  EXTERNAL_JS: 'change-external-js',
  RESETS_JOIN_NONCES: 'change-resets-join-nonces',
})

const reducer = (state, action) => {
  switch (action.type) {
    case actionTypes.EXTERNAL_JS:
      const externalJs = !state.externalJs
      return {
        externalJs,
        resetsJoinNonces: externalJs ? false : state.resetsJoinNonces,
      }
    case actionTypes.RESETS_JOIN_NONCES:
      return {
        ...state,
        resetsJoinNonces: !state.resetsJoinNonces,
      }
    default:
      return state
  }
}

const JoinServerForm = React.memo(props => {
  const { device, onSubmit, jsConfig } = props

  const isNewLorawanVersion = parseLorawanMacVersion(device.lorawan_version) >= 110

  // Setup and memoize initial reducer state.
  const initialState = React.useMemo(() => {
    const { resets_join_nonces: resetsJoinNonces = false } = device
    const externalJs = hasExternalJs(device)

    return {
      resetsJoinNonces,
      externalJs,
    }
  }, [device])
  const [state, dispatch] = React.useReducer(reducer, initialState)

  // Setup and memoize initial form state.
  const initialValues = React.useMemo(() => {
    const externalJs = hasExternalJs(device)
    const {
      root_keys = {
        nwk_key: {},
        app_key: {},
      },
      resets_join_nonces,
      join_server_address,
      lorawan_version,
    } = device

    return {
      join_server_address: externalJs ? undefined : join_server_address,
      resets_join_nonces,
      root_keys,
      _external_js: hasExternalJs(device),
      _lorawan_version: lorawan_version,
    }
  }, [device])

  // Setup and memoize callbacks for changes to `resets_join_nonces` and `_external_js`.
  const handleResetsJoinNoncesChange = React.useCallback(() => {
    dispatch({
      type: actionTypes.RESETS_JOIN_NONCES,
    })
  }, [])
  // Note: If the end device is provisioned on an external JS, we reset `root_keys` and
  // `resets_join_nonces` fields.
  const handleExternalJsChange = React.useCallback(
    evt => {
      dispatch({
        type: actionTypes.EXTERNAL_JS,
      })

      const { checked: externalJsChecked } = evt.target
      const { setValues, state: formState } = formRef.current

      if (externalJsChecked) {
        setValues({
          ...formState.values,
          root_keys: {
            nwk_key: {},
            app_key: {},
          },
          resets_join_nonces: false,
          join_server_address: undefined,
          _external_js: externalJsChecked,
        })
      } else {
        let { join_server_address } = initialValues
        const { resets_join_nonces, root_keys } = initialValues
        if (!Boolean(join_server_address)) {
          // always fallback to the default js address when resetting from
          // the 'provisioned by external js' option.
          join_server_address = new URL(jsConfig.base_url).hostname
        }

        setValues({
          ...formState.values,
          join_server_address,
          root_keys,
          _external_js: externalJsChecked,
          resets_join_nonces,
        })
      }
    },
    [initialValues, jsConfig.base_url],
  )

  const formRef = React.useRef(null)
  const [error, setError] = React.useState('')

  const onFormSubmit = React.useCallback(
    async (values, { setSubmitting, resetForm }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues, ['_external_js', '_lorawan_version'])

      setError('')
      try {
        await onSubmit(updatedValues)
        resetForm(castedValues)
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit],
  )

  const nwkKeyHidden = isNwkKeyHidden(device)
  const appKeyHidden = isAppKeyHidden(device)

  let appKeyPlaceholder = m.leaveBlankPlaceholder
  if (state.externalJs) {
    appKeyPlaceholder = sharedMessages.provisionedOnExternalJoinServer
  } else if (appKeyHidden) {
    appKeyPlaceholder = m.unexposed
  }

  let nwkKeyPlaceholder = m.leaveBlankPlaceholder
  if (state.externalJs) {
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
        title={m.externalJoinServer}
        description={m.externalJoinServerDescription}
        name="_external_js"
        onChange={handleExternalJsChange}
        component={Checkbox}
      />
      <Form.Field
        title={sharedMessages.joinServerAddress}
        placeholder={state.externalJs ? m.external : sharedMessages.addressPlaceholder}
        name="join_server_address"
        component={Input}
        disabled={state.externalJs}
      />
      <Form.Field
        title={sharedMessages.appKey}
        name="root_keys.app_key.key"
        type="byte"
        min={16}
        max={16}
        placeholder={appKeyPlaceholder}
        description={m.appKeyDescription}
        component={Input}
        disabled={state.externalJs || appKeyHidden}
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
          disabled={state.externalJs || nwkKeyHidden}
        />
      )}
      {isNewLorawanVersion && (
        <Form.Field
          title={m.resetsJoinNonces}
          onChange={handleResetsJoinNoncesChange}
          warning={state.resetsJoinNonces ? m.resetWarning : undefined}
          name="resets_join_nonces"
          component={Checkbox}
          disabled={state.externalJs}
        />
      )}
      <Form.Field
        title={sharedMessages.macVersion}
        name="_lorawan_version"
        component={Input}
        type="hidden"
        hidden
        disabled
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

JoinServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  onSubmit: PropTypes.func.isRequired,
}

export default JoinServerForm
