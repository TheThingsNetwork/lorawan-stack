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

import React, { useState } from 'react'
import { merge } from 'lodash'

import Radio from '@ttn-lw/components/radio-button'
import Form, { useFormContext } from '@ttn-lw/components/form'
import Modal from '@ttn-lw/components/modal'

import PropTypes from '@ttn-lw/lib/prop-types'

import m from '../messages'

import DeviceTypeRepositoryFormSection, {
  initialValues as repositoryInitialValues,
} from './device-type-repository-form-section'
import DeviceTypeManualFormSection, {
  initialValues as manualInitialValues,
} from './device-type-manual-form-section'

const initialValues = merge(
  {
    _inputMethod: 'device-repository',
  },
  manualInitialValues,
  repositoryInitialValues,
)

const DeviceTypeFormSection = props => {
  const { getRegistrationTemplate, appId } = props
  const {
    values: { _inputMethod, _isClaiming },
    resetForm,
  } = useFormContext()
  const [showModal, setShowModal] = useState(false)

  const handleMethodChange = React.useCallback(() => {
    setShowModal(true)
  }, [])

  const handleComplete = React.useCallback(() => {
    setShowModal(false)
    resetForm({ values: { ...initialValues, _inputMethod } })
  }, [resetForm, _inputMethod])

  return (
    <>
      {showModal && _isClaiming !== undefined && (
        <Modal
          buttonMessage={m.changeDeviceTypeButton}
          message={m.changeDeviceType}
          title={m.changeDeviceTypeButton}
          onComplete={handleComplete}
        />
      )}
      <Form.SubTitle title={m.endDeviceType} />
      <Form.Field
        title={m.inputMethod}
        name="_inputMethod"
        component={Radio.Group}
        onChange={handleMethodChange}
      >
        <Radio label={m.inputMethodDeviceRepo} value="device-repository" />
        <Radio label={m.inputMethodManual} value="manual" />
      </Form.Field>
      {_inputMethod === 'device-repository' && (
        <DeviceTypeRepositoryFormSection
          getRegistrationTemplate={getRegistrationTemplate}
          appId={appId}
        />
      )}
      {_inputMethod === 'manual' && <DeviceTypeManualFormSection />}
    </>
  )
}

DeviceTypeFormSection.propTypes = {
  appId: PropTypes.string,
  getRegistrationTemplate: PropTypes.func,
}
DeviceTypeFormSection.defaultProps = {
  appId: undefined,
  getRegistrationTemplate: () => null,
}

export { DeviceTypeFormSection as default, initialValues }
