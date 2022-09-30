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

import React, { useCallback, useRef, useState } from 'react'
import { merge } from 'lodash'

import Radio from '@ttn-lw/components/radio-button'
import Form, { useFormContext } from '@ttn-lw/components/form'
import PortalledModal from '@ttn-lw/components/modal/portalled'

import TOOLTIP_IDS from '@ttn-lw/lib/constants/tooltip-ids'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { hasSelectedDeviceRepositoryOther } from '@console/lib/device-utils'

import m from '../messages'
import { hasCompletedDeviceRepositorySelection } from '../utils'

import DeviceTypeRepositoryFormSection, {
  initialValues as repositoryInitialValues,
} from './repository-form-section'
import DeviceTypeManualFormSection, {
  initialValues as manualInitialValues,
} from './manual-form-section'

const initialValues = merge(
  {
    _inputMethod: 'device-repository',
  },
  manualInitialValues,
  repositoryInitialValues,
)

const DeviceTypeFormSection = () => {
  const { values, resetForm } = useFormContext()
  const { version_ids, frequency_plan_id, lorawan_version, lorawan_phy_version, _inputMethod } =
    values

  const isPristineForm =
    (!hasCompletedDeviceRepositorySelection(version_ids) ||
      hasSelectedDeviceRepositoryOther(version_ids)) &&
    !Boolean(frequency_plan_id) &&
    !Boolean(lorawan_phy_version) &&
    !Boolean(lorawan_version)

  const [showModal, setShowModal] = useState(false)

  const modalResolver = useRef()

  const handleInputMethodChange = useCallback(
    async ({ setFieldValue }, { name, value }) => {
      if (!isPristineForm) {
        setShowModal(true)
        const approved = await new Promise(resolve => {
          modalResolver.current = resolve
        })
        setShowModal(false)
        if (approved) {
          resetForm()
          return setFieldValue(name, value)
        }
      } else {
        return setFieldValue(name, value)
      }
    },
    [isPristineForm, resetForm],
  )

  const handleMethodModalComplete = React.useCallback(approved => {
    if (modalResolver && modalResolver.current) {
      modalResolver.current(approved)
    }
  }, [])

  return (
    <>
      <PortalledModal
        buttonMessage={m.changeDeviceTypeButton}
        message={m.changeDeviceType}
        title={m.changeDeviceTypeButton}
        onComplete={handleMethodModalComplete}
        visible={showModal}
      />
      <Form.SubTitle title={m.endDeviceType} />
      <Form.Field
        title={sharedMessages.inputMethod}
        component={Radio.Group}
        value={_inputMethod}
        name="_inputMethod"
        valueSetter={handleInputMethodChange}
        tooltipId={TOOLTIP_IDS.INPUT_METHOD}
      >
        <Radio label={m.inputMethodDeviceRepo} value="device-repository" />
        <Radio label={m.inputMethodManual} value="manual" />
      </Form.Field>
      {_inputMethod === 'device-repository' && <DeviceTypeRepositoryFormSection />}
      {_inputMethod === 'manual' && <DeviceTypeManualFormSection />}
    </>
  )
}

export { DeviceTypeFormSection as default, initialValues }
