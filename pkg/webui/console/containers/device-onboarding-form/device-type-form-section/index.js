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

import React, { useEffect } from 'react'
import { merge } from 'lodash'

import Radio from '@ttn-lw/components/radio-button'
import Form, { useFormContext } from '@ttn-lw/components/form'

import m from '../messages'

import DeviceTypeRepositoryFormSection from './device-type-repository-form-section'
import DeviceTypeManualFormSection from './device-type-manual-form-section'

const initialValues = {
  _inputMethod: 'device-repository',
}

const DeviceTypeFormSection = props => {
  const { getValues } = props
  const {
    values: { _inputMethod },
    setValues,
  } = useFormContext()

  useEffect(() => {
    // Set the section's initial values on mount.
    setValues(values => merge(values, initialValues))
  }, [getValues, setValues])

  return (
    <>
      <Form.Field title={m.inputMethod} name="_inputMethod" component={Radio.Group}>
        <Radio label={m.inputMethodDeviceRepo} value="device-repository" />
        <Radio label={m.inputMethodManual} value="manual" />
      </Form.Field>
      {_inputMethod === 'device-repository' && <DeviceTypeRepositoryFormSection />}
      {_inputMethod === 'manual' && <DeviceTypeManualFormSection />}
    </>
  )
}

export default DeviceTypeFormSection
