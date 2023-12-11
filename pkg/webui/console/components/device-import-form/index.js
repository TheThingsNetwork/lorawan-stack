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

import React, { useCallback, useState } from 'react'
import { defineMessages } from 'react-intl'
import { useFormikContext } from 'formik'
import { useSelector } from 'react-redux'

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
import { selectNsEnabled } from '@ttn-lw/lib/selectors/env'

import { frequencyPlanValueSetter, lorawanVersionValueSetter } from '@console/lib/device-utils'

import { selectNsFrequencyPlans } from '@console/store/selectors/configuration'

import FallbackVersionIdsSection from './fallback-version-ids-section'

import style from './device-import-form.styl'

const m = defineMessages({
  file: 'File',
  formatInfo: 'Format information',
  selectAFile: 'Please select a template file',
  fileInfoPlaceholder: 'Please select a template format',
  setClaimAuthCode: 'Set claim authentication code',
  targetedComponents: 'Targeted components',
  advancedSectionTitle: 'Advanced end device claiming settings',
  infoText:
    'You can use the import functionality to register multiple end devices at once by uploading a file containing the registration information in one of the available formats. For more information, see also our documentation on <DocLink>Importing End Devices</DocLink>.',
  fallbackValuesImport:
    'Fallback values can be defined in case the imported file does not provide them. Any device with a required value that is not provided as fallback or in the imported file will be skipped by the importer.',
  inputMethodDeviceRepo: 'Load end device profile from the LoRaWAN Device Repository',
  inputMethodManual: 'Enter LoRaWAN versions and frequency plan manually',
  fallbackValues: 'Fallback values',
  noFallback: 'Do not set any fallback values',
})

const validationSchema = Yup.object({
  format_id: Yup.string().required(sharedMessages.validateRequired),
  data: Yup.string().required(m.selectAFile),
  set_claim_auth_code: Yup.boolean(),
  frequency_plan_id: Yup.string(),
  lorawan_version: Yup.string(),
  lorawan_phy_version: Yup.string(),
  _inputMethod: Yup.string(),
  version_ids: Yup.object().when('_inputMethod', {
    is: 'device-repository',
    then: schema =>
      schema.shape({
        brand_id: Yup.string().required(),
        model_id: Yup.string().required(),
        hardware_version: Yup.string().required(),
        firmware_version: Yup.string().required(),
        band_id: Yup.string().required(),
      }),
  }),
})

const nsEnabled = selectNsEnabled

const DeviceBulkCreateFormInner = props => {
  const { warningSize, largeFileWarningMessage, jsEnabled } = props
  const {
    values: { format_id, _inputMethod, frequency_plan_id, lorawan_version },
  } = useFormikContext()

  const formatSelected = Boolean(format_id)
  const [allowedFileExtensions, setAllowedFileExtensions] = useState(undefined)
  const [formatDescription, setFormatDescription] = useState(undefined)
  const nsFrequencyPlans = useSelector(selectNsFrequencyPlans)

  const handleSelectChange = useCallback(value => {
    if (value && value.fileExtensions && value.fileExtensions instanceof Array) {
      setAllowedFileExtensions(value.fileExtensions.join(','))
    }
    if (value && value.description) {
      setFormatDescription(value.description)
    }
  }, [])

  return (
    <>
      <Message
        content={m.infoText}
        className={style.info}
        values={{
          DocLink: msg => (
            <Link.DocLink secondary path="/devices/adding-devices/adding-devices-in-bulk">
              {msg}
            </Link.DocLink>
          ),
        }}
      />
      <hr className={style.hRule} />
      <DeviceTemplateFormatSelect onChange={handleSelectChange} name="format_id" required />
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
          <hr />
          <Form.CollapseSection id="advanced-settings" title={m.advancedSectionTitle}>
            <Form.Field
              disabled={!jsEnabled}
              title={sharedMessages.claiming}
              label={m.setClaimAuthCode}
              component={Checkbox}
              name="set_claim_auth_code"
              tooltipId={tooltipIds.SET_CLAIM_AUTH_CODE}
            />
          </Form.CollapseSection>
          <hr />
          <Form.SubTitle title={m.fallbackValues} />
          <Notification small info content={m.fallbackValuesImport} />
          <Form.Field component={Radio.Group} name="_inputMethod">
            <Radio label={m.noFallback} value="no-fallback" />
            <Radio label={m.inputMethodDeviceRepo} value="device-repository" />
            <Radio label={m.inputMethodManual} value="manual" />
          </Form.Field>
          {_inputMethod === 'manual' && (
            <>
              {nsEnabled && nsFrequencyPlans && (
                <NsFrequencyPlansSelect
                  tooltipId={tooltipIds.FREQUENCY_PLAN}
                  name="frequency_plan_id"
                  valueSetter={frequencyPlanValueSetter}
                />
              )}
              <Form.Field
                title={sharedMessages.macVersion}
                name="lorawan_version"
                component={LorawanVersionInput}
                tooltipId={tooltipIds.LORAWAN_VERSION}
                frequencyPlan={frequency_plan_id}
                valueSetter={lorawanVersionValueSetter}
              />
              <Form.Field
                title={sharedMessages.phyVersion}
                name="lorawan_phy_version"
                component={PhyVersionInput}
                tooltipId={tooltipIds.REGIONAL_PARAMETERS}
                lorawanVersion={lorawan_version}
              />
            </>
          )}
          {_inputMethod === 'device-repository' && <FallbackVersionIdsSection />}
          <SubmitBar>
            <Form.Submit component={SubmitButton} message={sharedMessages.importDevices} />
          </SubmitBar>
        </>
      )}
    </>
  )
}

const DeviceBulkCreateForm = props => {
  const { initialValues, onSubmit, jsEnabled, warningSize, largeFileWarningMessage } = props

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
      <DeviceBulkCreateFormInner
        warningSize={warningSize}
        largeFileWarningMessage={largeFileWarningMessage}
        jsEnabled={jsEnabled}
      />
    </Form>
  )
}

DeviceBulkCreateFormInner.propTypes = {
  jsEnabled: PropTypes.bool.isRequired,
  largeFileWarningMessage: PropTypes.message,
  warningSize: PropTypes.number,
}

DeviceBulkCreateFormInner.defaultProps = {
  largeFileWarningMessage: undefined,
  warningSize: undefined,
}

DeviceBulkCreateForm.propTypes = {
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

DeviceBulkCreateForm.defaultProps = {
  largeFileWarningMessage: undefined,
  warningSize: undefined,
}

export default DeviceBulkCreateForm
