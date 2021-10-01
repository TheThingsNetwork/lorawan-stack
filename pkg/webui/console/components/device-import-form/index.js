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
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import { dict as componentDict } from '@ttn-lw/constants/components'

import Form from '@ttn-lw/components/form'
import FileInput from '@ttn-lw/components/file-input'
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import DeviceTemplateFormatSelect from '@console/containers/device-template-format-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './device-import-form.styl'

const m = defineMessages({
  file: 'File',
  formatInfo: 'Format information',
  selectAFile: 'Please select a template file',
  fileInfoPlaceholder: 'Please select a template format',
  claiming: 'Claiming',
  setClaimAuthCode: 'Set claim authentication code',
  targetedComponents: 'Targeted components',
  advancedSectionTitle: 'Advanced claiming and component settings',
  infoText:
    'You can use the import functionality to register multiple end devices at once by uploading a file containing the registration information in one of the available formats. For more information, see also our documentation on <DocLink>Importing End Devices</DocLink>.',
})

const validationSchema = Yup.object({
  format_id: Yup.string().required(sharedMessages.validateRequired),
  data: Yup.string().required(m.selectAFile),
  set_claim_auth_code: Yup.boolean(),
  components: Yup.object({
    is: Yup.boolean().required(),
    as: Yup.boolean(),
    js: Yup.boolean(),
    ns: Yup.boolean(),
  }).required(sharedMessages.validateRequired),
})

export default class DeviceBulkCreateForm extends Component {
  static propTypes = {
    components: PropTypes.components.isRequired,
    initialValues: PropTypes.shape({
      format_id: PropTypes.string,
      data: PropTypes.string,
      set_claim_auth_code: PropTypes.bool,
      components: PropTypes.shape({
        is: PropTypes.bool,
        ns: PropTypes.bool,
        js: PropTypes.bool,
        as: PropTypes.bool,
      }),
    }).isRequired,
    onSubmit: PropTypes.func.isRequired,
  }

  constructor(props) {
    super(props)

    const { initialValues } = props

    this.state = {
      allowedFileExtensions: undefined,
      formatDescription: undefined,
      formatSelected: false,
      jsSelected: Boolean(initialValues.components.js),
    }
    this.formRef = React.createRef()
  }

  @bind
  handleSelectChange(value) {
    const newState = { formatSelected: true }
    if (value && value.fileExtensions && value.fileExtensions instanceof Array) {
      newState.allowedFileExtensions = value.fileExtensions.join(',')
    }
    if (value && value.description) {
      newState.formatDescription = value.description
    }
    this.setState(newState)
  }

  @bind
  handleComponentChange(value) {
    const { jsSelected } = this.state
    const { values } = this.formRef.current
    const { js } = value

    if (js !== jsSelected) {
      this.setState({ jsSelected: js }, () => {
        if (values.set_claim_auth_code) {
          const { setFieldValue } = this.formRef.current

          // `claim_authentication_code` is stored in JS, so if the JS option is
          // not selected we dont want to include it in the payload.
          setFieldValue('set_claim_auth_code', false)
        }
      })
    }
  }

  render() {
    const { initialValues, onSubmit, components } = this.props
    const { allowedFileExtensions, formatSelected, formatDescription, jsSelected } = this.state
    return (
      <Form
        formikRef={this.formRef}
        onSubmit={onSubmit}
        validationSchema={validationSchema}
        submitEnabledWhenInvalid
        initialValues={initialValues}
      >
        <Message
          content={m.infoText}
          className={style.info}
          values={{
            DocLink: msg => (
              <Link.DocLink secondary path="/getting-started/migrating/import-devices/">
                {msg}
              </Link.DocLink>
            ),
          }}
        />
        <hr className={style.hRule} />
        <DeviceTemplateFormatSelect onChange={this.handleSelectChange} name="format_id" required />
        <Form.InfoField disabled={!formatSelected} title={m.formatInfo}>
          {formatDescription ? formatDescription : <Message content={m.fileInfoPlaceholder} />}
        </Form.InfoField>
        {formatSelected && (
          <>
            <Form.Field
              title={m.file}
              accept={allowedFileExtensions}
              component={FileInput}
              name="data"
              required
            />
            <Form.CollapseSection
              id="advanced-settings"
              title={m.advancedSectionTitle}
              errorTitles={['components', 'set_claim_auth_code']}
            >
              <Form.Field
                onChange={this.handleComponentChange}
                component={Checkbox.Group}
                name="components"
                title={m.targetedComponents}
                horizontal
              >
                {components.map(component => (
                  <Checkbox
                    disabled={component === 'is'}
                    key={component}
                    name={component}
                    label={componentDict[component]}
                  />
                ))}
              </Form.Field>
              <Form.Field
                disabled={!jsSelected}
                title={m.claiming}
                label={m.setClaimAuthCode}
                component={Checkbox}
                name="set_claim_auth_code"
                tooltipId={tooltipIds.SET_CLAIM_AUTH_CODE}
              />
            </Form.CollapseSection>
            <SubmitBar>
              <Form.Submit component={SubmitButton} message={sharedMessages.importDevices} />
            </SubmitBar>
          </>
        )}
      </Form>
    )
  }
}
