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

import React, { Component } from 'react'
import * as Yup from 'yup'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import Form from '../../../components/form'
import DeviceTemplateFormatSelect from '../../containers/device-template-format-select'
import FileInput from '../../../components/file-input'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import sharedMessages from '../../../lib/shared-messages'

const m = defineMessages({
  importFile: 'Import file',
  createDevices: 'Create Devices',
  selectAFile: 'Please select a template file',
})

const validationSchema = Yup.object({
  format_id: Yup.string().required(sharedMessages.validateRequired),
  data: Yup.string().required(m.selectAFile),
})

export default class DeviceBulkCreateForm extends Component {
  state = {
    allowedFileExtensions: undefined,
    formatSelected: false,
  }

  @bind
  handleSelectChange(value) {
    const newState = { formatSelected: true }
    if (value && value.fileExtensions && value.fileExtensions instanceof Array) {
      newState.allowedFileExtensions = value.fileExtensions.join(',')
    }
    this.setState(newState)
  }

  render() {
    const { initialValues, error, onSubmit } = this.props
    const { allowedFileExtensions, formatSelected } = this.state
    return (
      <Form
        error={error}
        onSubmit={onSubmit}
        validationSchema={validationSchema}
        submitEnabledWhenInvalid
        initialValues={initialValues}
      >
        <DeviceTemplateFormatSelect onChange={this.handleSelectChange} name="format_id" required />
        <Form.Field
          disabled={!formatSelected}
          title={m.importFile}
          accept={allowedFileExtensions}
          component={FileInput}
          name="data"
          required
        />
        <SubmitBar>
          <Form.Submit component={SubmitButton} message={m.createDevices} />
        </SubmitBar>
      </Form>
    )
  }
}
