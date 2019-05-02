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
import { connect } from 'react-redux'
import bind from 'autobind-decorator'

import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'
import FieldGroup from '../../../components/field/group'
import Message from '../../../lib/components/message'
import SubmitBar from '../../../components/submit-bar'
import FrequencyPlansSelect from '../../containers/freq-plans-select'

import sharedMessages from '../../../lib/shared-messages'
import m from './messages'
import validationSchema from './validation-schema'


@connect(function ({ configuration }, props) {
  return {
    nsFrequencyPlans: configuration.nsFrequencyPlans,
    frequencyPlanError: configuration.error,
  }
})
@bind
export default class DeviceDataForm extends Component {
  state = {
    error: '',
    otaa: true,
    resets_join_nonces: false,
    resets_f_cnt: false,
  }

  handleOTAASelect () {
    this.setState({ otaa: true })
  }

  handleABPSelect () {
    this.setState({ otaa: false })
  }

  handleResetsJoinNoncesChange (value) {
    this.setState({ resets_join_nonces: value })
  }

  handleResetsFrameCountersChange (value) {
    this.setState({ resets_f_cnt: value })
  }

  get ABPSection () {
    const { resets_f_cnt } = this.state
    return (
      <React.Fragment>
        <Field
          title={sharedMessages.devAddr}
          name="session.dev_addr"
          type="byte"
          min={4}
          max={4}
          placeholder={m.leaveBlankPlaceholder}
          description={m.deviceAddrDescription}
        />
        <Field
          title={sharedMessages.fwdNtwkKey}
          name="session.keys.f_nwk_s_int_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={m.leaveBlankPlaceholder}
          description={m.fwdNtwkKeyDescription}
        />
        <Field
          title={sharedMessages.sNtwkSIKey}
          name="session.keys.s_nwk_s_int_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={m.leaveBlankPlaceholder}
          description={m.sNtwkSIKeyDescription}
        />
        <Field
          title={sharedMessages.ntwkSEncKey}
          name="session.keys.nwk_s_enc_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={m.leaveBlankPlaceholder}
          description={m.ntwkSEncKeyDescription}
        />
        <Field
          title={sharedMessages.appSKey}
          name="session.keys.app_s_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={m.leaveBlankPlaceholder}
          description={m.appSKeyDescription}
        />
        <Field
          title={m.resetsFCnt}
          onChange={this.handleResetsFrameCountersChange}
          warning={resets_f_cnt ? m.resetWarning : undefined}
          name="mac_settings.resets_f_cnt"
          type="checkbox"
        />
      </React.Fragment>
    )
  }

  get OTAASection () {
    const { resets_join_nonces } = this.state
    return (
      <React.Fragment>
        <Field
          title={sharedMessages.joinEUI}
          name="ids.join_eui"
          type="byte"
          min={8}
          max={8}
          placeholder={m.joinEUIPlaceholder}
          required
        />
        <Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          placeholder={m.deviceEUIPlaceholder}
          description={m.deviceEUIDescription}
          required
        />
        <Field
          title={sharedMessages.nwkKey}
          name="root_keys.nwk_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={m.leaveBlankPlaceholder}
          description={m.nwkKeyDescription}
        />
        <Field
          title={sharedMessages.appKey}
          name="root_keys.app_key.key"
          type="byte"
          min={16}
          max={16}
          placeholder={m.leaveBlankPlaceholder}
          description={m.appKeyDescription}
        />
        <Field
          title={m.resetsJoinNonces}
          onChange={this.handleResetsJoinNoncesChange}
          warning={resets_join_nonces ? m.resetWarning : undefined}
          name="resets_join_nonces"
          type="checkbox"
        />
      </React.Fragment>
    )
  }

  render () {
    const { error, otaa } = this.state
    const { handleSubmit } = this.props

    return (
      <Form
        error={error}
        onSubmit={onSubmit}
        validationSchema={validationSchema}
        submitEnabledWhenInvalid
        isInitialValid={false}
        initialValues={
          {
            ids: {
              device_id: undefined,
              join_eui: undefined,
              dev_eui: undefined,
            },
            activation_mode: 'otaa',
            lorawan_version: undefined,
            lorawan_phy_version: undefined,
            frequency_plan_id: undefined,
            resets_join_nonces: false,
            root_keys: {},
            session: {},
            mac_settings: {
              resets_f_cnt: false,
            },
          }
        }
        mapErrorsToFields={{
          id_taken: 'application_id',
          identifiers: 'application_id',
        }}
        horizontal
      >
        <Message
          component="h4"
          content={sharedMessages.generalSettings}
        />
        <Field
          title={sharedMessages.devID}
          name="ids.device_id"
          placeholder={m.deviceIdPlaceholder}
          description={m.deviceIdDescription}
          autoFocus
          required
        />
        <Field
          title={sharedMessages.devName}
          name="name"
          placeholder={m.deviceNamePlaceholder}
          description={m.deviceNameDescription}
        />
        <Field
          title={sharedMessages.devDesc}
          name="description"
          type="textarea"
          description={m.deviceDescDescription}
        />
        <Message
          component="h4"
          content={m.lorawanOptions}
        />
        <Field
          title={sharedMessages.macVersion}
          name="lorawan_version"
          type="select"
          required
          options={[
            { value: 'MAC_V1_0', label: 'MAC V1.0' },
            { value: 'MAC_V1_0_1', label: 'MAC V1.0.1' },
            { value: 'MAC_V1_0_2', label: 'MAC V1.0.2' },
            { value: 'MAC_V1_1', label: 'MAC V1.1' },
          ]}
        />
        <Field
          title={sharedMessages.phyVersion}
          name="lorawan_phy_version"
          type="select"
          required
          options={[
            { value: 'PHY_V1_0', label: 'PHY V1.0' },
            { value: 'PHY_V1_0_1', label: 'PHY V1.0.1' },
            { value: 'PHY_V1_0_2_REV_A', label: 'PHY V1.0.2 REV A' },
            { value: 'PHY_V1_0_2_REV_B', label: 'PHY V1.0.2 REV B' },
            { value: 'PHY_V1_1_REV_A', label: 'PHY V1.1 REV A' },
            { value: 'PHY_V1_1_REV_B', label: 'PHY V1.1 REV B' },
          ]}
        />
        <FrequencyPlansSelect
          title={sharedMessages.frequencyPlan}
          source="ns"
          name="frequency_plan_id"
          horizontal
          required
        />
        <Field
          title={m.supportsClassC}
          name="supports_class_c"
          type="checkbox"
        />
        <Message
          component="h4"
          content={m.activationSettings}
        />
        <FieldGroup
          title={m.activationMode}
          name="activation_mode"
          columns
        >
          <Field
            title={m.otaa}
            value="otaa"
            type="radio"
            onChange={this.handleOTAASelect}
          />
          <Field
            title={m.abp}
            value="abp"
            type="radio"
            onChange={this.handleABPSelect}
          />
        </FieldGroup>
        {otaa ? this.OTAASection : this.ABPSection}
        <SubmitBar>
          <Button type="submit" message={m.createDevice} />
        </SubmitBar>
      </Form>
    )
  }
}

DeviceDataForm.propTypes = {
  onSubmit: PropTypes.func.isRequired,
  error: PropTypes.error,
  update: PropTypes.bool,
  initialValues: PropTypes.object,
}

DeviceDataForm.defaultProps = {
  initialValues: {},
  update: false,
  error: '',
}

export default DeviceDataForm
