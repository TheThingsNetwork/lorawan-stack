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

import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'
import Wizard from '@ttn-lw/components/wizard'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { parseLorawanMacVersion, generate16BytesKey } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const defaultInitialValues = {
  root_keys: {},
  net_id: undefined,
  resets_join_nonces: false,
  application_server_id: undefined,
  application_server_kek_label: undefined,
  network_server_kek_label: undefined,
}

const JoinSettingsForm = React.memo(props => {
  const { lorawanVersion, mayEditKeys, error } = props

  const validationContext = React.useMemo(
    () => ({
      mayEditKeys,
      lorawanVersion,
    }),
    [lorawanVersion, mayEditKeys],
  )

  const [resetsJoinNonces, setResetsJoinNonces] = React.useState(false)
  const handleResetsJoinNoncesChange = React.useCallback(
    evt => {
      setResetsJoinNonces(evt.target.checked)
    },
    [setResetsJoinNonces],
  )

  const lwVersion = parseLorawanMacVersion(lorawanVersion)

  return (
    <Wizard.Form
      error={error}
      initialValues={defaultInitialValues}
      validationContext={validationContext}
      validationSchema={validationSchema}
    >
      {mayEditKeys && (
        <>
          <Form.SubTitle title={sharedMessages.rootKeys} />
          <Form.Field
            required
            autoFocus={mayEditKeys}
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
          />
          {lwVersion >= 110 && (
            <Form.Field
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
      <Form.CollapseSection
        id="advanced-settings"
        title={sharedMessages.advancedSettings}
        initiallyCollapsed={mayEditKeys}
      >
        {lwVersion >= 110 && (
          <Form.Field
            title={sharedMessages.resetsJoinNonces}
            onChange={handleResetsJoinNoncesChange}
            warning={resetsJoinNonces ? sharedMessages.resetWarning : undefined}
            name="resets_join_nonces"
            component={Checkbox}
          />
        )}
        <Form.Field
          autoFocus={!mayEditKeys}
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
      </Form.CollapseSection>
    </Wizard.Form>
  )
})

JoinSettingsForm.propTypes = {
  error: PropTypes.error,
  lorawanVersion: PropTypes.string.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
}

JoinSettingsForm.defaultProps = {
  error: undefined,
}

const WrappedJoinSettingsForm = withBreadcrumb('device.add.steps.join', props => (
  <Breadcrumb path={props.match.url} content={props.title} />
))(JoinSettingsForm)

WrappedJoinSettingsForm.propTypes = {
  match: PropTypes.match.isRequired,
  title: PropTypes.message.isRequired,
}

export default WrappedJoinSettingsForm
