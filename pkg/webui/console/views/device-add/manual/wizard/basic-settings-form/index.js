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

import Input from '@ttn-lw/components/input'
import Wizard from '@ttn-lw/components/wizard'
import Form from '@ttn-lw/components/form'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import JoinEUIPRefixesInput from '@console/components/join-eui-prefixes-input'

import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { ACTIVATION_MODES, parseLorawanMacVersion } from '@console/lib/device-utils'

import validationSchema from './validation-schema'

const defaultInitialValues = {
  ids: {
    device_id: '',
    dev_eui: undefined,
    join_eui: undefined,
  },
  name: undefined,
  description: undefined,
}

const BasicSettingsForm = props => {
  const { lorawanVersion, activationMode, prefixes } = props

  const validationContext = React.useMemo(
    () => ({
      lorawanVersion,
      activationMode,
    }),
    [activationMode, lorawanVersion],
  )

  const lwVersion = parseLorawanMacVersion(lorawanVersion)
  const isNone = activationMode === ACTIVATION_MODES.NONE
  const isOTAA = activationMode === ACTIVATION_MODES.OTAA
  const isMulticast = activationMode === ACTIVATION_MODES.MULTICAST

  return (
    <Wizard.Form
      validationSchema={validationSchema}
      validationContext={validationContext}
      initialValues={defaultInitialValues}
    >
      <Form.Field
        required
        autoFocus
        title={sharedMessages.devID}
        name="ids.device_id"
        placeholder={sharedMessages.deviceIdPlaceholder}
        component={Input}
        glossaryId={glossaryIds.DEVICE_ID}
      />
      {!isNone && (
        <>
          {isOTAA && (
            <Form.Field
              title={lwVersion < 104 ? sharedMessages.appEUI : sharedMessages.joinEUI}
              component={JoinEUIPRefixesInput}
              name="ids.join_eui"
              prefixes={prefixes}
              required
              showPrefixes
              glossaryId={glossaryIds.JOIN_EUI}
            />
          )}
          {!isMulticast && (
            <Form.Field
              title={sharedMessages.devEUI}
              name="ids.dev_eui"
              type="byte"
              min={8}
              max={8}
              required={isOTAA || lwVersion === 104}
              component={Input}
              glossaryId={glossaryIds.DEV_EUI}
            />
          )}
        </>
      )}
      <Form.Field
        title={sharedMessages.devName}
        name="name"
        placeholder={sharedMessages.deviceNamePlaceholder}
        description={sharedMessages.deviceNameDescription}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.devDesc}
        name="description"
        type="textarea"
        placeholder={sharedMessages.deviceDescPlaceholder}
        description={sharedMessages.deviceDescDescription}
        component={Input}
      />
    </Wizard.Form>
  )
}

BasicSettingsForm.propTypes = {
  activationMode: PropTypes.oneOf(Object.values(ACTIVATION_MODES)).isRequired,
  lorawanVersion: PropTypes.string.isRequired,
  prefixes: PropTypes.euiPrefixes.isRequired,
}

const WrappedBasicSettingsForm = withBreadcrumb('device.add.steps.basic', props => (
  <Breadcrumb path={props.match.url} content={props.title} />
))(BasicSettingsForm)

WrappedBasicSettingsForm.propTypes = {
  match: PropTypes.match.isRequired,
  title: PropTypes.message.isRequired,
}

export default WrappedBasicSettingsForm
