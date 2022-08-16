// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import Form from '@ttn-lw/components/form'

import DevEUIComponent from '@console/containers/dev-eui-component'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from '../../messages'

const initialValues = {
  authenticated_identifiers: {
    dev_eui: '',
    authentication_code: '',
  },
  target_device_id: '',
}

const DeviceClaimingFormSection = () => {
  const idInputRef = React.useRef(null)

  return (
    <>
      <DevEUIComponent name="authenticated_identifiers.dev_eui" />
      <Form.Field
        title={sharedMessages.claimAuthCode}
        name="authenticated_identifiers.authentication_code"
        component={Input}
        sensitive
      />
      <Form.Field
        required
        title={sharedMessages.devID}
        name="target_device_id"
        placeholder={sharedMessages.deviceIdPlaceholder}
        component={Input}
        inputRef={idInputRef}
        tooltipId={tooltipIds.DEVICE_ID}
        description={m.deviceIdDescription}
      />
    </>
  )
}

export { DeviceClaimingFormSection as default, initialValues }
