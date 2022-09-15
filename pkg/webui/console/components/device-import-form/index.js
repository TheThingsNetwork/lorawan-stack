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

import Form from '@ttn-lw/components/form'
import FileInput from '@ttn-lw/components/file-input'
import Checkbox from '@ttn-lw/components/checkbox'
import Notification from '@ttn-lw/components/notification'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Link from '@ttn-lw/components/link'
import Radio from '@ttn-lw/components/radio-button'

import Message from '@ttn-lw/lib/components/message'

import PhyVersionInput from '@console/components/phy-version-input'
import LorawanVersionInput from '@console/components/lorawan-version-input'

import DeviceTemplateFormatSelect from '@console/containers/device-template-format-select'
import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { selectNsConfig } from '@ttn-lw/lib/selectors/env'

import style from './device-import-form.styl'

const m = defineMessages({
  file: 'File',
  formatInfo: 'Format information',
  selectAFile: 'Please select a template file',
  fileInfoPlaceholder: 'Please select a template format',
  claiming: 'Claiming',
  setClaimAuthCode: 'Set claim authentication code',
  targetedComponents: 'Targeted components',
  advancedSectionTitle: 'Advanced end device claiming settings',
  infoText:
    'You can use the import functionality to register multiple end devices at once by uploading a file containing the registration information in one of the available formats. For more information, see also our documentation on <DocLink>Importing End Devices</DocLink>.',
  fallbackValuesImport:
    'These values will be used in case the imported file does not provide them. They are not required, although if not provided here or in the imported file, the import of the end device will not be successful.',
  inputMethod: 'Input Method',
  inputMethodDeviceRepo: 'Select the end device in the LoRaWAN Device Repository',
  inputMethodManual: 'Enter end device specifics manually',
})

const validationSchema = Yup.object({
  format_id: Yup.string().required(sharedMessages.validateRequired),
  data: Yup.string().required(m.selectAFile),
  set_claim_auth_code: Yup.boolean(),
  frequency_plan_id: Yup.string(),
  lorawan_version: Yup.string(),
  lorawan_phy_version: Yup.string(),
})

const { enabled: nsEnabled } = selectNsConfig()

export default class DeviceBulkCreateForm extends Component {
  static propTypes = {
    initialValues: PropTypes.shape({
      format_id: PropTypes.string,
      data: PropTypes.string,
      set_claim_auth_code: PropTypes.bool,
      frequency_plan_id: PropTypes.string,
      lorawan_version: PropTypes.string,
      lorawan_phy_version: PropTypes.string,
    }).isRequired,
    jsEnabled: PropTypes.bool.isRequired,
    largeFileWarningMessage: PropTypes.message,
    onSubmit: PropTypes.func.isRequired,
    warningSize: PropTypes.number,
  }

  static defaultProps = {
    largeFileWarningMessage: undefined,
    warningSize: undefined,
  }

  state = {
    allowedFileExtensions: undefined,
    formatDescription: undefined,
    formatSelected: false,
    lorawanVersion: '',
    freqPlan: '',
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
  handleFreqPlanChange(option) {
    const { value: freqPlan } = option
    this.setState({ freqPlan })
  }

  @bind
  handleLorawanVersionChange(version) {
    this.setState({ lorawanVersion: version })
  }

  render() {
    const { initialValues, onSubmit, jsEnabled, warningSize, largeFileWarningMessage } = this.props
    const { allowedFileExtensions, formatSelected, formatDescription, freqPlan, lorawanVersion } =
      this.state
    let passedInitialValues = initialValues
    if (!jsEnabled && initialValues.set_claim_auth_code) {
      passedInitialValues = { ...initialValues, set_claim_auth_code: false }
    }
    return (
      <Form
        onSubmit={onSubmit}
        validationSchema={validationSchema}
        submitEnabledWhenInvalid
        initialValues={passedInitialValues}
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
              largeFileWarningMessage={largeFileWarningMessage}
              warningSize={warningSize}
              name="data"
              required
            />
            <Form.SubTitle title="Fallback values" />
            <Form.Field
              title={m.inputMethod}
              component={Radio.Group}
              value={_inputMethod}
              name="_inputMethod"
              valueSetter={handleInputMethodChange}
            >
              <Radio label={m.inputMethodDeviceRepo} value="device-repository" />
              <Radio label={m.inputMethodManual} value="manual" />
            </Form.Field>
            <Notification small info content={m.fallbackValuesImport} />
            {nsEnabled && (
              <NsFrequencyPlansSelect
                tooltipId={tooltipIds.FREQUENCY_PLAN}
                name="frequency_plan_id"
                onChange={this.handleFreqPlanChange}
              />
            )}
            <Form.Field
              title={sharedMessages.macVersion}
              name="lorawan_version"
              component={LorawanVersionInput}
              tooltipId={tooltipIds.LORAWAN_VERSION}
              onChange={this.handleLorawanVersionChange}
              frequencyPlan={freqPlan}
            />
            <Form.Field
              title={sharedMessages.phyVersion}
              name="lorawan_phy_version"
              component={PhyVersionInput}
              tooltipId={tooltipIds.REGIONAL_PARAMETERS}
              onChange={this.handlePhyVersionChange}
              lorawanVersion={lorawanVersion}
            />
            <Form.CollapseSection id="advanced-settings" title={m.advancedSectionTitle}>
              <Form.Field
                disabled={!jsEnabled}
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
