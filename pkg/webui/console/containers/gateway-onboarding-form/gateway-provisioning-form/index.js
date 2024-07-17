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

import React, { useCallback, useEffect, useMemo } from 'react'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'

import OwnersSelect from '@console/containers/owners-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import useDebounce from '@ttn-lw/lib/hooks/use-debounce'

import GatewayRegistrationFormSection from './gateway-registration-form-section'

// Empty strings will be interpreted as zero value (00 fill) by the backend
// For this reason, they need to be transformed to undefined instead,
// so that the value will be properly stripped when sent to the backend.
const gatewayEuiEncoder = value => (value === '' ? undefined : value)
const gatewayEuiDecoder = value => (value === undefined ? '' : value)

const GatewayProvisioningFormSection = () => {
  const { setFieldValue, values, setFieldTouched } = useFormContext()
  const isEuiMac = useMemo(() => values.ids.eui?.length === 12, [values.ids.eui])
  const debouncedEui = useDebounce(values.ids.eui, 3000)
  const isDebouncedEuiMac = useMemo(() => debouncedEui?.length === 12, [debouncedEui])
  const showMacConvert = isEuiMac && isDebouncedEuiMac

  useEffect(() => {
    if (showMacConvert) {
      setFieldTouched('ids.eui', true)
    } else if (isDebouncedEuiMac && !isEuiMac) {
      setFieldTouched('ids.eui', false)
    }
  }, [isDebouncedEuiMac, isEuiMac, setFieldTouched, showMacConvert])

  const handleConvertMacToEui = useCallback(() => {
    const euiValue = `${values.ids.eui.substring(0, 6)}FFFE${values.ids.eui.substring(6)}`
    setFieldValue('ids.eui', euiValue)
  }, [setFieldValue, values.ids.eui])

  return (
    <>
      <OwnersSelect name="_ownerId" required />
      <Form.Field
        title={sharedMessages.gatewayEUI}
        name="ids.eui"
        type="byte"
        min={8}
        max={8}
        placeholder={sharedMessages.gatewayEUI}
        component={Input}
        tooltipId={tooltipIds.GATEWAY_EUI}
        encode={gatewayEuiEncoder}
        decode={gatewayEuiDecoder}
        autoFocus
        action={
          showMacConvert
            ? {
                type: 'button',
                message: sharedMessages.convertMacToEui,
                onClick: handleConvertMacToEui,
              }
            : undefined
        }
      />
      <GatewayRegistrationFormSection />
    </>
  )
}

export { GatewayProvisioningFormSection as default }
